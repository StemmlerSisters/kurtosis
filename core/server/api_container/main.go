/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/free-ip-addr-tracker-lib/lib"
	"github.com/kurtosis-tech/kurtosis-core/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core/launcher/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis-core/launcher/args"
	"github.com/kurtosis-tech/kurtosis-core/launcher/enclave_container_launcher"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/external_container_store"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/module_store"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/module_store/module_launcher"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/user_service_launcher"
	"github.com/kurtosis-tech/kurtosis-core/server/api_container/server/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis-core/server/commons/enclave_data_directory"
	metrics_client "github.com/kurtosis-tech/metrics-library/golang/lib/client"
	"github.com/kurtosis-tech/metrics-library/golang/lib/source"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"
)

const (
	successExitCode = 0
	failureExitCode = 1

	grpcServerStopGracePeriod = 5 * time.Second

	shouldFlushMetricsClientQueueOnEachEvent = false
)

func main() {
	// NOTE: we'll want to change the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	err := runMain()
	if err != nil {
		logrus.Errorf("An error occurred when running the main function:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)

}

func runMain () error {
	serverArgs, err := args.GetArgsFromEnv()
	if err != nil {
		return stacktrace.Propagate(err, "Couldn't retrieve API container args from the environment")
	}

	logLevel, err := logrus.ParseLevel(serverArgs.LogLevel)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing the log level string '%v':", serverArgs.LogLevel)
	}
	logrus.SetLevel(logLevel)

	_, parsedSubnetMask, err := net.ParseCIDR(serverArgs.SubnetMask)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred parsing subnet CIDR string '%v'", serverArgs.SubnetMask)
	}
	freeIpAddrTracker := lib.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		parsedSubnetMask,
		serverArgs.TakenIpAddrs,
	)

	externalContainerStore := external_container_store.NewExternalContainerStore(freeIpAddrTracker)

	dockerManager, err := createDockerManager()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	enclaveDataDir := enclave_data_directory.NewEnclaveDataDirectory(serverArgs.EnclaveDataDirpathOnAPIContainer)

	serviceNetwork, moduleStore, err := createServiceNetworkAndModuleStore(dockerManager, enclaveDataDir, freeIpAddrTracker, serverArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the service network & module store")
	}

	metricsClient, err := metrics_client.CreateMetricsClient(source.KurtosisCoreSource, api_container_launcher.DefaultVersion, serverArgs.MetricsUserID, serverArgs.DidUserAcceptSendingMetrics, shouldFlushMetricsClientQueueOnEachEvent)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the metrics client")
	}

	//Creation of ApiContainerService
	apiContainerService, err := server.NewApiContainerService(
		enclaveDataDir,
		externalContainerStore,
		serviceNetwork,
		moduleStore,
		metricsClient,
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the API container service")
	}

	apiContainerServiceRegistrationFunc := func(grpcServer *grpc.Server) {
		kurtosis_core_rpc_api_bindings.RegisterApiContainerServiceServer(grpcServer, apiContainerService)
	}
	apiContainerServer := minimal_grpc_server.NewMinimalGRPCServer(
		serverArgs.GrpcListenPortNum,
		grpcServerStopGracePeriod,
		[]func(*grpc.Server){
			apiContainerServiceRegistrationFunc,
		},
	)

	logrus.Info("Running server...")
	if err := apiContainerServer.RunUntilInterrupted(); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the API container server")
	}

	return nil
}

func createDockerManager() (*docker_manager.DockerManager, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not initialize a Docker client from the environment")
	}

	dockerManager := docker_manager.NewDockerManager(logrus.StandardLogger(), dockerClient)
	return dockerManager, nil
}

func createServiceNetworkAndModuleStore(
		dockerManager *docker_manager.DockerManager,
		enclaveDataDir *enclave_data_directory.EnclaveDataDirectory,
		freeIpAddrTracker *lib.FreeIpAddrTracker,
		args *args.APIContainerArgs) (service_network.ServiceNetwork, *module_store.ModuleStore, error) {
	enclaveId := args.EnclaveId

	objAttrsProvider := schema.GetObjectAttributesProvider()
	enclaveObjAttrsProvider := objAttrsProvider.ForEnclave(enclaveId)

	// TODO We don't want to have the artifact cache inside the enclave data dir anymore - it should prob be a separate directory local filesystem
	//  This is because, with Kurtosis interactive, it will need to be independent of executions of Kurtosis
	filesArtifactCache, err := enclaveDataDir.GetFilesArtifactCache()
	if err != nil {
		return nil, nil, stacktrace.Propagate(err,"An error occurred getting the files artifact cache")
	}

	dockerNetworkId := args.NetworkId
	isPartitioningEnabled := args.IsPartitioningEnabled

	apiContainerSocketInsideNetwork := fmt.Sprintf(
		"%v:%v",
		args.ApiContainerIpAddr,
		args.GrpcListenPortNum,
	)

	filesArtifactExpander := files_artifact_expander.NewFilesArtifactExpander(
		args.EnclaveDataDirpathOnHostMachine,
		dockerManager,
		enclaveObjAttrsProvider,
		dockerNetworkId,
		freeIpAddrTracker,
		filesArtifactCache,
	)

	enclaveContainerLauncher := enclave_container_launcher.NewEnclaveContainerLauncher(
		dockerManager,
		enclaveObjAttrsProvider,
		args.EnclaveDataDirpathOnHostMachine,
	)

	userServiceLauncher := user_service_launcher.NewUserServiceLauncher(
		dockerManager,
		enclaveContainerLauncher,
		freeIpAddrTracker,
		filesArtifactExpander,
	)

	networkingSidecarManager := networking_sidecar.NewStandardNetworkingSidecarManager(
		dockerManager,
		enclaveObjAttrsProvider,
		freeIpAddrTracker,
		dockerNetworkId)

	serviceNetwork := service_network.NewServiceNetworkImpl(
		isPartitioningEnabled,
		freeIpAddrTracker,
		dockerManager,
		dockerNetworkId,
		enclaveDataDir,
		userServiceLauncher,
		networkingSidecarManager)

	moduleLauncher := module_launcher.NewModuleLauncher(
		enclaveId,
		dockerManager,
		apiContainerSocketInsideNetwork,
		enclaveContainerLauncher,
		freeIpAddrTracker,
		dockerNetworkId,
	)

	moduleStore := module_store.NewModuleStore(dockerManager, moduleLauncher)

	return serviceNetwork, moduleStore, nil
}
