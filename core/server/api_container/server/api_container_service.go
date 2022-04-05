/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	kurtosis_backend_service "github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis-core/launcher/enclave_container_launcher"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/bulk_command_execution_engine"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/bulk_command_execution_engine/v0_bulk_command_execution"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/module_store"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/module_store/module_store_types"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"math"
	"net/http"
	"time"
)

const (
	// Custom-set max size for logs coming back from docker exec.
	// Protobuf sets a maximum of 2GB for responses, in interest of keeping performance sane
	// we pick a reasonable limit of 10MB on log responses for docker exec.
	// See: https://stackoverflow.com/questions/34128872/google-protobuf-maximum-size/34186672
	maxLogOutputSizeBytes = 10 * 1024 * 1024

	// The string returned by the API if a service's public IP address doesn't exist
	missingPublicIpAddrStr = ""
)

// Guaranteed (by a unit test) to be a 1:1 mapping between API port protos and port spec protos
var apiContainerPortProtoToPortSpecPortProto = map[kurtosis_core_rpc_api_bindings.Port_Protocol]port_spec.PortProtocol{
	kurtosis_core_rpc_api_bindings.Port_TCP:  port_spec.PortProtocol_TCP,
	kurtosis_core_rpc_api_bindings.Port_SCTP: port_spec.PortProtocol_SCTP,
	kurtosis_core_rpc_api_bindings.Port_UDP:  port_spec.PortProtocol_UDP,
}

type ApiContainerService struct {
	// This embedding is required by gRPC
	kurtosis_core_rpc_api_bindings.UnimplementedApiContainerServiceServer

	enclaveDataDir *enclave_data_directory.EnclaveDataDirectory

	serviceNetwork service_network.ServiceNetwork

	moduleStore *module_store.ModuleStore

	bulkCmdExecEngine *bulk_command_execution_engine.BulkCommandExecutionEngine

	metricsClient client.MetricsClient
}

func NewApiContainerService(
	enclaveDirectory *enclave_data_directory.EnclaveDataDirectory,
	serviceNetwork service_network.ServiceNetwork,
	moduleStore *module_store.ModuleStore,
	metricsClient client.MetricsClient,
) (*ApiContainerService, error) {
	service := &ApiContainerService{
		enclaveDataDir:         enclaveDirectory,
		serviceNetwork:         serviceNetwork,
		moduleStore:            moduleStore,
		metricsClient:          metricsClient,
	}

	// NOTE: This creates a circular dependency between ApiContainerService <-> BulkCommandExecutionEngine, but out
	//  necessity: the API service must farm bulk commands out to the bulk command execution engine, which must call
	//  back to the API service to actually do work.
	v0BulkCmdProcessor, err := v0_bulk_command_execution.NewV0BulkCommandProcessor(serviceNetwork, service)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the v0 bulk command processor")
	}
	bulkCmdExecEngine := bulk_command_execution_engine.NewBulkCommandExecutionEngine(v0BulkCmdProcessor)
	service.bulkCmdExecEngine = bulkCmdExecEngine

	return service, nil
}

func (service ApiContainerService) LoadModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.LoadModuleArgs) (*kurtosis_core_rpc_api_bindings.LoadModuleResponse, error) {
	moduleId := module_store_types.ModuleID(args.ModuleId)
	image := args.ContainerImage
	serializedParams := args.SerializedParams

	if err := service.metricsClient.TrackLoadModule(args.ModuleId, image, serializedParams); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Errorf("An error occurred tracking load module event\n%v", err)
	}

	privateIpAddr, privateEnclaveContainerPort, publicIpAddr, publicEnclaveContainerPort, err := service.moduleStore.LoadModule(ctx, moduleId, image, serializedParams)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred loading module '%v' with container image '%v' and serialized params '%v'", moduleId, image, serializedParams)
	}
	// NON FUNCTIONING: Need to replace when doing refactoring work with modules
	privateApiPort, err := transformEnclaveContainerPortToApiPort(privateEnclaveContainerPort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the module's private enclave container port to an API port")
	}
	// NON FUNCTIONING: Need to replace when doing refactoring work with modules
	publicApiPort, err := transformEnclaveContainerPortToApiPort(publicEnclaveContainerPort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the module's public enclave container port to an API port")
	}

	result := binding_constructors.NewLoadModuleResponse(
		privateIpAddr.String(),
		privateApiPort,
		publicIpAddr.String(),
		publicApiPort,
	)
	return result, nil
}

func (service ApiContainerService) UnloadModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.UnloadModuleArgs) (*emptypb.Empty, error) {
	moduleId := module_store_types.ModuleID(args.ModuleId)

	if err := service.metricsClient.TrackUnloadModule(args.ModuleId); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Errorf("An error occurred tracking unload module event\n%v", err)
	}

	if err := service.moduleStore.UnloadModule(ctx, moduleId); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unloading module '%v' from the network", moduleId)
	}

	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) ExecuteModule(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecuteModuleArgs) (*kurtosis_core_rpc_api_bindings.ExecuteModuleResponse, error) {
	moduleId := module_store_types.ModuleID(args.ModuleId)
	serializedParams := args.SerializedParams

	if err := service.metricsClient.TrackExecuteModule(args.ModuleId, serializedParams); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Errorf("An error occurred tracking execute module event\n%v", err)
	}

	serializedResult, err := service.moduleStore.ExecuteModule(ctx, moduleId, serializedParams)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing module '%v' with serialized params '%v'", moduleId, serializedParams)
	}

	resp := &kurtosis_core_rpc_api_bindings.ExecuteModuleResponse{SerializedResult: serializedResult}
	return resp, nil
}

func (service ApiContainerService) GetModuleInfo(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetModuleInfoArgs) (*kurtosis_core_rpc_api_bindings.GetModuleInfoResponse, error) {
	moduleIdStr := args.ModuleId
	privateIpAddr, privateEnclaveContainerPort, publicIpAddr, publicEnclaveContainerPort, err := service.moduleStore.GetModuleInfo(module_store_types.ModuleID(moduleIdStr))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the IP address for module '%v'", moduleIdStr)
	}
	// NON FUNCTIONING: Need to replace when refactoring modules to use backend
	privateApiPort, err := transformEnclaveContainerPortToApiPort(privateEnclaveContainerPort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the module's private enclave container port to an API port")
	}
	// NON FUNCTIONING: Need to replace when refactoring modules to use backend
	publicApiPort, err := transformEnclaveContainerPortToApiPort(publicEnclaveContainerPort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the module's public enclave container port to an API port")
	}
	response := binding_constructors.NewGetModuleInfoResponse(
		privateIpAddr.String(),
		privateApiPort,
		publicIpAddr.String(),
		publicApiPort,
	)
	return response, nil
}

func (service ApiContainerService) RegisterFilesArtifacts(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RegisterFilesArtifactsArgs) (*emptypb.Empty, error) {
	filesArtifactCache, err := service.enclaveDataDir.GetFilesArtifactCache()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the files artifact cache")
	}

	// TODO PERF: Do these in parallel
	logrus.Debug("Downloading files artifacts to the files artifact cache...")
	for artifactId, url := range args.FilesArtifactUrls {
		if err := filesArtifactCache.DownloadFilesArtifact(artifactId, url); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred downloading files artifact '%v' from URL '%v'", artifactId, url)
		}
	}
	logrus.Debug("Files artifacts downloaded successfully")

	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) RegisterService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RegisterServiceArgs) (*kurtosis_core_rpc_api_bindings.RegisterServiceResponse, error) {
	serviceId := kurtosis_backend_service.ServiceID(args.ServiceId)
	partitionId := service_network_types.PartitionID(args.PartitionId)

	privateIpAddr, relativeServiceDirpath, err := service.serviceNetwork.RegisterService(serviceId, partitionId)
	if err != nil {
		// TODO IP: Leaks internal information about API container
		return nil, stacktrace.Propagate(err, "An error occurred registering service '%v' in the service network", serviceId)
	}

	return &kurtosis_core_rpc_api_bindings.RegisterServiceResponse{
		PrivateIpAddr:          privateIpAddr.String(),
		RelativeServiceDirpath: relativeServiceDirpath,
	}, nil
}

func (service ApiContainerService) StartService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StartServiceArgs) (*kurtosis_core_rpc_api_bindings.StartServiceResponse, error) {
	logrus.Debugf("Received request to start service with the following args: %+v", args)
	serviceId := kurtosis_backend_service.ServiceID(args.ServiceId)
	privateApiPorts := args.PrivatePorts
	privateServicePorts := map[string]*port_spec.PortSpec{}
	for portId, privateApiPort := range privateApiPorts {
		privateEnclaveContainerPort, err := transformApiPortToPortSpec(privateApiPort)
		if err != nil {
			return nil, stacktrace.NewError("An error occurred transforming the API port for private port '%v' into an enclave container port", portId)
		}
		privateServicePorts[portId] = privateEnclaveContainerPort
	}
	maybePublicIpAddr, publicServicePorts, err := service.serviceNetwork.StartService(
		ctx,
		serviceId,
		args.DockerImage,
		// TODO: Update this method
		privateServicePorts,
		args.EntrypointArgs,
		args.CmdArgs,
		args.DockerEnvVars,
		args.EnclaveDataDirMntDirpath,
		args.FilesArtifactMountDirpaths)
	if err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred starting the service in the service network")
	}
	publicApiPorts, err := transformPortSpecMapToApiPortsMap(publicServicePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the service's public port specs to API ports")
	}
	publicIpAddrStr := missingPublicIpAddrStr
	if maybePublicIpAddr != nil {
		publicIpAddrStr = maybePublicIpAddr.String()
	}
	response := binding_constructors.NewStartServiceResponse(publicIpAddrStr, publicApiPorts)

	serviceStartLoglineSuffix := ""
	if len(publicServicePorts) > 0 {
		serviceStartLoglineSuffix = fmt.Sprintf(
			" with the following public ports: %+v",
			publicServicePorts,
		)
	}
	logrus.Infof("Started service '%v'%v", serviceId, serviceStartLoglineSuffix)

	return response, nil
}

func (service ApiContainerService) GetServiceInfo(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetServiceInfoArgs) (*kurtosis_core_rpc_api_bindings.GetServiceInfoResponse, error) {
	serviceIdStr := args.GetServiceId()
	serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
	privateIpAddr, relativeServiceDirpath, err := service.serviceNetwork.GetServiceRegistrationInfo(serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the registration info for service '%v'", serviceIdStr)
	}

	privateServicePorts, maybePublicIpAddr, publicServicePorts, enclaveDataDirMntDirpath, err := service.serviceNetwork.GetServiceRunInfo(serviceId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the run info for service '%v'", serviceIdStr)
	}
	publicIpAddrStr := missingPublicIpAddrStr
	if maybePublicIpAddr != nil {
		publicIpAddrStr = maybePublicIpAddr.String()
	}
	privateApiPorts, err := transformPortSpecMapToApiPortsMap(privateServicePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the service's private enclave container ports to API ports")
	}
	publicApiPorts, err := transformPortSpecMapToApiPortsMap(publicServicePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the service's public enclave container ports to API ports")
	}

	serviceInfoResponse := binding_constructors.NewGetServiceInfoResponse(
		privateIpAddr.String(),
		privateApiPorts,
		publicIpAddrStr,
		publicApiPorts,
		enclaveDataDirMntDirpath,
		relativeServiceDirpath,
	)
	return serviceInfoResponse, nil
}

func (service ApiContainerService) RemoveService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RemoveServiceArgs) (*emptypb.Empty, error) {
	serviceId := kurtosis_backend_service.ServiceID(args.ServiceId)

	containerStopTimeoutSeconds := args.ContainerStopTimeoutSeconds
	containerStopTimeout := time.Duration(containerStopTimeoutSeconds) * time.Second

	if err := service.serviceNetwork.RemoveService(ctx, serviceId, containerStopTimeout); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred removing service with ID '%v'", serviceId)
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) Repartition(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RepartitionArgs) (*emptypb.Empty, error) {
	// No need to check for dupes here - that happens at the lowest-level call to ServiceNetwork.Repartition (as it should)
	partitionServices := map[service_network_types.PartitionID]map[kurtosis_backend_service.ServiceID]bool{}
	for partitionIdStr, servicesInPartition := range args.PartitionServices {
		partitionId := service_network_types.PartitionID(partitionIdStr)
		serviceIdSet := map[kurtosis_backend_service.ServiceID]bool{}
		for serviceIdStr := range servicesInPartition.ServiceIdSet {
			serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
			serviceIdSet[serviceId] = true
		}
		partitionServices[partitionId] = serviceIdSet
	}

	partitionConnections := map[service_network_types.PartitionConnectionID]partition_topology.PartitionConnection{}
	for partitionAStr, partitionBToConnection := range args.PartitionConnections {
		partitionAId := service_network_types.PartitionID(partitionAStr)
		for partitionBStr, connectionInfo := range partitionBToConnection.ConnectionInfo {
			partitionBId := service_network_types.PartitionID(partitionBStr)
			partitionConnectionId := *service_network_types.NewPartitionConnectionID(partitionAId, partitionBId)
			if _, found := partitionConnections[partitionConnectionId]; found {
				return nil, stacktrace.NewError(
					"Partition connection '%v' <-> '%v' was defined twice (possibly in reverse order)",
					partitionAId,
					partitionBId)
			}
			partitionConnection := partition_topology.PartitionConnection{
				PacketLossPercentage: connectionInfo.PacketLossPercentage,
			}
			partitionConnections[partitionConnectionId] = partitionConnection
		}
	}

	defaultConnectionInfo := args.DefaultConnection
	defaultConnection := partition_topology.PartitionConnection{
		PacketLossPercentage: defaultConnectionInfo.PacketLossPercentage,
	}

	if err := service.serviceNetwork.Repartition(
		ctx,
		partitionServices,
		partitionConnections,
		defaultConnection); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred repartitioning the test network")
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) ExecCommand(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecCommandArgs) (*kurtosis_core_rpc_api_bindings.ExecCommandResponse, error) {
	serviceIdStr := args.ServiceId
	serviceId := kurtosis_backend_service.ServiceID(serviceIdStr)
	command := args.CommandArgs
	exitCode, logOutput, err := service.serviceNetwork.ExecCommand(ctx, serviceId, command)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running exec command '%v' against service '%v' in the service network",
			command,
			serviceId)
	}
	numLogOutputBytes := len(logOutput)
	if numLogOutputBytes > maxLogOutputSizeBytes {
		return nil, stacktrace.NewError(
			"Log output from docker exec command '%+v' was %v bytes, but maximum size allowed by Kurtosis is %v",
			command,
			numLogOutputBytes,
			maxLogOutputSizeBytes,
		)
	}
	resp := &kurtosis_core_rpc_api_bindings.ExecCommandResponse{
		ExitCode:  exitCode,
		LogOutput: logOutput,
	}
	return resp, nil
}

func (service ApiContainerService) WaitForHttpGetEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpGetEndpointAvailabilityArgs) (*emptypb.Empty, error) {

	serviceIdStr := args.ServiceId

	if err := service.waitForEndpointAvailability(
		serviceIdStr,
		http.MethodGet,
		args.Port,
		args.Path,
		args.InitialDelayMilliseconds,
		args.Retries,
		args.RetriesDelayMilliseconds,
		"",
		args.BodyText); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for HTTP endpoint '%v' to become available",
			args.Path,
		)
	}

	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) WaitForHttpPostEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForHttpPostEndpointAvailabilityArgs) (*emptypb.Empty, error) {

	serviceIdStr := args.ServiceId

	if err := service.waitForEndpointAvailability(
		serviceIdStr,
		http.MethodPost,
		args.Port,
		args.Path,
		args.InitialDelayMilliseconds,
		args.Retries,
		args.RetriesDelayMilliseconds,
		args.RequestBody,
		args.BodyText); err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred waiting for HTTP endpoint '%v' to become available",
			args.Path,
		)
	}

	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) ExecuteBulkCommands(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecuteBulkCommandsArgs) (*emptypb.Empty, error) {
	if err := service.bulkCmdExecEngine.Process(ctx, []byte(args.SerializedCommands)); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing the bulk commands")
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) GetServices(ctx context.Context, empty *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.GetServicesResponse, error) {

	serviceIDs := make(map[string]bool, len(service.serviceNetwork.GetServiceIDs()))

	for serviceID := range service.serviceNetwork.GetServiceIDs() {
		serviceIDStr := string(serviceID)
		if _, ok := serviceIDs[serviceIDStr]; !ok {
			serviceIDs[serviceIDStr] = true
		}
	}

	resp := &kurtosis_core_rpc_api_bindings.GetServicesResponse{
		ServiceIds: serviceIDs,
	}
	return resp, nil
}

func (service ApiContainerService) GetModules(ctx context.Context, empty *emptypb.Empty) (*kurtosis_core_rpc_api_bindings.GetModulesResponse, error) {

	allModuleIDs := make(map[string]bool, len(service.moduleStore.GetModules()))

	for moduleID, _ := range service.moduleStore.GetModules() {
		moduleIDStr := string(moduleID)
		if _, ok := allModuleIDs[moduleIDStr]; !ok {
			allModuleIDs[moduleIDStr] = true
		}
	}

	resp := &kurtosis_core_rpc_api_bindings.GetModulesResponse{
		ModuleIds: allModuleIDs,
	}
	return resp, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func transformApiPortToPortSpec(port *kurtosis_core_rpc_api_bindings.Port) (*port_spec.PortSpec, error) {
	portNumUint32 := port.GetNumber()
	apiProto := port.GetProtocol()
	if portNumUint32 > math.MaxUint16 {
		return nil, stacktrace.NewError(
			"API port num '%v' is bigger than max allowed enclave container port num '%v'",
			portNumUint32,
			math.MaxUint16,
		)
	}
	portNumUint16 := uint16(portNumUint32)
	portSpecProto, found := apiContainerPortProtoToPortSpecPortProto[apiProto]
	if !found {
		return nil, stacktrace.NewError("Couldn't find an enclave container port proto for API port proto '%v'; this should never happen, and is a bug in Kurtosis!", apiProto.String())
	}
	result, err := port_spec.NewPortSpec(portNumUint16, portSpecProto)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating enclave container port object with port num '%v' and protocol '%v'",
			portNumUint16,
			portSpecProto,
		)
	}
	return result, nil
}

func transformPortSpecToApiPort(port *port_spec.PortSpec) (*kurtosis_core_rpc_api_bindings.Port, error) {
	portNumUint16 := port.GetNumber()
	portSpecProto := port.GetProtocol()
	// Yes, this isn't the most efficient way to do this, but the map is tiny so it doesn't matter
	var apiProto kurtosis_core_rpc_api_bindings.Port_Protocol
	foundApiProto := false
	for mappedApiProto, mappedPortSpecProto := range apiContainerPortProtoToPortSpecPortProto {
		if portSpecProto == mappedPortSpecProto {
			apiProto = mappedApiProto
			foundApiProto = true
			break
		}
	}
	if !foundApiProto {
		return nil, stacktrace.NewError("Couldn't find an API port proto for port spec port proto '%v'; this should never happen, and is a bug in Kurtosis!", portSpecProto)
	}
	result := binding_constructors.NewPort(uint32(portNumUint16), apiProto)
	return result, nil
}

func transformPortSpecMapToApiPortsMap(apiPorts map[string]*port_spec.PortSpec) (map[string]*kurtosis_core_rpc_api_bindings.Port, error) {
	result := map[string]*kurtosis_core_rpc_api_bindings.Port{}
	for portId, portSpec := range apiPorts {
		publicApiPort, err := transformPortSpecToApiPort(portSpec)
		if err != nil {
			return nil, stacktrace.NewError("An error occurred transforming port spec for port '%v' into an API port", portId)
		}
		result[portId] = publicApiPort
	}
	return result, nil
}

func (service ApiContainerService) waitForEndpointAvailability(
	serviceIdStr string,
	httpMethod string,
	port uint32,
	path string,
	initialDelayMilliseconds uint32,
	retries uint32,
	retriesDelayMilliseconds uint32,
	requestBody string,
	bodyText string) error {

	var (
		resp *http.Response
		err  error
	)

	privateServiceIp, _, err := service.serviceNetwork.GetServiceRegistrationInfo(kurtosis_backend_service.ServiceID(serviceIdStr))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the registration info for service '%v'", serviceIdStr)
	}

	url := fmt.Sprintf("http://%v:%v/%v", privateServiceIp, port, path)

	time.Sleep(time.Duration(initialDelayMilliseconds) * time.Millisecond)

	for i := uint32(0); i < retries; i++ {
		resp, err = makeHttpRequest(httpMethod, url, requestBody)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(retriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return stacktrace.Propagate(
			err,
			"The HTTP endpoint '%v' didn't return a success code, even after %v retries with %v milliseconds in between retries",
			url,
			retries,
			retriesDelayMilliseconds,
		)
	}

	if bodyText != "" {
		body := resp.Body
		defer body.Close()

		bodyBytes, err := ioutil.ReadAll(body)

		if err != nil {
			return stacktrace.Propagate(err,
				"An error occurred reading the response body from endpoint '%v'", url)
		}

		bodyStr := string(bodyBytes)

		if bodyStr != bodyText {
			return stacktrace.NewError("Expected response body text '%v' from endpoint '%v' but got '%v' instead", bodyText, url, bodyStr)
		}
	}

	return nil
}

func makeHttpRequest(httpMethod string, url string, body string) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	if httpMethod == http.MethodPost {
		var bodyByte = []byte(body)
		resp, err = http.Post(url, "application/json", bytes.NewBuffer(bodyByte))
	} else if httpMethod == http.MethodGet {
		resp, err = http.Get(url)
	} else {
		return nil, stacktrace.NewError("HTTP method '%v' not allowed", httpMethod)
	}

	if err != nil {
		return nil, stacktrace.Propagate(err, "An HTTP error occurred when sending GET request to endpoint '%v'", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, stacktrace.NewError("Received non-OK status code: '%v'", resp.StatusCode)
	}
	return resp, nil
}
