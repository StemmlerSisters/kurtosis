/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package kubernetes_manager

import (
	"context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

const (
	protocolTCP = "TCP"
)

var (
	removeServiceDeletePolicy  = metav1.DeletePropagationForeground
	removeServiceDeleteOptions = metav1.DeleteOptions{
		PropagationPolicy: &removeServiceDeletePolicy,
	}
)

type KubernetesManager struct {
	// The logger that all log messages will be written to
	log *logrus.Logger // NOTE: This log should be used for all log statements - the system-wide logger should NOT be used!

	// The underlying K8s client that will be used to modify the K8s environment
	kubernetesClientSet *kubernetes.Clientset
}

/*
NewKubernetesManager
Creates a new K8s manager for manipulating the k8s cluster using the given client.

Args:
	log: The logger that this K8s manager will write all its log messages to.
	kubernetesClientSet: The k8s client that will be used when interacting with the underlying k8s cluster.
*/
func NewKubernetesManager(log *logrus.Logger, kubernetesClientSet *kubernetes.Clientset) *KubernetesManager {
	return &KubernetesManager{
		log:                 log,
		kubernetesClientSet: kubernetesClientSet,
	}
}

// ---------------------------Deployments------------------------------------------------------------------------------

/*
CreateDeployment
Creates a new k8s deployment with the given parameters

Args:


Returns:
	id: The deployment ID
*/
func (manager *KubernetesManager) CreateDeployment(ctx context.Context, deploymentName string, namespace string, labels map[string]string, containerImage string, replicas int32, volumes []apiv1.Volume, volumeMounts []apiv1.VolumeMount, envVars map[string]string, containerName string) (*appsv1.Deployment, error) {
	deploymentsClient := manager.kubernetesClientSet.AppsV1().Deployments(namespace)

	var podEnvVars []apiv1.EnvVar

	for k, v := range envVars {
		envVar := apiv1.EnvVar{
			Name:  k,
			Value: v,
		}
		podEnvVars = append(podEnvVars, envVar)
	}

	objectMeta := metav1.ObjectMeta{
		Name:   deploymentName,
		Labels: labels,
	}

	selector := &metav1.LabelSelector{
		MatchLabels: labels,
	}

	containers := []apiv1.Container{
		{
			Name:         containerName,
			Image:        containerImage,
			VolumeMounts: volumeMounts,
			Env:          podEnvVars,
		},
	}

	podSpec := apiv1.PodSpec{
		Volumes:    volumes,
		Containers: containers,
	}

	podTemplateSpec := apiv1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: podSpec,
	}

	deploymentSpec := appsv1.DeploymentSpec{
		Replicas: manager.int32Ptr(replicas),
		Selector: selector,
		Template: podTemplateSpec,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: objectMeta,
		Spec:       deploymentSpec,
	}

	deploymentResult, err := deploymentsClient.Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create deployment '%s' in namespace '%s'", deploymentName, namespace)
	}

	return deploymentResult, nil
}

func (manager *KubernetesManager) ListDeployments(ctx context.Context, namespace string) (*appsv1.DeploymentList, error) {
	deploymentsClient := manager.kubernetesClientSet.AppsV1().Deployments(namespace)

	list, err := deploymentsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list deployments for namespace '%s'", namespace)
	}

	return list, nil
}

func (manager *KubernetesManager) RemoveDeployment(ctx context.Context, namespace string, name string) error {
	deploymentsClient := manager.kubernetesClientSet.AppsV1().Deployments(namespace)

	if err := deploymentsClient.Delete(ctx, name, removeServiceDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete deployment '%s' with delete options '%v' and for namespace '%s'", name, removeServiceDeleteOptions, namespace)
	}
	return nil
}

func (manager *KubernetesManager) GetDeploymentByName(ctx context.Context, namespace string, name string) (*appsv1.Deployment, error) {
	deploymentsClient := manager.kubernetesClientSet.AppsV1().Deployments(namespace)

	deployment, err := deploymentsClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get deployment '%s' with namespace '%s'", name, namespace)
	}

	return deployment, nil
}

func (manager *KubernetesManager) GetDeploymentsByLabels(ctx context.Context, namespace string, deploymentLabels map[string]string) (*appsv1.DeploymentList, error) {
	deploymentsClient := manager.kubernetesClientSet.AppsV1().Deployments(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(deploymentLabels).String(),
	}

	deployments, err := deploymentsClient.List(ctx, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get deployment by labels '%+v' in namespace '%s'", deploymentLabels, namespace)
	}

	return deployments, nil
}

func (manager *KubernetesManager) UpdateDeploymentReplicas(ctx context.Context, namespace string, deploymentLabels map[string]string, replicas int32) error {
	deploymentsClient := manager.kubernetesClientSet.AppsV1().Deployments(apiv1.NamespaceDefault)

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, err := manager.GetDeploymentsByLabels(ctx, namespace, deploymentLabels)
		if err != nil {
			return stacktrace.Propagate(err, "Failed to get deployments by labels %v", deploymentLabels)
		}

		for _, deployment := range result.Items {
			deployment.Spec.Replicas = manager.int32Ptr(replicas)
			_, err = deploymentsClient.Update(ctx, &deployment, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
		}
		return err
	})
	if retryErr != nil {
		stacktrace.Propagate(retryErr, "Failed to update deployment replicas by labels '%+v' and namespace '%s'", deploymentLabels, namespace)
	}

	return nil
}

// ---------------------------Services------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateService(ctx context.Context, name string, namespace string, serviceLabels map[string]string, serviceType apiv1.ServiceType, port int32, targetPort int32) (*apiv1.Service, error) {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	objectMeta := metav1.ObjectMeta{
		Name:   name,
		Labels: serviceLabels,
	}

	ports := []apiv1.ServicePort{
		{
			Protocol: protocolTCP,
			Port:     port,
			TargetPort: intstr.IntOrString{
				IntVal: targetPort, // internal container port
			},
		},
	}

	serviceSpec := apiv1.ServiceSpec{
		Ports:    ports,
		Selector: serviceLabels, // these labels are used to match with the Pod
		Type:     serviceType,
	}

	service := &apiv1.Service{
		ObjectMeta: objectMeta,
		Spec:       serviceSpec,
	}

	serviceResult, err := servicesClient.Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create service '%s' in namespace '%s'", name, namespace)
	}

	return serviceResult, nil
}

func (manager *KubernetesManager) RemoveService(ctx context.Context, namespace string, name string) error {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	if err := servicesClient.Delete(ctx, name, removeServiceDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete service '%s' with delete options '%v' and namespace '%s'", name, removeServiceDeleteOptions, namespace)
	}

	return nil
}

func (manager *KubernetesManager) GetServiceByName(ctx context.Context, namespace string, name string) (*apiv1.Service, error) {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	serviceResult, err := servicesClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get service '%s' for namespace '%s'", name, namespace)
	}

	return serviceResult, nil
}

func (manager *KubernetesManager) ListServices(ctx context.Context, namespace string) (*apiv1.ServiceList, error) {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	serviceResult, err := servicesClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list services for namespace '%s'", namespace)
	}

	return serviceResult, nil
}

func (manager *KubernetesManager) GetServicesByLabels(ctx context.Context, serviceLabels map[string]string, namespace string) (*apiv1.ServiceList, error) {
	servicesClient := manager.kubernetesClientSet.CoreV1().Services(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(serviceLabels).String(),
	}

	serviceResult, err := servicesClient.List(ctx, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list services with labels '%+v' for namespace '%s'", serviceLabels, namespace)
	}

	return serviceResult, nil
}

// ---------------------------Volumes------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateStorageClass(ctx context.Context, name string, provisioner string, volumeBindingMode storagev1.VolumeBindingMode) (*storagev1.StorageClass, error) {
	storageClassClient := manager.kubernetesClientSet.StorageV1().StorageClasses()

	//volumeBindingMode := storagev1.VolumeBindingWaitForFirstConsumer
	//provisioner := "kubernetes.io/no-provisioner"

	storageClass := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Provisioner:       provisioner,
		VolumeBindingMode: &volumeBindingMode,
	}

	storageClassResult, err := storageClassClient.Create(ctx, storageClass, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create storage class with name '%s'", name)
	}

	return storageClassResult, nil
}

func (manager *KubernetesManager) RemoveStorageClass(ctx context.Context, name string) error {
	storageClassClient := manager.kubernetesClientSet.StorageV1().StorageClasses()

	err := storageClassClient.Delete(ctx, name, removeServiceDeleteOptions)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to delete storage class with name '%s' with delete options '%v'", name, removeServiceDeleteOptions)
	}

	return nil
}

func (manager *KubernetesManager) GetStorageClass(ctx context.Context, name string) (*storagev1.StorageClass, error) {
	storageClassClient := manager.kubernetesClientSet.StorageV1().StorageClasses()

	storageClassResult, err := storageClassClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get storage class with name '%s'", name)
	}

	return storageClassResult, nil
}

func (manager *KubernetesManager) CreatePersistentVolume(ctx context.Context, volumeName string, volumeLabels map[string]string, quantity string, path string, storageClassName string) (*apiv1.PersistentVolume, error) {
	volumesClient := manager.kubernetesClientSet.CoreV1().PersistentVolumes()

	//quantity := "100Gi"
	//storageClassName := "my-local-storage"
	//path := "/Users/mariofernandez/Library/Application Support/kurtosis/engine-data"

	persistentVolume := &apiv1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:   volumeName,
			Labels: volumeLabels,
		},
		Spec: apiv1.PersistentVolumeSpec{
			Capacity: map[apiv1.ResourceName]resource.Quantity{
				apiv1.ResourceStorage: resource.MustParse(quantity),
			},
			PersistentVolumeSource: apiv1.PersistentVolumeSource{
				HostPath: &apiv1.HostPathVolumeSource{
					Path: path,
				},
			},
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce,
			},
			StorageClassName: storageClassName,
		},
	}

	persistentVolumeResult, err := volumesClient.Create(ctx, persistentVolume, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create persistent volume with name '%s'", volumeName)
	}

	return persistentVolumeResult, nil
}

func (manager *KubernetesManager) RemovePersistentVolume(ctx context.Context, volumeName string) error {
	volumesClient := manager.kubernetesClientSet.CoreV1().PersistentVolumes()

	if err := volumesClient.Delete(ctx, volumeName, removeServiceDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to create persistent volume with name '%s' and deleteOptions '%v'", volumeName, removeServiceDeleteOptions)
	}

	return nil
}

func (manager *KubernetesManager) GetPersistentVolume(ctx context.Context, volumeName string) (*apiv1.PersistentVolume, error) {
	volumesClient := manager.kubernetesClientSet.CoreV1().PersistentVolumes()

	persistentVolumeResult, err := volumesClient.Get(ctx, volumeName, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create persistent volume with name '%s'", volumeName)
	}

	return persistentVolumeResult, nil
}

func (manager *KubernetesManager) ListPersistentVolumes(ctx context.Context, volumeName string) (*apiv1.PersistentVolumeList, error) {
	volumesClient := manager.kubernetesClientSet.CoreV1().PersistentVolumes()

	persistentVolumesResult, err := volumesClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list persistent volumes")
	}

	return persistentVolumesResult, nil
}

func (manager *KubernetesManager) GetPersistentVolumesByLabels(ctx context.Context, persistentVolumeLabels map[string]string) (*apiv1.PersistentVolumeList, error) {
	volumesClient := manager.kubernetesClientSet.CoreV1().PersistentVolumes()

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(persistentVolumeLabels).String(),
	}

	persistentVolumesResult, err := volumesClient.List(ctx, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get persistent volumes by labels '%+v'", persistentVolumeLabels)
	}

	return persistentVolumesResult, nil
}

func (manager *KubernetesManager) CreatePersistentVolumeClaim(ctx context.Context, namespace string, persistentVolumeClaimName string, persistentVolumeClaimLabels map[string]string, quantity string, storageClassName string) (*apiv1.PersistentVolumeClaim, error) {
	volumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	//storageClassName := "my-local-storage"
	//quantity := "10Gi"

	persistentVolumeClaim := &apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   persistentVolumeClaimName,
			Labels: persistentVolumeClaimLabels,
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce,
			},
			StorageClassName: &storageClassName,
			Resources: apiv1.ResourceRequirements{
				Requests: map[apiv1.ResourceName]resource.Quantity{
					apiv1.ResourceStorage: resource.MustParse(quantity),
				},
			},
		},
	}

	persistentVolumeClaimResult, err := volumeClaimsClient.Create(ctx, persistentVolumeClaim, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create persistent volume claim with name '%s' with namespace '%s'", persistentVolumeClaimName, namespace)
	}

	return persistentVolumeClaimResult, nil
}

func (manager *KubernetesManager) RemovePersistentVolumeClaim(ctx context.Context, namespace string, persistentVolumeClaimName string) error {
	volumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	if err := volumeClaimsClient.Delete(ctx, persistentVolumeClaimName, removeServiceDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete persistent volume claim with name '%s' with delete options '%v' and namespace '%s'", persistentVolumeClaimName, removeServiceDeleteOptions, namespace)
	}

	return nil
}

func (manager *KubernetesManager) GetPersistentVolumeClaim(ctx context.Context, namespace string, persistentVolumeClaimName string) (*apiv1.PersistentVolumeClaim, error) {
	persistentVolumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	volumeClaim, err := persistentVolumeClaimsClient.Get(ctx, persistentVolumeClaimName, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get persistent volume claim with name '%s' and with namespace '%s'", persistentVolumeClaimName, namespace)
	}

	return volumeClaim, nil
}

func (manager *KubernetesManager) ListPersistentVolumeClaims(ctx context.Context, namespace string) (*apiv1.PersistentVolumeClaimList, error) {
	persistentVolumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	persistentVolumeClaimsResult, err := persistentVolumeClaimsClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list persistent volume claims for namespace '%s'", namespace)
	}

	return persistentVolumeClaimsResult, nil
}

func (manager *KubernetesManager) GetPersistentVolumeClaimsByLabels(ctx context.Context, namespace string, persistentVolumeClaimLabels map[string]string) (*apiv1.PersistentVolumeClaimList, error) {
	persistentVolumeClaimsClient := manager.kubernetesClientSet.CoreV1().PersistentVolumeClaims(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(persistentVolumeClaimLabels).String(),
	}

	persistentVolumeClaimsResult, err := persistentVolumeClaimsClient.List(ctx, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to Get persistent volume claim with labels '%+v' and namespace '%s'", persistentVolumeClaimLabels, namespace)
	}

	return persistentVolumeClaimsResult, nil
}

// ---------------------------namespaces------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateNamespace(ctx context.Context, name string, namespaceLabels map[string]string) (*apiv1.Namespace, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	namespace := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: namespaceLabels,
		},
	}

	namespaceResult, err := namespaceClient.Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create namespace with name '%s'", name)
	}

	return namespaceResult, nil
}

func (manager *KubernetesManager) RemoveNamespace(ctx context.Context, name string) error {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	if err := namespaceClient.Delete(ctx, name, removeServiceDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete namespace with name '%s' with delete options '%v'", name, removeServiceDeleteOptions)
	}

	return nil
}

func (manager *KubernetesManager) GetNamespace(ctx context.Context, name string) (*apiv1.Namespace, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	namespace, err := namespaceClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get namespace with name '%s'", name)
	}

	return namespace, nil
}

func (manager *KubernetesManager) ListNamespaces(ctx context.Context) (*apiv1.NamespaceList, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	namespaces, err := namespaceClient.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list namespaces")
	}

	return namespaces, nil
}

func (manager *KubernetesManager) GetNamespacesByLabels(ctx context.Context, namespaceLabels map[string]string) (*apiv1.NamespaceList, error) {
	namespaceClient := manager.kubernetesClientSet.CoreV1().Namespaces()

	listOptions := metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(namespaceLabels).String(),
	}

	namespaces, err := namespaceClient.List(ctx, listOptions)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to list namespaces with labels '%+v'", namespaceLabels)
	}

	return namespaces, nil
}

// ---------------------------DaemonSets------------------------------------------------------------------------------

func (manager *KubernetesManager) CreateDaemonSet(ctx context.Context, namespace string, name string, daemonSetLabels map[string]string) (*appsv1.DaemonSet, error) {
	daemonSetClient := manager.kubernetesClientSet.AppsV1().DaemonSets(namespace)

	daemonSet := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: daemonSetLabels,
		},
		Spec: appsv1.DaemonSetSpec{},
	}

	daemonSetResult, err := daemonSetClient.Create(ctx, daemonSet, metav1.CreateOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create daemonSet with name '%s' and namespace '%s'", name, namespace)
	}

	return daemonSetResult, nil
}

func (manager *KubernetesManager) RemoveDaemonSet(ctx context.Context, name string, namespace string) error {

	daemonSetClient := manager.kubernetesClientSet.AppsV1().DaemonSets(namespace)
	if err := daemonSetClient.Delete(ctx, name, removeServiceDeleteOptions); err != nil {
		return stacktrace.Propagate(err, "Failed to delete daemonSet with name '%s' with delete options '%v'", name, removeServiceDeleteOptions)
	}

	return nil
}

func (manager *KubernetesManager) GetDaemonSet(ctx context.Context, name string, namespace string) (*appsv1.DaemonSet, error) {
	daemonSetClient := manager.kubernetesClientSet.AppsV1().DaemonSets(namespace)

	daemonSet, err := daemonSetClient.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get daemonSet with name '%s' in namespace '%s'", name, namespace)
	}

	return daemonSet, nil
}

// Private functions
func (manager *KubernetesManager) int32Ptr(i int32) *int32 { return &i }
