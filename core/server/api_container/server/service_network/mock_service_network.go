/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package service_network

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-core/launcher/enclave_container_launcher"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"time"
)

type MockServiceNetwork struct {
	servicePrivateIps                map[service_network_types.ServiceID]net.IP
	serviceEnclaveDataDirMntDirpaths map[service_network_types.ServiceID]string
}

func NewMockServiceNetwork(serviceIps map[service_network_types.ServiceID]net.IP, serviceEnclaveDataDirMntDirpaths map[service_network_types.ServiceID]string) *MockServiceNetwork {
	return &MockServiceNetwork{servicePrivateIps: serviceIps, serviceEnclaveDataDirMntDirpaths: serviceEnclaveDataDirMntDirpaths}
}

func (m MockServiceNetwork) Repartition(ctx context.Context, newPartitionServices map[service_network_types.PartitionID]*service_network_types.ServiceIDSet, newPartitionConnections map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection, newDefaultConnection partition_topology.PartitionConnection) error {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) RegisterService(serviceId service_network_types.ServiceID, partitionId service_network_types.PartitionID) (net.IP, string, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) StartService(ctx context.Context, serviceId service_network_types.ServiceID, imageName string, privatePorts map[string]*enclave_container_launcher.EnclaveContainerPort, entrypointArgs []string, cmdArgs []string, dockerEnvVars map[string]string, enclaveDataDirMountDirpath string, filesArtifactMountDirpaths map[string]string) (resultMaybePublicIpAddr net.IP, resultPublicPorts map[string]*enclave_container_launcher.EnclaveContainerPort, resultErr error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) RemoveService(ctx context.Context, serviceId service_network_types.ServiceID, containerStopTimeout time.Duration) error {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) ExecCommand(ctx context.Context, serviceId service_network_types.ServiceID, command []string) (int32, string, error) {
	panic("This is unimplemented for the mock network")
}

func (m MockServiceNetwork) GetServiceRegistrationInfo(serviceId service_network_types.ServiceID) (privateIpAddr net.IP, relativeServiceDirpath string, resultErr error) {
	ip, found := m.servicePrivateIps[serviceId]
	if !found {
		return nil, "", stacktrace.NewError("No private IP defined for service with ID '%v'", serviceId)
	}
	return ip, "", nil
}

func (m MockServiceNetwork) GetServiceRunInfo(serviceId service_network_types.ServiceID) (privatePorts map[string]*enclave_container_launcher.EnclaveContainerPort, maybePublicIpAddr net.IP, publicPorts map[string]*enclave_container_launcher.EnclaveContainerPort, enclaveDataDirMntDirpath string, resultErr error) {
	dataDirMntDirpath, found := m.serviceEnclaveDataDirMntDirpaths[serviceId]
	if !found {
		return nil, nil, nil, "", stacktrace.NewError("No enclave data directory mount dirpath defined for service with ID '%v'", serviceId)
	}
	return nil, nil, nil, dataDirMntDirpath, nil
}

func (m MockServiceNetwork) GetServiceIDs() map[service_network_types.ServiceID]bool {
	panic("This is unimplemented for the mock network")
}
