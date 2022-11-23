package upload_files

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

const (
	UploadFilesBuiltinName = "upload_files"

	srcArgName = "src"

	artifactIdArgName            = "artifact_id?"
	nonOptionalArtifactIdArgName = "artifact_id"

	ensureCompressedFileIsLesserThanGRPCLimit = false

	emptyStarlarkString = starlark.String("")
)

type UploadFilesInstruction struct {
	serviceNetwork service_network.ServiceNetwork
	provider       startosis_modules.ModuleContentProvider

	position   *kurtosis_instruction.InstructionPosition
	src        string
	artifactId enclave_data_directory.FilesArtifactUUID

	pathOnDisk string
}

func GenerateUploadFilesBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, provider startosis_modules.ModuleContentProvider, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		srcPath, artifactUuid, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		pathOnDisk, interpretationError := provider.GetOnDiskAbsoluteFilePath(srcPath)
		if interpretationError != nil {
			return nil, interpretationError
		}
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		uploadInstruction := NewUploadFilesInstruction(instructionPosition, serviceNetwork, provider, srcPath, pathOnDisk, artifactUuid)
		*instructionsQueue = append(*instructionsQueue, uploadInstruction)
		return starlark.String(artifactUuid), nil
	}
}

func NewUploadFilesInstruction(position *kurtosis_instruction.InstructionPosition, serviceNetwork service_network.ServiceNetwork, provider startosis_modules.ModuleContentProvider, src string, pathOnDisk string, artifactId enclave_data_directory.FilesArtifactUUID) *UploadFilesInstruction {
	return &UploadFilesInstruction{
		position:       position,
		serviceNetwork: serviceNetwork,
		src:            src,
		provider:       provider,
		pathOnDisk:     pathOnDisk,
		artifactId:     artifactId,
	}
}

func (instruction *UploadFilesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *UploadFilesInstruction) GetCanonicalInstruction() string {
	return shared_helpers.CanonicalizeInstruction(UploadFilesBuiltinName, kurtosis_instruction.NoArgs, instruction.getKwargs())
}

func (instruction *UploadFilesInstruction) Execute(_ context.Context) (*string, error) {
	compressedData, err := shared_utils.CompressPath(instruction.pathOnDisk, ensureCompressedFileIsLesserThanGRPCLimit)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while compressing the files '%v'", instruction.pathOnDisk)
	}
	err = instruction.serviceNetwork.UploadFilesArtifactToTargetArtifactUUID(compressedData, instruction.artifactId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while uploading the compressed contents\n'%v'", compressedData)
	}
	logrus.Infof("Succesfully uploaded files from instruction '%v' to '%v'", instruction.position.String(), instruction.artifactId)
	return nil, nil
}

func (instruction *UploadFilesInstruction) String() string {
	return instruction.GetCanonicalInstruction()
}

func (instruction *UploadFilesInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	// this doesn't do anything but can't return an error as the validator runs this regardless
	// this is a no-op
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (string, enclave_data_directory.FilesArtifactUUID, *startosis_errors.InterpretationError) {

	var srcPathArg starlark.String
	var artifactUuidArg = emptyStarlarkString
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, srcArgName, &srcPathArg, artifactIdArgName, &artifactUuidArg); err != nil {
		return "", "", startosis_errors.NewInterpretationError(err.Error())
	}

	if artifactUuidArg == emptyStarlarkString {
		placeHolderArtifactUuid, err := enclave_data_directory.NewFilesArtifactUUID()
		if err != nil {
			return "", "", startosis_errors.NewInterpretationError("An empty or no artifact_uuid was passed, we tried creating one but failed")
		}
		artifactUuidArg = starlark.String(placeHolderArtifactUuid)
	}

	srcPath, interpretationErr := kurtosis_instruction.ParseNonEmptyString(srcArgName, srcPathArg)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}

	artifactUuid, interpretationErr := kurtosis_instruction.ParseArtifactUuid(nonOptionalArtifactIdArgName, artifactUuidArg)
	if interpretationErr != nil {
		return "", "", interpretationErr
	}

	return srcPath, artifactUuid, nil
}

func (instruction *UploadFilesInstruction) getKwargs() starlark.StringDict {
	return starlark.StringDict{
		srcArgName:                   starlark.String(instruction.src),
		nonOptionalArtifactIdArgName: starlark.String(instruction.artifactId),
	}
}
