package kurtosis_command

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/args"
	"github.com/kurtosis-tech/kurtosis-cli/cli/command_framework/kurtosis_command/flags"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

const (
	shouldLogCompletionDebugMessagesToStderr = true

	uintBase = 10
	uint32Bits = 32
)

// This is a struct intended to abstract away much of the details of creating a Cobra command that does what we want,
//  so that Kurtosis devs can talk in higher-level notions
// E.g. simply by providing the flags and args, the usage string will be automatically generated for the Kurtosis dev
type KurtosisCommand struct {
	// The string for the command (e.g. "inspect" or "ls")
	CommandStr string

	// Will be used when displaying the command for tab completion
	ShortDescription string

	LongDescription string

	// Order isn't important here
	Flags []*flags.FlagConfig

	// Order IS important here
	Args []*args.ArgConfig

	// The actual logic that the command will run
	RunFunc func(flags *flags.ParsedFlags, args *args.ParsedArgs) error
}

// Gets a Cobra command represnting the KurtosisCommand
// This function is intended to be run in an init() (i.e. before the program runs any logic), so it will panic if
//  any errors occur
func (kurtosisCmd *KurtosisCommand) MustGetCobraCommand() *cobra.Command {
	// Verify no duplicate flag keys
	usedFlagKeys := map[string]bool{}
	for _, flagConfig := range kurtosisCmd.Flags {
		key := flagConfig.Key
		if len(strings.TrimSpace(key)) == 0 {
			panic(stacktrace.NewError(
				"Empty flag key defined for command '%v'",
				kurtosisCmd.CommandStr,
			))
		}
		if _, found := usedFlagKeys[key]; found {
			panic(stacktrace.NewError(
				"Found duplicate flags with key '%v' for command '%v'",
				key,
				kurtosisCmd.CommandStr,
			))
		}
		usedFlagKeys[key] = true
	}

	// Verify no duplicate arg keys
	usedArgKeys := map[string]bool{}
	for _, argConfig := range kurtosisCmd.Args {
		key := argConfig.Key
		if len(strings.TrimSpace(key)) == 0 {
			panic(stacktrace.NewError(
				"Empty arg key defined for command '%v'",
				kurtosisCmd.CommandStr,
			))
		}
		if _, found := usedArgKeys[key]; found {
			panic(stacktrace.NewError(
				"Found duplicate args with key '%v' for command '%v'",
				key,
				kurtosisCmd.CommandStr,
			))
		}
		usedArgKeys[key] = true
	}

	// Verify that we don't have any invalid positional arg combinations, e.g.:
	//  - Any arg after an optional arg (the parser wouldn't know whether you want the optional arg or the one after it)
	//  - Any arg after an arg that consumes N args (since the CLI couldn't know where the greedy arg stops and the required arg begins)
	terminalArgKey := ""
	for _, argConfig := range kurtosisCmd.Args {
		key := argConfig.Key
		if terminalArgKey != "" {
			panic(stacktrace.NewError(
				"Arg '%v' for command '%v' must be the last argument because it's either optional or greedy, but arg '%v' was declared after it",
				terminalArgKey,
				kurtosisCmd.CommandStr,
				key,
			))
		}
		if argConfig.IsOptional || argConfig.IsGreedy {
			terminalArgKey = key
		}
	}

	// Based on digging through the Cobra source code, the toComplete string is theoretically the string that the user
	//  is in the process of typing when they press TAB. However, in my tests on Bash, the shell will automatically
	//  filter the results based off the partialStr without us needing to filter them ~ ktoday, 2022-02-02
	getCompletionsFunc := func(cmd *cobra.Command, previousArgStrs []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		parsedFlags := flags.NewParsedFlags(cmd.Flags())

		parsedArgs, argToComplete := args.ParseArgsForCompletion(kurtosisCmd.Args, previousArgStrs)
		if argToComplete == nil {
			// NOTE: We can't just use logrus because anything printed to STDOUT will be interpreted as a completion
			// See:
			//  https://github.com/spf13/cobra/blob/master/shell_completions.md#:~:text=the%20RunE%20function.-,Debugging,ShellCompDirectiveNoFileComp%20%23%20This%20is%20on%20stderr
			cobra.CompDebugln("Not completing because no argument needs completion", shouldLogCompletionDebugMessagesToStderr)
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		completionFunc := argToComplete.CompletionsFunc
		if completionFunc == nil {
			// NOTE: We can't just use logrus because anything printed to STDOUT will be interpreted as a completion
			// See:
			//  https://github.com/spf13/cobra/blob/master/shell_completions.md#:~:text=the%20RunE%20function.-,Debugging,ShellCompDirectiveNoFileComp%20%23%20This%20is%20on%20stderr
			cobra.CompDebugln(
				fmt.Sprintf(
					"Not completing because arg needing completion '%v' doesn't have a custom completion function",
					argToComplete.Key,
				),
				shouldLogCompletionDebugMessagesToStderr,
			)
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		completions, err := argToComplete.CompletionsFunc(parsedFlags, parsedArgs)
		if err != nil {
			// NOTE: We can't just use logrus because anything printed to STDOUT will be interpreted as a completion
			// See:
			//  https://github.com/spf13/cobra/blob/master/shell_completions.md#:~:text=the%20RunE%20function.-,Debugging,ShellCompDirectiveNoFileComp%20%23%20This%20is%20on%20stderr
			cobra.CompDebugln(
				fmt.Sprintf(
					"An error occurred running the completions function with previous arg strs '%+v' and toComplete string '%v':\n%v",
					previousArgStrs,
					toComplete,
					err,
				),
				shouldLogCompletionDebugMessagesToStderr,
			)
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return completions, cobra.ShellCompDirectiveNoFileComp
	}

	// Prepare the run function to be slotted into the Cobra command, which will do both arg validation & logic execution
	cobraRunFunc := func(cmd *cobra.Command, allArgs []string) error {
		parsedFlags := flags.NewParsedFlags(cmd.Flags())

		parsedArgs, err := args.ParseArgsForValidation(kurtosisCmd.Args, allArgs)
		if err != nil {
			logrus.Debugf("An error occurred while parsing args '%+v':\n%v", allArgs, err)

			// NOTE: This is a VERY special instance where we don't wrap the error with stacktrace.Propagate, because
			//  the errors returned by this function will *only* be arg-parsing errors and the stacktrace just adds
			//  clutter & confusion to what the user sees without providing any useful information
			return err
		}

		// Validate all the args
		for _, config := range kurtosisCmd.Args {
			validationFunc := config.ValidationFunc
			if validationFunc == nil {
				continue
			}
			if err := validationFunc(parsedFlags, parsedArgs); err != nil {
				return stacktrace.Propagate(err, "An error occurred validating arg '%v'", config.Key)
			}
		}

		if err := kurtosisCmd.RunFunc(parsedFlags, parsedArgs); err != nil {
			return stacktrace.Propagate(err, "An error occurred running command '%v'", kurtosisCmd.CommandStr)
		}

		return nil
	}

	// Build usage string
	allArgUsageStrs := []string{}
	for _, argConfig := range kurtosisCmd.Args {
		argUsageStr := renderArgUsageStr(argConfig)
		allArgUsageStrs = append(allArgUsageStrs, argUsageStr)
	}
	usageStr := fmt.Sprintf(
		"%v [flags] %v",
		kurtosisCmd.CommandStr,
		strings.Join(allArgUsageStrs, " "),
	)

	result := &cobra.Command{
		Use:                   usageStr,
		DisableFlagsInUseLine: true, // Not needed since we manually add the string in the usage string
		Short:                 kurtosisCmd.ShortDescription,
		Long:                  kurtosisCmd.LongDescription,
		ValidArgsFunction:     getCompletionsFunc,
		RunE: cobraRunFunc,
	}

	// Validates that the default values for the declared flags match the declard types, and add them to the Cobra command
	// Verify all flag default values match their declared types
	resultFlags := result.Flags()
	for _, flagConfig := range kurtosisCmd.Flags {
		key := flagConfig.Key
		usage := flagConfig.Usage
		defaultValStr := flagConfig.Default

		typeStr := flagConfig.Type.AsString()
		defaultValueDoesntMatchType := false
		switch typeStr {
		case flags.FlagType_String.AsString():
			// No validation needed because the default type is already string
			resultFlags.String(
				key,
				defaultValStr,
				usage,
			)
		case flags.FlagType_Bool.AsString():
			defaultValue, err := strconv.ParseBool(defaultValStr)
			if err != nil {
				defaultValueDoesntMatchType = true
				break
			}
			resultFlags.Bool(
				key,
				defaultValue,
				usage,
			)
		case flags.FlagType_Uint32.AsString():
			defaultValueUint64, err := strconv.ParseUint(defaultValStr, uintBase, uint32Bits)
			if err != nil {
				defaultValueDoesntMatchType = true
				break
			}
			resultFlags.Uint32(
				key,
				uint32(defaultValueUint64),
				usage,
			)
		default:
			panic(stacktrace.NewError("Flag '%v' on command '%v' is of unrecognized type '%v'", key, kurtosisCmd.CommandStr, typeStr))
		}
		if defaultValueDoesntMatchType {
			panic(stacktrace.NewError(
				"Default value of flag '%v' on command '%v' is '%v', which doesn't match the flag's declared type of '%v'",
				key,
				kurtosisCmd.CommandStr,
				defaultValStr,
				typeStr,
			))
		}
	}

	return result
}


// ====================================================================================================
//                                   Private Helper Functions
// ====================================================================================================

func renderArgUsageStr(arg *args.ArgConfig) string {
	result := arg.Key
	if arg.IsGreedy {
		result = result + "..."
	}
	if arg.IsOptional {
		result = "[" + result + "]"
	}
	return result
}