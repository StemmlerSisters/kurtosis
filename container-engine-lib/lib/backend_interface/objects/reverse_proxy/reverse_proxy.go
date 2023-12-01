package reverse_proxy

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/container"
	"net"
)

// This component is responsible for routing http traffic to the services
type ReverseProxy struct {
	status container.ContainerStatus

	// This will be nil if the container is not running
	maybePrivateIpAddr net.IP

	// HTTP port
	httpPort uint16

	// Dashboard port
	dashboardPort uint16
}

func NewReverseProxy(
	status container.ContainerStatus,
	maybePrivateIpAddr net.IP,
	httpPort uint16,
	dashboardPort uint16) *ReverseProxy {
	return &ReverseProxy{
		status:               status,
		maybePrivateIpAddr:   maybePrivateIpAddr,
		httpPort:             httpPort,
		dashboardPort:        dashboardPort,
	}
}

func (reverseProxy *ReverseProxy) GetStatus() container.ContainerStatus {
	return reverseProxy.status
}

func (reverseProxy *ReverseProxy) GetMaybePrivateIpAddr() net.IP {
	return reverseProxy.maybePrivateIpAddr
}

func (reverseProxy *ReverseProxy) GetHttpPort() uint16 {
	return reverseProxy.httpPort
}

func (reverseProxy *ReverseProxy) GetDashboardPort() uint16 {
	return reverseProxy.dashboardPort
}
