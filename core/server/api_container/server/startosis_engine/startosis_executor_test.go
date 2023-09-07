package startosis_engine

import (
	"context"
	"errors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/mock_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
	"go.starlark.net/starlark"
	"os"
	"strings"
	"testing"
)

const (
	executeSuccessfully = true
	throwOnExecute      = false

	doDryRun       = true
	executeForReal = false
	isSkipped      = false

	noScriptOutputObject = ""
	noParallelism        = 1
)

var (
	dummyPosition               = kurtosis_starlark_framework.NewKurtosisBuiltinPosition("dummyFile", 12, 1)
	noInstructionArgsForTesting []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg
)

func TestExecuteKurtosisInstructions_ExecuteForReal_Success(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)
	runtimeValueStore, createRuntimeValueStoreErr := runtime_value_store.CreateRuntimeValueStore(nil, enclaveDb)
	require.NoError(t, createRuntimeValueStoreErr)

	executor := NewStartosisExecutor(runtimeValueStore)
	require.NotNil(t, executor)

	instructionsPlan := instructions_plan.NewInstructionsPlan()
	instruction1 := createMockInstruction(t, "instruction1", executeSuccessfully)
	scheduledInstruction1 := instructions_plan.NewScheduledInstruction("instruction1", instruction1, starlark.None).Executed(true)
	instructionsPlan.AddScheduledInstruction(scheduledInstruction1)

	instruction2 := createMockInstruction(t, "instruction2", executeSuccessfully)
	instruction3 := createMockInstruction(t, "instruction3", executeSuccessfully)
	_, addInstErr := instructionsPlan.AddInstruction(instruction2, starlark.None)
	require.NoError(t, addInstErr)
	_, addInstErr = instructionsPlan.AddInstruction(instruction3, starlark.None)
	require.NoError(t, addInstErr)

	require.Equal(t, executor.enclavePlan.Size(), 0) // check that the enclave plan is empty prior to execution

	_, serializedInstruction, err := executeSynchronously(t, executor, executeForReal, instructionsPlan)
	instruction1.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction1.AssertNumberOfCalls(t, "Execute", 0) // not executed as it was already executed
	instruction2.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction2.AssertNumberOfCalls(t, "Execute", 1)
	instruction3.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction3.AssertNumberOfCalls(t, "Execute", 1)

	require.Nil(t, err)

	expectedSerializedInstructions := []*kurtosis_core_rpc_api_bindings.StarlarkInstruction{
		binding_constructors.NewStarlarkInstruction(
			dummyPosition.ToAPIType(), "instruction1", "instruction1()", noInstructionArgsForTesting, isSkipped),
		binding_constructors.NewStarlarkInstruction(
			dummyPosition.ToAPIType(), "instruction2", "instruction2()", noInstructionArgsForTesting, isSkipped),
		binding_constructors.NewStarlarkInstruction(
			dummyPosition.ToAPIType(), "instruction3", "instruction3()", noInstructionArgsForTesting, isSkipped),
	}
	require.Equal(t, expectedSerializedInstructions, serializedInstruction)
	require.Equal(t, executor.enclavePlan.Size(), 3) // check that the enclave plan now contains the 4 instructions
}

func TestExecuteKurtosisInstructions_ExecuteForReal_FailureHalfWay(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)
	runtimeValueStore, err := runtime_value_store.CreateRuntimeValueStore(nil, enclaveDb)
	require.NoError(t, err)

	executor := NewStartosisExecutor(runtimeValueStore)
	require.NotNil(t, executor)

	instruction1 := createMockInstruction(t, "instruction1", executeSuccessfully)
	instruction2 := createMockInstruction(t, "instruction2", throwOnExecute)
	instruction3 := createMockInstruction(t, "instruction3", executeSuccessfully)
	instructionsPlan := instructions_plan.NewInstructionsPlan()
	_, addInsErr := instructionsPlan.AddInstruction(instruction1, starlark.None)
	require.NoError(t, addInsErr)
	_, addInsErr = instructionsPlan.AddInstruction(instruction2, starlark.None)
	require.NoError(t, addInsErr)
	_, addInsErr = instructionsPlan.AddInstruction(instruction3, starlark.None)
	require.NoError(t, addInsErr)

	_, serializedInstruction, executionError := executeSynchronously(t, executor, executeForReal, instructionsPlan)
	instruction1.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction1.AssertNumberOfCalls(t, "Execute", 1)
	instruction2.AssertNumberOfCalls(t, "String", 1)
	instruction2.AssertNumberOfCalls(t, "Execute", 1)
	// nothing called for instruction 3 because instruction 2 threw an error
	instruction3.AssertNumberOfCalls(t, "GetCanonicalInstruction", 0)
	instruction3.AssertNumberOfCalls(t, "Execute", 0)

	expectedErrorMsgPrefix := `An error occurred executing instruction (number 2) at dummyFile[12:1]:
instruction2()
 --- at`
	expectedLowLevelErrorMessage := "expected error for test"
	require.NotNil(t, executionError)
	require.Contains(t, executionError.GetErrorMessage(), expectedErrorMsgPrefix)
	require.Contains(t, executionError.GetErrorMessage(), expectedLowLevelErrorMessage)

	expectedSerializedInstructions := []*kurtosis_core_rpc_api_bindings.StarlarkInstruction{
		// only instruction 1 and 2 because it failed at instruction 2
		binding_constructors.NewStarlarkInstruction(
			dummyPosition.ToAPIType(), "instruction1", "instruction1()", noInstructionArgsForTesting, isSkipped),
		binding_constructors.NewStarlarkInstruction(
			dummyPosition.ToAPIType(), "instruction2", "instruction2()", noInstructionArgsForTesting, isSkipped),
	}
	require.Equal(t, expectedSerializedInstructions, serializedInstruction)
}

func TestExecuteKurtosisInstructions_DoDryRun(t *testing.T) {
	enclaveDb := getEnclaveDBForTest(t)
	runtimeValueStore, createRuntimeValueStoreErr := runtime_value_store.CreateRuntimeValueStore(nil, enclaveDb)
	require.NoError(t, createRuntimeValueStoreErr)

	executor := NewStartosisExecutor(runtimeValueStore)
	require.NotNil(t, executor)

	instruction1 := createMockInstruction(t, "instruction1", executeSuccessfully)
	instruction2 := createMockInstruction(t, "instruction2", executeSuccessfully)
	instructionsPlan := instructions_plan.NewInstructionsPlan()
	_, addInsErr := instructionsPlan.AddInstruction(instruction1, starlark.None)
	require.NoError(t, addInsErr)
	_, addInsErr = instructionsPlan.AddInstruction(instruction2, starlark.None)
	require.NoError(t, addInsErr)

	_, serializedInstruction, err := executeSynchronously(t, executor, doDryRun, instructionsPlan)
	instruction1.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	instruction2.AssertNumberOfCalls(t, "GetCanonicalInstruction", 1)
	// both execute never called because dry run = true
	instruction1.AssertNumberOfCalls(t, "Execute", 0)
	instruction2.AssertNumberOfCalls(t, "Execute", 0)

	require.Nil(t, err)

	expectedSerializedInstructions := []*kurtosis_core_rpc_api_bindings.StarlarkInstruction{
		binding_constructors.NewStarlarkInstruction(
			dummyPosition.ToAPIType(), "instruction1", "instruction1()", noInstructionArgsForTesting, isSkipped),
		binding_constructors.NewStarlarkInstruction(
			dummyPosition.ToAPIType(), "instruction2", "instruction2()", noInstructionArgsForTesting, isSkipped),
	}
	require.Equal(t, serializedInstruction, expectedSerializedInstructions)
}

func createMockInstruction(t *testing.T, instructionName string, executeSuccessfully bool) *mock_instruction.MockKurtosisInstruction {
	instruction := mock_instruction.NewMockKurtosisInstruction(t)

	stringifiedInstruction := instructionName + "()"
	canonicalInstruction := binding_constructors.NewStarlarkInstruction(
		dummyPosition.ToAPIType(), instructionName, stringifiedInstruction, noInstructionArgsForTesting, isSkipped)
	instruction.EXPECT().GetCanonicalInstruction(mock.Anything).Maybe().Return(canonicalInstruction)
	instruction.EXPECT().GetPositionInOriginalScript().Maybe().Return(dummyPosition)
	instruction.EXPECT().String().Maybe().Return(stringifiedInstruction)

	if executeSuccessfully {
		instruction.EXPECT().Execute(mock.Anything).Maybe().Return(nil, nil)
	} else {
		instruction.EXPECT().Execute(mock.Anything).Maybe().Return(nil, errors.New("expected error for test"))
	}

	return instruction
}

func executeSynchronously(t *testing.T, executor *StartosisExecutor, dryRun bool, instructionsPlan *instructions_plan.InstructionsPlan) (string, []*kurtosis_core_rpc_api_bindings.StarlarkInstruction, *kurtosis_core_rpc_api_bindings.StarlarkExecutionError) {
	scriptOutput := strings.Builder{}
	var serializedInstructions []*kurtosis_core_rpc_api_bindings.StarlarkInstruction

	scheduledInstructions, err := instructionsPlan.GeneratePlan()
	require.Nil(t, err)

	executionResponseLines := executor.Execute(context.Background(), dryRun, noParallelism, 0, scheduledInstructions, noScriptOutputObject)
	for executionResponseLine := range executionResponseLines {
		if executionResponseLine.GetError() != nil {
			return scriptOutput.String(), serializedInstructions, executionResponseLine.GetError().GetExecutionError()
		}
		if executionResponseLine.GetInstruction() != nil {
			executedKurtosisInstruction := executionResponseLine.GetInstruction()
			serializedInstructions = append(serializedInstructions, executedKurtosisInstruction)
		}
		if executionResponseLine.GetInstructionResult() != nil {
			if _, err := scriptOutput.WriteString(executionResponseLine.GetInstructionResult().GetSerializedInstructionResult()); err != nil {
				require.Nil(t, err)
			}
		}
	}
	return scriptOutput.String(), serializedInstructions, nil
}

func getEnclaveDBForTest(t *testing.T) *enclave_db.EnclaveDB {
	file, err := os.CreateTemp("/tmp", "*.db")
	defer func() {
		err = os.Remove(file.Name())
		require.NoError(t, err)
	}()

	require.NoError(t, err)
	db, err := bolt.Open(file.Name(), 0666, nil)
	require.NoError(t, err)
	enclaveDb := &enclave_db.EnclaveDB{
		DB: db,
	}

	return enclaveDb
}
