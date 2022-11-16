package service_network

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"net"
	"net/http"
)

type ServiceNetwork interface {
	Repartition(
		ctx context.Context,
		newPartitionServices map[service_network_types.PartitionID]map[service.ServiceID]bool,
		newPartitionConnections map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection,
		newDefaultConnection partition_topology.PartitionConnection,
	) error

	StartServices(
		ctx context.Context,
		serviceConfigs map[service.ServiceID]*kurtosis_core_rpc_api_bindings.ServiceConfig,
		partitionID service_network_types.PartitionID,
	) (
		map[service.ServiceID]*service.Service,
		map[service.ServiceID]error,
		error,
	)

	RemoveService(
		ctx context.Context,
		serviceId service.ServiceID,
	) (service.ServiceGUID, error)

	PauseService(
		ctx context.Context,
		serviceId service.ServiceID,
	) error

	UnpauseService(
		ctx context.Context,
		serviceId service.ServiceID,
	) error

	ExecCommand(
		ctx context.Context,
		serviceId service.ServiceID,
		command []string,
	) (int32, string, error)

	HttpRequestService(
		ctx context.Context,
		serviceId service.ServiceID,
		portId string,
		method string,
		contentType string,
		endpoint string,
		body string,
	) (*http.Response, error)

	GetService(ctx context.Context, serviceId service.ServiceID) (
		*service.Service,
		error,
	)

	CopyFilesFromService(
		ctx context.Context,
		serviceId service.ServiceID,
		srcPath string,
	) (
		enclave_data_directory.FilesArtifactUUID,
		error,
	)
	CopyFilesFromServiceToTargetArtifactUUID(
		ctx context.Context,
		serviceId service.ServiceID,
		srcPath string,
		filesArtifactUuid enclave_data_directory.FilesArtifactUUID,
	) (
		enclave_data_directory.FilesArtifactUUID,
		error,
	)

	GetServiceIDs() map[service.ServiceID]bool

	GetIPAddressForService(serviceID service.ServiceID) (net.IP, bool)

	RenderTemplates(templatesAndDataByDestinationRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData) (enclave_data_directory.FilesArtifactUUID, error)
	RenderTemplatesToTargetFilesArtifactUUID(templatesAndDataByDestinationRelFilepath map[string]*kurtosis_core_rpc_api_bindings.RenderTemplatesToFilesArtifactArgs_TemplateAndData, filesArtifactUuid enclave_data_directory.FilesArtifactUUID) (enclave_data_directory.FilesArtifactUUID, error)

	UploadFilesArtifact(data []byte) (enclave_data_directory.FilesArtifactUUID, error)
	UploadFilesArtifactToTargetArtifactUUID(data []byte, targetFilesArtifactUuid enclave_data_directory.FilesArtifactUUID) error
}
