package user_services_functions

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_kurtosis_backend/shared_helpers"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/kubernetes_manager"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_annotation_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/kubernetes/object_attributes_provider/kubernetes_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/container-engine-lib/lib/operation_parallelizer"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	applyconfigurationsv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

const(
	userServiceContainerName = "user-service-container"
	// Our user services don't need service accounts
	userServiceServiceAccountName = ""

	megabytesToBytesFactor = 1_000_000
)

// Completeness enforced via unit test
var kurtosisPortProtocolToKubernetesPortProtocolTranslator = map[port_spec.PortProtocol]apiv1.Protocol{
	port_spec.PortProtocol_TCP: apiv1.ProtocolTCP,
	port_spec.PortProtocol_UDP: apiv1.ProtocolUDP,
	port_spec.PortProtocol_SCTP: apiv1.ProtocolSCTP,
}

func StartUserServices(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	services map[service.ServiceGUID]*service.ServiceConfig,
	cliModeArgs *shared_helpers.CliModeArgs,
	apiContainerModeArgs *shared_helpers.ApiContainerModeArgs,
	engineServerModeArgs *shared_helpers.EngineServerModeArgs,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceGUID]service.Service,
	map[service.ServiceGUID]error,
	error,
) {
	// Sanity check for port bindings on all services
	//TODO this is a huge hack to temporarily enable static ports for NEAR until we have a more productized solution
	for _, config := range services {
		publicPorts := config.GetPublicPorts()
		if publicPorts != nil && len(publicPorts) > 0 {
			logrus.Warn("The Kubernetes Kurtosis backend doesn't support defining static ports for services; the public ports will be ignored")
		}
	}

	// Check all services already have registrations attached
	serviceGUIDs := map[service.ServiceGUID]bool{}
	for guid := range services {
		serviceGUIDs[guid] = true
	}
	preexistingServicesFilters := &service.ServiceFilters{
		GUIDs: serviceGUIDs,
	}
	preexistingObjectsAndResources, err := shared_helpers.GetMatchingUserServiceObjectsAndKubernetesResources(
		ctx,
		enclaveID,
		preexistingServicesFilters,
		cliModeArgs,
		apiContainerModeArgs,
		engineServerModeArgs,
		kubernetesManager,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user service objects and Kubernetes resources matching service GUIDs '%v'", serviceGUIDs)
	}
	for guid, _ := range serviceGUIDs {
		if _, found := preexistingObjectsAndResources[guid]; !found {
			return nil, nil, stacktrace.NewError("Couldn't find any service registrations matching service GUID '%v'", guid)
		}
	}
	if len(preexistingObjectsAndResources) > len(serviceGUIDs) {
		// Should never happen because service GUIDs should be unique
		return nil, nil, stacktrace.NewError("Found more than one service registration matching service GUIDs; this is a bug in Kurtosis")
	}

	successfulStarts, failedStarts, err := runStartServiceOperationsInParallel(
		ctx,
		enclaveID,
		services,
		preexistingObjectsAndResources,
		kubernetesManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to start services in parallel.")
	}

	return successfulStarts, failedStarts, nil
}

// ====================================================================================================
//                       Private helper functions
// ====================================================================================================
func runStartServiceOperationsInParallel(
	ctx context.Context,
	enclaveID enclave.EnclaveID,
	services map[service.ServiceGUID]*service.ServiceConfig,
	servicesObjectsAndResources map[service.ServiceGUID]*shared_helpers.UserServiceObjectsAndKubernetesResources,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	map[service.ServiceGUID]service.Service,
	map[service.ServiceGUID]error,
	error,
) {
	startServiceOperations := map[operation_parallelizer.OperationID]operation_parallelizer.Operation{}
	for guid, config := range services {
		startServiceOperations[operation_parallelizer.OperationID(guid)] = createStartServiceOperation(
			ctx,
			guid,
			config,
			servicesObjectsAndResources,
			enclaveID,
			kubernetesManager)
	}

	successfulServiceObjs, failedOperations := operation_parallelizer.RunOperationsInParallel(startServiceOperations)

	successfulServices := map[service.ServiceGUID]service.Service{}
	failedServices := map[service.ServiceGUID]error{}

	for id, data := range successfulServiceObjs {
		serviceGUID := service.ServiceGUID(id)
		serviceObj, ok := data.(service.Service)
		if !ok {
			return nil, nil, stacktrace.NewError("Casting to a service object on data returned from the operation to start service with guid `%v` failed.", serviceGUID)
		}
		successfulServices[serviceGUID] = serviceObj
	}

	for id, err := range failedOperations {
		serviceGUID := service.ServiceGUID(id)
		failedServices[serviceGUID] = err
	}

	return successfulServices, failedServices, nil
}

func createStartServiceOperation(
	ctx context.Context,
	serviceGUID service.ServiceGUID,
	serviceConfig *service.ServiceConfig,
	servicesObjectsAndResources map[service.ServiceGUID]*shared_helpers.UserServiceObjectsAndKubernetesResources,
	enclaveID enclave.EnclaveID,
	kubernetesManager *kubernetes_manager.KubernetesManager) operation_parallelizer.Operation {
	return func() (interface{}, error) {
		filesArtifactsExpansion := serviceConfig.GetFilesArtifactsExpansion()
		containerImageName := serviceConfig.GetContainerImageName()
		privatePorts := serviceConfig.GetPrivatePorts()
		entrypointArgs := serviceConfig.GetEntrypointArgs()
		cmdArgs := serviceConfig.GetCmdArgs()
		envVars := serviceConfig.GetEnvVars()
		cpuAllocationMillicpus := serviceConfig.GetCPUAllocationMillicpus()
		memoryAllocationMegabytes := serviceConfig.GetMemoryAllocationMegabytes()

		matchingObjectAndResources, found := servicesObjectsAndResources[serviceGUID]
		if !found {
			return nil, stacktrace.NewError("Even though we pulled back some Kubernetes resources, no Kubernetes resources were available for requested service GUID '%v'; this is a bug in Kurtosis", serviceGUID)
		}
		kubernetesService := matchingObjectAndResources.KubernetesResources.Service
		serviceObj := matchingObjectAndResources.Service
		if serviceObj != nil {
			return nil, stacktrace.NewError("Cannot start service with GUID '%v' because the service has already been started previously", serviceGUID)
		}

		namespaceName := kubernetesService.GetNamespace()
		serviceRegistrationObj := matchingObjectAndResources.ServiceRegistration

		var podInitContainers []apiv1.Container
		var podVolumes []apiv1.Volume
		var userServiceContainerVolumeMounts []apiv1.VolumeMount
		if filesArtifactsExpansion != nil {
			podVolumes, userServiceContainerVolumeMounts, podInitContainers, _ = prepareFilesArtifactsExpansionResources(
				filesArtifactsExpansion.ExpanderImage,
				filesArtifactsExpansion.ExpanderEnvVars,
				filesArtifactsExpansion.ExpanderDirpathsToServiceDirpaths,
			)
		}

		objectAttributesProvider := object_attributes_provider.GetKubernetesObjectAttributesProvider()
		enclaveObjAttributesProvider := objectAttributesProvider.ForEnclave(enclaveID)

		// Create the pod
		podAttributes, err := enclaveObjAttributesProvider.ForUserServicePod(serviceGUID, serviceRegistrationObj.GetID(), privatePorts)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting attributes for new pod for service with GUID '%v'", serviceGUID)
		}
		podLabelsStrs := shared_helpers.GetStringMapFromLabelMap(podAttributes.GetLabels())
		podAnnotationsStrs := shared_helpers.GetStringMapFromAnnotationMap(podAttributes.GetAnnotations())

		podContainers, err := getUserServicePodContainerSpecs(
			containerImageName,
			entrypointArgs,
			cmdArgs,
			envVars,
			privatePorts,
			userServiceContainerVolumeMounts,
			cpuAllocationMillicpus,
			memoryAllocationMegabytes,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the container specs for the user service pod with image '%v'", containerImageName)
		}

		podName := podAttributes.GetName().GetString()
		createdPod, err := kubernetesManager.CreatePod(
			ctx,
			namespaceName,
			podName,
			podLabelsStrs,
			podAnnotationsStrs,
			podInitContainers,
			podContainers,
			podVolumes,
			userServiceServiceAccountName,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating pod '%v' using image '%v'", podName, containerImageName)
		}
		shouldDestroyPod := true
		defer func() {
			if shouldDestroyPod {
				if err := kubernetesManager.RemovePod(ctx, createdPod); err != nil {
					logrus.Errorf("Starting service didn't complete successfully so we tried to remove the pod we created but doing so threw an error:\n%v", err)
					logrus.Errorf("ACTION REQUIRED: You'll need to remove pod '%v' in '%v' manually!!!", podName, namespaceName)
				}
			}
		}()

		updatedService, undoServiceUpdateFunc, err := updateServiceWhenContainerStarted(ctx, namespaceName, kubernetesService, privatePorts, kubernetesManager)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred updating service '%v' to reflect its new ports: %+v", kubernetesService.GetName(), privatePorts)
		}
		shouldUndoServiceUpdate := true
		defer func() {
			if shouldUndoServiceUpdate {
				undoServiceUpdateFunc()
			}
		}()

		kubernetesResources := map[service.ServiceGUID]*shared_helpers.UserServiceKubernetesResources{
			serviceGUID: {
				Service: updatedService,
				Pod:     createdPod,
			},
		}

		convertedObjects, err := shared_helpers.GetUserServiceObjectsFromKubernetesResources(enclaveID, kubernetesResources)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting a service object from the Kubernetes service and newly-created pod")
		}
		objectsAndResources, found := convertedObjects[serviceGUID]
		if !found {
			return nil, stacktrace.NewError(
				"Successfully converted the Kubernetes service + pod representing a running service with GUID '%v' to a "+
					"Kurtosis object, but couldn't find that key in the resulting map; this is a bug in Kurtosis",
				serviceGUID,
			)
		}

		shouldDestroyPod = false
		shouldUndoServiceUpdate = false
		return objectsAndResources.Service, nil
	}
}

// Update the service to:
// - Set the service ports appropriately
// - Irrevocably record that a pod is bound to the service (so that even if the pod is deleted, the service won't
// 	 be usable again
func updateServiceWhenContainerStarted(
	ctx context.Context,
	namespaceName string,
	kubernetesService *apiv1.Service,
	privatePorts map[string]*port_spec.PortSpec,
	kubernetesManager *kubernetes_manager.KubernetesManager,
) (
	*apiv1.Service,
	func(), // function to undo the update
	error,
) {
	serviceName := kubernetesService.GetName()

	serializedPortSpecs, err := kubernetes_port_spec_serializer.SerializePortSpecs(privatePorts)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred serializing the following private port specs: %+v", privatePorts)
	}

	// We only need to modify the ports from the default (unbound) ports if the user actually declares ports
	newServicePorts := kubernetesService.Spec.Ports
	if len(privatePorts) > 0 {
		candidateNewServicePorts, err := getKubernetesServicePortsFromPrivatePortSpecs(privatePorts)
		if err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes service ports for the following private port specs: %+v", privatePorts)
		}
		newServicePorts = candidateNewServicePorts
	}

	newAnnotations := kubernetesService.Annotations
	if newAnnotations == nil {
		newAnnotations = map[string]string{}
	}
	newAnnotations[kubernetes_annotation_key_consts.PortSpecsKubernetesAnnotationKey.GetString()] = serializedPortSpecs.GetString()

	updatingConfigurator := func(updatesToApply *applyconfigurationsv1.ServiceApplyConfiguration) {
		specUpdateToApply := applyconfigurationsv1.ServiceSpec()
		for _, newServicePort := range newServicePorts {
			// You'd *think* that we could just feed in &newServicePort below, since the struct below needs pointer
			// args. However, this would be a bug: Go for loop iteration variables are updated in-place (probably to
			// save on memory), so if you pass around references to the for loop iteration variable then all dereferences
			// done after the loop will get the same value (the value of the last iteration). Therefore, we copy the variable
			// on each loop so that we have a fixed moment-in-time value.
			newServicePortCopy := newServicePort
			portUpdateToApply := &applyconfigurationsv1.ServicePortApplyConfiguration{
				Name:     &newServicePortCopy.Name,
				Protocol: &newServicePortCopy.Protocol,
				// TODO fill this out for an app port!
				AppProtocol: nil,
				Port:        &newServicePortCopy.Port,
			}
			specUpdateToApply.WithPorts(portUpdateToApply)
		}
		updatesToApply.WithSpec(specUpdateToApply)

		updatesToApply.WithAnnotations(newAnnotations)
	}

	undoUpdateFunc := func() {
		undoingConfigurator := func(reversionToApply *applyconfigurationsv1.ServiceApplyConfiguration) {
			specReversionToApply := applyconfigurationsv1.ServiceSpec()
			for _, oldServicePort := range newServicePorts {
				portUpdateToApply := &applyconfigurationsv1.ServicePortApplyConfiguration{
					Name:     &oldServicePort.Name,
					Protocol: &oldServicePort.Protocol,
					// TODO fill this out for an app port!
					AppProtocol: nil,
					Port:        &oldServicePort.Port,
				}
				specReversionToApply.WithPorts(portUpdateToApply)
			}
			reversionToApply.WithSpec(specReversionToApply)

			reversionToApply.WithAnnotations(kubernetesService.Annotations)
		}
		if _, err := kubernetesManager.UpdateService(ctx, namespaceName, serviceName, undoingConfigurator); err != nil {
			logrus.Errorf(
				"An error occurred updating Kubernetes service '%v' in namespace '%v' to open the service ports "+
					"and add the serialized private port specs annotation so we tried to revert the service to "+
					"its values before the update, but an error occurred; this means the service is likely in "+
					"an inconsistent state:\n%v",
				serviceName,
				namespaceName,
				err,
			)
			// Nothing we can really do here to recover
		}
	}

	updatedService, err := kubernetesManager.UpdateService(ctx, namespaceName, serviceName, updatingConfigurator)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred updating Kubernetes service '%v' in namespace '%v' to open the service ports and add "+
				"the serialized private port specs annotation",
			serviceName,
			namespaceName,
		)
	}
	shouldUndoUpdate := true
	defer func() {
		if shouldUndoUpdate {
			undoUpdateFunc()
		}
	}()

	shouldUndoUpdate = false
	return updatedService, undoUpdateFunc, nil
}

func getUserServicePodContainerSpecs(
	image string,
	entrypointArgs []string,
	cmdArgs []string,
	envVarStrs map[string]string,
	privatePorts map[string]*port_spec.PortSpec,
	containerMounts []apiv1.VolumeMount,
	cpuAllocationMillicpus uint64,
	memoryAllocationMegabytes uint64,
) (
	[]apiv1.Container,
	error,
) {
	var containerEnvVars []apiv1.EnvVar
	for varName, varValue := range envVarStrs {
		envVar := apiv1.EnvVar{
			Name:  varName,
			Value: varValue,
		}
		containerEnvVars = append(containerEnvVars, envVar)
	}

	kubernetesContainerPorts, err := getKubernetesContainerPortsFromPrivatePortSpecs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting Kubernetes container ports from the private port specs map")
	}

	resourceLimitsList := apiv1.ResourceList{}
	resourceRequestsList := apiv1.ResourceList{}
	// 0 is considered the empty value (meaning the field was never set), so if either fields are 0, that resource is left unbounded
	if cpuAllocationMillicpus != 0 {
		resourceLimitsList[apiv1.ResourceCPU] = *resource.NewMilliQuantity(int64(cpuAllocationMillicpus), resource.DecimalSI)
		resourceRequestsList[apiv1.ResourceCPU] = *resource.NewMilliQuantity(int64(cpuAllocationMillicpus), resource.DecimalSI)
	}
	if memoryAllocationMegabytes != 0 {
		memoryAllocationInBytes := convertMegabytesToBytes(memoryAllocationMegabytes)
		resourceLimitsList[apiv1.ResourceMemory] = *resource.NewQuantity(int64(memoryAllocationInBytes), resource.DecimalSI)
		resourceRequestsList[apiv1.ResourceMemory] = *resource.NewQuantity(int64(memoryAllocationInBytes), resource.DecimalSI)
	}
	resourceRequirements := apiv1.ResourceRequirements{
		Limits: resourceLimitsList,
		Requests: resourceRequestsList,
	}

	// TODO create networking sidecars here
	containers := []apiv1.Container{
		{
			Name:  userServiceContainerName,
			Image: image,
			// Yes, even though this is called "command" it actually corresponds to the Docker ENTRYPOINT
			Command:      entrypointArgs,
			Args:         cmdArgs,
			Ports:        kubernetesContainerPorts,
			Env:          containerEnvVars,
			VolumeMounts: containerMounts,
			Resources:    resourceRequirements,

			// NOTE: There are a bunch of other interesting Container options that we omitted for now but might
			// want to specify in the future
		},
	}

	return containers, nil
}

func getKubernetesServicePortsFromPrivatePortSpecs(privatePorts map[string]*port_spec.PortSpec) ([]apiv1.ServicePort, error) {
	result := []apiv1.ServicePort{}
	for portId, portSpec := range privatePorts {
		kurtosisProtocol := portSpec.GetProtocol()
		kubernetesProtocol, found := kurtosisPortProtocolToKubernetesPortProtocolTranslator[kurtosisProtocol]
		if !found {
			// Should never happen because we enforce completeness via unit test
			return nil, stacktrace.NewError("No Kubernetes port protocol was defined for Kurtosis port protocol '%v'; this is a bug in Kurtosis", kurtosisProtocol)
		}

		kubernetesPortObj := apiv1.ServicePort{
			Name:        portId,
			Protocol:    kubernetesProtocol,
			// TODO Specify this!!! Will make for a really nice user interface (e.g. "https")
			AppProtocol: nil,
			// Safe to cast because max uint16 < int32
			Port:        int32(portSpec.GetNumber()),
		}
		result = append(result, kubernetesPortObj)
	}
	return result, nil
}

func getKubernetesContainerPortsFromPrivatePortSpecs(privatePorts map[string]*port_spec.PortSpec) ([]apiv1.ContainerPort, error) {
	result := []apiv1.ContainerPort{}
	for portId, portSpec := range privatePorts {
		kurtosisProtocol := portSpec.GetProtocol()
		kubernetesProtocol, found := kurtosisPortProtocolToKubernetesPortProtocolTranslator[kurtosisProtocol]
		if !found {
			// Should never happen because we enforce completeness via unit test
			return nil, stacktrace.NewError("No Kubernetes port protocol was defined for Kurtosis port protocol '%v'; this is a bug in Kurtosis", kurtosisProtocol)
		}

		kubernetesPortObj := apiv1.ContainerPort{
			Name:          portId,
			// Safe to do because max uint16 < int32
			ContainerPort: int32(portSpec.GetNumber()),
			Protocol:      kubernetesProtocol,
		}
		result = append(result, kubernetesPortObj)
	}
	return result, nil
}

func convertMegabytesToBytes(value uint64) uint64 {
	return value * megabytesToBytesFactor
}