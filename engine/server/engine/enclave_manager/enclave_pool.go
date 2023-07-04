package enclave_manager

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

const (
	// This is the same default value used in the `kurtosis enclave add` CLI command
	defaultApiContainerLogLevel = logrus.DebugLevel

	// The partitioning feature is disabled by default because the Enclave Pool
	// is enabled only for K8s and the network partitioning feature is not implemented yet
	defaultIsPartitioningEnabled = false

	fill = true
)

type EnclavePool struct {
	kurtosisBackend         backend_interface.KurtosisBackend
	enclaveCreator          *EnclaveCreator
	idleEnclavesChan        chan enclave.EnclaveUUID
	fillChan                chan bool
	engineVersion           string
	cancelSubRoutineCtxFunc context.CancelFunc
}

// CreateEnclavePool will do the following:
// 1- Will remove idle enclaves from previous engine runs even if the pool is not activated (this is for removing
// any resource leak after an engine restar without this feature enabled or after an engine crash)
// 2- Wil create a new enclave pool object, if pool size > 1, return nil if pool size = 0, or return an error
// 3- Will start a subroutine in charge of filling the pool
func CreateEnclavePool(
	kurtosisBackend backend_interface.KurtosisBackend,
	kurtosisBackendType args.KurtosisBackendType,
	enclaveCreator *EnclaveCreator,
	poolSize uint8,
	engineVersion string,
) (*EnclavePool, error) {

	//TODO the current implementation only removes the previous idle enclave, it's pending to implement the reusable feature
	//TODO the reuse logic is not enable yet because we ned to store the APIC version on the APIContainer object in container-engine-lib
	//TODO in order to using it for comparing it with the expected version

	// iterate on all the existing enclaves in order to find idle enclaves already created
	// and reuse or destroy them if these were created from old Kurtosis version
	if err := destroyIdleEnclaves(kurtosisBackend); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred destroying previous idle enclave before creating the enclave pool")
	}

	// validations
	// poolSize = 0 means that the Enclave Pool won't be activated, it returns nil with no error
	if poolSize == 0 {
		return nil, nil
	}

	// this channel is the repository of idle enclave UUIDs
	idleEnclavesChan := make(chan enclave.EnclaveUUID, poolSize)

	// This channel is used as a signal to tell to the sub-routine that one idle enclave
	// has been allocated from the pool
	// It has the capacity = poolSize for not blocking the caller (for concurrent requests)
	fillChan := make(chan bool, poolSize)

	ctxWithCancel, cancelCtxFunc := context.WithCancel(context.Background())

	enclavePool := &EnclavePool{
		kurtosisBackend:         kurtosisBackend,
		enclaveCreator:          enclaveCreator,
		idleEnclavesChan:        idleEnclavesChan,
		fillChan:                fillChan,
		engineVersion:           engineVersion,
		cancelSubRoutineCtxFunc: cancelCtxFunc,
	}

	go enclavePool.run(ctxWithCancel)

	enclavePool.init(poolSize)

	return enclavePool, nil
}

// GetEnclave returns the first idle enclave from the pool, and the enclave is renamed with the
// name set by the caller before returning it. It returns nil if there is no enclave on the pool
// or if the requested enclave params are different from the enclave in the pool params
func (pool *EnclavePool) GetEnclave(
	ctx context.Context,
	newEnclaveName string,
	engineVersion string,
	apiContainerVersion string,
	apiContainerLogLevel logrus.Level,
	isPartitioningEnabled bool,
) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {

	logrus.Debugf(
		"Requesting enclave from pool using params: engine version '%s', api container version '%s', api container log level '%s' and is partitioning enabled '%v'...",
		engineVersion,
		apiContainerVersion,
		apiContainerLogLevel,
		isPartitioningEnabled,
	)

	// TODO change the logLevel value ?? it's pending to check if it's possible
	// The enclaves in the pool are already configured with defaults params and there is no way to update
	// this config, so we have to check if the requested enclave params are equal to the enclaves stored
	// in the pool before returning one
	if !areRequestedEnclaveParamsEqualToEnclaveInThePoolParams(
		engineVersion,
		apiContainerVersion,
		apiContainerLogLevel,
		isPartitioningEnabled,
	) {
		logrus.Debugf("The requested enclave params are different from the enclave in the pool params")
		return nil, nil
	}

	// If there is no idle enclave in the pool returns nil
	// for not to block the caller
	if len(pool.idleEnclavesChan) == 0 {
		return nil, nil
	}
	enclaveUUID, ok := <-pool.idleEnclavesChan
	if !ok {
		return nil, stacktrace.NewError("A new enclave can't be returned from the pool because the internal channel is closed, it shouldn't happen; this is a bug in Kurtosis")
	}
	// let the subroutine knows that one idle enclave has been taken from the pool,
	// and it has to fill the pool again
	pool.fillChan <- fill

	shouldDestroyEnclaveBecauseSomethingFails := true
	defer func() {
		if shouldDestroyEnclaveBecauseSomethingFails {
			idleEnclavesToRemove := map[enclave.EnclaveUUID]bool{
				enclaveUUID: true,
			}
			if err := destroyEnclavesByUUID(ctx, pool.kurtosisBackend, idleEnclavesToRemove); err != nil {
				logrus.Errorf(
					"Something fail while getting an enclave from the pool, we tried to destroy the "+
						"enclave that was taken from it, for avoiding a resource leak, but this also fails, "+
						"so you will have to manually destroy the enclave with UUUID '%v'. Error:\n%v", enclaveUUID, err)
			}
		}
	}()

	enclaveObj, err := pool.getRunningEnclave(ctx, enclaveUUID)
	if err != nil {
		logrus.Debugf("The idle enclave with UUID '%v' is not longer running or have been destroyed", enclaveUUID)
		return nil, stacktrace.Propagate(err, "An error occurred getting a running enclave with UUID '%v'", enclaveUUID)
	}

	newCreationTime := time.Now()

	if err := pool.kurtosisBackend.UpdateEnclave(ctx, enclaveUUID, newEnclaveName, &newCreationTime); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred updating enclave with UUID '%v', trying to update name to '%s' and creation time to '%v'", enclaveUUID, newEnclaveName, newCreationTime)
	}

	enclaveInfo, err := getEnclaveInfoForEnclave(ctx, pool.kurtosisBackend, enclaveObj)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclave info for enclave '%v'", enclaveObj)
	}
	enclaveInfo.Name = newEnclaveName

	logrus.Debugf("Returning enclave Info '%+v' for requested enclave name '%s'", enclaveInfo, newEnclaveName)

	shouldDestroyEnclaveBecauseSomethingFails = false
	return enclaveInfo, nil
}

// Close stop the EnclavePool subroutine, in charge of filling the pool,
// and removes all the idle enclaves already created
func (pool *EnclavePool) Close() error {

	defer close(pool.idleEnclavesChan)
	defer close(pool.fillChan)

	// will terminate running processes in the subroutine
	pool.cancelSubRoutineCtxFunc()

	// destroy all the idle enclaves
	if err := destroyIdleEnclaves(pool.kurtosisBackend); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying idle enclave")
	}

	return nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
func (pool *EnclavePool) init(poolSize uint8) {
	for i := uint8(0); i < poolSize; i++ {
		pool.fillChan <- fill
	}
}

// run is executed in a subroutine and wait for any of these two signals:
// 1- for creating and add a new idle enclave in the pool
// 2- for closing the subroutine
func (pool *EnclavePool) run(ctx context.Context) {

	for {
		// wait until receive the re-fill signal or the ctx has done signal
		select {
		case <-pool.fillChan:
			if err := pool.createAndAddOneIdleEnclaveIfNeeded(ctx); err != nil {
				if err == context.Canceled {
					logrus.Debug("The subroutine context has been canceled")
				} else {
					logrus.Errorf("An error occurred filling the enclave pool. Error\n%v", err)
				}
				break
			}
		case <-ctx.Done():
			logrus.Debug("The subroutine context has done")
			logrus.Debug("Enclave pool sub-routine stopped")
			return
		}
	}
}

func (pool *EnclavePool) createAndAddOneIdleEnclaveIfNeeded(ctx context.Context) error {

	newEnclaveInfo, err := pool.createNewIdleEnclave(ctx)
	if err != nil {
		if err == context.Canceled {
			return nil
		}
		return stacktrace.Propagate(err, "An error occurred creating a new idle enclave.")
	}

	enclaveUUID := enclave.EnclaveUUID(newEnclaveInfo.EnclaveUuid)
	pool.idleEnclavesChan <- enclaveUUID
	logrus.Debugf("Enclave with UUID '%v' was added intho the pool channel", enclaveUUID)

	return nil
}

func (pool *EnclavePool) createNewIdleEnclave(ctx context.Context) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {

	enclaveName, err := GetRandomIdleEnclaveName()
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred generating a random name for a new idle enclave.",
		)
	}

	apiContainerVersion := pool.engineVersion

	newEnclaveInfo, err := pool.enclaveCreator.CreateEnclave(
		ctx,
		apiContainerVersion,
		defaultApiContainerLogLevel,
		enclaveName,
		defaultIsPartitioningEnabled,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating idle enclave with name '%s', api container version '%v' and container log level '%s'",
			enclaveName,
			apiContainerVersion,
			defaultApiContainerLogLevel,
		)
	}

	logrus.Debugf("New idle enclave created '%+v'", newEnclaveInfo)
	return newEnclaveInfo, nil
}

func (pool *EnclavePool) getRunningEnclave(ctx context.Context, enclaveUUID enclave.EnclaveUUID) (*enclave.Enclave, error) {
	filters := &enclave.EnclaveFilters{
		UUIDs: map[enclave.EnclaveUUID]bool{
			enclaveUUID: true,
		},
		Statuses: map[enclave.EnclaveStatus]bool{
			enclave.EnclaveStatus_Running: true,
		},
	}

	enclaves, err := pool.kurtosisBackend.GetEnclaves(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclaves using filters '%+v'", filters)
	}
	enclavesLen := len(enclaves)

	if enclavesLen == 0 {
		return nil, stacktrace.NewError("There is not any running enclave with UUID '%v', it could have been stopped or destroyed", enclaveUUID)
	}
	if enclavesLen > 1 {
		return nil, stacktrace.NewError("Expected to find only one running enclave with UUID '%v', but '%v' were found, this is a bug in Kurtosis", enclaveUUID, enclavesLen)
	}

	enclaveObj, found := enclaves[enclaveUUID]
	if !found {
		return nil, stacktrace.NewError("Expected to find an enclave with UUID '%v' in enclave map '%+v', but none was found; this is a bug in Kurtosis", enclaveUUID, enclaves)
	}

	return enclaveObj, nil
}

func areRequestedEnclaveParamsEqualToEnclaveInThePoolParams(
	engineVersion string,
	apiContainerVersion string,
	apiContainerLogLevel logrus.Level,
	isPartitioningEnabled bool,
) bool {

	// if the api container version is empty string means that will be executed with the default version
	// which is the same that the current engine version
	if apiContainerVersion == "" {
		apiContainerVersion = engineVersion
	}

	if engineVersion == apiContainerVersion &&
		apiContainerLogLevel == defaultApiContainerLogLevel &&
		!isPartitioningEnabled {
		return true
	}

	return false
}

func destroyIdleEnclaves(kurtosisBackend backend_interface.KurtosisBackend) error {
	ctx := context.Background()

	filters := &enclave.EnclaveFilters{
		UUIDs: map[enclave.EnclaveUUID]bool{},
		Statuses: map[enclave.EnclaveStatus]bool{
			enclave.EnclaveStatus_Running: true,
		},
	}

	enclaves, err := kurtosisBackend.GetEnclaves(ctx, filters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting enclaves using filters '%+v'", filters)
	}

	idleEnclavesToRemove := map[enclave.EnclaveUUID]bool{}

	for enclaveUUID, enclaveObj := range enclaves {
		enclaveName := enclaveObj.GetName()
		// is it an idle enclave from a previous run?
		if strings.HasPrefix(enclaveName, idleEnclaveNamePrefix) {
			idleEnclavesToRemove[enclaveUUID] = true
		}
	}

	if err := destroyEnclavesByUUID(ctx, kurtosisBackend, idleEnclavesToRemove); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying enclaves with UUIDs '%v'", idleEnclavesToRemove)
	}

	return nil
}

func destroyEnclavesByUUID(
	ctx context.Context,
	kurtosisBackend backend_interface.KurtosisBackend,
	enclavesToRemove map[enclave.EnclaveUUID]bool,
) error {

	if len(enclavesToRemove) < 1 {
		return nil
	}

	destroyEnclaveFilters := &enclave.EnclaveFilters{
		UUIDs:    enclavesToRemove,
		Statuses: map[enclave.EnclaveStatus]bool{},
	}

	logrus.Debugf("Destroying enclaves '%+v'", enclavesToRemove)
	_, destroyEnclaveErrs, err := kurtosisBackend.DestroyEnclaves(ctx, destroyEnclaveFilters)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying enclaves using filters '%v'", destroyEnclaveFilters)
	}
	if len(destroyEnclaveErrs) > 0 {
		logrus.Errorf("Errors occurred removing the following enclaves...")
		var removalErrorStrings []string
		for enclaveUUID, destroyEnclaveErr := range destroyEnclaveErrs {
			logrus.Errorf("Enclave wit UUID '%v' error:\n %v", enclaveUUID, destroyEnclaveErr.Error())
			resultErrStr := fmt.Sprintf(">>>>>>>>>>>>>>>>> ERROR on Enclave %v <<<<<<<<<<<<<<<<<\n%v", enclaveUUID, destroyEnclaveErr.Error())
			removalErrorStrings = append(removalErrorStrings, resultErrStr)
		}
		logrus.Errorf("...you should have to manually remove all these errored enclaves.")
		errorSeparator := "\n"
		joinedRemovalErrors := strings.Join(removalErrorStrings, errorSeparator)
		return stacktrace.NewError("Following errors occurred while removing idle enclaves :\n%v", joinedRemovalErrors)
	}
	return nil
}
