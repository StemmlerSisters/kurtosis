package test_helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_consts"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"os"
	"time"
)

const (
	configFilepathRelativeToSharedDirRoot = "config-file.txt"

	datastoreImage = "kurtosistech/example-datastore-server"
	apiServiceImage = "kurtosistech/example-api-server"

	datastoreWaitForStartupMaxPolls = 10
	datastoreWaitForStartupDelayMilliseconds = 1000

	apiWaitForStartupMaxPolls = 10
	apiWaitForStartupDelayMilliseconds = 1000

	defaultPartitionId = ""
)
var datastorePortStr = fmt.Sprintf("%v/%v", datastore_rpc_api_consts.ListenPort, datastore_rpc_api_consts.ListenProtocol)
var apiPortStr = fmt.Sprintf("%v/%v", example_api_server_rpc_api_consts.ListenPort, example_api_server_rpc_api_consts.ListenProtocol)

type GrpcAvailabilityChecker interface {
	IsAvailable(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type datastoreConfig struct {
	DatastoreIp   string `json:"datastoreIp"`
	DatastorePort uint16    `json:"datastorePort"`
}

func AddDatastoreService(ctx context.Context, serviceId services.ServiceID, enclaveCtx *enclaves.EnclaveContext) (*services.ServiceContext, datastore_rpc_api_bindings.DatastoreServiceClient, func(), error) {
	containerConfigSupplier := getDatastoreContainerConfigSupplier()

	serviceCtx, hostPortBindings, err := enclaveCtx.AddService(serviceId, containerConfigSupplier)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	hostPortBinding, found := hostPortBindings[datastorePortStr]
	if !found {
		return nil, nil, nil, stacktrace.NewError("No datastore host port binding found for port string '%v'", datastorePortStr)
	}

	datastoreIp := hostPortBinding.InterfaceIp
	datastorePortNumStr := hostPortBinding.InterfacePort
	client, clientCloseFunc, err := CreateDatastoreClient(datastoreIp, datastorePortNumStr)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating the datastore client for IP '%v' and port '%v'",
			datastoreIp,
			datastorePortNumStr,
		)
	}

	if err := WaitForHealthy(ctx, client, datastoreWaitForStartupMaxPolls, datastoreWaitForStartupDelayMilliseconds); err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}
	return serviceCtx, client, clientCloseFunc, nil
}

func CreateDatastoreClient(ipAddr string, portNum string) (datastore_rpc_api_bindings.DatastoreServiceClient, func(), error) {
	url := fmt.Sprintf("%v:%v", ipAddr, portNum)
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred connecting to datastore service on URL '%v'", url)
	}
	clientCloseFunc := func() {
		if err := conn.Close(); err != nil {
			logrus.Warnf("We tried to close the datastore client, but doing so threw an error:\n%v", err)
		}
	}
	client := datastore_rpc_api_bindings.NewDatastoreServiceClient(conn)
	return client, clientCloseFunc, nil
}

func AddAPIService(ctx context.Context, serviceId services.ServiceID, enclaveCtx *enclaves.EnclaveContext, datastoreIPInsideNetwork string) (*services.ServiceContext, example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, func(), error) {
	serviceCtx, client, clientCloseFunc, err := AddAPIServiceToPartition(ctx, serviceId, enclaveCtx, datastoreIPInsideNetwork, defaultPartitionId)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred adding API service to default partition")
	}
	return serviceCtx, client, clientCloseFunc, nil
}


func AddAPIServiceToPartition(ctx context.Context, serviceId services.ServiceID, enclaveCtx *enclaves.EnclaveContext, datastoreIPInsideNetwork string, partitionId enclaves.PartitionID) (*services.ServiceContext, example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, func(), error) {
	containerConfigSupplier := getApiServiceContainerConfigSupplier(datastoreIPInsideNetwork)

	serviceCtx, hostPortBindings, err := enclaveCtx.AddServiceToPartition(serviceId, partitionId, containerConfigSupplier)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	hostPortBinding, found := hostPortBindings[apiPortStr]
	if !found {
		return nil, nil, nil, stacktrace.NewError("No API service host port binding found for port string '%v'", apiPortStr)
	}

	url := fmt.Sprintf("%v:%v", hostPortBinding.InterfaceIp, hostPortBinding.InterfacePort)
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred connecting to API service on URL '%v'", url)
	}
	clientCloseFunc := func() {
		if err := conn.Close(); err != nil {
			logrus.Warnf("We tried to close the API service client, but doing so threw an error:\n%v", err)
		}
	}
	client := example_api_server_rpc_api_bindings.NewExampleAPIServerServiceClient(conn)

	if err := WaitForHealthy(context.Background(), client, apiWaitForStartupMaxPolls, apiWaitForStartupDelayMilliseconds); err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the API service to become available")
	}
	return serviceCtx, client, clientCloseFunc, nil
}

func WaitForHealthy(ctx context.Context, client GrpcAvailabilityChecker, retries uint32, retriesDelayMilliseconds uint32) error {
	var (
		emptyArgs = &empty.Empty{}
		err       error
	)

	for i := uint32(0); i < retries; i++ {
		_, err = client.IsAvailable(ctx, emptyArgs)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(retriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return stacktrace.Propagate(
			err,
			"The service didn't return a success code, even after %v retries with %v milliseconds in between retries",
			retries,
			retriesDelayMilliseconds,
		)
	}

	return nil
}

// ====================================================================================================
//                                      Private Helper Methods
// ====================================================================================================
func getDatastoreContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			datastoreImage,
		).WithUsedPorts(map[string]bool{
			datastorePortStr: true,
		}).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func getApiServiceContainerConfigSupplier(datastoreIPInsideNetwork string) func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

		datastoreConfigFileFilePath, err := createDatastoreConfigFileInServiceDirectory(datastoreIPInsideNetwork, sharedDirectory)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating data store config file in service container")
		}

		startCmd := []string{
			"./example-api-server.bin",
			"--config",
			datastoreConfigFileFilePath.GetAbsPathOnServiceContainer(),
		}

		containerConfig := services.NewContainerConfigBuilder(
			apiServiceImage,
		).WithUsedPorts(map[string]bool{
			apiPortStr: true,
		}).WithCmdOverride(startCmd).Build()

		return containerConfig, nil
	}

	return containerConfigSupplier
}



func createDatastoreConfigFileInServiceDirectory(datastoreIP string, sharedDirectory *services.SharedPath) (*services.SharedPath, error) {
	configFileFilePath := sharedDirectory.GetChildPath(configFilepathRelativeToSharedDirRoot)

	logrus.Infof("Config file absolute path on this container: %v , on service container: %v", configFileFilePath.GetAbsPathOnThisContainer(), configFileFilePath.GetAbsPathOnServiceContainer())

	logrus.Debugf("Datastore IP: %v , port: %v", datastoreIP, datastore_rpc_api_consts.ListenPort)

	configObj := datastoreConfig{
		DatastoreIp:   datastoreIP,
		DatastorePort: datastore_rpc_api_consts.ListenPort,
	}
	configBytes, err := json.Marshal(configObj)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the config to JSON")
	}

	logrus.Debugf("API config JSON: %v", string(configBytes))

	if err := ioutil.WriteFile(configFileFilePath.GetAbsPathOnThisContainer(), configBytes, os.ModePerm); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred writing the serialized config JSON to file")
	}

	return configFileFilePath, nil
}
