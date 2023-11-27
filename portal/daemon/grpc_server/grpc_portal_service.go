package grpc_server

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/portal/kurtosis_portal_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/portal/daemon/port_forward_manager"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

const (
	PortalServiceGrpcPort = 9502
)

type GrpcPortalService struct {
	sync.RWMutex

	portForwardManager *port_forward_manager.PortForwardManager
}

func NewPortalService(manager *port_forward_manager.PortForwardManager) *GrpcPortalService {
	return &GrpcPortalService{
		RWMutex:            sync.RWMutex{},
		portForwardManager: manager,
	}
}

func (service *GrpcPortalService) Ping(ctx context.Context, ping *kurtosis_portal_rpc_api_bindings.PortalPing) (*kurtosis_portal_rpc_api_bindings.PortalPong, error) {
	err := service.portForwardManager.Ping(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Portal Daemon is running but the Port Forward Manager failed to respond to the ping")
	}
	return &kurtosis_portal_rpc_api_bindings.PortalPong{}, nil
}

func (service *GrpcPortalService) CreateUserServicePortForward(ctx context.Context, args *kurtosis_portal_rpc_api_bindings.CreateUserServicePortForwardArgs) (*kurtosis_portal_rpc_api_bindings.CreateUserServicePortForwardResponse, error) {
	enclaveServicePort := toInternalEnclaveServicePort(args.GetEnclaveServicePortId())
	localPort, err := service.portForwardManager.ForwardUserServiceToPort(ctx, enclaveServicePort, uint16(args.GetLocalPortNumber()))

	if err != nil {
		return nil, err
	}
	return &kurtosis_portal_rpc_api_bindings.CreateUserServicePortForwardResponse{LocalPortNumber: uint32(localPort)}, nil
}

func (service *GrpcPortalService) Close() error {
	service.Lock()
	defer service.Unlock()

	// TODO(omar): implement

	return nil
}

func toInternalEnclaveServicePort(esp *kurtosis_portal_rpc_api_bindings.EnclaveServicePortId) port_forward_manager.EnclaveServicePort {
	return port_forward_manager.NewEnclaveServicePort(esp.GetEnclaveId(), esp.GetServiceId(), esp.GetPortId())
}
