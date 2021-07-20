/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/server/bulk_command_execution_engine"
	"github.com/kurtosis-tech/kurtosis/api_container/server/bulk_command_execution_engine/v0_bulk_command_execution"
	"github.com/kurtosis-tech/kurtosis/api_container/server/lambda_store"
	"github.com/kurtosis-tech/kurtosis/api_container/server/lambda_store/lambda_store_types"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/partition_topology"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	// Custom-set max size for logs coming back from docker exec.
	// Protobuf sets a maximum of 2GB for responses, in interest of keeping performance sane
	// we pick a reasonable limit of 10MB on log responses for docker exec.
	// See: https://stackoverflow.com/questions/34128872/google-protobuf-maximum-size/34186672
	maxLogOutputSizeBytes = 10 * 1024 * 1024
)

type ApiContainerService struct {
	// This embedding is required by gRPC
	kurtosis_core_rpc_api_bindings.UnimplementedApiContainerServiceServer

	serviceNetwork service_network.ServiceNetwork

	lambdaStore *lambda_store.LambdaStore

	bulkCmdExecEngine *bulk_command_execution_engine.BulkCommandExecutionEngine
}

func NewApiContainerService(serviceNetwork service_network.ServiceNetwork, lambdaStore *lambda_store.LambdaStore) (*ApiContainerService, error) {
	service := &ApiContainerService{
		serviceNetwork: serviceNetwork,
		lambdaStore:    lambdaStore,
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

func (service ApiContainerService) LoadLambda(ctx context.Context, args *kurtosis_core_rpc_api_bindings.LoadLambdaArgs) (*emptypb.Empty, error) {
	lambdaId := lambda_store_types.LambdaID(args.LambdaId)
	image := args.ContainerImage
	serializedParams := args.SerializedParams
	if err := service.lambdaStore.LoadLambda(ctx, lambdaId, image, serializedParams); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred loading Lambda '%v' with container image '%v' and serialized params '%v'", lambdaId, image, serializedParams)
	}
	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) ExecuteLambda(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecuteLambdaArgs) (*kurtosis_core_rpc_api_bindings.ExecuteLambdaResponse, error) {
	lambdaId := lambda_store_types.LambdaID(args.LambdaId)
	serializedParams := args.SerializedParams
	serializedResult, err := service.lambdaStore.ExecuteLambda(ctx, lambdaId, serializedParams)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing Lambda '%v' with serialized params '%v'", lambdaId, serializedParams)
	}
	resp := &kurtosis_core_rpc_api_bindings.ExecuteLambdaResponse{SerializedResult: serializedResult}
	return resp, nil
}

func (service ApiContainerService) GetLambdaInfo(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetLambdaInfoArgs) (*kurtosis_core_rpc_api_bindings.GetLambdaInfoResponse, error) {
	lambdaIdStr := args.LambdaId
	ipAddr, err := service.lambdaStore.GetLambdaIPAddrByID(lambda_store_types.LambdaID(lambdaIdStr))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the IP address for Lambda '%v'", lambdaIdStr)
	}
	response := &kurtosis_core_rpc_api_bindings.GetLambdaInfoResponse{IpAddr: ipAddr.String()}
	return response, nil
}

func (service ApiContainerService) RegisterService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RegisterServiceArgs) (*kurtosis_core_rpc_api_bindings.RegisterServiceResponse, error) {
	serviceId := service_network_types.ServiceID(args.ServiceId)
	partitionId := service_network_types.PartitionID(args.PartitionId)

	ip, err := service.serviceNetwork.RegisterService(serviceId, partitionId)
	if err != nil {
		// TODO IP: Leaks internal information about API container
		return nil, stacktrace.Propagate(err, "An error occurred registering service '%v' in the service network", serviceId)
	}

	return &kurtosis_core_rpc_api_bindings.RegisterServiceResponse{
		IpAddr:                          ip.String(),
	}, nil
}

func (service ApiContainerService) GenerateFiles(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GenerateFilesArgs) (*kurtosis_core_rpc_api_bindings.GenerateFilesResponse, error) {
	serviceId := service_network_types.ServiceID(args.ServiceId)
	filesToGenerate := args.FilesToGenerate
	generatedFileRelativeFilepaths, err := service.serviceNetwork.GenerateFiles(serviceId, filesToGenerate)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating files for service '%v'", serviceId)
	}
	return &kurtosis_core_rpc_api_bindings.GenerateFilesResponse{
		GeneratedFileRelativeFilepaths: generatedFileRelativeFilepaths,
	}, nil
}

func (service ApiContainerService) LoadStaticFiles(ctx context.Context, args *kurtosis_core_rpc_api_bindings.LoadStaticFilesArgs) (*kurtosis_core_rpc_api_bindings.LoadStaticFilesResponse, error) {
	serviceId := service_network_types.ServiceID(args.ServiceId)
	copiedStaticFileRelativeFilepaths, err := service.serviceNetwork.LoadStaticFiles(serviceId, args.StaticFiles)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred loading static files for service '%v'", serviceId)
	}
	return &kurtosis_core_rpc_api_bindings.LoadStaticFilesResponse{
		CopiedStaticFileRelativeFilepaths: copiedStaticFileRelativeFilepaths,
	}, nil
}

func (service ApiContainerService) StartService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.StartServiceArgs) (*kurtosis_core_rpc_api_bindings.StartServiceResponse, error) {
	logrus.Debugf("Received request to start service with the following args: %+v", args)

	usedPorts := map[nat.Port]bool{}
	portObjToPortSpecStr := map[nat.Port]string{}
	for portSpecStr := range args.UsedPorts {
		// NOTE: this function, frustratingly, doesn't return an error on failure - just emptystring
		protocol, portNumberStr := nat.SplitProtoPort(portSpecStr)
		if protocol == "" {
			return nil, stacktrace.NewError(
				"Could not split port specification string '%s' into protocol & number strings",
				portSpecStr)
		}
		portObj, err := nat.NewPort(protocol, portNumberStr)
		if err != nil {
			// TODO IP: Leaks internal information about the API container
			return nil, stacktrace.Propagate(
				err,
				"An error occurred constructing a port object out of protocol '%v' and port number string '%v'",
				protocol,
				portNumberStr)
		}
		usedPorts[portObj] = true
		portObjToPortSpecStr[portObj] = portSpecStr
	}

	serviceId := service_network_types.ServiceID(args.ServiceId)

	hostPortBindings, err := service.serviceNetwork.StartService(
		ctx,
		serviceId,
		args.DockerImage,
		usedPorts,
		args.EntrypointArgs,
		args.CmdArgs,
		args.DockerEnvVars,
		args.SuiteExecutionVolMntDirpath,
		args.FilesArtifactMountDirpaths)
	if err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred starting the service in the service network")
	}

	// We strip out ports with nil host port bindings to make it easier to iterate over this map on the client side
	responseHostPortBindings := map[string]*kurtosis_core_rpc_api_bindings.PortBinding{}
	for portObj, hostPortBinding := range hostPortBindings {
		portSpecStr, found := portObjToPortSpecStr[portObj]
		if !found {
			return nil, stacktrace.NewError(
				"Found a port object, %+v, that doesn't correspond to a spec string as passed in via the args; this is very strange!",
				portObj,
			)
		}
		if hostPortBinding != nil {
			responseBinding := &kurtosis_core_rpc_api_bindings.PortBinding{
				InterfaceIp:   hostPortBinding.HostIP,
				InterfacePort: hostPortBinding.HostPort,
			}
			responseHostPortBindings[portSpecStr] = responseBinding
		}
	}
	response := kurtosis_core_rpc_api_bindings.StartServiceResponse{
		UsedPortsHostPortBindings: responseHostPortBindings,
	}

	serviceStartLoglineSuffix := ""
	if len(responseHostPortBindings) > 0 {
		serviceStartLoglineSuffix = fmt.Sprintf(
			" with the following service-port-to-host-port bindings: %+v",
			responseHostPortBindings,
		)
	}
	logrus.Infof("Started service '%v'%v", serviceId, serviceStartLoglineSuffix)

	return &response, nil
}

func (service ApiContainerService) GetServiceInfo(ctx context.Context, args *kurtosis_core_rpc_api_bindings.GetServiceInfoArgs) (*kurtosis_core_rpc_api_bindings.GetServiceInfoResponse, error) {
	serviceIP, err := service.getServiceIPByServiceId(args.ServiceId)
	if err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred when trying to get the service IP address by service ID: '%v'",
			args.ServiceId)
	}

	serviceID := service_network_types.ServiceID(args.ServiceId)
	suiteExecutionVolMntDirpath, err :=service.serviceNetwork.GetServiceSuiteExecutionVolMntDirpath(serviceID)
	if err != nil {
		return nil, stacktrace.Propagate(err,"An error occurred when trying to get service suite execution volume directory path by service ID: '%v'",
			serviceID)
	}

	serviceInfoResponse := &kurtosis_core_rpc_api_bindings.GetServiceInfoResponse{
		IpAddr: serviceIP.String(),
		SuiteExecutionVolumeMountDirpath: suiteExecutionVolMntDirpath,
	}
	return serviceInfoResponse, nil
}

func (service ApiContainerService) RemoveService(ctx context.Context, args *kurtosis_core_rpc_api_bindings.RemoveServiceArgs) (*emptypb.Empty, error) {
	serviceId := service_network_types.ServiceID(args.ServiceId)

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
	partitionServices := map[service_network_types.PartitionID]*service_network_types.ServiceIDSet{}
	for partitionIdStr, servicesInPartition := range args.PartitionServices {
		partitionId := service_network_types.PartitionID(partitionIdStr)
		serviceIdSet := service_network_types.NewServiceIDSet()
		for serviceIdStr := range servicesInPartition.ServiceIdSet {
			serviceId := service_network_types.ServiceID(serviceIdStr)
			serviceIdSet.AddElem(serviceId)
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
				IsBlocked: connectionInfo.IsBlocked,
			}
			partitionConnections[partitionConnectionId] = partitionConnection
		}
	}

	defaultConnectionInfo := args.DefaultConnection
	defaultConnection := partition_topology.PartitionConnection{
		IsBlocked: defaultConnectionInfo.IsBlocked,
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
	serviceId := service_network_types.ServiceID(serviceIdStr)
	command := args.CommandArgs
	exitCode, logOutput, err := service.serviceNetwork.ExecCommand(ctx, serviceId, command)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred running exec command '%v' against service '%v' in the service network",
			command,
			serviceId)
	}
	logOutputSize := logOutput.Len()
	if logOutputSize > maxLogOutputSizeBytes {
		return nil, stacktrace.NewError("Log output from docker exec command %+v was %v bytes, but maximum size allowed by Kurtosis is %v",
			command,
			logOutputSize,
			maxLogOutputSizeBytes,
		)
	}
	resp := &kurtosis_core_rpc_api_bindings.ExecCommandResponse{
		ExitCode: exitCode,
		LogOutput: logOutput.Bytes(),
	}
	return resp, nil
}

func (service ApiContainerService) WaitForEndpointAvailability(ctx context.Context, args *kurtosis_core_rpc_api_bindings.WaitForEndpointAvailabilityArgs) (*emptypb.Empty, error) {
	var(
		resp *http.Response
		err error
	)

	serviceIdStr := args.ServiceId
	serviceIP, err := service.getServiceIPByServiceId(serviceIdStr)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred when trying to get the IP address for service '%v'",
			serviceIdStr,
		)
	}

	url := fmt.Sprintf("http://%v:%v/%v", serviceIP, args.Port, args.Path)

	time.Sleep(time.Duration(args.InitialDelaySeconds) * time.Second)

	for i := uint32(0); i < args.Retries; i++ {
		resp, err = makeHttpGetRequest(url)
		if err == nil  {
			break
		}
		time.Sleep(time.Duration(args.RetriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"The HTTP endpoint '%v' didn't return a success code, even after %v retries with %v milliseconds in between retries",
			url,
			args.Retries,
			args.RetriesDelayMilliseconds,
		)
	}

	if args.BodyText != "" {
		body := resp.Body
		defer body.Close()

		bodyBytes, err := ioutil.ReadAll(body)

		if err != nil {
			return nil, stacktrace.Propagate(err,
				"An error occurred reading the response body from endpoint '%v'", url)
		}

		bodyStr := string(bodyBytes)

		if bodyStr != args.BodyText {
			return nil, stacktrace.NewError("Expected response body text '%v' from endpoint '%v' but got '%v' instead", args.BodyText, url, bodyStr)
		}
	}

	return &emptypb.Empty{}, nil
}

func (service ApiContainerService) ExecuteBulkCommands(ctx context.Context, args *kurtosis_core_rpc_api_bindings.ExecuteBulkCommandsArgs) (*emptypb.Empty, error) {
	if err := service.bulkCmdExecEngine.Process(ctx, []byte(args.SerializedCommands)); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred executing the bulk commands")
	}
	return &emptypb.Empty{}, nil
}

// ====================================================================================================
// 									   Private helper methods
// ====================================================================================================
func makeHttpGetRequest(url string) (*http.Response, error){
	resp, err := http.Get(url)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An HTTP error occurred when sending GET request to endpoint '%v'", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, stacktrace.NewError("Received non-OK status code: '%v'", resp.StatusCode)
	}
	return resp, nil
}

func (service ApiContainerService) getServiceIPByServiceId(serviceId string) (net.IP, error){
	serviceID := service_network_types.ServiceID(serviceId)
	serviceIP, err := service.serviceNetwork.GetServiceIP(serviceID)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"An error occurred when trying to get the service IP address by service ID: '%v'",
			serviceId)
	}
	return serviceIP, nil
}

