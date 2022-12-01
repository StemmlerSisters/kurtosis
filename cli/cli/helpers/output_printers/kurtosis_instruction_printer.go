package output_printers

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/interactive_terminal_decider"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

const (
	bazelBuildDefaultFilename = ""

	instructionPrefixString = "> "
	resultPrefixString      = ""

	progressBarLength = 20       // in characters
	progressBarChar   = "\u2588" // unicode for: █

	codeCommentPrefix = "# "
)

var (
	colorizeInstruction      = color.New(color.FgCyan).SprintfFunc()
	colorizeResult           = color.New(color.FgWhite).SprintfFunc()
	colorizeError            = color.New(color.FgRed).SprintfFunc()
	colorizeRunSuccessfulMsg = color.New(color.FgGreen).SprintfFunc()

	colorizeProgressBarIsDone    = color.New(color.FgGreen).SprintfFunc()
	colorizeProgressBarRemaining = color.New(color.FgWhite).SprintfFunc()
)

var (
	writer = logrus.StandardLogger().Out

	spinnerChar  = spinner.CharSets[11]
	spinnerSpeed = 250 * time.Millisecond
	spinnerColor = spinner.WithColor("yellow")
)

type ExecutionPrinter struct {
	lock               *sync.Mutex
	isSpinnerBeingUsed bool
	spinner            *spinner.Spinner
	isStarted          bool
}

func NewExecutionPrinter() *ExecutionPrinter {
	return &ExecutionPrinter{
		lock:               &sync.Mutex{},
		isSpinnerBeingUsed: false,
		spinner:            nil,
		isStarted:          false,
	}
}

func (printer *ExecutionPrinter) Start() error {
	if printer.isStarted {
		return stacktrace.NewError("printer already started")
	}
	printer.isStarted = true
	if interactive_terminal_decider.IsInteractiveTerminal() {
		printer.isSpinnerBeingUsed = false
		logrus.Infof("Kurtosis CLI is running in a non interactive terminal. Everything will work but progress information and the progress bar will not be displayed.")
		return nil
	}
	printer.isSpinnerBeingUsed = true
	printer.spinner = spinner.New(spinnerChar, spinnerSpeed, spinnerColor, spinner.WithWriter(writer))
	printer.spinner.Start()
	return nil
}

func (printer *ExecutionPrinter) Stop() {
	if printer.isSpinnerBeingUsed && printer.isStarted {
		if printer.spinner != nil && printer.spinner.Active() {
			printer.spinner.Stop()
		}
	}
	printer.isStarted = false
}

// PrintKurtosisExecutionResponseLineToStdOut format and prints the instruction to StdOut.
func (printer *ExecutionPrinter) PrintKurtosisExecutionResponseLineToStdOut(responseLine *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, verbosity run.Verbosity) error {
	// Printing is a 3 phase operation:
	// 1. stop spinner to clear the ephemeral progress info
	// 2. print whatever needs to be printed, could be nothing
	// 3. restart the spinner, potentially with an updated content
	// To avoid conflicts, we take a lock out of cautiousness (this method shouldn't be called concurrently anyway)
	printer.lock.Lock()
	defer printer.lock.Unlock()

	if !printer.isStarted {
		return stacktrace.NewError("Cannot print with a non started printer")
	}

	// process response payload
	if responseLine.GetInstruction() != nil {
		formattedInstruction := formatInstruction(responseLine.GetInstruction(), verbosity)
		// we separate each tuple (instruction, result) with an additional newline
		formattedInstructionWithNewline := fmt.Sprintf("\n%s", formattedInstruction)
		if err := printer.printPersistentLineToStdOut(formattedInstructionWithNewline); err != nil {
			return stacktrace.Propagate(err, "Error printing Kurtosis instruction: \n%v", formattedInstruction)
		}
	} else if responseLine.GetInstructionResult() != nil {
		formattedInstructionResult := formatInstructionResult(responseLine.GetInstructionResult())
		if err := printer.printPersistentLineToStdOut(formattedInstructionResult); err != nil {
			return stacktrace.Propagate(err, "Error printing Kurtosis instruction result: \n%v", formattedInstructionResult)
		}
	} else if responseLine.GetError() != nil {
		var errorMsg string
		if responseLine.GetError().GetInterpretationError() != nil {
			errorMsg = fmt.Sprintf("There was an error interpreting Starlark code \n%v", responseLine.GetError().GetInterpretationError().GetErrorMessage())
		} else if responseLine.GetError().GetValidationError() != nil {
			errorMsg = fmt.Sprintf("There was an error validating Starlark code \n%v", responseLine.GetError().GetValidationError().GetErrorMessage())
		} else if responseLine.GetError().GetExecutionError() != nil {
			errorMsg = fmt.Sprintf("There was an error executing Starlark code \n%v", responseLine.GetError().GetExecutionError().GetErrorMessage())
		}
		formattedError := formatError(errorMsg)
		if err := printer.printPersistentLineToStdOut(formattedError); err != nil {
			return stacktrace.Propagate(err, "An error happened executing Starlark code but the error couldn't be printed to the CLI output. Error message was: \n%v", errorMsg)
		}
	} else if responseLine.GetProgressInfo() != nil {
		if printer.isSpinnerBeingUsed {
			progress := responseLine.GetProgressInfo()
			progressBarStr := formatProgressBar(progress.GetCurrentStepNumber(), progress.GetTotalSteps())
			spinnerInfoString := fmt.Sprintf("   %s %s", progressBarStr, progress.GetCurrentStepInfo())
			printer.spinner.Suffix = spinnerInfoString
		}
	} else if responseLine.GetRunFinishedEvent() != nil {
		formattedRunOutputMessage := formatRunOutput(responseLine.GetRunFinishedEvent())
		formattedRunOutputMessageWithNewline := fmt.Sprintf("\n%s", formattedRunOutputMessage)
		if err := printer.printPersistentLineToStdOut(formattedRunOutputMessageWithNewline); err != nil {
			return stacktrace.Propagate(err, "Unable to print the success output message containing the serialized output object. Message was: \n%v", formattedRunOutputMessage)
		}
	}
	return nil
}

func (printer *ExecutionPrinter) printPersistentLineToStdOut(lineToPrint string) error {
	// If spinner is being used, we have to stop spinner -> print -> start spinner in order to keep the spinner at the bottom of the output
	if printer.isSpinnerBeingUsed {
		printer.spinner.Stop()
	}
	if _, err := fmt.Fprintln(writer, lineToPrint); err != nil {
		return stacktrace.Propagate(err, "An error happened printing a Starlark run response line. Line was:\n%s", lineToPrint)
	}
	if printer.isSpinnerBeingUsed {
		printer.spinner.Start()
	}
	return nil
}

func formatError(errorMessage string) string {
	return colorizeError(errorMessage)
}

func formatInstruction(instruction *kurtosis_core_rpc_api_bindings.StarlarkInstruction, verbosity run.Verbosity) string {
	var serializedInstruction string
	switch verbosity {
	case run.Brief:
		serializedInstruction = formatInstructionToReadableString(instruction, false)
	case run.Detailed:
		serializedInstruction = formatInstructionToReadableString(instruction, true)
	case run.Executable:
		serializedInstruction = formatInstructionToExecutable(instruction)
	default:
		logrus.Warnf("Unsupported verbosity flag: '%s'. Supported values are: (%s). The instruction will be printed with the default verbosity '%s'",
			verbosity.String(), strings.Join(run.VerbosityStrings(), ", "), run.Brief.String())
		serializedInstruction = formatInstructionToReadableString(instruction, false)
	}
	return colorizeInstruction(serializedInstruction)
}

func formatInstructionResult(instructionResult *kurtosis_core_rpc_api_bindings.StarlarkInstructionResult) string {
	serializedInstructionResult := fmt.Sprintf("%s%s", resultPrefixString, instructionResult.GetSerializedInstructionResult())
	return colorizeResult(serializedInstructionResult)
}

func formatInstructionToReadableString(instruction *kurtosis_core_rpc_api_bindings.StarlarkInstruction, exhaustive bool) string {
	serializedInstructionComponents := []string{instruction.GetInstructionName()}
	for _, arg := range instruction.GetArguments() {
		if exhaustive || arg.GetIsRepresentative() {
			var serializedArg string
			if arg.ArgName != nil {
				serializedArg = fmt.Sprintf("%s=%s", arg.GetArgName(), arg.GetSerializedArgValue())
			} else {
				serializedArg = arg.GetSerializedArgValue()
			}
			serializedInstructionComponents = append(serializedInstructionComponents, serializedArg)
		}
	}

	var serializedInstruction string
	if exhaustive {
		separator := fmt.Sprintf("\n%s\t", instructionPrefixString)
		serializedInstruction = strings.Join(serializedInstructionComponents, separator)
	} else {
		separator := " "
		serializedInstruction = strings.Join(serializedInstructionComponents, separator)
	}
	return fmt.Sprintf("%s%s", instructionPrefixString, serializedInstruction)
}

func formatInstructionToExecutable(instruction *kurtosis_core_rpc_api_bindings.StarlarkInstruction) string {
	serializedInstruction := fmt.Sprintf(
		"from %s[%d:%d]\n%s",
		instruction.GetPosition().GetFilename(),
		instruction.GetPosition().GetLine(),
		instruction.GetPosition().GetColumn(),
		instruction.GetExecutableInstruction(),
	)
	serializedInstructionWithComment := fmt.Sprintf("%s%s", codeCommentPrefix, serializedInstruction)

	parsedInstruction, err := build.ParseDefault(bazelBuildDefaultFilename, []byte(serializedInstructionWithComment))
	if err != nil {
		logrus.Warnf("Unable to format instruction. Will print it with no indentation. Problematic instruction was: \n%v", serializedInstructionWithComment)
		return serializedInstructionWithComment
	}

	multiLineInstruction := strings.Builder{}
	for _, statement := range parsedInstruction.Stmt {
		multiLineInstruction.WriteString(build.FormatString(statement))
	}
	return multiLineInstruction.String()
}

func formatProgressBar(currentStep uint32, totalSteps uint32) string {
	progressBar := strings.Builder{}
	threshold := currentStep * progressBarLength
	for i := uint32(0); i < progressBarLength; i++ {
		if i*totalSteps < threshold {
			progressBar.WriteString(colorizeProgressBarIsDone(progressBarChar))
		} else {
			progressBar.WriteString(colorizeProgressBarRemaining(progressBarChar))
		}
	}
	return progressBar.String()
}

func formatRunOutput(runFinishedEvent *kurtosis_core_rpc_api_bindings.StarlarkRunFinishedEvent) string {
	if !runFinishedEvent.GetIsRunSuccessful() {
		return colorizeError("Error encountered running Starlark code")
	}
	var runSuccessMsg string
	if runFinishedEvent.GetSerializedOutput() != "" {
		runSuccessMsg = fmt.Sprintf("Starlark code successfully executed. Output was: \n%v", runFinishedEvent.GetSerializedOutput())
	} else {
		runSuccessMsg = "Starlark code successfully executed. No output was returned"
	}
	return colorizeRunSuccessfulMsg(runSuccessMsg)
}
