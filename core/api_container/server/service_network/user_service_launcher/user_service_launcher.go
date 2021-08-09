/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package user_service_launcher

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/api_container/server/optional_host_port_binding_supplier"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/user_service_launcher/files_artifact_expander"
	"github.com/kurtosis-tech/kurtosis/commons"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/object_name_providers"
	"github.com/palantir/stacktrace"
	"net"
)

/*
Convenience struct whose only purpose is launching user services
 */
type UserServiceLauncher struct {
	dockerManager *docker_manager.DockerManager

	enclaveObjNameProvider *object_name_providers.EnclaveObjectNameProvider

	freeIpAddrTracker *commons.FreeIpAddrTracker

	optionalHostPortBindingSupplier *optional_host_port_binding_supplier.OptionalHostPortBindingSupplier

	filesArtifactExpander *files_artifact_expander.FilesArtifactExpander

	dockerNetworkId string

	// The name of the Docker volume containing data for the enclave
	enclaveDataVolName string
}

func NewUserServiceLauncher(dockerManager *docker_manager.DockerManager, enclaveObjNameProvider *object_name_providers.EnclaveObjectNameProvider, freeIpAddrTracker *commons.FreeIpAddrTracker, optionalHostPortBindingSupplier *optional_host_port_binding_supplier.OptionalHostPortBindingSupplier, filesArtifactExpander *files_artifact_expander.FilesArtifactExpander, dockerNetworkId string, enclaveDataVolName string) *UserServiceLauncher {
	return &UserServiceLauncher{dockerManager: dockerManager, enclaveObjNameProvider: enclaveObjNameProvider, freeIpAddrTracker: freeIpAddrTracker, optionalHostPortBindingSupplier: optionalHostPortBindingSupplier, filesArtifactExpander: filesArtifactExpander, dockerNetworkId: dockerNetworkId, enclaveDataVolName: enclaveDataVolName}
}

/**
Launches a testnet service with the given parameters

Returns:
	* The container ID of the newly-launched service
	* The mapping of used_port -> host_port_binding (if no host port is bound, then the value will be nil)
 */
func (launcher UserServiceLauncher) Launch(
		ctx context.Context,
		serviceId service_network_types.ServiceID,
		ipAddr net.IP,
		imageName string,
		usedPorts map[nat.Port]bool,
		entrypointArgs []string,
		cmdArgs []string,
		dockerEnvVars map[string]string,
		enclaveDataVolMntDirpath string,
		// Mapping files artifact ID -> mountpoint on the container to launch
		filesArtifactIdsToMountpoints map[string]string) (string, map[nat.Port]*nat.PortBinding, error) {

	usedArtifactIdSet := map[string]bool{}
	for artifactId := range filesArtifactIdsToMountpoints {
		usedArtifactIdSet[artifactId] = true
	}

	// First expand the files artifacts into volumes, so that any errors get caught early
	// NOTE: if users don't need to investigate the volume contents, we could keep track of the volumes we create
	//  and delete them at the end of the test to keep things cleaner
	artifactIdsToVolumes, err := launcher.filesArtifactExpander.ExpandArtifactsIntoVolumes(ctx, serviceId, usedArtifactIdSet)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred expanding the requested files artifacts into volumes")
	}

	artifactVolumeMounts := map[string]string{}
	for artifactId, mountpoint := range filesArtifactIdsToMountpoints {
		artifactVolume, found := artifactIdsToVolumes[artifactId]
		if !found {
			return "", nil, stacktrace.NewError(
				"Even though we declared that we need files artifact '%v' to be expanded, no volume containing the " +
					"expanded contents was found; this is a bug in Kurtosis",
				artifactId,
			)
		}
		artifactVolumeMounts[artifactVolume] = mountpoint
	}

	usedPortsWithHostBindings, err := launcher.optionalHostPortBindingSupplier.BindPortsToHostIfNeeded(usedPorts)
	if err != nil {
		return "", nil, stacktrace.Propagate(
			err,
			"An error occurred getting used ports with host bindings",
		)
	}

	volumeMounts := map[string]string{
		launcher.enclaveDataVolName: enclaveDataVolMntDirpath,
	}
	for artifactVolName, mountpoint := range artifactVolumeMounts {
		volumeMounts[artifactVolName] = mountpoint
	}

	containerId, err := launcher.dockerManager.CreateAndStartContainer(
		ctx,
		imageName,
		launcher.enclaveObjNameProvider.ForUserServiceContainer(serviceId),
		launcher.dockerNetworkId,
		ipAddr,
		map[docker_manager.ContainerCapability]bool{},
		docker_manager.DefaultNetworkMode,
		usedPortsWithHostBindings,
		entrypointArgs,
		cmdArgs,
		dockerEnvVars,
		map[string]string{}, // no bind mounts for services created via the Kurtosis API
		volumeMounts,
		false,		// User services definitely shouldn't be able to access the Docker host machine
	)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred starting the Docker container for service with image '%v'", imageName)
	}
	return containerId, usedPortsWithHostBindings, nil
}
