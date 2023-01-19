package add_service

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers/magic_string_helper"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	AddServiceBuiltinName = "add_service"

	serviceIdArgName = "service_id"

	serviceConfigArgName = "config"
)

func GenerateAddServiceBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		addServiceInstruction := newEmptyAddServiceInstruction(serviceNetwork, instructionPosition, runtimeValueStore)
		if interpretationError := addServiceInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, addServiceInstruction)
		returnValue, interpretationError := addServiceInstruction.makeAddServiceInterpretationReturnValue()
		if interpretationError != nil {
			return nil, interpretationError
		}
		return returnValue, nil
	}
}

type AddServiceInstruction struct {
	serviceNetwork    service_network.ServiceNetwork
	runtimeValueStore *runtime_value_store.RuntimeValueStore

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	serviceName   kurtosis_backend_service.ServiceName
	serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig
}

func newEmptyAddServiceInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, runtimeValueStore *runtime_value_store.RuntimeValueStore) *AddServiceInstruction {
	return &AddServiceInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		starlarkKwargs:    starlark.StringDict{},
		serviceName:       "",
		serviceConfig:     nil,
		runtimeValueStore: runtimeValueStore,
	}
}

func NewAddServiceInstruction(serviceNetwork service_network.ServiceNetwork, position *kurtosis_instruction.InstructionPosition, serviceName kurtosis_backend_service.ServiceName, serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig, starlarkKwargs starlark.StringDict, runtimeValueStore *runtime_value_store.RuntimeValueStore) *AddServiceInstruction {
	return &AddServiceInstruction{
		serviceNetwork:    serviceNetwork,
		position:          position,
		serviceName:       serviceName,
		serviceConfig:     serviceConfig,
		starlarkKwargs:    starlarkKwargs,
		runtimeValueStore: runtimeValueStore,
	}
}

func (instruction *AddServiceInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *AddServiceInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[serviceIdArgName]), serviceIdArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[serviceConfigArgName]), serviceConfigArgName, kurtosis_instruction.NotRepresentative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), AddServiceBuiltinName, instruction.String(), args)
}

func (instruction *AddServiceInstruction) Execute(ctx context.Context) (*string, error) {
	serviceIdStr, err := magic_string_helper.ReplaceRuntimeValueInString(string(instruction.serviceName), instruction.runtimeValueStore)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error occurred while replacing facts in service id for '%v'", instruction.serviceName)
	}
	instruction.serviceName = kurtosis_backend_service.ServiceName(serviceIdStr)
	err = instruction.replaceMagicStrings()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred replacing IP Address with actual values in add service instruction for service '%v'", instruction.serviceName)
	}

	startedService, err := instruction.serviceNetwork.StartService(ctx, instruction.serviceName, instruction.serviceConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed adding service '%s' to enclave with an unexpected error", instruction.serviceName)
	}
	instructionResult := fmt.Sprintf("Service '%s' added with service GUID '%s'", instruction.serviceName, startedService.GetRegistration().GetUUID())
	return &instructionResult, nil
}

func (instruction *AddServiceInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if partition_topology.ParsePartitionId(instruction.serviceConfig.Subnetwork) != partition_topology.DefaultPartitionId {
		if !environment.IsNetworkPartitioningEnabled() {
			return startosis_errors.NewValidationError("Service was about to be started inside subnetwork '%s' but the Kurtosis enclave was started with subnetwork capabilities disabled. Make sure to run the Starlark script with subnetwork enabled.", *instruction.serviceConfig.Subnetwork)
		}
	}
	if environment.DoesServiceIdExist(instruction.serviceName) {
		return startosis_errors.NewValidationError("There was an error validating '%v' as service ID '%v' already exists", AddServiceBuiltinName, instruction.serviceName)
	}
	for _, artifactName := range instruction.serviceConfig.FilesArtifactMountpoints {
		if !environment.DoesArtifactNameExist(artifactName) {
			return startosis_errors.NewValidationError("There was an error validating '%v' as artifact name '%v' does not exist", AddServiceBuiltinName, artifactName)
		}
	}
	environment.AddServiceId(instruction.serviceName)
	environment.AppendRequiredContainerImage(instruction.serviceConfig.ContainerImageName)
	return nil
}

func (instruction *AddServiceInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(AddServiceBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *AddServiceInstruction) replaceMagicStrings() error {
	serviceIdStr := string(instruction.serviceName)
	entryPointArgs := instruction.serviceConfig.EntrypointArgs
	for index, entryPointArg := range entryPointArgs {
		entryPointArgWithIPAddressReplaced, err := magic_string_helper.ReplaceIPAddressInString(entryPointArg, instruction.serviceNetwork, serviceIdStr)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing IP address in entry point args for '%v'", entryPointArg)
		}
		entryPointArgWithIPAddressAndRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(entryPointArgWithIPAddressReplaced, instruction.runtimeValueStore)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing runtime value in entry point args for '%v'", entryPointArg)
		}
		entryPointArgs[index] = entryPointArgWithIPAddressAndRuntimeValueReplaced
	}

	cmdArgs := instruction.serviceConfig.CmdArgs
	for index, cmdArg := range cmdArgs {
		cmdArgWithIPAddressReplaced, err := magic_string_helper.ReplaceIPAddressInString(cmdArg, instruction.serviceNetwork, serviceIdStr)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing IP address in command args for '%v'", cmdArg)
		}
		cmdArgWithIPAddressAndRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(cmdArgWithIPAddressReplaced, instruction.runtimeValueStore)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing runtime value in command args for '%v'", cmdArg)
		}
		cmdArgs[index] = cmdArgWithIPAddressAndRuntimeValueReplaced
	}

	envVars := instruction.serviceConfig.EnvVars
	for envVarName, envVarValue := range envVars {
		envVarValueWithIPAddressReplaced, err := magic_string_helper.ReplaceIPAddressInString(envVarValue, instruction.serviceNetwork, serviceIdStr)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing IP address in env vars for '%v'", envVarValue)
		}
		envVarValueWithIPAddressAndRuntimeValueReplaced, err := magic_string_helper.ReplaceRuntimeValueInString(envVarValueWithIPAddressReplaced, instruction.runtimeValueStore)
		if err != nil {
			return stacktrace.Propagate(err, "Error occurred while replacing runtime value in command args for '%v'", envVars)
		}
		envVars[envVarName] = envVarValueWithIPAddressAndRuntimeValueReplaced
	}

	return nil
}

func (instruction *AddServiceInstruction) makeAddServiceInterpretationReturnValue() (*kurtosis_types.Service, *startosis_errors.InterpretationError) {
	ports := instruction.serviceConfig.GetPrivatePorts()
	portSpecsDict := starlark.NewDict(len(ports))
	for portId, port := range ports {
		number := port.GetNumber()
		transportProtocol := port.GetTransportProtocol()
		maybeApplicationProtocol := port.GetMaybeApplicationProtocol()

		portSpec := kurtosis_types.NewPortSpec(number, transportProtocol, maybeApplicationProtocol)
		if err := portSpecsDict.SetKey(starlark.String(portId), portSpec); err != nil {
			return nil, startosis_errors.NewInterpretationError("An error occurred while creating a port spec for values "+
				"(number: '%v', transport_protocol: '%v', application_protocol: '%v') the add instruction return value",
				number, transportProtocol, maybeApplicationProtocol)
		}
	}
	ipAddress := starlark.String(fmt.Sprintf(magic_string_helper.IpAddressReplacementPlaceholderFormat, instruction.serviceName))
	returnValue := kurtosis_types.NewService(ipAddress, portSpecsDict)
	return returnValue, nil
}

func (instruction *AddServiceInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	// TODO(gb): this now handles both serviceConfig being an unnamed struct or a fully fleshed ServiceConfig type
	//  Remove code to handle struc once downstream code is migrated
	var serviceIdArg starlark.String
	var serviceConfigArgAsStruct *starlarkstruct.Struct
	var serviceConfigArg *kurtosis_types.ServiceConfig
	var isServiceConfigGenericStruct bool
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, serviceConfigArgName, &serviceConfigArg); err != nil {
		if errParsingConfigAsStruct := starlark.UnpackArgs(b.Name(), args, kwargs, serviceIdArgName, &serviceIdArg, serviceConfigArgName, &serviceConfigArgAsStruct); errParsingConfigAsStruct != nil {
			return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", AddServiceBuiltinName, args, kwargs)
		}
		isServiceConfigGenericStruct = true
	} else {
		isServiceConfigGenericStruct = false
	}
	instruction.starlarkKwargs[serviceIdArgName] = serviceIdArg
	if isServiceConfigGenericStruct {
		instruction.starlarkKwargs[serviceConfigArgName] = serviceConfigArgAsStruct
	} else {
		instruction.starlarkKwargs[serviceConfigArgName] = serviceConfigArg
	}
	instruction.starlarkKwargs.Freeze()

	serviceName, interpretationErr := kurtosis_instruction.ParseServiceId(serviceIdArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	var serviceConfig *kurtosis_core_rpc_api_bindings.ServiceConfig
	if isServiceConfigGenericStruct {
		serviceConfig, interpretationErr = kurtosis_instruction.ParseServiceConfigArg(serviceConfigArgAsStruct)
	} else {
		serviceConfig, interpretationErr = serviceConfigArg.ToKurtosisType()
	}
	if interpretationErr != nil {
		return interpretationErr
	}
	instruction.serviceName = serviceName
	instruction.serviceConfig = serviceConfig
	return nil
}
