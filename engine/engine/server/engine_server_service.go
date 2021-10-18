package server

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/engine/enclave_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

type EngineServerService struct {
	// This embedding is required by gRPC
	kurtosis_engine_rpc_api_bindings.UnimplementedEngineServiceServer

	enclaveManager *enclave_manager.EnclaveManager
}

func NewEngineServerService(enclaveManager *enclave_manager.EnclaveManager) *EngineServerService {
	service := &EngineServerService{
		enclaveManager: enclaveManager,
	}
	return service
}

func (service *EngineServerService) CreateEnclave(ctx context.Context, args *kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs) (*kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse, error) {
	logrus.Debugf("Received request to create new enclave with the following args: %+v", args)

	apiContainerLogLevel, err := logrus.ParseLevel(args.ApiContainerLogLevel)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", args.ApiContainerLogLevel)
	}

	networkId, networkIpAndMask, apiContainerId, apiContainerIpAddr, apiContainerHostPortBinding, err := service.enclaveManager.CreateEnclave(
		ctx,
		args.ApiContainerImage,
		apiContainerLogLevel,
		args.EnclaveId,
		args.IsPartitioningEnabled,
		args.ShouldPublishAllPorts,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating new enclave with ID '%v'", args.EnclaveId)
	}

	response := &kurtosis_engine_rpc_api_bindings.CreateEnclaveResponse{
		NetworkId:                   networkId,
		NetworkCidr:                 networkIpAndMask.String(),
		ApiContainerId:              apiContainerId,
		ApiContainerIpInsideNetwork: apiContainerIpAddr.String(),
		ApiContainerHostIp:          apiContainerHostPortBinding.HostIP,
		ApiContainerHostPort:        apiContainerHostPortBinding.HostPort,
	}

	return response, nil
}
