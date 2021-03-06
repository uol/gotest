package http_test

import (
	"math"
	"net/http"
	"strings"
	"testing"
	"time"

	randomdata "github.com/Pallinder/go-randomdata"
	"github.com/stretchr/testify/assert"
	gotesthttp "github.com/uol/gotest/http"
)

/**
* The tests for the http server used by tests.
* @author rnojiri
**/

var defaultConf gotesthttp.Configuration = gotesthttp.Configuration{
	Host:        "localhost",
	Port:        18080,
	ChannelSize: 5,
}

// createDummyResponse - creates a dummy response data
func createDummyResponse() gotesthttp.ResponseData {

	headers := http.Header{}
	headers.Add("Content-type", "text/plain; charset=utf-8")
	headers.Add("X-custom", randomdata.Adjective())

	return gotesthttp.ResponseData{
		RequestData: gotesthttp.RequestData{
			URI:     "/" + strings.ToLower(randomdata.SillyName()),
			Body:    randomdata.City(),
			Method:  randomdata.StringSample("GET", "POST", "PUT"),
			Headers: headers,
		},
		Status: http.StatusOK,
	}
}

// Test404 - tests when a non mapped response is called
func Test404(t *testing.T) {

	defaultConf.Responses = map[string][]gotesthttp.ResponseData{
		"default": {createDummyResponse()},
	}

	server := gotesthttp.NewServer(&defaultConf)
	defer server.Close()

	response := gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, &gotesthttp.RequestData{
		URI:    "/not",
		Method: "GET",
	})

	assert.Equal(t, http.StatusNotFound, response.Status, "expected 404 status")

	response = gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, &gotesthttp.RequestData{
		URI:    "/test",
		Method: "POST",
	})

	assert.Equal(t, http.StatusNotFound, response.Status, "expected 404 status")

	response = gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, &gotesthttp.RequestData{
		URI:    defaultConf.Responses["default"][0].URI,
		Method: defaultConf.Responses["default"][0].Method,
	})

	assert.Equal(t, http.StatusOK, response.Status, "expected 200 status")
}

// TestSuccess - tests when everything goes right
func TestSuccess(t *testing.T) {

	defaultConf.Responses = map[string][]gotesthttp.ResponseData{
		"default": {createDummyResponse()},
	}

	server := gotesthttp.NewServer(&defaultConf)
	defer server.Close()

	clientRequest := &gotesthttp.RequestData{
		URI:     defaultConf.Responses["default"][0].URI,
		Body:    defaultConf.Responses["default"][0].Body,
		Method:  defaultConf.Responses["default"][0].Method,
		Headers: defaultConf.Responses["default"][0].Headers,
	}

	serverResponse := gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, clientRequest)
	if !compareResponses(t, &defaultConf.Responses["default"][0], serverResponse) {
		return
	}

	serverRequest := gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest, serverRequest)
}

// TestMultipleResponses - tests when everything goes right with multiple responses
func TestMultipleResponses(t *testing.T) {

	configuredResponse1 := createDummyResponse()
	configuredResponse1.URI = "/text"
	configuredResponse1.Method = "POST"

	configuredResponse2 := createDummyResponse()
	configuredResponse2.URI = "/json"
	configuredResponse2.Method = "PUT"
	configuredResponse2.Status = http.StatusCreated
	configuredResponse2.Body = `{"metric": "test-metric", "value": 1.0}`
	configuredResponse2.Headers.Del("Content-type")
	configuredResponse2.Headers.Set("Content-type", "application/json")

	defaultConf.Responses = map[string][]gotesthttp.ResponseData{
		"default": {configuredResponse1, configuredResponse2},
	}

	server := gotesthttp.NewServer(&defaultConf)
	defer server.Close()

	clientRequest1 := &gotesthttp.RequestData{
		URI:     configuredResponse1.URI,
		Body:    configuredResponse1.Body,
		Method:  configuredResponse1.Method,
		Headers: configuredResponse1.Headers,
	}

	serverResponse := gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, clientRequest1)
	if !compareResponses(t, &configuredResponse1, serverResponse) {
		return
	}

	serverRequest := gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest1, serverRequest)

	clientRequest2 := &gotesthttp.RequestData{
		URI:     configuredResponse2.URI,
		Body:    configuredResponse2.Body,
		Method:  configuredResponse2.Method,
		Headers: configuredResponse2.Headers,
	}

	serverResponse = gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, clientRequest2)
	if !compareResponses(t, &configuredResponse2, serverResponse) {
		return
	}

	serverRequest = gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest2, serverRequest)
}

// compareResponses - compares two responses
func compareResponses(t *testing.T, r1 *gotesthttp.ResponseData, r2 *gotesthttp.ResponseData) bool {

	result := true

	result = result && assert.Equal(t, r1.Body, r2.Body, "same body expected")
	result = result && containsHeaders(t, r1.Headers, r2.Headers)
	result = result && assert.Equal(t, r1.Method, r2.Method, "same method expected")
	result = result && assert.Equal(t, r1.Status, r2.Status, "same status expected")
	result = result && assert.Equal(t, defaultConf.Host, r2.Host, "same URI expected")
	result = result && assert.Equal(t, defaultConf.Port, r2.Port, "same URI expected")

	return result
}

// compareRequests - compares two requests
func compareRequests(t *testing.T, r1 *gotesthttp.RequestData, r2 *gotesthttp.RequestData) bool {

	result := true

	result = result && assert.Equal(t, r1.Body, r2.Body, "same body expected")
	result = result && containsHeaders(t, r1.Headers, r2.Headers)
	result = result && assert.Equal(t, r1.Method, r2.Method, "same method expected")
	result = result && assert.Equal(t, r1.URI, r2.URI, "same URI expected")
	result = result && assert.Equal(t, defaultConf.Host, r2.Host, "same URI expected")
	result = result && assert.Equal(t, defaultConf.Port, r2.Port, "same URI expected")

	return result
}

// containsHeaders - checks for the headers
func containsHeaders(t *testing.T, mustExist, fullSet http.Header) bool {

	if mustExist == nil {
		return true
	}

	assert.NotNil(t, fullSet, "the full set of headers must not be null")

	for mustExistHeader, mustExistValues := range mustExist {

		if !assert.Truef(t, len(fullSet[mustExistHeader]) > 0, "expected a list of values for the header: %s", mustExistHeader) {
			return false
		}

		if !assert.Equal(t, fullSet[mustExistHeader], mustExistValues, "expected some headers") {
			return false
		}
	}

	return true
}

// TestSuccessMultiModes - tests when everything goes right
func TestSuccessMultiModes(t *testing.T) {

	r1 := createDummyResponse()
	r2 := createDummyResponse()

	defaultConf.Responses = map[string][]gotesthttp.ResponseData{
		"mode1": {r1},
		"mode2": {r2},
	}

	server := gotesthttp.NewServer(&defaultConf)
	defer server.Close()

	reqHeader := http.Header{}
	reqHeader.Add("Content-type", "text/plain; charset=utf-8")

	clientRequest1 := &gotesthttp.RequestData{
		URI:     r1.URI,
		Body:    r1.Body,
		Method:  r1.Method,
		Headers: r1.Headers,
	}

	clientRequest2 := &gotesthttp.RequestData{
		URI:     r2.URI,
		Body:    r2.Body,
		Method:  r2.Method,
		Headers: r2.Headers,
	}

	server.SetMode("mode2")

	serverResponse := gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, clientRequest2)
	if !compareResponses(t, &defaultConf.Responses["mode2"][0], serverResponse) {
		return
	}

	serverRequest := gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest2, serverRequest)

	server.SetMode("mode1")

	serverResponse = gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, clientRequest1)
	if !compareResponses(t, &defaultConf.Responses["mode1"][0], serverResponse) {
		return
	}

	serverRequest = gotesthttp.WaitForServerRequest(server, time.Second, 10*time.Second)
	compareRequests(t, clientRequest1, serverRequest)
}

// TestWaitResponse - tests the wait parameter
func TestWaitResponse(t *testing.T) {

	randomSeconds := randomdata.Number(1, 5)

	resp := createDummyResponse()
	resp.Wait = time.Duration(randomSeconds) * time.Second

	defaultConf.Responses = map[string][]gotesthttp.ResponseData{
		"default": {resp},
	}

	server := gotesthttp.NewServer(&defaultConf)
	defer server.Close()

	clientRequest := &gotesthttp.RequestData{
		URI:     defaultConf.Responses["default"][0].URI,
		Body:    defaultConf.Responses["default"][0].Body,
		Method:  defaultConf.Responses["default"][0].Method,
		Headers: defaultConf.Responses["default"][0].Headers,
	}

	start := time.Now()

	serverResponse := gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, clientRequest)
	if !compareResponses(t, &defaultConf.Responses["default"][0], serverResponse) {
		return
	}

	requestTime := time.Since(start)

	if !assert.EqualValues(t, float64(randomSeconds), math.Floor(requestTime.Seconds()), "expected same amount of time") {
		return
	}

	serverRequest := gotesthttp.WaitForServerRequest(server, time.Duration(randomSeconds+1)*time.Second, 10*time.Second)
	compareRequests(t, clientRequest, serverRequest)
}
