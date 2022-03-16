package label_value_consts

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
)

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// If these value change, it will lead to the Kurtosis engine losing track of old containers
	//  which will cause a resource leak on the user's system!
	//
	//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	//
	appIdLabelValueStr               = "kurtosis"
	engineContainerTypeLabelValueStr = "kurtosis-engine"
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	apiContainerContainerTypeLabelValueStr = "api-container"
	userServiceContainerTypeLabelValueStr = "user-service"
	networkingSidecarContainerTypeLabelValueStr = "networking-sidecar"

	trueValueStr = "true"

	falseValueStr = "false"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old containers
//  which will cause a resource leak on the user's system!
//
//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
//
var AppIDLabelValue = docker_label_value.MustCreateNewDockerLabelValue(appIdLabelValueStr)
var EngineContainerTypeLabelValue = docker_label_value.MustCreateNewDockerLabelValue(engineContainerTypeLabelValueStr)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

var APIContainerContainerTypeLabelValue = docker_label_value.MustCreateNewDockerLabelValue(apiContainerContainerTypeLabelValueStr)
var UserServiceContainerTypeLabelValue = docker_label_value.MustCreateNewDockerLabelValue(userServiceContainerTypeLabelValueStr)
var NetworkingSidecarContainerTypeLabelValue = docker_label_value.MustCreateNewDockerLabelValue(networkingSidecarContainerTypeLabelValueStr)
var NetworkPartitioningEnabledLabelValue = docker_label_value.MustCreateNewDockerLabelValue(trueValueStr)
var NetworkPartitioningDisabledLabelValue = docker_label_value.MustCreateNewDockerLabelValue(falseValueStr)
var NetworkDestroyed = docker_label_value.MustCreateNewDockerLabelValue(trueValueStr)
var NetworkIsNotDestroyed = docker_label_value.MustCreateNewDockerLabelValue(falseValueStr)
