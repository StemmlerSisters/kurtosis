// Package kurtosis_core_http_api_bindings provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version (devel) DO NOT EDIT.
package kurtosis_core_http_api_bindings

import (
	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
)

// Defines values for Connect.
const (
	CONNECT   Connect = "CONNECT"
	NOCONNECT Connect = "NO_CONNECT"
)

// Defines values for ContainerStatus.
const (
	ContainerStatusRUNNING ContainerStatus = "RUNNING"
	ContainerStatusSTOPPED ContainerStatus = "STOPPED"
	ContainerStatusUNKNOWN ContainerStatus = "UNKNOWN"
)

// Defines values for ImageDownloadMode.
const (
	ImageDownloadModeALWAYS  ImageDownloadMode = "ALWAYS"
	ImageDownloadModeMISSING ImageDownloadMode = "MISSING"
)

// Defines values for KurtosisFeatureFlag.
const (
	NOINSTRUCTIONSCACHING KurtosisFeatureFlag = "NO_INSTRUCTIONS_CACHING"
)

// Defines values for PortTransportProtocol.
const (
	SCTP PortTransportProtocol = "SCTP"
	TCP  PortTransportProtocol = "TCP"
	UDP  PortTransportProtocol = "UDP"
)

// Defines values for RestartPolicy.
const (
	RestartPolicyALWAYS RestartPolicy = "ALWAYS"
	RestartPolicyNEVER  RestartPolicy = "NEVER"
)

// Defines values for ServiceStatus.
const (
	ServiceStatusRUNNING ServiceStatus = "RUNNING"
	ServiceStatusSTOPPED ServiceStatus = "STOPPED"
	ServiceStatusUNKNOWN ServiceStatus = "UNKNOWN"
)

// Defines values for WaitForEndpointAvailabilityArgsHttpMethod.
const (
	GET  WaitForEndpointAvailabilityArgsHttpMethod = "GET"
	POST WaitForEndpointAvailabilityArgsHttpMethod = "POST"
)

// Connect 0 - CONNECT // Best effort port forwarding
// 1 - NO_CONNECT // Port forwarding disabled
type Connect string

// ConnectServicesArgs defines model for ConnectServicesArgs.
type ConnectServicesArgs struct {
	// Connect 0 - CONNECT // Best effort port forwarding
	// 1 - NO_CONNECT // Port forwarding disabled
	Connect *Connect `json:"connect,omitempty"`
}

// ConnectServicesResponse defines model for ConnectServicesResponse.
type ConnectServicesResponse = map[string]interface{}

// Container defines model for Container.
type Container struct {
	CmdArgs        *[]string          `json:"cmd_args,omitempty"`
	EntrypointArgs *[]string          `json:"entrypoint_args,omitempty"`
	EnvVars        *map[string]string `json:"env_vars,omitempty"`
	ImageName      *string            `json:"image_name,omitempty"`

	// Status 0 - STOPPED
	// 1 - RUNNING
	// 2 - UNKNOWN
	Status *ContainerStatus `json:"status,omitempty"`
}

// ContainerStatus 0 - STOPPED
// 1 - RUNNING
// 2 - UNKNOWN
type ContainerStatus string

// DataChunkMetadata defines model for DataChunkMetadata.
type DataChunkMetadata struct {
	Name *string `json:"name,omitempty"`
}

// ExecCommandArgs Exec Command
type ExecCommandArgs struct {
	CommandArgs *[]string `json:"command_args,omitempty"`
}

// ExecCommandResponse defines model for ExecCommandResponse.
type ExecCommandResponse struct {
	ExitCode *int32 `json:"exit_code,omitempty"`

	// LogOutput Assumes UTF-8 encoding
	LogOutput *string `json:"log_output,omitempty"`
}

// FileArtifactContentsFileDescription defines model for FileArtifactContentsFileDescription.
type FileArtifactContentsFileDescription struct {
	// Path Path relative to the file artifact
	Path *string `json:"path,omitempty"`

	// Size Size of the file, in bytes
	Size *int64 `json:"size,omitempty"`

	// TextPreview A bit of text content, if the file allows (similar to UNIX's 'head')
	TextPreview *string `json:"text_preview,omitempty"`
}

// FileArtifactReference Files Artifact identifier
type FileArtifactReference struct {
	// Name UUID of the files artifact, for use when referencing it in the future
	Name *string `json:"name,omitempty"`

	// Uuid UUID of the files artifact, for use when referencing it in the future
	Uuid *string `json:"uuid,omitempty"`
}

// GetExistingAndHistoricalServiceIdentifiersResponse defines model for GetExistingAndHistoricalServiceIdentifiersResponse.
type GetExistingAndHistoricalServiceIdentifiersResponse struct {
	AllIdentifiers *[]ServiceIdentifiers `json:"allIdentifiers,omitempty"`
}

// GetServicesResponse defines model for GetServicesResponse.
type GetServicesResponse struct {
	ServiceInfo *ServiceInfo `json:"service_info,omitempty"`
}

// GetStarlarkRunResponse defines model for GetStarlarkRunResponse.
type GetStarlarkRunResponse struct {
	ExperimentalFeatures   *[]KurtosisFeatureFlag `json:"experimental_features,omitempty"`
	MainFunctionName       *string                `json:"main_function_name,omitempty"`
	PackageId              *string                `json:"package_id,omitempty"`
	Parallelism            *int32                 `json:"parallelism,omitempty"`
	RelativePathToMainFile *string                `json:"relative_path_to_main_file,omitempty"`

	// RestartPolicy 0 - NEVER
	// 1 - ALWAYS
	RestartPolicy    *RestartPolicy `json:"restart_policy,omitempty"`
	SerializedParams *string        `json:"serialized_params,omitempty"`
	SerializedScript *string        `json:"serialized_script,omitempty"`
}

// ImageDownloadMode 0 - ALWAYS
// 1 - MISSING
type ImageDownloadMode string

// InspectFilesArtifactContentsResponse defines model for InspectFilesArtifactContentsResponse.
type InspectFilesArtifactContentsResponse struct {
	FileDescriptions *[]FileArtifactContentsFileDescription `json:"file_descriptions,omitempty"`
}

// KurtosisFeatureFlag 0 - NO_INSTRUCTIONS_CACHING
type KurtosisFeatureFlag string

// ListFilesArtifactNamesAndUuidsResponse defines model for ListFilesArtifactNamesAndUuidsResponse.
type ListFilesArtifactNamesAndUuidsResponse struct {
	FileNamesAndUuids *[]FileArtifactReference `json:"file_names_and_uuids,omitempty"`
}

// Port Shared Objects (Used By Multiple Endpoints)
type Port struct {
	MaybeApplicationProtocol *string `json:"maybe_application_protocol,omitempty"`

	// MaybeWaitTimeout The wait timeout duration in string
	MaybeWaitTimeout *string `json:"maybe_wait_timeout,omitempty"`
	Number           int32   `json:"number"`

	// TransportProtocol 0 - TCP
	// 1 - SCTP
	// 2 - UDP
	TransportProtocol PortTransportProtocol `json:"transport_protocol"`
}

// PortTransportProtocol 0 - TCP
// 1 - SCTP
// 2 - UDP
type PortTransportProtocol string

// RestartPolicy 0 - NEVER
// 1 - ALWAYS
type RestartPolicy string

// RunStarlarkPackageArgs defines model for RunStarlarkPackageArgs.
type RunStarlarkPackageArgs struct {
	// ClonePackage Whether the package should be cloned or not.
	// If false, then the package will be pulled from the APIC local package store. If it's a local package then is must
	// have been uploaded using UploadStarlarkPackage prior to calling RunStarlarkPackage.
	// If true, then the package will be cloned from GitHub before execution starts
	ClonePackage *bool `json:"clone_package,omitempty"`

	// CloudInstanceId Defaults to empty
	CloudInstanceId *string `json:"cloud_instance_id,omitempty"`

	// CloudUserId Defaults to empty
	CloudUserId *string `json:"cloud_user_id,omitempty"`

	// DryRun Defaults to false
	DryRun               *bool                  `json:"dry_run,omitempty"`
	ExperimentalFeatures *[]KurtosisFeatureFlag `json:"experimental_features,omitempty"`

	// ImageDownloadMode 0 - ALWAYS
	// 1 - MISSING
	ImageDownloadMode *ImageDownloadMode `json:"image_download_mode,omitempty"`

	// Local the payload of the local module
	Local *[]byte `json:"local,omitempty"`

	// MainFunctionName The name of the main function, the default value is "run"
	MainFunctionName *string `json:"main_function_name,omitempty"`
	PackageId        *string `json:"package_id,omitempty"`

	// Parallelism Defaults to 4
	Parallelism *int32 `json:"parallelism,omitempty"`

	// RelativePathToMainFile The relative main file filepath, the default value is the "main.star" file in the root of a package
	RelativePathToMainFile *string `json:"relative_path_to_main_file,omitempty"`

	// Remote just a flag to indicate the module must be cloned inside the API
	Remote *bool `json:"remote,omitempty"`

	// SerializedParams Serialized parameters data for the Starlark package main function
	// This should be a valid JSON string
	SerializedParams *string `json:"serialized_params,omitempty"`
}

// RunStarlarkScriptArgs defines model for RunStarlarkScriptArgs.
type RunStarlarkScriptArgs struct {
	// CloudInstanceId Defaults to empty
	CloudInstanceId *string `json:"cloud_instance_id,omitempty"`

	// CloudUserId Defaults to empty
	CloudUserId *string `json:"cloud_user_id,omitempty"`

	// DryRun Defaults to false
	DryRun               *bool                  `json:"dry_run,omitempty"`
	ExperimentalFeatures *[]KurtosisFeatureFlag `json:"experimental_features,omitempty"`

	// ImageDownloadMode 0 - ALWAYS
	// 1 - MISSING
	ImageDownloadMode *ImageDownloadMode `json:"image_download_mode,omitempty"`

	// MainFunctionName The name of the main function, the default value is "run"
	MainFunctionName *string `json:"main_function_name,omitempty"`

	// Parallelism Defaults to 4
	Parallelism      *int32  `json:"parallelism,omitempty"`
	SerializedParams *string `json:"serialized_params,omitempty"`
	SerializedScript *string `json:"serialized_script,omitempty"`
}

// ServiceIdentifiers An service identifier is a collection of uuid, name and shortened uuid
type ServiceIdentifiers struct {
	// Name Name of the service
	Name *string `json:"name,omitempty"`

	// ServiceUuid UUID of the service
	ServiceUuid *string `json:"service_uuid,omitempty"`

	// ShortenedUuid The shortened uuid of the service
	ShortenedUuid *string `json:"shortened_uuid,omitempty"`
}

// ServiceInfo defines model for ServiceInfo.
type ServiceInfo struct {
	Container *Container `json:"container,omitempty"`

	// MaybePublicIpAddr Public IP address *outside* the enclave where the service is reachable
	// NOTE: Will be empty if the service isn't running, the service didn't define any ports, or the backend doesn't support reporting public service info
	MaybePublicIpAddr *string `json:"maybe_public_ip_addr,omitempty"`

	// MaybePublicPorts Shared Objects (Used By Multiple Endpoints)
	MaybePublicPorts *Port `json:"maybe_public_ports,omitempty"`

	// Name Name of the service
	Name *string `json:"name,omitempty"`

	// PrivateIpAddr The IP address of the service inside the enclave
	PrivateIpAddr *string `json:"private_ip_addr,omitempty"`

	// PrivatePorts Shared Objects (Used By Multiple Endpoints)
	PrivatePorts *Port `json:"private_ports,omitempty"`

	// ServiceStatus 0 - STOPPED
	// 1 - RUNNING
	// 2 - UNKNOWN
	ServiceStatus *ServiceStatus `json:"service_status,omitempty"`

	// ServiceUuid UUID of the service
	ServiceUuid *string `json:"service_uuid,omitempty"`

	// ShortenedUuid Shortened uuid of the service
	ShortenedUuid *string `json:"shortened_uuid,omitempty"`
}

// ServiceStatus 0 - STOPPED
// 1 - RUNNING
// 2 - UNKNOWN
type ServiceStatus string

// StarlarkError defines model for StarlarkError.
type StarlarkError struct {
	ExecutionError      *StarlarkExecutionError      `json:"execution_error,omitempty"`
	InterpretationError *StarlarkInterpretationError `json:"interpretation_error,omitempty"`
	ValidationError     *StarlarkValidationError     `json:"validation_error,omitempty"`
}

// StarlarkExecutionError defines model for StarlarkExecutionError.
type StarlarkExecutionError struct {
	ErrorMessage *string `json:"error_message,omitempty"`
}

// StarlarkInfo defines model for StarlarkInfo.
type StarlarkInfo struct {
	InfoMessage *string `json:"info_message,omitempty"`
}

// StarlarkInstruction defines model for StarlarkInstruction.
type StarlarkInstruction struct {
	Arguments             *[]StarlarkInstructionArg    `json:"arguments,omitempty"`
	ExecutableInstruction *string                      `json:"executable_instruction,omitempty"`
	InstructionName       *string                      `json:"instruction_name,omitempty"`
	IsSkipped             *bool                        `json:"is_skipped,omitempty"`
	Position              *StarlarkInstructionPosition `json:"position,omitempty"`
}

// StarlarkInstructionArg defines model for StarlarkInstructionArg.
type StarlarkInstructionArg struct {
	ArgName            *string `json:"arg_name,omitempty"`
	IsRepresentative   *bool   `json:"is_representative,omitempty"`
	SerializedArgValue *string `json:"serialized_arg_value,omitempty"`
}

// StarlarkInstructionPosition defines model for StarlarkInstructionPosition.
type StarlarkInstructionPosition struct {
	Column   *int32  `json:"column,omitempty"`
	Filename *string `json:"filename,omitempty"`
	Line     *int32  `json:"line,omitempty"`
}

// StarlarkInstructionResult defines model for StarlarkInstructionResult.
type StarlarkInstructionResult struct {
	SerializedInstructionResult *string `json:"serialized_instruction_result,omitempty"`
}

// StarlarkInterpretationError defines model for StarlarkInterpretationError.
type StarlarkInterpretationError struct {
	ErrorMessage *string `json:"error_message,omitempty"`
}

// StarlarkRunFinishedEvent defines model for StarlarkRunFinishedEvent.
type StarlarkRunFinishedEvent struct {
	IsRunSuccessful  *bool   `json:"is_run_successful,omitempty"`
	SerializedOutput *string `json:"serialized_output,omitempty"`
}

// StarlarkRunProgress defines model for StarlarkRunProgress.
type StarlarkRunProgress struct {
	CurrentStepInfo   *[]string `json:"current_step_info,omitempty"`
	CurrentStepNumber *int32    `json:"current_step_number,omitempty"`
	TotalSteps        *int32    `json:"total_steps,omitempty"`
}

// StarlarkRunResponseLine Starlark Execution Response
type StarlarkRunResponseLine struct {
	Error             *StarlarkError             `json:"error,omitempty"`
	Info              *StarlarkInfo              `json:"info,omitempty"`
	Instruction       *StarlarkInstruction       `json:"instruction,omitempty"`
	InstructionResult *StarlarkInstructionResult `json:"instruction_result,omitempty"`
	ProgressInfo      *StarlarkRunProgress       `json:"progress_info,omitempty"`
	RunFinishedEvent  *StarlarkRunFinishedEvent  `json:"run_finished_event,omitempty"`
	Warning           *StarlarkWarning           `json:"warning,omitempty"`
}

// StarlarkValidationError defines model for StarlarkValidationError.
type StarlarkValidationError struct {
	ErrorMessage *string `json:"error_message,omitempty"`
}

// StarlarkWarning defines model for StarlarkWarning.
type StarlarkWarning struct {
	WarningMessage *string `json:"warning_message,omitempty"`
}

// StoreFilesArtifactFromServiceArgs defines model for StoreFilesArtifactFromServiceArgs.
type StoreFilesArtifactFromServiceArgs struct {
	// Name The name of the files artifact
	Name string `json:"name"`

	// SourcePath The absolute source path where the source files will be copied from
	SourcePath string `json:"source_path"`
}

// StoreWebFilesArtifactArgs Store Web Files Artifact
type StoreWebFilesArtifactArgs struct {
	// Name The name of the files artifact
	Name string `json:"name"`

	// Url URL to download the artifact from
	Url string `json:"url"`
}

// StreamedDataChunk Streamed Data Chunk
type StreamedDataChunk struct {
	// Data Chunk of the overall files artifact bytes
	Data     *[]byte            `json:"data,omitempty"`
	Metadata *DataChunkMetadata `json:"metadata,omitempty"`

	// PreviousChunkHash Hash of the PREVIOUS chunk, or empty string is this is the first chunk
	// Referencing the previous chunk via its hash allows Kurtosis to validate
	// the consistency of the data in case some chunk were not received
	PreviousChunkHash *string `json:"previous_chunk_hash,omitempty"`
}

// WaitForEndpointAvailabilityArgs Wait For HTTP Endpoint Availability
type WaitForEndpointAvailabilityArgs struct {
	// BodyText If the endpoint returns this value, the service will be marked as available (e.g. Hello World).
	BodyText   *string                                    `json:"body_text,omitempty"`
	HttpMethod *WaitForEndpointAvailabilityArgsHttpMethod `json:"http_method,omitempty"`

	// InitialDelayMilliseconds The number of milliseconds to wait until executing the first HTTP call
	InitialDelayMilliseconds *int32 `json:"initial_delay_milliseconds,omitempty"`

	// Path The path of the service to check. It mustn't start with the first slash. For instance `service/health`
	Path *string `json:"path,omitempty"`

	// Retries Max number of HTTP call attempts that this will execute until giving up and returning an error
	Retries *int32 `json:"retries,omitempty"`

	// RetriesDelayMilliseconds Number of milliseconds to wait between retries
	RetriesDelayMilliseconds *int32 `json:"retries_delay_milliseconds,omitempty"`
}

// WaitForEndpointAvailabilityArgsHttpMethod defines model for WaitForEndpointAvailabilityArgs.HttpMethod.
type WaitForEndpointAvailabilityArgsHttpMethod string

// ArtifactIdentifier defines model for artifact_identifier.
type ArtifactIdentifier = string

// EnclaveIdentifier The package identifier that will be executed
type EnclaveIdentifier = string

// PackageId defines model for package_id.
type PackageId = string

// PortNumber defines model for port_number.
type PortNumber = int32

// ServiceIdentifier defines model for service_identifier.
type ServiceIdentifier = string

// PostEnclavesEnclaveIdentifierArtifactsLocalFileMultipartBody defines parameters for PostEnclavesEnclaveIdentifierArtifactsLocalFile.
type PostEnclavesEnclaveIdentifierArtifactsLocalFileMultipartBody = openapi_types.File

// GetEnclavesEnclaveIdentifierServicesServiceIdentifierParams defines parameters for GetEnclavesEnclaveIdentifierServicesServiceIdentifier.
type GetEnclavesEnclaveIdentifierServicesServiceIdentifierParams struct {
	// AdditionalProperties Additional properties
	AdditionalProperties *string `form:"additional-properties,omitempty" json:"additional-properties,omitempty"`
}

// PostEnclavesEnclaveIdentifierStarlarkPackagesMultipartBody defines parameters for PostEnclavesEnclaveIdentifierStarlarkPackages.
type PostEnclavesEnclaveIdentifierStarlarkPackagesMultipartBody = openapi_types.File

// PostEnclavesEnclaveIdentifierArtifactsLocalFileMultipartRequestBody defines body for PostEnclavesEnclaveIdentifierArtifactsLocalFile for multipart/form-data ContentType.
type PostEnclavesEnclaveIdentifierArtifactsLocalFileMultipartRequestBody = PostEnclavesEnclaveIdentifierArtifactsLocalFileMultipartBody

// PutEnclavesEnclaveIdentifierArtifactsRemoteFileJSONRequestBody defines body for PutEnclavesEnclaveIdentifierArtifactsRemoteFile for application/json ContentType.
type PutEnclavesEnclaveIdentifierArtifactsRemoteFileJSONRequestBody = StoreWebFilesArtifactArgs

// PutEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifierJSONRequestBody defines body for PutEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifier for application/json ContentType.
type PutEnclavesEnclaveIdentifierArtifactsServicesServiceIdentifierJSONRequestBody = StoreFilesArtifactFromServiceArgs

// PostEnclavesEnclaveIdentifierServicesConnectionJSONRequestBody defines body for PostEnclavesEnclaveIdentifierServicesConnection for application/json ContentType.
type PostEnclavesEnclaveIdentifierServicesConnectionJSONRequestBody = ConnectServicesArgs

// PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommandJSONRequestBody defines body for PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommand for application/json ContentType.
type PostEnclavesEnclaveIdentifierServicesServiceIdentifierCommandJSONRequestBody = ExecCommandArgs

// PostEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailabilityJSONRequestBody defines body for PostEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailability for application/json ContentType.
type PostEnclavesEnclaveIdentifierServicesServiceIdentifierEndpointsPortNumberAvailabilityJSONRequestBody = WaitForEndpointAvailabilityArgs

// PostEnclavesEnclaveIdentifierStarlarkPackagesMultipartRequestBody defines body for PostEnclavesEnclaveIdentifierStarlarkPackages for multipart/form-data ContentType.
type PostEnclavesEnclaveIdentifierStarlarkPackagesMultipartRequestBody = PostEnclavesEnclaveIdentifierStarlarkPackagesMultipartBody

// PostEnclavesEnclaveIdentifierStarlarkPackagesPackageIdJSONRequestBody defines body for PostEnclavesEnclaveIdentifierStarlarkPackagesPackageId for application/json ContentType.
type PostEnclavesEnclaveIdentifierStarlarkPackagesPackageIdJSONRequestBody = RunStarlarkPackageArgs

// PostEnclavesEnclaveIdentifierStarlarkScriptsJSONRequestBody defines body for PostEnclavesEnclaveIdentifierStarlarkScripts for application/json ContentType.
type PostEnclavesEnclaveIdentifierStarlarkScriptsJSONRequestBody = RunStarlarkScriptArgs
