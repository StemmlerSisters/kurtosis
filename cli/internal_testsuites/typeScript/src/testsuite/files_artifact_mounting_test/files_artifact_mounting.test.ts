import { ContainerConfig, ContainerConfigBuilder, FilesArtifactID, PortProtocol, PortSpec, ServiceID, SharedPath } from "kurtosis-core-api-lib"
import log from "loglevel";
import { Result, ok, err } from "neverthrow";
import axios from "axios"

import { createEnclave } from "../../test_helpers/enclave_setup";

const TEST_NAME = "files-artifact-mounting-test"
const IS_PARTITIONING_ENABLED = false

const FILE_SERVER_SERVICE_IMAGE = "flashspys/nginx-static"
const FILE_SERVER_SERVICE_ID: ServiceID = "file-server"
const FILE_SERVER_PORT_ID = "http"
const FILE_SERVER_PRIVATE_PORT_NUM = 80

const WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS = 500
const WAIT_FOR_STARTUP_MAX_RETRIES = 15
const WAIT_INITIAL_DELAY_MILLISECONDS = 0

const TEST_FILES_ARTIFACT_ID: FilesArtifactID = "test-files-artifact"
const TEST_FILES_ARTIFACT_URL = "https://kurtosis-public-access.s3.us-east-1.amazonaws.com/test-artifacts/static-fileserver-files.tgz"

// Filenames & contents for the files stored in the files artifact
const FILE1_FILENAME = "file1.txt"
const FILE2_FILENAME = "file2.txt"

const EXPECTED_FILE1_CONTENTS = "file1\n"
const EXPECTED_FILE2_CONTENTS = "file2\n"

const FILE_SERVER_PORT_SPEC = new PortSpec( FILE_SERVER_PRIVATE_PORT_NUM, PortProtocol.TCP )

jest.setTimeout(40000)

test("Test files artifact mounting", async () => {
    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {

        // ------------------------------------- TEST SETUP ----------------------------------------------
        const filesArtifacts = new Map<string,FilesArtifactID>()
        filesArtifacts.set(TEST_FILES_ARTIFACT_ID, TEST_FILES_ARTIFACT_URL)
        const registerFilesArtifactsResult = await enclaveContext.registerFilesArtifacts(filesArtifacts);

        if(registerFilesArtifactsResult.isErr()) { throw registerFilesArtifactsResult.error }

        const fileServerContainerConfigSupplier = getFileServerContainerConfigSupplier()

        const addServiceResult = await enclaveContext.addService(FILE_SERVER_SERVICE_ID, fileServerContainerConfigSupplier)

        if(addServiceResult.isErr()){ throw addServiceResult.error }

        const serviceContext = addServiceResult.value
        const publicPort = serviceContext.getPublicPorts().get(FILE_SERVER_PORT_ID)
        if(publicPort === undefined){
            throw new Error(`Expected to find public port for ID ${FILE_SERVER_PORT_ID}, but none was found`)
        }

        const fileServerPublicIp = serviceContext.getMaybePublicIPAddress();
        const fileServerPublicPortNum = publicPort.number

        const waitForHttpGetEndpointAvailabilityResult = await enclaveContext.waitForHttpGetEndpointAvailability(
            FILE_SERVER_SERVICE_ID, 
            FILE_SERVER_PRIVATE_PORT_NUM,
            FILE1_FILENAME, 
            WAIT_INITIAL_DELAY_MILLISECONDS, 
            WAIT_FOR_STARTUP_MAX_RETRIES, 
            WAIT_FOR_STARTUP_TIME_BETWEEN_POLLS, 
            ""
        );

        if(waitForHttpGetEndpointAvailabilityResult.isErr()){
            throw waitForHttpGetEndpointAvailabilityResult.error
        }

        log.info(`Added file server service with public IP ${fileServerPublicIp} and port ${fileServerPublicPortNum}`)

        // ------------------------------------- TEST RUN ----------------------------------------------

        const file1ContentsResult = await getFileContents(
            fileServerPublicIp,
            fileServerPublicPortNum,
            FILE1_FILENAME
        )
        if(file1ContentsResult.isErr()){
            log.error("An error occurred getting file 1's contents")
            throw file1ContentsResult.error
        }

        const file1Contents = file1ContentsResult.value
        if(file1Contents !== EXPECTED_FILE1_CONTENTS){
            throw new Error(`Actual file 1 contents ${file1Contents} != expected file 1 contents ${EXPECTED_FILE1_CONTENTS}`)
        }

        const file2ContentsResult = await getFileContents(
            fileServerPublicIp,
            fileServerPublicPortNum,
            FILE2_FILENAME
        )

        if(file2ContentsResult.isErr()){
            log.error("An error occurred getting file 2's contents")
            throw file2ContentsResult.error
        }

        const file2Contents = file2ContentsResult.value
        if(file2Contents !== EXPECTED_FILE2_CONTENTS){
            throw new Error(`Actual file 2 contents ${file2Contents} != expected file 2 contents ${EXPECTED_FILE2_CONTENTS}`)
        }


    }finally{
        stopEnclaveFunction()
    }
    jest.clearAllTimers()
})

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

function getFileServerContainerConfigSupplier(): (ipAddr: string, sharedDirectory: SharedPath) => Result<ContainerConfig, Error> {
	
    const containerConfigSupplier = (ipAddr:string, sharedDirectory: SharedPath): Result<ContainerConfig, Error> => {

        const usedPorts = new Map<string, PortSpec>()
        usedPorts.set(FILE_SERVER_PORT_ID, FILE_SERVER_PORT_SPEC)

        const filesArtifacts = new Map<string, FilesArtifactID>()
        filesArtifacts.set(TEST_FILES_ARTIFACT_ID, "/static")

        const containerConfig = new ContainerConfigBuilder(FILE_SERVER_SERVICE_IMAGE)
            .withUsedPorts(usedPorts)
            .withFilesArtifacts(filesArtifacts)
            .build()

        return ok(containerConfig)
    }

    return containerConfigSupplier
}

async function getFileContents(ipAddress: string, portNum: number, filename: string): Promise<Result<string, Error>> {
	let response;
    try {
        response = await axios(`http://${ipAddress}:${portNum}/${filename}`)
    }catch(error){
        log.error(`An error occurred getting the contents of file ${filename}`)
        if(error instanceof Error){
            return err(error)
        }else{
            return err(new Error("An error occurred getting the contents of file, but the error wasn't of type Error"))
        }
    }
    const bodyStr = String(response.data)
	return ok(bodyStr)
}