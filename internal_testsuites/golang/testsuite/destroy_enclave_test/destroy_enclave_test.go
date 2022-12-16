package destroy_enclave_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "destroy-enclave"
	isPartitioningEnabled = false

	fileServerServiceImage                      = "flashspys/nginx-static"
	fileServerServiceId      services.ServiceID = "file-server"
	fileServerPortId                            = "http"
	fileServerPrivatePortNum                    = 80

	testFilesArtifactUrl = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"

	filesArtifactMountpoint  = "/static"
	emptyApplicationProtocol = ""
)

var fileServerPortSpec = services.NewPortSpec(fileServerPrivatePortNum, services.TransportProtocol_TCP, emptyApplicationProtocol)

func TestDestroyEnclave(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	shouldStopEnclaveAtTheEnd := true
	defer func() {
		if shouldStopEnclaveAtTheEnd {
			stopEnclaveFunc()
		}
	}()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	filesArtifactUuid, err := enclaveCtx.StoreWebFiles(ctx, testFilesArtifactUrl)
	require.NoError(t, err, "An error occurred storing the files artifact")

	fileServerContainerConfig := getFileServerContainerConfig(filesArtifactUuid)
	_, err = enclaveCtx.AddService(fileServerServiceId, fileServerContainerConfig)
	require.NoError(t, err, "An error occurred adding the file server service")

	err = destroyEnclaveFunc()
	require.NoErrorf(t, err, "An error occurred destroying enclave with ID '%v'", enclaveCtx.GetEnclaveID())
	shouldStopEnclaveAtTheEnd = false
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

func getFileServerContainerConfig(filesArtifactUuid services.FilesArtifactUUID) *services.ContainerConfig {

	containerConfig := services.NewContainerConfigBuilder(
		fileServerServiceImage,
	).WithUsedPorts(map[string]*services.PortSpec{
		fileServerPortId: fileServerPortSpec,
	}).WithFiles(map[string]services.FilesArtifactUUID{
		filesArtifactMountpoint: filesArtifactUuid,
	}).Build()
	return containerConfig
}
