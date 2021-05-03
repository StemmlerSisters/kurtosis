/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package test_execution

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/api/bindings"
	"github.com/kurtosis-tech/kurtosis/api_container/api_container_docker_consts/api_container_exit_codes"
	service_network2 "github.com/kurtosis-tech/kurtosis/api_container/service_network"
	partition_topology2 "github.com/kurtosis-tech/kurtosis/api_container/service_network/partition_topology"
	service_network_types2 "github.com/kurtosis-tech/kurtosis/api_container/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

const (
	// Custom-set max size for logs coming back from docker exec.
	// Protobuf sets a maximum of 2GB for responses, in interest of keeping performance sane
	// we pick a reasonable limit of 10MB on log responses for docker exec.
	// See: https://stackoverflow.com/questions/34128872/google-protobuf-maximum-size/34186672
	maxLogOutputSizeBytes = 10 * 1024 * 1024

	// The amount of time a testsuite container has after registering itself with the API container to register
	//  a test setup (there should be no reason that registering test setup doesn't happen immediately)
	testSetupRegistrationTimeout = 10 * time.Second

	// TODO Implement a formalized debug mode that doesn't use this timeout
	// The amount of time a testsuite container has after completing test setup to register
	//  a test execution (the only reason registration wouldn't happen immediately is if the testsuite
	//  is running inside a debugger that's waiting for a connection)
	testExecutionRegistrationTimeout = 20 * time.Second
)


type testExecutionService struct {
	dockerManager             *docker_manager.DockerManager
	serviceNetwork            *service_network2.ServiceNetwork
	testName                  string
}

func newTestExecutionService(
		dockerManager *docker_manager.DockerManager,
		serviceNetwork *service_network2.ServiceNetwork,
		testName string,
		testSetupTimeoutInSeconds uint32,
		testRunTimeoutInSeconds uint32,
		testSuiteContainerId string,
		shutdownChan chan int) *testExecutionService {
	return &testExecutionService{
		dockerManager:             dockerManager,
		serviceNetwork:            serviceNetwork,
		testName:                  testName,
		testSetupTimeoutInSeconds: testSetupTimeoutInSeconds,
		testRunTimeoutInSeconds:   testRunTimeoutInSeconds,
		testSuiteContainerId:      testSuiteContainerId,
		stateMachine:              newTestExecutionServiceStateMachine(),
		shutdownChan:              shutdownChan,
	}
}

func (service *testExecutionService) HandleSuiteRegistrationEvent() error {
	if err := service.stateMachine.assertAndAdvance(waitingForSuiteRegistration); err != nil {
		return stacktrace.Propagate(
			err,
			"Cannot register test suite; an error occurred advancing the state machine")
	}

	// Launch timeout thread that will error if a test setup isn't registered soon
	go func() {
		time.Sleep(testSetupRegistrationTimeout)
		if err := service.stateMachine.assertOneOfSet(map[serviceState]bool{waitingForTestSetupRegistration: true}); err == nil {
			service.shutdownChan <- api_container_exit_codes.NoTestSetupRegistered
		}
	}()

	return nil
}

func (service *testExecutionService) RegisterTestSetup(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := service.stateMachine.assertAndAdvance(waitingForTestSetupRegistration); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "Cannot register test setup; an error occurred advancing the state machine")
	}

	timeoutSeconds := service.testSetupTimeoutInSeconds
	timeout := time.Duration(timeoutSeconds) * time.Second

	// If the testsuite throws an error during setup and exits, the user would have to wait for the setup
	// timeout (which can be very long). To speed things up, we'll monitor the testsuite container to
	// ensure that an error is thrown if the testsuite exits during the setup phase
	go func() {
		// We use the background context so that waiting continues even when the request finishes
		_, waitForSuiteExitErr := service.dockerManager.WaitForExit(context.Background(), service.testSuiteContainerId)
		if assertIsSetupPhaseErr := service.stateMachine.assertOneOfSet(map[serviceState]bool{waitingForTestSetupCompletion: true}); assertIsSetupPhaseErr == nil {
			if waitForSuiteExitErr != nil {
				logrus.Warnf("The testsuite container was determined to have exited while execution was still in the " +
					"test setup phase which should never happen, but the following error occurred while waiting for " +
					"the testsuite container to exit so the determination that the testsuite exited may be spurious:")
				fmt.Fprintln(logrus.StandardLogger().Out, waitForSuiteExitErr)
			}
			logrus.Errorf("The testsuite container exited during the test setup phase, which should never happen")
			service.shutdownChan <- api_container_exit_codes.TestsuiteExitedDuringSetup
		}
	}()

	// Launch timeout thread that will error if the test setup doesn't complete within the allotted time limit
	go func() {
		time.Sleep(timeout)
		if err := service.stateMachine.assertOneOfSet(map[serviceState]bool{waitingForTestSetupCompletion: true}); err == nil {
			service.shutdownChan <- api_container_exit_codes.TestHitSetupTimeout
		}
	}()

	return &emptypb.Empty{}, nil
}

func (service *testExecutionService) RegisterTestSetupCompletion(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := service.stateMachine.assertAndAdvance(waitingForTestSetupCompletion); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "Cannot register test setup completion; an error occurred advancing the state machine")
	}

	// Launch timeout thread that will error if a test execution isn't registered soon
	go func() {
		time.Sleep(testExecutionRegistrationTimeout)
		if err := service.stateMachine.assertOneOfSet(map[serviceState]bool{waitingForTestExecutionRegistration: true}); err == nil {
			service.shutdownChan <- api_container_exit_codes.NoTestExecutionRegistered
		}
	}()

	return &emptypb.Empty{}, nil
}

func (service *testExecutionService) RegisterTestExecution(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := service.stateMachine.assertAndAdvance(waitingForTestExecutionRegistration); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "Cannot register test execution; an error occurred advancing the state machine")
	}

	timeoutSeconds := service.testRunTimeoutInSeconds
	timeout := time.Duration(timeoutSeconds) * time.Second

	// Launch timeout thread that will error if the test execution doesn't complete within the allotted time limit
	go func() {
		time.Sleep(timeout)
		if err := service.stateMachine.assertOneOfSet(map[serviceState]bool{waitingForExecutionCompletion: true}); err == nil {
			service.shutdownChan <- api_container_exit_codes.TestHitExecutionTimeout
		}
	}()

	// Launch thread to monitor the state of the testsuite container
	go func() {
		// We use the background context so that waiting continues even when the request finishes
		if _, err := service.dockerManager.WaitForExit(context.Background(), service.testSuiteContainerId); err != nil {
			logrus.Errorf("An error occurred waiting for the testsuite container to exit:")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
			service.shutdownChan <- api_container_exit_codes.ErrWaitingForSuiteContainerExit
			return
		}
		if err := service.stateMachine.assertAndAdvance(waitingForExecutionCompletion); err != nil {
			logrus.Warnf("The testsuite container exited, but an error occurred advancing the state machine to its final state")
			fmt.Fprintln(logrus.StandardLogger().Out, err)
		}
		service.shutdownChan <- api_container_exit_codes.SuccessfulExit
	}()

	return &emptypb.Empty{}, nil
}

func (service *testExecutionService) RegisterService(_ context.Context, args *bindings.RegisterServiceArgs) (*bindings.RegisterServiceResponse, error) {
	expectedStateSet := map[serviceState]bool{waitingForTestSetupCompletion: true, waitingForExecutionCompletion: true};
	if err := service.stateMachine.assertOneOfSet(expectedStateSet); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "Cannot register service; test execution service wasn't in one of the expected states '%+v'", expectedStateSet)
	}

	serviceId := service_network_types2.ServiceID(args.ServiceId)
	partitionId := service_network_types2.PartitionID(args.PartitionId)

	ip, err := service.serviceNetwork.RegisterService(serviceId, partitionId)
	if err != nil {
		// TODO IP: Leaks internal information about API container
		return nil, stacktrace.Propagate(err, "An error occurred registering service '%v' in the service network", serviceId)
	}

	return &bindings.RegisterServiceResponse{
		IpAddr:                          ip.String(),
	}, nil
}

func (service *testExecutionService) GenerateFiles(_ context.Context, args *bindings.GenerateFilesArgs) (*bindings.GenerateFilesResponse, error) {
	serviceId := service_network_types2.ServiceID(args.ServiceId)
	filesToGenerate := args.FilesToGenerate
	generatedFileRelativeFilepaths, err := service.serviceNetwork.GenerateFiles(serviceId, filesToGenerate)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating files for service '%v'", serviceId)
	}
	return &bindings.GenerateFilesResponse{
		GeneratedFileRelativeFilepaths: generatedFileRelativeFilepaths,
	}, nil
}

func (service *testExecutionService) StartService(ctx context.Context, args *bindings.StartServiceArgs) (*bindings.StartServiceResponse, error) {
	logrus.Debugf("Received request to start service with the following args: %+v", args)

	expectedStateSet := map[serviceState]bool{waitingForTestSetupCompletion: true, waitingForExecutionCompletion: true};
	if err := service.stateMachine.assertOneOfSet(expectedStateSet); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "Cannot start service; test execution service wasn't in one of the expected states '%+v'", expectedStateSet)
	}

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

	serviceId := service_network_types2.ServiceID(args.ServiceId)

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

	responseHostPortBindings := map[string]*bindings.PortBinding{}
	for portObj, hostPortBinding := range hostPortBindings {
		portSpecStr, found := portObjToPortSpecStr[portObj]
		if !found {
			return nil, stacktrace.NewError(
				"Found a port object, %+v, that doesn't correspond to a spec string as passed in via the args; this is very strange!",
				portObj,
			)
		}
		responseBinding := &bindings.PortBinding{
			InterfaceIp:   hostPortBinding.HostIP,
			InterfacePort: hostPortBinding.HostPort,
		}
		responseHostPortBindings[portSpecStr] = responseBinding
	}
	response := bindings.StartServiceResponse{
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

func (service *testExecutionService) RemoveService(ctx context.Context, args *bindings.RemoveServiceArgs) (*emptypb.Empty, error) {
	if err := service.stateMachine.assertOneOfSet(map[serviceState]bool{waitingForExecutionCompletion: true}); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "Cannot remove service; test execution service wasn't in expected state '%v'", waitingForExecutionCompletion)
	}

	serviceId := service_network_types2.ServiceID(args.ServiceId)

	containerStopTimeoutSeconds := args.ContainerStopTimeoutSeconds
	containerStopTimeout := time.Duration(containerStopTimeoutSeconds) * time.Second

	if err := service.serviceNetwork.RemoveService(ctx, serviceId, containerStopTimeout); err != nil {
		// TODO IP: Leaks internal information about the API container
		return nil, stacktrace.Propagate(err, "An error occurred removing service with ID '%v'", serviceId)
	}
	return &emptypb.Empty{}, nil
}

func (service *testExecutionService) Repartition(ctx context.Context, args *bindings.RepartitionArgs) (*emptypb.Empty, error) {
	// No need to check for dupes here - that happens at the lowest-level call to ServiceNetwork.Repartition (as it should)
	partitionServices := map[service_network_types2.PartitionID]*service_network_types2.ServiceIDSet{}
	for partitionIdStr, servicesInPartition := range args.PartitionServices {
		partitionId := service_network_types2.PartitionID(partitionIdStr)
		serviceIdSet := service_network_types2.NewServiceIDSet()
		for serviceIdStr := range servicesInPartition.ServiceIdSet {
			serviceId := service_network_types2.ServiceID(serviceIdStr)
			serviceIdSet.AddElem(serviceId)
		}
		partitionServices[partitionId] = serviceIdSet
	}

	partitionConnections := map[service_network_types2.PartitionConnectionID]partition_topology2.PartitionConnection{}
	for partitionAStr, partitionBToConnection := range args.PartitionConnections {
		partitionAId := service_network_types2.PartitionID(partitionAStr)
		for partitionBStr, connectionInfo := range partitionBToConnection.ConnectionInfo {
			partitionBId := service_network_types2.PartitionID(partitionBStr)
			partitionConnectionId := *service_network_types2.NewPartitionConnectionID(partitionAId, partitionBId)
			if _, found := partitionConnections[partitionConnectionId]; found {
				return nil, stacktrace.NewError(
					"Partition connection '%v' <-> '%v' was defined twice (possibly in reverse order)",
					partitionAId,
					partitionBId)
			}
			partitionConnection := partition_topology2.PartitionConnection{
				IsBlocked: connectionInfo.IsBlocked,
			}
			partitionConnections[partitionConnectionId] = partitionConnection
		}
	}

	defaultConnectionInfo := args.DefaultConnection
	defaultConnection := partition_topology2.PartitionConnection{
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

func (service *testExecutionService) ExecCommand(ctx context.Context, args *bindings.ExecCommandArgs) (*bindings.ExecCommandResponse, error) {
	serviceIdStr := args.ServiceId
	serviceId := service_network_types2.ServiceID(serviceIdStr)
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
	resp := &bindings.ExecCommandResponse{
		ExitCode: exitCode,
		LogOutput: logOutput.Bytes(),
	}
	return resp, nil
}



