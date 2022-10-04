package validator_state

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
)

type StartosisValidatorState struct {
	requiredDockerImages map[string]bool
	kurtosisBackend      *backend_interface.KurtosisBackend
}

func NewStartosisValidatorState(kurtosisBackend *backend_interface.KurtosisBackend) *StartosisValidatorState {
	return &StartosisValidatorState{
		map[string]bool{},
		kurtosisBackend,
	}
}

func (validatorState *StartosisValidatorState) AppendRequiredDockerImage(dockerImage string) {
	validatorState.requiredDockerImages[dockerImage] = true
}

func (validatorState *StartosisValidatorState) Validate(ctx context.Context) error {
	err := validatorState.validateDockerImages(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed while validating images")
	}
	return nil
}

func (validatorState *StartosisValidatorState) validateDockerImages(ctx context.Context) error {
	// TODO(victor.colombo): Parallelize pull image calls
	for image := range validatorState.requiredDockerImages {
		err := (*validatorState.kurtosisBackend).PullImage(ctx, image)
		if err != nil {
			return stacktrace.Propagate(err, "Failed fetching the required image %v", image)
		}
	}
	return nil
}
