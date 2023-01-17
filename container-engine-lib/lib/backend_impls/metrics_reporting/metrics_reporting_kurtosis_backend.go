package metrics_reporting

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/exec_result"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/networking_sidecar"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"net"
)

// TODO CALL THE METRICS LIBRARY EVENT-REGISTRATION FUNCTIONS HERE!!!!
type MetricsReportingKurtosisBackend struct {
	underlying backend_interface.KurtosisBackend
}

func NewMetricsReportingKurtosisBackend(underlying backend_interface.KurtosisBackend) *MetricsReportingKurtosisBackend {
	return &MetricsReportingKurtosisBackend{underlying: underlying}
}

func (backend *MetricsReportingKurtosisBackend) FetchImage(ctx context.Context, image string) error {
	if err := backend.underlying.FetchImage(ctx, image); err != nil {
		return stacktrace.Propagate(err, "An error occurred pulling image '%v'", image)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) CreateEngine(
	ctx context.Context,
	imageOrgAndRepo string,
	imageVersionTag string,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	envVars map[string]string,
) (*engine.Engine, error) {
	result, err := backend.underlying.CreateEngine(
		ctx,
		imageOrgAndRepo,
		imageVersionTag,
		grpcPortNum,
		grpcProxyPortNum,
		envVars,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine using image '%v' with tag '%v'", imageOrgAndRepo, imageVersionTag)
	}
	return result, nil
}

// Gets point-in-time data about engines matching the given filters
func (backend *MetricsReportingKurtosisBackend) GetEngines(ctx context.Context, filters *engine.EngineFilters) (map[engine.EngineGUID]*engine.Engine, error) {
	engines, err := backend.underlying.GetEngines(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting engines using filters: %+v", filters)
	}
	return engines, nil
}

func (backend *MetricsReportingKurtosisBackend) StopEngines(ctx context.Context, filters *engine.EngineFilters) (
	successfulIds map[engine.EngineGUID]bool,
	failedIds map[engine.EngineGUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.StopEngines(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping engines using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyEngines(ctx context.Context, filters *engine.EngineFilters) (
	successfulIds map[engine.EngineGUID]bool,
	failedIds map[engine.EngineGUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.DestroyEngines(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying engines using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateEnclave(ctx context.Context, enclaveUuid enclave.EnclaveUUID, enclaveName string, isPartitioningEnabled bool) (*enclave.Enclave, error) {
	result, err := backend.underlying.CreateEnclave(ctx, enclaveUuid, enclaveName, isPartitioningEnabled)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating enclave with UUID '%v' and is-partitioning-enabled value '%v'", enclaveUuid, isPartitioningEnabled)
	}
	return result, nil
}

func (backend *MetricsReportingKurtosisBackend) GetEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	map[enclave.EnclaveUUID]*enclave.Enclave,
	error,
) {
	results, err := backend.underlying.GetEnclaves(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting enclaves using filters: %+v", filters)
	}
	return results, nil
}

func (backend *MetricsReportingKurtosisBackend) StopEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[enclave.EnclaveUUID]bool,
	erroredEnclaveIds map[enclave.EnclaveUUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.StopEnclaves(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping enclaves using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DumpEnclave(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	outputDirpath string,
) error {
	if err := backend.underlying.DumpEnclave(ctx, enclaveId, outputDirpath); err != nil {
		return stacktrace.Propagate(err, "An error occurred dumping enclave '%v' to path '%v'", enclaveId, outputDirpath)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyEnclaves(
	ctx context.Context,
	filters *enclave.EnclaveFilters,
) (
	successfulEnclaveIds map[enclave.EnclaveUUID]bool,
	erroredEnclaveIds map[enclave.EnclaveUUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.DestroyEnclaves(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying enclaves using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateAPIContainer(
	ctx context.Context,
	image string,
	enclaveId enclave.EnclaveUUID,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	enclaveDataVolumeDirpath string,
	ownIpEnvVar string,
	customEnvVars map[string]string,
) (*api_container.APIContainer, error) {
	if _, found := customEnvVars[ownIpEnvVar]; found {
		return nil, stacktrace.NewError("Requested own IP environment variable '%v' conflicts with custom environment variable", ownIpEnvVar)
	}

	result, err := backend.underlying.CreateAPIContainer(
		ctx,
		image,
		enclaveId,
		grpcPortNum,
		grpcProxyPortNum,
		enclaveDataVolumeDirpath,
		ownIpEnvVar,
		customEnvVars,
	)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating an API container from image '%v' with envvars: %+v",
			image,
			customEnvVars,
		)
	}
	return result, nil
}

func (backend *MetricsReportingKurtosisBackend) GetAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (map[enclave.EnclaveUUID]*api_container.APIContainer, error) {
	results, err := backend.underlying.GetAPIContainers(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting API containers matching filters: %+v", filters)
	}
	return results, nil
}

func (backend *MetricsReportingKurtosisBackend) StopAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveUUID]bool, erroredApiContainerIds map[enclave.EnclaveUUID]error, resultErr error) {
	successes, failures, err := backend.underlying.StopAPIContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping API containers using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyAPIContainers(ctx context.Context, filters *api_container.APIContainerFilters) (successfulApiContainerIds map[enclave.EnclaveUUID]bool, erroredApiContainerIds map[enclave.EnclaveUUID]error, resultErr error) {
	successes, failures, err := backend.underlying.DestroyAPIContainers(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying API containers using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) RegisterUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceID]bool) (map[service.ServiceID]*service.ServiceRegistration, map[service.ServiceID]error, error) {
	successes, failures, err := backend.underlying.RegisterUserServices(ctx, enclaveUuid, services)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred registering services to enclave '%v' with the following service ids: %+v", enclaveUuid, services)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) UnregisterUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceGUID]bool) (map[service.ServiceGUID]bool, map[service.ServiceGUID]error, error) {
	successes, failures, err := backend.underlying.UnregisterUserServices(ctx, enclaveUuid, services)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred unregistering services from enclave '%v' with the following service guids: %+v", enclaveUuid, services)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) StartRegisteredUserServices(ctx context.Context, enclaveUuid enclave.EnclaveUUID, services map[service.ServiceGUID]*service.ServiceConfig) (map[service.ServiceGUID]*service.Service, map[service.ServiceGUID]error, error) {
	successes, failures, err := backend.underlying.StartRegisteredUserServices(ctx, enclaveUuid, services)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred starting services in enclave '%v' with the following service ids: %+v", enclaveUuid, services)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) GetUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	filters *service.ServiceFilters,
) (
	map[service.ServiceGUID]*service.Service,
	error,
) {
	services, err := backend.underlying.GetUserServices(ctx, enclaveId, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user services in enclave '%v' using filters '%+v'", enclaveId, filters)
	}
	return services, nil
}

func (backend *MetricsReportingKurtosisBackend) GetUserServiceLogs(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	filters *service.ServiceFilters,
	shouldFollowLogs bool,
) (
	map[service.ServiceGUID]io.ReadCloser,
	map[service.ServiceGUID]error,
	error,
) {
	userServiceLogs, erroredUserServices, err := backend.underlying.GetUserServiceLogs(ctx, enclaveId, filters, shouldFollowLogs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred getting user service logs in enclave '%v' using filters '%+v'", enclaveId, filters)
	}
	return userServiceLogs, erroredUserServices, nil
}

func (backend *MetricsReportingKurtosisBackend) PauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceId service.ServiceGUID,
) error {
	err := backend.underlying.PauseService(ctx, enclaveId, serviceId)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to pause service '%v' in enclave '%v'", serviceId, enclaveId)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) UnpauseService(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceId service.ServiceGUID,
) error {
	err := backend.underlying.UnpauseService(ctx, enclaveId, serviceId)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to unpause service '%v' in enclave '%v'", serviceId, enclaveId)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) RunUserServiceExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	userServiceCommands map[service.ServiceGUID][]string,
) (
	succesfulUserServiceExecResults map[service.ServiceGUID]*exec_result.ExecResult,
	erroredUserServiceGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	succesfulUserServiceExecResults, erroredUserServiceGuids, err := backend.underlying.RunUserServiceExecCommands(ctx, enclaveId, userServiceCommands)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred running user service exec commands '%+v' on enclave '%v'",
			userServiceCommands,
			enclaveId,
		)
	}
	return succesfulUserServiceExecResults, erroredUserServiceGuids, nil
}

func (backend *MetricsReportingKurtosisBackend) GetConnectionWithUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceGUID service.ServiceGUID,
) (
	resultConn net.Conn,
	resultErr error,
) {
	newConn, err := backend.underlying.GetConnectionWithUserService(ctx, enclaveId, serviceGUID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting connection with user service with GUID '%v'", serviceGUID)
	}
	return newConn, nil
}

func (backend *MetricsReportingKurtosisBackend) CopyFilesFromUserService(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceGuid service.ServiceGUID,
	srcPath string,
	output io.Writer,
) error {
	if err := backend.underlying.CopyFilesFromUserService(ctx, enclaveId, serviceGuid, srcPath, output); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred copying files from sourcepath '%v' in user service with GUID '%v' in enclave with ID '%v'",
			srcPath,
			serviceGuid,
			enclaveId,
		)
	}
	return nil
}

func (backend *MetricsReportingKurtosisBackend) StopUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	filters *service.ServiceFilters,
) (
	successfulUserServiceGuids map[service.ServiceGUID]bool,
	erroredUserServiceGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.StopUserServices(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping user services in enclave '%v' using filters: %+v", enclaveId, filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyUserServices(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	filters *service.ServiceFilters,
) (
	successfulUserServiceGuids map[service.ServiceGUID]bool,
	erroredUserServiceGuids map[service.ServiceGUID]error,
	resultErr error,
) {
	successes, failures, err := backend.underlying.DestroyUserServices(ctx, enclaveId, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying user services using filters: %+v", filters)
	}
	return successes, failures, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateNetworkingSidecar(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	serviceGuid service.ServiceGUID,
) (
	*networking_sidecar.NetworkingSidecar,
	error,
) {
	networkingSidecar, err := backend.underlying.CreateNetworkingSidecar(ctx, enclaveId, serviceGuid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating networking sidecar for user service with GUID '%v' in enclave with ID '%v'", serviceGuid, enclaveId)
	}
	return networkingSidecar, nil
}

func (backend *MetricsReportingKurtosisBackend) GetNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceGUID]*networking_sidecar.NetworkingSidecar,
	error,
) {
	networkingSidecars, err := backend.underlying.GetNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting networking sidecars using filters '%+v'", filters)
	}
	return networkingSidecars, nil
}

func (backend *MetricsReportingKurtosisBackend) RunNetworkingSidecarExecCommands(
	ctx context.Context,
	enclaveId enclave.EnclaveUUID,
	networkingSidecarsCommands map[service.ServiceGUID][]string,
) (
	map[service.ServiceGUID]*exec_result.ExecResult,
	map[service.ServiceGUID]error,
	error,
) {
	successfulNetworkingSidecarExecResults, erroredUserServiceGuids, err := backend.underlying.RunNetworkingSidecarExecCommands(ctx, enclaveId, networkingSidecarsCommands)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred running networking sidecar exec commands '%+v' in enclave with ID '%v'", networkingSidecarsCommands, enclaveId)
	}
	return successfulNetworkingSidecarExecResults, erroredUserServiceGuids, nil
}

func (backend *MetricsReportingKurtosisBackend) StopNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {
	successfulUserServiceGuids, erroredUserServiceGuids, err := backend.underlying.StopNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred stopping networking sidecars using filters '%+v'", filters)
	}
	return successfulUserServiceGuids, erroredUserServiceGuids, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyNetworkingSidecars(
	ctx context.Context,
	filters *networking_sidecar.NetworkingSidecarFilters,
) (
	map[service.ServiceGUID]bool,
	map[service.ServiceGUID]error,
	error,
) {
	successfulUserServiceGuids, erroredUserServiceGuids, err := backend.underlying.DestroyNetworkingSidecars(ctx, filters)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred destroying networking sidecars using filters '%+v'", filters)
	}
	return successfulUserServiceGuids, erroredUserServiceGuids, nil
}

func (backend *MetricsReportingKurtosisBackend) CreateLogsDatabase(
	ctx context.Context,
	logsDatabaseHttpPortNumber uint16,
) (
	*logs_database.LogsDatabase,
	error,
) {

	logsDatabase, err := backend.underlying.CreateLogsDatabase(ctx, logsDatabaseHttpPortNumber)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the logs database with HTTP port number '%v'", logsDatabaseHttpPortNumber)
	}

	return logsDatabase, nil
}

// if nothing is found returns nil
func (backend *MetricsReportingKurtosisBackend) GetLogsDatabase(
	ctx context.Context,
) (
	resultMaybeLogsDatabase *logs_database.LogsDatabase,
	resultErr error,
) {
	maybeLogsDatabase, err := backend.underlying.GetLogsDatabase(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database")
	}

	return maybeLogsDatabase, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyLogsDatabase(
	ctx context.Context,
) error {
	if err := backend.underlying.DestroyLogsDatabase(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the logs database")
	}

	return nil
}

func (backend *MetricsReportingKurtosisBackend) CreateLogsCollector(
	ctx context.Context,
	logsCollectorTcpPortNumber uint16,
	logsCollectorHttpPortNumber uint16,
) (
	*logs_collector.LogsCollector,
	error,
) {

	logsCollector, err := backend.underlying.CreateLogsCollector(ctx, logsCollectorTcpPortNumber, logsCollectorHttpPortNumber)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the logs collector with TCP port number '%v' and HTTP port number '%v'", logsCollectorTcpPortNumber, logsCollectorHttpPortNumber)
	}

	return logsCollector, nil
}

// if nothing is found returns nil
func (backend *MetricsReportingKurtosisBackend) GetLogsCollector(
	ctx context.Context,
) (
	resultMaybeLogsCollector *logs_collector.LogsCollector,
	resultErr error,
) {
	maybeLogsCollector, err := backend.underlying.GetLogsCollector(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs collector")
	}

	return maybeLogsCollector, nil
}

func (backend *MetricsReportingKurtosisBackend) DestroyLogsCollector(
	ctx context.Context,
) error {

	if err := backend.underlying.DestroyLogsCollector(ctx); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying the logs collector")
	}

	return nil
}
