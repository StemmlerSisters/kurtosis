/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_container_launcher

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"strconv"
)

const (
	dockerSocketFilepath = "/var/run/docker.sock"

	enclaveContainerPortNumUintBase = 10
	encalveContainerPortNumUintBits = 16

	uninitializedPublicIpAddrStrValue = ""
)

// EnclaveContainerLauncher
// Class that abstracts away the dirty work of launching a container and getting the public ports back from
//  the launched container
type EnclaveContainerLauncher struct {
	dockerManager *docker_manager.DockerManager
	
	enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider

	// The enclave data directory path on the host machine, so the launcher can bind-mount it to enclave
	//  containers
	enclaveDataDirpathOnHostMachine string
}

func NewEnclaveContainerLauncher(dockerManager *docker_manager.DockerManager, enclaveObjAttrsProvider schema.EnclaveObjectAttributesProvider, enclaveDataDirpathOnHostMachine string) *EnclaveContainerLauncher {
	return &EnclaveContainerLauncher{dockerManager: dockerManager, enclaveObjAttrsProvider: enclaveObjAttrsProvider, enclaveDataDirpathOnHostMachine: enclaveDataDirpathOnHostMachine}
}

// NOTE: Will return a nil IP & empty public ports map if no private ports are supplied
func (launcher *EnclaveContainerLauncher) Launch(
	ctx context.Context,
	image string, // The image to start the container with
	shouldPullContainerImageBeforeLaunch bool,
	ipAddr net.IP,
	dockerNetworkId string,
	enclaveDataDirMountDirpath string, // The location where the enclave data dir will be bind-mounted on the container
	privatePorts map[string]*EnclaveContainerPort,
	objectAttributesSupplier func(provider schema.EnclaveObjectAttributesProvider) (schema.ObjectAttributes, error),
	environmentVariables map[string]string,
	shouldBindMountDockerSocket bool,
	maybeAlias string,	// Leave as emptystring to not set an alias
	maybeEntrypointArgs []string, // Leave nil to not set ENTRYPOINT args
	maybeCmdArgs []string, // Leave nil to not set CMD args
	maybeVolumeMounts map[string]string, // Leave nil to not set any volume mounts
) (
	resultContainerId string,
	resultMaybePublicIpAddr net.IP,	// Will be nil if len(privatePorts) == 0
	resultPublicPorts map[string]*EnclaveContainerPort, // Will be empty if len(privatePorts) == 0
	resultErr error,
){
	// Best-effort pull attempt
	if shouldPullContainerImageBeforeLaunch {
		if err := launcher.dockerManager.PullImage(ctx, image); err != nil {
			logrus.Warnf(
				"Failed to pull the latest version of image '%v'; you may be running an out-of-date version",
				image,
			)
		}
	}

	portIdsForDockerPortObjs, publishSpecs, err := getPortMapsBeforeContainerStart(privatePorts)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the ports maps required for starting an enclave container")
	}

	bindMounts := map[string]string{
		launcher.enclaveDataDirpathOnHostMachine: enclaveDataDirMountDirpath,
	}
	if shouldBindMountDockerSocket {
		bindMounts[dockerSocketFilepath] = dockerSocketFilepath
	}
	
	objectAttributes, err := objectAttributesSupplier(launcher.enclaveObjAttrsProvider)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred getting the container attributes using the supplier")
	}

	containerName := objectAttributes.GetName()
	containerLabels := objectAttributes.GetLabels()
	createAndStartArgsBuilder := docker_manager.NewCreateAndStartContainerArgsBuilder(
		image,
		containerName,
		dockerNetworkId,
	).WithStaticIP(
		ipAddr,
	).WithUsedPorts(
		publishSpecs,
	).WithEnvironmentVariables(
		environmentVariables,
	).WithBindMounts(
		bindMounts,
	).WithLabels(
		containerLabels,
	)
	if maybeAlias != "" {
		createAndStartArgsBuilder.WithAlias(maybeAlias)
	}
	if maybeEntrypointArgs != nil {
		createAndStartArgsBuilder.WithEntrypointArgs(maybeEntrypointArgs)
	}
	if maybeCmdArgs != nil {
		createAndStartArgsBuilder.WithCmdArgs(maybeCmdArgs)
	}
	if maybeVolumeMounts != nil {
		createAndStartArgsBuilder.WithVolumeMounts(maybeVolumeMounts)
	}
	createAndStartArgs := createAndStartArgsBuilder.Build()
	containerId, hostPortBindingsByPortObj, err := launcher.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return "", nil, nil, stacktrace.Propagate(err, "An error occurred starting the Docker container for service with image '%v'", image)
	}
	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			if err := launcher.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Error("Launching the service container failed, but an error occurred killing container we started:")
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually kill container with ID '%v'", containerId)
			}
		}
	}()

	var maybePublicIpAddr net.IP = nil
	publicPorts := map[string]*EnclaveContainerPort{}
	if len(privatePorts) > 0 {
		maybePublicIpAddr, publicPorts, err = condensePublicNetworkInfoFromHostMachineBindings(
			hostPortBindingsByPortObj,
			privatePorts,
			portIdsForDockerPortObjs,
		)
		if err != nil {
			return "", nil, nil, stacktrace.Propagate(err, "An error occurred extracting public IP addr & ports from the host machine ports returned by the container engine")
		}
	}

	shouldKillContainer = false
	return containerId, maybePublicIpAddr, publicPorts, nil
}

// ====================================================================================================
//                                      Private Helper Functions
// ====================================================================================================
// Takes in the ports used by a container and provides the necessary maps required for:
//  1) getting the container's labels
//  2) starting the service
//  3) getting the service's host machine port bindings after the service is started
func getPortMapsBeforeContainerStart(
	privatePorts map[string]*EnclaveContainerPort,
) (
	resultPortIdsForDockerPortObjs map[nat.Port]string,
	resultPublishSpecs map[nat.Port]docker_manager.PortPublishSpec, // Used by container engine
	resultErr error,
) {
	portIdsForDockerPortObjs := map[nat.Port]string{}
	publishSpecs := map[nat.Port]docker_manager.PortPublishSpec{}
	for portId, enclaveContainerPort := range privatePorts {
		portNum := enclaveContainerPort.GetNumber()
		portProto := enclaveContainerPort.GetProtocol()

		dockerPortObj, err := nat.NewPort(
			string(portProto),
			fmt.Sprintf("%v", portNum),
		)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"An error occurred creating a Docker port object using port num '%v' and protocol string '%v'",
				portNum,
				portProto,
			)
		}

		if preexistingPortId, found := portIdsForDockerPortObjs[dockerPortObj]; found {
			return nil, nil, stacktrace.NewError(
				"Port '%v' declares Docker port spec '%v', but this port spec is already in use by port '%v'",
				portId,
				dockerPortObj,
				preexistingPortId,
			)
		}
		portIdsForDockerPortObjs[dockerPortObj] = portId

		publishSpecs[dockerPortObj] = docker_manager.NewAutomaticPublishingSpec()
	}
	return portIdsForDockerPortObjs, publishSpecs, nil
}

// condensePublicNetworkInfoFromHostMachineBindings
// Condenses declared private port bindings and the host machine port bindings returned by the container engine lib into:
//  1) a single host machine IP address
//  2) a map of private port binding IDs -> public ports
// An error is thrown if there are multiple host machine IP addresses
func condensePublicNetworkInfoFromHostMachineBindings(
	hostMachinePortBindings map[nat.Port]*nat.PortBinding,
	privatePorts map[string]*EnclaveContainerPort,
	portIdsForDockerPortObjs map[nat.Port]string,
) (
	resultPublicIpAddr net.IP,
	resultPublicPorts map[string]*EnclaveContainerPort,
	resultErr error,
) {
	if len(hostMachinePortBindings) == 0 {
		return nil, nil, stacktrace.NewError("Cannot condense public network info if no host machine port bindings are provided")
	}

	publicIpAddrStr := uninitializedPublicIpAddrStrValue
	publicPorts := map[string]*EnclaveContainerPort{}
	for dockerPortObj, hostPortBinding := range hostMachinePortBindings {
		portId, found := portIdsForDockerPortObjs[dockerPortObj]
		if !found {
			// If the container engine reports a host port binding that wasn't declared in the input used-ports object, ignore it
			// This could happen if a port is declared in the Dockerfile
			continue
		}

		privatePort, found := privatePorts[portId]
		if !found {
			return nil,  nil, stacktrace.NewError(
				"The container engine returned a host machine port binding for Docker port spec '%v', but this port spec didn't correspond to any port ID; this is very likely a bug in Kurtosis",
				dockerPortObj,
			)
		}

		hostIpAddr := hostPortBinding.HostIP
		if publicIpAddrStr == uninitializedPublicIpAddrStrValue {
			publicIpAddrStr = hostIpAddr
		} else if publicIpAddrStr != hostIpAddr {
			return nil, nil, stacktrace.NewError(
				"A public IP address '%v' was already declared for the service, but Docker port object '%v' declares a different public IP address '%v'",
				publicIpAddrStr,
				dockerPortObj,
				hostIpAddr,
			)
		}

		hostPortStr := hostPortBinding.HostPort
		hostPortUint64, err := strconv.ParseUint(hostPortStr, enclaveContainerPortNumUintBase, encalveContainerPortNumUintBits)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"An error occurred parsing host machine port string '%v' into a uint with %v bits and base %v",
				hostPortStr,
				encalveContainerPortNumUintBits,
				enclaveContainerPortNumUintBase,
			)
		}
		hostPortUint16 := uint16(hostPortUint64) // Safe to do because our ParseUint declares the expected number of bits
		portProto := privatePort.GetProtocol()
		publicPort, err := NewEnclaveContainerPort(hostPortUint16, portProto)
		if err != nil {
			return nil, nil, stacktrace.Propagate(
				err,
				"An error occurred creating public port object with num '%v' and protocol '%v'; this should never happen and likely means a bug in Kurtosis",
				hostPortUint16,
				portProto,
			)
		}
		publicPorts[portId] = publicPort
	}
	if publicIpAddrStr == uninitializedPublicIpAddrStrValue {
		return nil, nil, stacktrace.NewError("No public IP address string was retrieved from host port bindings: %+v", hostMachinePortBindings)
	}
	publicIpAddr := net.ParseIP(publicIpAddrStr)
	if publicIpAddr == nil {
		return nil, nil, stacktrace.NewError("Couldn't parse service's public IP address string '%v' to an IP object", publicIpAddrStr)
	}
	return publicIpAddr, publicPorts, nil
}
