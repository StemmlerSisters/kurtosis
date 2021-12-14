import {
    ServiceID,
    EnclaveContext,
    SharedPath,
    ContainerConfig,
    ContainerConfigBuilder,
    ServiceContext,
    PartitionID,
    PortSpec,
    PortProtocol,
} from "kurtosis-core-api-lib";
import * as datastoreApi from "example-datastore-server-api-lib";
import * as serverApi from "example-api-server-api-lib";
import { err, ok, Result } from "neverthrow";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as grpc from "@grpc/grpc-js";
import log from "loglevel";
import * as fs from 'fs';

const CONFIG_FILEPATH_RELATIVE_TO_SHARED_DIR_ROOT = "config-file.txt";

const DATASTORE_IMAGE = "kurtosistech/example-datastore-server";
const API_SERVICE_IMAGE = "kurtosistech/example-api-server";

const DATASTORE_PORT_ID = "rpc";
const API_PORT_ID = "rpc";

const DATASTORE_WAIT_FOR_STARTUP_MAX_POLLS = 10;
const DATASTORE_WAIT_FOR_STARTUP_DELAY_MILLISECONDS = 1000;

const API_WAIT_FOR_STARTUP_MAX_POLLS = 10;
const API_WAIT_FOR_STARTUP_DELAY_MILLISECONDS = 1000;

const DEFAULT_PARTITION_ID = "";

const DATASTORE_PORT_SPEC = new PortSpec(
    datastoreApi.LISTEN_PORT,
    PortProtocol.TCP,
)
const API_PORT_SPEC = new PortSpec(
    serverApi.LISTEN_PORT,
    PortProtocol.TCP,
)

export async function addDatastoreService(serviceId: ServiceID, enclaveContext: EnclaveContext):
    Promise<Result<{
        serviceContext: ServiceContext;
        client: datastoreApi.DatastoreServiceClient;
        clientCloseFunction: () => void;
    },Error>> {
    
    const containerConfigSupplier = getDatastoreContainerConfigSupplier();

    const addServiceResult = await enclaveContext.addService(serviceId, containerConfigSupplier);
    if (addServiceResult.isErr()) {
        log.error("An error occurred adding the datastore service");
        return err(addServiceResult.error);
    }
    const serviceContext = addServiceResult.value;

    const publicPort: PortSpec | undefined = serviceContext.getPublicPorts().get(DATASTORE_PORT_ID);
    if (publicPort === undefined) {
        return err(new Error(`No datastore public port found for port ID '${DATASTORE_PORT_ID}'`))
    }

    const publicIp = serviceContext.getMaybePublicIPAddress();
    const publicPortNum = publicPort.number;
    const { client, clientCloseFunction } = createDatastoreClient(publicIp, publicPortNum);

    const waitForHealthyResult = await waitForHealthy(
        client,
        DATASTORE_WAIT_FOR_STARTUP_MAX_POLLS,
        DATASTORE_WAIT_FOR_STARTUP_DELAY_MILLISECONDS
    );

    if (waitForHealthyResult.isErr()) {
        log.error("An error occurred waiting for the datastore service to become available")
        return err(waitForHealthyResult.error);
    }

    return ok({ serviceContext, client, clientCloseFunction });
};

function createDatastoreClient(ipAddr: string, portNum: number): { client: datastoreApi.DatastoreServiceClient; clientCloseFunction: () => void } {
    const url = `${ipAddr}:${portNum}`;
    const client = new datastoreApi.DatastoreServiceClient(url, grpc.credentials.createInsecure());
    const clientCloseFunction = () => client.close();

    return { client, clientCloseFunction }
};

export async function addAPIService( serviceId: ServiceID, enclaveContext: EnclaveContext, datastoreIPInsideNetwork: string):
    Promise<Result<{
        serviceContext: ServiceContext;
        client: serverApi.ExampleAPIServerServiceClient;
        clientCloseFunction: () => void;
    }, Error>> {
  
    const addAPIServiceToPartitionResult = await addAPIServiceToPartition(
        serviceId,
        enclaveContext,
        datastoreIPInsideNetwork,
        DEFAULT_PARTITION_ID
    );
    if(addAPIServiceToPartitionResult.isErr()) return err(addAPIServiceToPartitionResult.error)
    const { serviceContext, client, clientCloseFunction} = addAPIServiceToPartitionResult.value
  
    return ok({ serviceContext, client, clientCloseFunction})
}

export async function addAPIServiceToPartition( serviceId: ServiceID, enclaveContext: EnclaveContext, datastoreIPInsideNetwork: string, partitionId: PartitionID):
    Promise<Result<{
        serviceContext: ServiceContext;
        client: serverApi.ExampleAPIServerServiceClient;
        clientCloseFunction: () => void;
    },Error>> {
  
    const containerConfigSupplier = getApiServiceContainerConfigSupplier(datastoreIPInsideNetwork)

    const addServiceToPartitionResult = await enclaveContext.addServiceToPartition(serviceId, partitionId, containerConfigSupplier)
    if(addServiceToPartitionResult.isErr()) return err(addServiceToPartitionResult.error)

    const serviceContext = addServiceToPartitionResult.value;

    const publicPort: PortSpec | undefined = serviceContext.getPublicPorts().get(API_PORT_ID);
    if (publicPort === undefined) {
        return err(new Error(`No API service public port found for port ID '${API_PORT_ID}'`));
    }
  
    const url = `${serviceContext.getMaybePublicIPAddress()}:${publicPort.number}`;
    const client = new serverApi.ExampleAPIServerServiceClient(url, grpc.credentials.createInsecure());
    const clientCloseFunction = () => client.close();

    const waitForHealthyResult = await waitForHealthy(client, API_WAIT_FOR_STARTUP_MAX_POLLS, API_WAIT_FOR_STARTUP_DELAY_MILLISECONDS)

    if(waitForHealthyResult.isErr()) {
        log.error("An error occurred waiting for the API service to become available")
        return err(waitForHealthyResult.error)
    }
  
    return ok({ serviceContext, client, clientCloseFunction })
};

async function waitForHealthy(
  client: datastoreApi.DatastoreServiceClient | serverApi.ExampleAPIServerServiceClient,
  retries: number,
  retriesDelayMilliseconds: number
): Promise<Result<null, Error>> {

    const emptyArgs: google_protobuf_empty_pb.Empty = new google_protobuf_empty_pb.Empty();

    const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

    const checkClientAvailabilityResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise(async (resolve, _unusedReject) => {
        client.isAvailable(emptyArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
        if (error === null) {
            if (!response) {
                resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
            } else {
                resolve(ok(response));
            }
        } else {
            resolve(err(error));
        }
        });
    });

    for (let i = 0; i < retries; i++) {
        const checkClientAvailabilityResult = await checkClientAvailabilityResultPromise;
        if (checkClientAvailabilityResult.isOk()) {
            return ok(null)
        } else {
            log.debug(checkClientAvailabilityResult.error)
            await sleep(retriesDelayMilliseconds);
        }
    }

    return err(new Error(`The service didn't return a success code, even after ${retries} retries with ${retriesDelayMilliseconds} milliseconds in between retries`));

}

// ====================================================================================================
//                                      Private Helper Methods
// ====================================================================================================

function getDatastoreContainerConfigSupplier(): ( ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> {

    const containerConfigSupplier = ( ipAddr: string, sharedDirectory: SharedPath): Result<ContainerConfig, Error> => {
        const usedPorts = new Map<string, PortSpec>();
        usedPorts.set(DATASTORE_PORT_ID, DATASTORE_PORT_SPEC);

        const containerConfig = new ContainerConfigBuilder(DATASTORE_IMAGE).withUsedPorts(usedPorts).build();

        return ok(containerConfig);
    };

    return containerConfigSupplier;
}

function getApiServiceContainerConfigSupplier(datastoreIPInsideNetwork:string):
    (ipAddr:string,sharedDirectory: SharedPath)=> Result<ContainerConfig, Error> {

    const containerConfigSupplier = (ipAddr:string,sharedDirectory: SharedPath): Result<ContainerConfig, Error> =>{
        const datastoreConfigFileFilePathResult = createDatastoreConfigFileInServiceDirectory(datastoreIPInsideNetwork, sharedDirectory)
        if(datastoreConfigFileFilePathResult.isErr()) { return err(datastoreConfigFileFilePathResult.error) }
       
        const datastoreConfigFileFilePath = datastoreConfigFileFilePathResult.value
  
        const usedPorts = new Map<string, PortSpec>();
        usedPorts.set(API_PORT_ID, API_PORT_SPEC);
        const startCmd:string[] = ["./example-api-server.bin", "--config", datastoreConfigFileFilePath.getAbsPathOnServiceContainer()]
  
        const containerConfig = new ContainerConfigBuilder(API_SERVICE_IMAGE)
            .withUsedPorts(usedPorts)
            .withCmdOverride(startCmd)
            .build()
  
        return ok(containerConfig)
    }

    return containerConfigSupplier;

  }
  
  function createDatastoreConfigFileInServiceDirectory(datastoreIP: string, sharedDirectory: SharedPath): Result<SharedPath,Error>{
    const configFileFilePath = sharedDirectory.getChildPath(CONFIG_FILEPATH_RELATIVE_TO_SHARED_DIR_ROOT)
  
    log.info(`Config file absolute path on this container: ${configFileFilePath.getAbsPathOnThisContainer()}, on service container:${configFileFilePath.getAbsPathOnServiceContainer()}`)
    
    log.debug(`Datastore IP:${datastoreIP} , port: ${datastoreApi.LISTEN_PORT}`)
    
    const config = {
        datastoreIp:   datastoreIP,
        datastorePort: datastoreApi.LISTEN_PORT,
    };

    const configJSONStringified = JSON.stringify(config);
  
    log.debug(`API config JSON: ${configJSONStringified}`)

    try {
        fs.writeFileSync(configFileFilePath.getAbsPathOnThisContainer(), configJSONStringified);
    }catch(error) {
        log.error("An error occurred writing the serialized config JSON to file")
        if(error instanceof Error){
            return err(error)
        }else {
            return err(new Error("Encountered error while writing the file, but the error wasn't of type Error"))
        }
    }
  
    return ok(configFileFilePath);
}

