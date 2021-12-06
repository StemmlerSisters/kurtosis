import { ServiceContext } from "kurtosis-core-api-lib"
import * as serverApi from "example-api-server-api-lib";
import { Result, ok, err } from "neverthrow"
import log from "loglevel"
import * as grpc from "grpc"
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";

import { createEnclave } from "../../test_helpers/enclave_setup";
import { addAPIService, addDatastoreService } from "../../test_helpers/test_helpers";

const TEST_NAME = "basic-datastore-and-api-test";
const IS_PARTITIONING_ENABLED = false;
const DATASTORE_SERVICE_ID = "datastore";
const API_SERVICE_ID = "api";
const TEST_PERSON_ID = "23";
const TEST_NUM_BOOKS_READ = 3;

jest.setTimeout(30000)

test("Test basic data store and API", async () => {

    // ------------------------------------- ENGINE SETUP ----------------------------------------------
    const createEnclaveResult = await createEnclave(TEST_NAME, IS_PARTITIONING_ENABLED)

    if(createEnclaveResult.isErr()) { throw createEnclaveResult.error }

    const { enclaveContext, stopEnclaveFunction } = createEnclaveResult.value

    try {
        // ------------------------------------- TEST SETUP ----------------------------------------------

        log.info("Adding datastore service...")

        const addDatastoreServiceResult = await addDatastoreService(DATASTORE_SERVICE_ID, enclaveContext)

        if(addDatastoreServiceResult.isErr()) { throw addDatastoreServiceResult.error }

        const { 
            serviceContext: datastoreServiceContext, 
            clientCloseFunction:datastoreClientCloseFunction 
        } = addDatastoreServiceResult.value

        log.info("Added datastore service")

        log.info("Adding API service...")
        const apiClientResult: Result<{
            serviceContext: ServiceContext;
            client: serverApi.ExampleAPIServerServiceClient;
            clientCloseFunction: () => void;
          }, Error> = await addAPIService(API_SERVICE_ID, enclaveContext, datastoreServiceContext.getIPAddress())
		
          if(apiClientResult.isErr()){ throw apiClientResult.error }

        const { 
            client: apiClient, 
            clientCloseFunction: apiClientCloseFunction  
        } = apiClientResult.value

		log.info("Added API service")
        
        
        try {
            // ------------------------------------- TEST RUN ----------------------------------------------
            log.info(`Verifying that person with test ID ${TEST_PERSON_ID} doesn't already exist...`);
            
            const getPersonArgs = new serverApi.GetPersonArgs()
            getPersonArgs.setPersonId(TEST_PERSON_ID)

            const getPersonResultPromise: Promise<Result<serverApi.GetPersonResponse, Error>> = new Promise((resolve, _unusedReject) => {
                    apiClient.getPerson(getPersonArgs, (error: grpc.ServiceError | null, response?: serverApi.GetPersonResponse) => {
                        if (error === null) {
                            if (!response) {
                                resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                            } else {
                                resolve(ok(response!));
                            }
                        } else {
                            resolve(err(error));
                        }
                    })
            })

            const getPersonResult = await getPersonResultPromise;
            if(getPersonResult.isOk()) { 
                throw new Error("Expected an error trying to get a person who doesn't exist yet, but didn't receive one")
            }
            log.info("Verified that test person doesn't already exist")
            
            
            log.info(`Adding test person with ID ${TEST_PERSON_ID}...`)
            const addPersonArgs = new serverApi.AddPersonArgs()
            addPersonArgs.setPersonId(TEST_PERSON_ID)

            const addPersonResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                apiClient.addPerson(addPersonArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
                    if (error === null) {
                        if (!response) {
                            resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                        } else {
                            resolve(ok(response!));
                        }
                    } else {
                        resolve(err(error));
                    }
                })
            })

            const addPersonResult = await addPersonResultPromise;
            if(addPersonResult.isErr()) {
                log.error(addPersonResult.error)
                throw new Error(`An error occurred adding test person with ID ${TEST_PERSON_ID}`) 
            }
            log.info("Test person added")
            
            log.info(`Incrementing test person's number of books read by ${TEST_NUM_BOOKS_READ}...`)
            
            const incrementBooksReadArgs = new serverApi.IncrementBooksReadArgs()
            incrementBooksReadArgs.setPersonId(TEST_PERSON_ID)
            
            for (let i = 0; i < TEST_NUM_BOOKS_READ; i++) {
                const incrementBooksReadResultPromise: Promise<Result<google_protobuf_empty_pb.Empty, Error>> = new Promise((resolve, _unusedReject) => {
                        apiClient.incrementBooksRead(incrementBooksReadArgs, (error: grpc.ServiceError | null, response?: google_protobuf_empty_pb.Empty) => {
                            if (error === null) {
                                if (!response) {
                                    resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                                } else {
                                    resolve(ok(response!));
                                }
                            } else {
                                resolve(err(error));
                            }
                        })
                })
                const incrementBooksReadResult = await incrementBooksReadResultPromise;
                if(incrementBooksReadResult.isErr()) { 
                    log.error(incrementBooksReadResult.error)
                    throw new Error("An error occurred incrementing the number of books read")
                }
            }
            
            log.info("Incremented number of books read")
            
            log.info("Retrieving test person to verify number of books read...")

            const getNewPersonResultPromise: Promise<Result<serverApi.GetPersonResponse, Error>> = new Promise((resolve, _unusedReject) => {
                apiClient.getPerson(getPersonArgs, (error: grpc.ServiceError | null, response?: serverApi.GetPersonResponse) => {
                    if (error === null) {
                        if (!response) {
                            resolve(err(new Error("No error was encountered but the response was still falsy; this should never happen")));
                        } else {
                            resolve(ok(response!));
                        }
                    } else {
                        resolve(err(error));
                    }
                })
            })
            const  getNewPersonResult = await getNewPersonResultPromise
            if(getNewPersonResult.isErr()){
                log.error("An error occurred getting the test person to verify the number of books read")
                throw getNewPersonResult.error
            }
            log.info("Retrieved test person")

            const newPersonBooksRead = getNewPersonResult.value.getBooksRead()
            
            if(TEST_NUM_BOOKS_READ !== newPersonBooksRead){
                throw new Error(`Expected number of book read ${TEST_NUM_BOOKS_READ} != actual number of books read ${newPersonBooksRead}`)
            }
            
        }finally{
            apiClientCloseFunction()
            datastoreClientCloseFunction()
        }
        
    }finally{
        stopEnclaveFunction()
    }
    
})

