package logs_database_functions

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_database"
	"github.com/kurtosis-tech/stacktrace"
)

func GetLogsDatabase(
	ctx context.Context,
	filters *logs_database.LogsDatabaseFilters,
	dockerManager *docker_manager.DockerManager,
) (*logs_database.LogsDatabase, error){

	logsDatabaseObject, _, err := getLogsDatabaseObjectAndContainerIdMatching(ctx, filters, dockerManager)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the logs database using filters '%+v'", filters)
	}

	return logsDatabaseObject, nil
}
