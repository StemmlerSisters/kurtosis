package centralized_logs

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/centralized_logs/mocks"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

const (
	fakeLogsDatabaseAddress = "1.2.3.4:8080"

	testEnclaveId        = "test-enclave"
	testUserService1Guid = "test-user-service-1"
	testUserService2Guid = "test-user-service-2"
	testUserService3Guid = "test-user-service-3"

	filterText = "first"

	//Expected values
	expectedFirstLogLineOnEachService               = "This is the first log line."
	expectedOrganizationIdHttpHeaderKey             = "X-Scope-Orgid"
	expectedStartTimeQueryParamKey                  = "start"
	expectedQueryLogsQueryParamKey                  = "query"
	expectedEntriesLimitQueryParamKey               = "limit"
	expectedDirectionQueryParamKey                  = "direction"
	expectedKurtosisContainerTypeLokiTagKey         = "comKurtosistechContainerType"
	expectedKurtosisGuidLokiTagKey                  = "comKurtosistechGuid"
	expectedURLScheme                               = "http"
	expectedQueryRangeURLPath                       = "/loki/api/v1/query_range"
	expectedQueryLogsQueryParamValueRegex           = `{comKurtosistechContainerType="user-service",comKurtosistechGuid=~"test-user-service-[1-3]\|test-user-service-[1-3]\|test-user-service-[1-3]"}`
	expectedQueryLogsWithFilterQueryParamValueRegex = expectedQueryLogsQueryParamValueRegex + "|= " + filterText
	expectedEntriesLimitQueryParamValue             = "4000"
	expectedDirectionQueryParamValue                = "forward"
	expectedAmountQueryParams                       = 4

	userServiceContainerType = "user-service"

	testTimeOut = 30 * time.Second
)

func TestGetUserServiceLogsWithoutFilter_ValidResponse(t *testing.T) {
	enclaveId := enclave.EnclaveUUID(testEnclaveId)
	userServiceGuids := map[service.ServiceUUID]bool{
		testUserService1Guid: true,
		testUserService2Guid: true,
		testUserService3Guid: true,
	}
	mockHttpClient := mocks.NewMockHttpClient(t)
	mockHttpClient.EXPECT().Do(mock.Anything).Run(func(request *http.Request) {
		// Here we validate the shape of the query matches our expectations and return true only if it's the case
		require.Equal(t, expectedURLScheme, request.URL.Scheme)
		require.Equal(t, fakeLogsDatabaseAddress, request.URL.Host)
		require.Equal(t, expectedQueryRangeURLPath, request.URL.Path)
		require.Equal(t, http.MethodGet, request.Method)

		organizationIds, found := request.Header[expectedOrganizationIdHttpHeaderKey]
		require.True(t, found, "Expected to find header key '%v' in request header '%+v', but it was not found", expectedOrganizationIdHttpHeaderKey, request.Header)

		expectedEnclaveId := enclaveId
		var foundExpectedEnclaveId bool
		for _, organizationId := range organizationIds {
			enclaveIdObj := enclave.EnclaveUUID(organizationId)
			if enclaveIdObj == expectedEnclaveId {
				foundExpectedEnclaveId = true
			}
		}
		require.True(t, foundExpectedEnclaveId, "Expected to find enclave ID '%v' in request header values '%+v' for header with key '%v', but it was not found", expectedEnclaveId, organizationIds, expectedOrganizationIdHttpHeaderKey)

		require.Equal(t, expectedAmountQueryParams, len(request.URL.Query()), "Expected to request contains '%v' query params, but '%v' query params were found", expectedAmountQueryParams, len(request.URL.Query()))

		found = request.URL.Query().Has(expectedStartTimeQueryParamKey)
		require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedStartTimeQueryParamKey, request.Form)

		found = request.URL.Query().Has(expectedQueryLogsQueryParamKey)
		require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedStartTimeQueryParamKey, request.Form)

		queryLogsQueryParams := request.URL.Query().Get(expectedQueryLogsQueryParamKey)
		require.Regexp(t, regexp.MustCompile(expectedQueryLogsQueryParamValueRegex), queryLogsQueryParams)

		var (
			foundExpectedKurtosisContainerTypeLokiTagKey bool
			foundExpectedKurtosisGuidLokiTagKey          bool
		)

		foundKurtosisContainerTypeLokiTagKey := strings.Contains(queryLogsQueryParams, expectedKurtosisContainerTypeLokiTagKey)
		if foundKurtosisContainerTypeLokiTagKey {
			foundExpectedKurtosisContainerTypeLokiTagKey = true
		}
		foundKurtosisGuidLokiTagKey := strings.Contains(queryLogsQueryParams, expectedKurtosisGuidLokiTagKey)
		if foundKurtosisGuidLokiTagKey {
			foundExpectedKurtosisGuidLokiTagKey = true
		}

		require.True(t, foundExpectedKurtosisContainerTypeLokiTagKey, "Expected to find Loki's tag key key '%v' in request query params '%+v', but it was not found", expectedKurtosisContainerTypeLokiTagKey, queryLogsQueryParams)
		require.True(t, foundExpectedKurtosisGuidLokiTagKey, "Expected to find Loki's tag key key '%v' in request query params '%+v', but it was not found", expectedKurtosisGuidLokiTagKey, queryLogsQueryParams)

		found = request.URL.Query().Has(expectedEntriesLimitQueryParamKey)
		require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedEntriesLimitQueryParamKey, request.Form)
		limitQueryParam := request.URL.Query().Get(expectedEntriesLimitQueryParamKey)
		require.Equal(t, expectedEntriesLimitQueryParamValue, limitQueryParam)

		found = request.URL.Query().Has(expectedDirectionQueryParamKey)
		require.True(t, found, "Expected to find query param with key '%v' in request form values '%+v', but it was not found", expectedDirectionQueryParamKey, request.Form)
		directionQueryParam := request.URL.Query().Get(expectedDirectionQueryParamKey)
		require.Equal(t, expectedDirectionQueryParamValue, directionQueryParam)
	}).Return(&http.Response{
		Status:           "",
		StatusCode:       http.StatusOK,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             io.NopCloser(strings.NewReader(mocks.MockedResponseBodyWithSeveralValuesStr)),
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}, nil)

	logsDatabaseClient := NewLokiLogsDatabaseClient(fakeLogsDatabaseAddress, mockHttpClient)

	ctx := context.Background()

	expectedUserServiceAmountLogLinesByUserServiceGuid := map[service.ServiceUUID]int{
		testUserService1Guid: 3,
		testUserService2Guid: 4,
		testUserService3Guid: 2,
	}

	var emptyLokiLineFilters []LokiLineFilter
	emptyLogPipeline := NewLokiLogPipeline(emptyLokiLineFilters)

	userServiceLogsByGuidChan, errChan, closeStreamFunc, err := logsDatabaseClient.GetUserServiceLogs(ctx, enclaveId, userServiceGuids, emptyLogPipeline)
	defer closeStreamFunc()

	require.NoError(t, err, "An error occurred getting user service logs for UUIDs '%+v' in enclave '%v'", userServiceGuids, enclaveId)
	require.NotNil(t, userServiceLogsByGuidChan, "Received a nil user service logs channel, but a non-nil value was expected")
	require.Nil(t, errChan, "Received a not nil error channel, but a nil value was expected")

	var testEvaluationErr error

	shouldReceiveStream := true
	for shouldReceiveStream {
		select {
		case <-time.Tick(testTimeOut):
			testEvaluationErr = stacktrace.NewError("Receiving stream logs in the test has reached the '%v' time out", testTimeOut)
			shouldReceiveStream = false
			break
		case userServiceLogsByGuid, isChanOpen := <-userServiceLogsByGuidChan:
			if !isChanOpen {
				shouldReceiveStream = false
				break
			}

			require.Equal(t, len(userServiceGuids), len(userServiceLogsByGuid))

			for userServiceGuid := range userServiceGuids {
				logLines, found := userServiceLogsByGuid[userServiceGuid]
				require.True(t, found)

				expectedAmountLogLines, found := expectedUserServiceAmountLogLinesByUserServiceGuid[userServiceGuid]
				require.True(t, found)

				require.Equal(t, expectedAmountLogLines, len(logLines))

				require.Equal(t, expectedFirstLogLineOnEachService, logLines[0].GetContent())
			}

			shouldReceiveStream = false
			break
		}
	}

	require.NoError(t, testEvaluationErr)

}

func TestGetUserServiceLogsWithFilter_ValidResponse(t *testing.T) {
	enclaveId := enclave.EnclaveUUID(testEnclaveId)
	userServiceGuids := map[service.ServiceUUID]bool{
		testUserService1Guid: true,
		testUserService2Guid: true,
		testUserService3Guid: true,
	}
	mockHttpClient := mocks.NewMockHttpClient(t)
	mockHttpClient.EXPECT().Do(mock.Anything).Run(func(request *http.Request) {

		queryLogsQueryParams := request.URL.Query().Get(expectedQueryLogsQueryParamKey)
		require.Regexp(t, regexp.MustCompile(expectedQueryLogsWithFilterQueryParamValueRegex), queryLogsQueryParams)

	}).Return(&http.Response{
		Status:           "",
		StatusCode:       http.StatusOK,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             io.NopCloser(strings.NewReader(mocks.MockedResponseBodyWithOneLineValuesStr)),
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}, nil)

	logsDatabaseClient := NewLokiLogsDatabaseClient(fakeLogsDatabaseAddress, mockHttpClient)

	ctx := context.Background()

	expectedUserServiceAmountLogLinesByUserServiceGuid := map[service.ServiceUUID]int{
		testUserService1Guid: 1,
		testUserService2Guid: 1,
		testUserService3Guid: 1,
	}

	lokiLineFilter := NewDoesContainTextLokiLineFilter(filterText)

	lokiLineFilters := []LokiLineFilter{
		*lokiLineFilter,
	}
	logPipeline := NewLokiLogPipeline(lokiLineFilters)

	userServiceLogsByGuidChan, errChan, closeStreamFunc, err := logsDatabaseClient.GetUserServiceLogs(ctx, enclaveId, userServiceGuids, logPipeline)
	defer closeStreamFunc()

	require.NoError(t, err, "An error occurred getting user service logs for UUIDs '%+v' using log pipeline '%v' in enclave '%v'", userServiceGuids, logPipeline, enclaveId)
	require.NotNil(t, userServiceLogsByGuidChan, "Received a nil user service logs channel, but a non-nil value was expected")
	require.Nil(t, errChan, "Received a not nil error channel, but a nil value was expected")

	var testEvaluationErr error

	shouldReceiveStream := true
	for shouldReceiveStream {
		select {
		case <-time.Tick(testTimeOut):
			testEvaluationErr = stacktrace.NewError("Receiving stream logs in the test has reached the '%v' time out", testTimeOut)
			shouldReceiveStream = false
			break
		case userServiceLogsByGuid, isChanOpen := <-userServiceLogsByGuidChan:
			if !isChanOpen {
				shouldReceiveStream = false
				break
			}

			require.Equal(t, len(userServiceGuids), len(userServiceLogsByGuid))

			for userServiceGuid := range userServiceGuids {
				logLines, found := userServiceLogsByGuid[userServiceGuid]
				require.True(t, found)

				expectedAmountLogLines, found := expectedUserServiceAmountLogLinesByUserServiceGuid[userServiceGuid]
				require.True(t, found)

				require.Equal(t, expectedAmountLogLines, len(logLines))

				require.Equal(t, expectedFirstLogLineOnEachService, logLines[0].GetContent())
			}

			shouldReceiveStream = false
			break
		}
	}

	require.NoError(t, testEvaluationErr)

}

func TestNewUserServiceLogLinesByUserServiceGuidFromLokiStreamsReturnSuccessfullyForLogTailJsonResponseBody(t *testing.T) {

	expectedLogLines := []string{"kurtosis", "test", "running", "successfully"}
	userServiceGuidStr := "stream-logs-test-service-1666785469"
	userServiceGuid := service.ServiceUUID(userServiceGuidStr)

	expectedValuesInStream1 := [][]string{
		{"1666785473000000000", "{\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\",\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"kurtosis\",\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\"}"},
	}

	expectedValuesInStream2 := [][]string{
		{"1666785473000000000", "{\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\",\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"test\",\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\"}"},
	}

	expectedValuesInStream3 := [][]string{
		{"1666785473000000000", "{\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\",\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\",\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"running\"}"},
	}

	expectedValuesInStream4 := [][]string{
		{"1666785473000000000", "{\"container_name\":\"/ts-testsuite.stream-logs-test.1666785464--user-service--stream-logs-test-service-1666785469\",\"source\":\"stdout\",\"log\":\"successfully\",\"comKurtosistechGuid\":\"stream-logs-test-service-1666785469\",\"comKurtosistechContainerType\":\"user-service\",\"com.kurtosistech.enclave-id\":\"ts-testsuite.stream-logs-test.1666785464\",\"container_id\":\"b0735bc50a76a0476928607aca13a4c73c814036bdbf8b989c2f3b458cc21eab\"}"},
	}

	lokiStreams1 := newLokiStreamValueForTest(userServiceGuid, expectedValuesInStream1)
	lokiStreams2 := newLokiStreamValueForTest(userServiceGuid, expectedValuesInStream2)
	lokiStreams3 := newLokiStreamValueForTest(userServiceGuid, expectedValuesInStream3)
	lokiStreams4 := newLokiStreamValueForTest(userServiceGuid, expectedValuesInStream4)

	lokiStreams := []lokiStreamValue{
		lokiStreams1,
		lokiStreams2,
		lokiStreams3,
		lokiStreams4,
	}

	resultLogsByKurtosisUserServiceGuid, err := newUserServiceLogLinesByUserServiceGuidFromLokiStreams(lokiStreams)
	require.NoError(t, err)
	require.NotNil(t, resultLogsByKurtosisUserServiceGuid)
	require.Equal(t, len(lokiStreams), len(resultLogsByKurtosisUserServiceGuid[userServiceGuid]))
	for expectedLogLineIndex, expectedLogLine := range expectedLogLines {
		actualLogLine := resultLogsByKurtosisUserServiceGuid[userServiceGuid][expectedLogLineIndex].GetContent()
		require.Equal(t, expectedLogLine, actualLogLine)
	}

}

func TestFilterExistingServiceGuids_FilteringWorksAsExpected(t *testing.T) {
	mockHttpClient := mocks.NewMockHttpClient(t)

	jsonResponse := `{"status": "` + lokiSuccessStatusInResponse + `", "data": ["` + testUserService1Guid + `", "` + testUserService2Guid + `"]}`
	mockHttpClient.EXPECT().Do(mock.MatchedBy(func(req *http.Request) bool {
		expectedQueryPrefix := startTimeQueryParamKey + "="
		expectedPath := fmt.Sprintf(baseLokiApiPath+queryListLabelValuesWithinRangeEndpoint, kurtosisGuidLokiTagKey)
		return req.Method == http.MethodGet &&
			req.URL.Scheme == httpScheme &&
			req.URL.Host == fakeLogsDatabaseAddress &&
			req.URL.Path == expectedPath &&
			strings.HasPrefix(req.URL.RawQuery, expectedQueryPrefix)
	})).Return(&http.Response{
		Status:           "",
		StatusCode:       http.StatusOK,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             io.NopCloser(strings.NewReader(jsonResponse)),
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}, nil)

	lokiDbClient := NewLokiLogsDatabaseClient(fakeLogsDatabaseAddress, mockHttpClient)

	ctx := context.Background()
	requestedServiceGuids := map[service.ServiceUUID]bool{
		service.ServiceUUID(testUserService1Guid): true,
		service.ServiceUUID(testUserService2Guid): true,
		service.ServiceUUID(testUserService3Guid): true,
	}
	result, err := lokiDbClient.FilterExistingServiceGuids(ctx, testEnclaveId, requestedServiceGuids)
	require.Nil(t, err)
	require.Contains(t, result, service.ServiceUUID(testUserService1Guid))
	require.Contains(t, result, service.ServiceUUID(testUserService2Guid))
	require.NotContains(t, result, service.ServiceUUID(testUserService3Guid))
}

func TestFilterExistingServiceGuids_LokiServerNotFound(t *testing.T) {
	mockHttpClient := mocks.NewMockHttpClient(t)

	mockHttpClient.EXPECT().Do(mock.Anything).Return(&http.Response{
		Status:           "",
		StatusCode:       http.StatusNotFound,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             io.NopCloser(strings.NewReader("{}")),
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}, nil)

	lokiDbClient := NewLokiLogsDatabaseClient(fakeLogsDatabaseAddress, mockHttpClient)

	ctx := context.Background()
	requestedServiceGuids := map[service.ServiceUUID]bool{
		service.ServiceUUID(testUserService1Guid): true,
	}
	result, err := lokiDbClient.FilterExistingServiceGuids(ctx, testEnclaveId, requestedServiceGuids)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "An error occurred doing HTTP request ")
}

func TestFilterExistingServiceGuids_LokiServerReturnsErrorStatus(t *testing.T) {
	mockHttpClient := mocks.NewMockHttpClient(t)

	jsonResponse := `{"status": "ERROR_STATUS", "data": []}`
	mockHttpClient.EXPECT().Do(mock.Anything).Return(&http.Response{
		Status:           "",
		StatusCode:       http.StatusOK,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             io.NopCloser(strings.NewReader(jsonResponse)),
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}, nil)

	lokiDbClient := NewLokiLogsDatabaseClient(fakeLogsDatabaseAddress, mockHttpClient)

	ctx := context.Background()
	requestedServiceGuids := map[service.ServiceUUID]bool{
		service.ServiceUUID(testUserService1Guid): true,
	}
	result, err := lokiDbClient.FilterExistingServiceGuids(ctx, testEnclaveId, requestedServiceGuids)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "The logs database returns an error status when fetching the existing service UUIDs. Response was: ")
}

func TestFilterExistingServiceGuids_UnexpectedResponseObjectShape(t *testing.T) {
	mockHttpClient := mocks.NewMockHttpClient(t)

	jsonResponse := `{"UNEXPECTED_JSONS": ""}`
	mockHttpClient.EXPECT().Do(mock.Anything).Return(&http.Response{
		Status:           "",
		StatusCode:       http.StatusOK,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             io.NopCloser(strings.NewReader(jsonResponse)),
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}, nil)

	lokiDbClient := NewLokiLogsDatabaseClient(fakeLogsDatabaseAddress, mockHttpClient)

	ctx := context.Background()
	requestedServiceGuids := map[service.ServiceUUID]bool{
		service.ServiceUUID(testUserService1Guid): true,
	}
	result, err := lokiDbClient.FilterExistingServiceGuids(ctx, testEnclaveId, requestedServiceGuids)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "The logs database returns an error status when fetching the existing service UUIDs. Response was: ")
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func newLokiStreamValueForTest(userServiceGuid service.ServiceUUID, expectedValues [][]string) lokiStreamValue {
	newLokiStreamValue := lokiStreamValue{
		Stream: struct {
			KurtosisContainerType string `json:"comKurtosistechContainerType"`
			KurtosisGUID          string `json:"comKurtosistechGuid"`
		}(struct {
			KurtosisContainerType string
			KurtosisGUID          string
		}{KurtosisContainerType: userServiceContainerType, KurtosisGUID: string(userServiceGuid)}),
		Values: expectedValues,
	}
	return newLokiStreamValue
}
