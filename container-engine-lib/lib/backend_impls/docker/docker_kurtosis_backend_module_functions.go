package docker

import (
	"context"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager/types"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

const (
	// The module container uses gRPC so MUST listen on TCP (no other protocols are supported)
	moduleContainerPortProtocol          = port_spec.PortProtocol_TCP

	maxWaitForModuleContainerAvailabilityRetries         = 10
	timeBetweenWaitForModuleContainerAvailabilityRetries = 1 * time.Second
)

func (backend *DockerKurtosisBackend) CreateModule(
	ctx context.Context,
	image string,
	enclaveId enclave.EnclaveID,
	id module.ModuleID,
	guid module.ModuleGUID,
	ipAddr net.IP, // TODO REMOVE THIS ONCE WE FIX THE STATIC IP PROBLEM!!
	grpcPortNum uint16,
	enclaveDataDirpathOnHostMachine string,
	envVars map[string]string,
)(
	newModule *module.Module,
	resultErr error,
) {
	// Verify no module container with the given GUID already exists in the enclave
	preexistingModuleFilters := &module.ModuleFilters{
		EnclaveIDs: map[enclave.EnclaveID]bool{
			enclaveId: true,
		},
		GUIDs: map[module.ModuleGUID]bool{
			guid: true,
		},
	}
	preexistingModules, err := backendCore.GetModules(ctx, preexistingModuleFilters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting preexisting modules in enclave '%v' with GUID '%v'", enclaveId, guid)
	}
	if len(preexistingModules) > 0 {
		return nil, stacktrace.NewError("Found existing module container(s) in enclave '%v' with GUID '%v'; cannot start a new one", enclaveId, guid)
	}


	// Get the Docker network ID where we'll start the new API container
	matchingNetworks, err := backendCore.getEnclaveNetworksByEnclaveIds(ctx, map[enclave.EnclaveID]bool{
		enclaveId: true,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave networks for enclave ID '%v'", enclaveId)
	}
	numMatchingNetworks := len(matchingNetworks)
	if numMatchingNetworks == 0 {
		return nil, stacktrace.NewError("No network found for enclave with ID '%v'", enclaveId)
	}
	if numMatchingNetworks > 1 {
		return nil, stacktrace.NewError("Found '%v' enclave networks with ID '%v', which shouldn't happen", numMatchingNetworks, enclaveId)
	}
	enclaveNetwork := matchingNetworks[0]

	privateGrpcPortSpec, err := port_spec.NewPortSpec(grpcPortNum, moduleContainerPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating the module container's private grpc port spec object using number '%v' and protocol '%v'",
			grpcPortNum,
			enginePortProtocol.String(),
		)
	}

	enclaveObjAttrProvider, err := backendCore.objAttrsProvider.ForEnclave(enclaveId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Couldn't get an object attribute provider for enclave '%v'", enclaveId)
	}

	moduleContainerAttrs, err := enclaveObjAttrProvider.ForModuleContainer(
		ipAddr,
		string(id),
		string(guid),
		kurtosisInternalContainerGrpcPortId,
		privateGrpcPortSpec,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the object attributes for the module container")
	}

	privateGrpcDockerPort, err := transformPortSpecToDockerPort(privateGrpcPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred transforming the private grpc port spec to a Docker port")
	}
	usedPorts := map[nat.Port]docker_manager.PortPublishSpec{
		privateGrpcDockerPort:      docker_manager.NewAutomaticPublishingSpec(),
	}

	bindMounts := map[string]string{
		enclaveDataDirpathOnHostMachine: enclaveDataDirpathOnAPIContainer,
	}

	labelStrs := map[string]string{}
	for labelKey, labelValue := range moduleContainerAttrs.GetLabels() {
		labelStrs[labelKey.GetString()] = labelValue.GetString()
	}

	// Best-effort pull attempt
	if err = backendCore.dockerManager.PullImage(ctx, image); err != nil {
		logrus.Warnf("Failed to pull the latest version of module container image '%v'; you may be running an out-of-date version", image)
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		image,
		moduleContainerAttrs.GetName().GetString(),
		enclaveNetwork.GetId(),
	).WithEnvironmentVariables(
		envVars,
	).WithBindMounts(
		bindMounts,
	).WithUsedPorts(
		usedPorts,
	).WithLabels(
		labelStrs,
	).Build()

	containerId, hostMachinePortBindings, err := backendCore.dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred starting the module container")
	}
	shouldKillContainer := true
	defer func() {
		if shouldKillContainer {
			// NOTE: We use the background context here so that the kill will still go off even if the reason for
			// the failure was the original context being cancelled
			if err := backendCore.dockerManager.KillContainer(context.Background(), containerId); err != nil {
				logrus.Errorf(
					"Launching module container '%v' with container ID '%v' didn't complete successfully so we " +
						"tried to kill the container we started, but doing so exited with an error:\n%v",
					moduleContainerAttrs.GetName(),
					containerId,
					err,
				)
				logrus.Errorf("ACTION REQUIRED: You'll need to manually stop module container with ID '%v'!!!!!!", containerId)
			}
		}
	}()

	if err := waitForPortAvailabilityUsingNetstat(
		ctx,
		backendCore.dockerManager,
		containerId,
		privateGrpcPortSpec,
		maxWaitForModuleContainerAvailabilityRetries,
		timeBetweenWaitForModuleContainerAvailabilityRetries,
	); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred waiting for the module container's grpc port to become available")
	}

	result, err := getModuleObjectFromContainerInfo(containerId, labelStrs, types.ContainerStatus_Running, hostMachinePortBindings)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a module container object from container with ID '%v'", containerId)
	}

	shouldKillContainer = false
	return result, nil
}

func (backend *DockerKurtosisBackend) GetModules(
	ctx context.Context,
	filters *module.ModuleFilters,
) (
	map[module.ModuleGUID]*module.Module,
	error,
) {
	matchingModuleContainers, err := backendCore.getMatchingModules(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting module containers matching the following filters: %+v", filters)
	}

	matchingModuleContainersByModuleID := map[module.ModuleGUID]*module.Module{}
	for _, moduleObj := range matchingModuleContainers {
		matchingModuleContainersByModuleID[moduleObj.GetGUID()] = moduleObj
	}

	return matchingModuleContainersByModuleID, nil
}

func (backend *DockerKurtosisBackend) DestroyModules(
	ctx context.Context,
	filters *module.ModuleFilters,
) (
	successfulModuleIds map[module.ModuleGUID]bool,
	erroredModuleIds map[module.ModuleGUID]error,
	resultErr error,
) {
	matchingModuleContainersByContainerId, err := backendCore.getMatchingModules(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting module containers matching the following filters: %+v", filters)
	}

	successIds := map[module.ModuleGUID]bool{}
	errorIds := map[module.ModuleGUID]error{}
	for containerId, moduleObj := range matchingModuleContainersByContainerId {
		moduleGuid := moduleObj.GetGUID()
		enclaveId := moduleObj.GetEnclaveID()
		if err := backendCore.dockerManager.RemoveContainer(ctx, containerId); err != nil {
			wrappedErr := stacktrace.Propagate(
				err,
				"An error occurred removing module container for GUID '%v' with container ID '%v' from enclave '%v'",
				moduleGuid,
				containerId,
				enclaveId,
			)
			errorIds[moduleGuid] = wrappedErr
		} else {
			successIds[moduleGuid] = true
		}
	}
	return successIds, errorIds, nil
}

// ====================================================================================================
//                                     Private Helper Methods
// ====================================================================================================
// Gets modules matching the search filters, indexed by their container ID
func (backendCore *DockerKurtosisBackend) getMatchingModules(ctx context.Context, filters *module.ModuleFilters) (map[string]*module.Module, error) {
	moduleContainerSearchLabels := map[string]string{
		label_key_consts.AppIDLabelKey.GetString():         label_value_consts.AppIDLabelValue.GetString(),
		label_key_consts.ContainerTypeLabelKey.GetString(): label_value_consts.ModuleContainerTypeLabelValue.GetString(),
	}
	matchingModuleContainers, err := backendCore.dockerManager.GetContainersByLabels(ctx, moduleContainerSearchLabels, shouldFetchAllContainersWhenRetrievingContainers)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred fetching module containers using labels: %+v", moduleContainerSearchLabels)
	}

	matchingModuleObjects := map[string]*module.Module{}
	for _, moduleContainer := range matchingModuleContainers {
		containerId := moduleContainer.GetId()
		moduleObj, err := getModuleObjectFromContainerInfo(
			containerId,
			moduleContainer.GetLabels(),
			moduleContainer.GetStatus(),
			moduleContainer.GetHostPortBindings(),
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred converting container with ID '%v' into a module object", moduleContainer.GetId())
		}

		if filters.EnclaveIDs != nil && len(filters.EnclaveIDs) > 0 {
			if _, found := filters.EnclaveIDs[moduleObj.GetEnclaveID()]; !found {
				continue
			}
		}

		// If the ID filter is specified, drop modules not matching it
		if filters.GUIDs != nil && len(filters.GUIDs) > 0 {
			if _, found := filters.GUIDs[moduleObj.GetGUID()]; !found {
				continue
			}
		}

		// If status filter is specified, drop modules	 not matching it
		if filters.Statuses != nil && len(filters.Statuses) > 0 {
			if _, found := filters.Statuses[moduleObj.GetStatus()]; !found {
				continue
			}
		}

		matchingModuleObjects[containerId] = moduleObj
	}

	return matchingModuleObjects, nil
}

func getModuleObjectFromContainerInfo(
	containerId string,
	labels map[string]string,
	containerStatus types.ContainerStatus,
	allHostMachinePortBindings map[nat.Port]*nat.PortBinding,
) (*module.Module, error) {
	enclaveId, found := labels[label_key_consts.EnclaveIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected the module's enclave ID to be found under label '%v' but the label wasn't present", label_key_consts.EnclaveIDLabelKey.GetString())
	}

	id, found := labels[label_key_consts.IDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find module ID label key '%v' but none was found", label_key_consts.IDLabelKey.GetString())
	}

	guid, found := labels[label_key_consts.GUIDLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find module GUID label key '%v' but none was found", label_key_consts.GUIDLabelKey.GetString())
	}

	var privateIpAddr net.IP
	privateIpAddrStr, found := labels[label_key_consts.PrivateIPLabelKey.GetString()]
	// UNCOMMENT THIS AFTER 2022-06-30 WHEN NOBODY HAS MODULES WITHOUT THE PRIVATE IP ADDRESS LABEL
	/*
		if !found {
			return nil, stacktrace.NewError("Expected to find module private IP label key '%v' but none was found", label_key_consts.PrivateIPLabelKey.GetString())
		}
	*/
	if found {
		candidatePrivateIpAddr := net.ParseIP(privateIpAddrStr)
		if candidatePrivateIpAddr == nil {
			return nil, stacktrace.NewError("Couldn't parse private IP address string '%v' to an IP", privateIpAddrStr)
		}
		privateIpAddr = candidatePrivateIpAddr
	}

	privateGrpcPortSpec, err := getPrivateModulePorts(labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the module container's private port specs from container '%v' with labels: %+v", containerId, labels)
	}

	isContainerRunning, found := isContainerRunningDeterminer[containerStatus]
	if !found {
		// This should never happen because we enforce completeness in a unit test
		return nil, stacktrace.NewError("No is-running designation found for module container status '%v'; this is a bug in Kurtosis!", containerStatus.String())
	}
	var moduleStatus container_status.ContainerStatus
	if isContainerRunning {
		moduleStatus = container_status.ContainerStatus_Running
	} else {
		moduleStatus = container_status.ContainerStatus_Stopped
	}

	var publicIpAddr net.IP
	var publicGrpcPortSpec *port_spec.PortSpec
	if moduleStatus == container_status.ContainerStatus_Running {
		publicGrpcPortIpAddr, candidatePublicGrpcPortSpec, err := getPublicPortBindingFromPrivatePortSpec(privateGrpcPortSpec, allHostMachinePortBindings)
		if err != nil {
			return nil, stacktrace.Propagate(err, "The module is running, but an error occurred getting the public port spec for the module's grpc private port spec")
		}
		publicGrpcPortSpec = candidatePublicGrpcPortSpec
		publicIpAddr = publicGrpcPortIpAddr
	}

	result := module.NewModule(
		enclave.EnclaveID(enclaveId),
		module.ModuleID(id),
		module.ModuleGUID(guid),
		moduleStatus,
		privateIpAddr,
		privateGrpcPortSpec,
		publicIpAddr,
		publicGrpcPortSpec,
	)

	return result, nil
}

func getPrivateModulePorts(containerLabels map[string]string) (
	resultGrpcPortSpec *port_spec.PortSpec,
	resultErr error,
) {
	serializedPortSpecs, found := containerLabels[label_key_consts.PortSpecsLabelKey.GetString()]
	if !found {
		return nil, stacktrace.NewError("Expected to find port specs label '%v' but none was found", label_key_consts.PortSpecsLabelKey.GetString())
	}

	portSpecs, err := port_spec_serializer.DeserializePortSpecs(serializedPortSpecs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing port specs string '%v'", serializedPortSpecs)
	}

	grpcPortSpec, foundGrpcPort := portSpecs[kurtosisInternalContainerGrpcPortId]
	if !foundGrpcPort {
		return nil, stacktrace.NewError("No grpc port with ID '%v' found in the port specs", kurtosisInternalContainerGrpcPortId)
	}

	return grpcPortSpec, nil
}
