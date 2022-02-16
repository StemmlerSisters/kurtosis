package kurtosis_backend_core

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/kubernetes_manager"
	"github.com/kurtosis-tech/object-attributes-schema-lib/schema"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"net"
	"strconv"
)

const (
	kurtosisEngineNamespace   = "kurtosis-namespace"
	kurtosisEngineReplicas    = 1
	storageClass              = "standard"
	defaultQuantity           = "10Gi"
	defaultHostPathInMinikube = "/host/data/engine-data"
	externalServiceType       = "LoadBalancer"
	zeroReplicas              = 0
)

type KurtosisKubernetesBackendCore struct {
	log *logrus.Logger

	kubernetesManager *kubernetes_manager.KubernetesManager

	objAttrsProvider schema.ObjectAttributesProvider
}

func NewKurtosisKubernetesBackendCore(log *logrus.Logger, k8sManager *kubernetes_manager.KubernetesManager, objAttrsProvider schema.ObjectAttributesProvider) *KurtosisKubernetesBackendCore {
	return &KurtosisKubernetesBackendCore{
		log: log,

		kubernetesManager: k8sManager,
		objAttrsProvider:  objAttrsProvider,
	}
}

func (backendCore KurtosisKubernetesBackendCore) CreateEngine(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	listenPortNum uint16,
	engineDataDirpathOnHostMachine string,
	imageOrgAndRepo string,
	serializedEnvVars map[string]string,
) (
	resultPublicIpAddr net.IP,
	resultPublicPortNum uint16,
	resultErr error,
) {
	// getting the object attributes for the engine server
	engineAttrs, err := backendCore.objAttrsProvider.ForEngineServer(listenPortNum) // TODO we should probably create a new function for labels that make sense for kubernetes deployment
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred getting the engine server container attributes using port num '%v'", listenPortNum)
	}

	// getting the object attributes for the engine server
	engineAttrsForPod, err := backendCore.objAttrsProvider.ForEngineServer(listenPortNum) // TODO we should probably create a new function for labels that make sense for kubernetes pod
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred getting the engine server container attributes using port num '%v'", listenPortNum)
	}

	engineDataDirpathOnHostMachine = defaultHostPathInMinikube

	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred creating the engine server args")
	}

	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		imageOrgAndRepo,
		imageVersionTag,
	)

	// checking if the kurtosis namespace already exists and creating it otherwise
	kurtosisNamespaceList, err := backendCore.kubernetesManager.GetNamespacesByLabels(ctx, engineLabels)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred when trying to get the kurtosis engine namespace by labels '%+v'",
			engineAttrs.GetLabels())
	}
	if len(kurtosisNamespaceList.Items) == 0 {
		_, err = backendCore.kubernetesManager.CreateNamespace(ctx, kurtosisEngineNamespace, engineLabels)
		if err != nil {
			return nil, 0, stacktrace.Propagate(
				err,
				"An error occurred when trying to create the kurtosis engine namespace to be named '%v'",
				kurtosisEngineNamespace)
		}
	}

	// creating persistent volume
	_, err = backendCore.kubernetesManager.CreatePersistentVolume(ctx, engineAttrs.GetName(), engineAttrs.GetLabels(), defaultQuantity, engineDataDirpathOnHostMachine, storageClass)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred when trying to create the persistent volume to be named '%v'",
			engineAttrs.GetName())
	}

	// creating persistent volume claim
	_, err = backendCore.kubernetesManager.CreatePersistentVolumeClaim(ctx, kurtosisEngineNamespace, engineAttrs.GetName(), engineAttrs.GetLabels(), defaultQuantity, storageClass)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred when trying to create the persistent volume claim to be named '%v'",
			engineAttrs.GetName())
	}

	// defining the volumes for the deployment
	volumes := []apiv1.Volume{
		{
			Name: engineAttrs.GetName(),
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: engineAttrs.GetName(),
				},
			},
		},
	}

	volumeMounts := []apiv1.VolumeMount{
		{
			Name:      engineAttrs.GetName(),
			MountPath: EngineDataDirpathOnEngineServerContainer,
		},
	}

	// creating deployment
	_, err = backendCore.kubernetesManager.CreateDeployment(ctx, engineAttrs.GetName(), kurtosisEngineNamespace, engineAttrs.GetLabels(), engineAttrsForPod.GetLabels(), containerImageAndTag, kurtosisEngineReplicas, volumes, volumeMounts, serializedEnvVars, engineAttrs.GetName())
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred generating the engine server's environment variables")
	}

	// creating service
	service, err := backendCore.kubernetesManager.CreateService(ctx, engineAttrs.GetName(), kurtosisEngineNamespace, engineAttrs.GetLabels(), externalServiceType, int32(listenPortNum), int32(listenPortNum))
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred generating the engine server's environment variables")
	}

	publicIpAddr := net.ParseIP(service.Spec.ClusterIP)

	publicPortNumStr := string(service.Spec.Ports[0].NodePort)
	publicPortNumUint64, err := strconv.ParseUint(publicPortNumStr, publicPortNumParsingBase, publicPortNumParsingUintBits)
	if err != nil {
		return nil, 0, stacktrace.Propagate(
			err,
			"An error occurred parsing engine server public port string '%v' using base '%v' and uint bits '%v'",
			publicPortNumStr,
			publicPortNumParsingBase,
			publicPortNumParsingUintBits,
		)
	}
	publicPortNumUint16 := uint16(publicPortNumUint64) // Safe to do because we pass the requisite number of bits into the parse command

	return publicIpAddr, publicPortNumUint16, nil
}

func (backendCore KurtosisKubernetesBackendCore) StopEngine(ctx context.Context) error {
	err := backendCore.kubernetesManager.UpdateDeploymentReplicas(ctx, kurtosisEngineNamespace, engineLabels, int32(zeroReplicas))
	if err != nil {
		stacktrace.Propagate(err, "An error occurred while trying to stop the engine server with labels %v", engineLabels)
	}

	return nil
}

func (backendCore KurtosisKubernetesBackendCore) CleanStoppedEngines(ctx context.Context) ([]string, []error, error) {
	deploymentsList, err := backendCore.kubernetesManager.GetDeploymentsByLabels(ctx, kurtosisEngineNamespace, engineLabels)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred while trying to clean the the stopped engine containers with labels %+v", engineLabels)
	}

	successfullyDestroyedEngineNames := []string{}
	removeEngineErrors := []error{}

	if len(deploymentsList.Items) > 0 {
		for _, deployment := range deploymentsList.Items {
			if !backendCore.checkIfContainerisRunning(deployment) {
				err = backendCore.cleanEngineServer(ctx, deployment.Name)
				if err != nil {
					removeEngineErrors = append(removeEngineErrors, err)
				} else {
					successfullyDestroyedEngineNames = append(successfullyDestroyedEngineNames, deployment.Name)
				}
			}
		}
	}

	return successfullyDestroyedEngineNames, removeEngineErrors, nil
}

func (backendCore KurtosisKubernetesBackendCore) checkIfContainerisRunning(deployment v1.Deployment) bool {
	return *deployment.Spec.Replicas > 0
}

func (backendCore KurtosisKubernetesBackendCore) cleanEngineServer(ctx context.Context, name string) error {
	err := backendCore.kubernetesManager.RemoveDeployment(ctx, kurtosisEngineNamespace, name)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to delete the deployment from the engine with name %s", name)
	}

	err = backendCore.kubernetesManager.RemovePersistentVolumeClaim(ctx, kurtosisEngineNamespace, name)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to delete the persistent volume claim from the engine with name %s", name)
	}

	err = backendCore.kubernetesManager.RemovePersistentVolume(ctx, name)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to delete the persistent volume from the engine with name %s", name)
	}

	err = backendCore.kubernetesManager.RemoveService(ctx, kurtosisEngineNamespace, name)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while trying to delete the service from the engine with name %s", name)
	}

	return nil
}

func (backendCore KurtosisKubernetesBackendCore) GetEngineStatus(
	ctx context.Context,
) (engineStatus string, ipAddr net.IP, portNum uint16, err error) {
	deploymentList, err := backendCore.kubernetesManager.GetDeploymentsByLabels(ctx, kurtosisEngineNamespace, engineLabels)
	if err != nil {
		return "", nil, 0, stacktrace.Propagate(err, "An error occurred getting Kurtosis engine deployments")
	}

	var deployments []v1.Deployment
	for _, deployment := range deploymentList.Items {
		if *deployment.Spec.Replicas > 0 {
			deployments = append(deployments, deployment)
		}
	}

	numRunningEngines := len(deployments)
	if numRunningEngines > 1 {
		return "", nil, 0, stacktrace.NewError("Cannot report engine status because we found %v running Kurtosis engines; this is very strange as there should never be more than one", numRunningEngines)
	}

	if numRunningEngines == 0 {
		return EngineStatus_Stopped, nil, 0, nil
	}

	engineDeployment := deployments[0]

	service, err := backendCore.kubernetesManager.GetServiceByName(ctx, kurtosisEngineNamespace, engineDeployment.Name)

	publicIpAddr := net.ParseIP(service.Spec.ClusterIP)

	publicPortNumStr := string(service.Spec.Ports[0].NodePort)
	publicPortNumUint64, err := strconv.ParseUint(publicPortNumStr, publicPortNumParsingBase, publicPortNumParsingUintBits)
	if err != nil {
		return "", nil, 0, stacktrace.Propagate(
			err,
			"An error occurred parsing engine server public port string '%v' using base '%v' and uint bits '%v'",
			publicPortNumStr,
			publicPortNumParsingBase,
			publicPortNumParsingUintBits,
		)
	}
	publicPortNumUint16 := uint16(publicPortNumUint64) // Safe to do because we pass the requisite number of bits into the parse command

	return "", publicIpAddr, publicPortNumUint16, nil
}
