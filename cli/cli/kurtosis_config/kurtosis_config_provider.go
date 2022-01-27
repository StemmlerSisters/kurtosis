package kurtosis_config

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type KurtosisConfigProvider struct {
	configStore *kurtosisConfigStore
}

func NewKurtosisConfigProvider(configStore *kurtosisConfigStore) *KurtosisConfigProvider {
	return &KurtosisConfigProvider{configStore: configStore}
}

func NewDefaultKurtosisConfigProvider() *KurtosisConfigProvider {
	configStore := GetKurtosisConfigStore()
	configProvider := NewKurtosisConfigProvider(configStore)
	return configProvider
}

func (configProvider *KurtosisConfigProvider) GetOrInitializeConfig() (*KurtosisConfig, error){

	var (
		kurtosisConfig *KurtosisConfig
		err            error
	)

	doesKurtosisConfigAlreadyExists, err := configProvider.configStore.HasConfig()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking if Kurtosis config already exists")
	}
	if doesKurtosisConfigAlreadyExists {
		kurtosisConfig, err = configProvider.configStore.GetConfig()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting config")
		}
	} else {
		kurtosisConfig, err = initInteractiveConfig()
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred executing init interactive config")
		}

		if err = configProvider.configStore.SetConfig(kurtosisConfig); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred setting Kurtosis config")
		}
	}
	logrus.Debugf("Loaded Kurtosis Config  %+v", kurtosisConfig)
	return kurtosisConfig, nil
}
