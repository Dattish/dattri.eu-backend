package handler

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestHttpsRedirectHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler := HttpsRedirect()

	handler.ServeHTTP(recorder, req)

	if status := recorder.Code; status != http.StatusMovedPermanently {
		t.Errorf("Handler did not return correct status: got %v want %v",
			status, http.StatusMovedPermanently)
	}

	httpsPrefix := "https://"
	if location := recorder.HeaderMap.Get("location"); !strings.HasPrefix(location, httpsPrefix) {
		t.Errorf("Handler did not return correct location prefix: got %v want %v",
			location, httpsPrefix)
	}
}

func TestFileEndpointFromConfig(t *testing.T) {
	jsonBytes := []byte(
		`[{"endpoint" : "/test", "resourceType" : "file", "path" : "./test/test.txt"}]`)
	serveMux := http.NewServeMux()
	err := EndpointsFromConfig(serveMux, jsonBytes)

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	serveMux.ServeHTTP(recorder, req)

	expectedBody := "hello!"
	if body := recorder.Body; body.String() != expectedBody {
		t.Errorf("Handler did not return correct body: got %v want %v",
			body, expectedBody)
	}
}

func TestDirectoryEndpointFromConfig(t *testing.T) {
	jsonBytes := []byte(
		`[{"endpoint" : "/test/", "resourceType" : "directory", "path" : "./test"}]`)
	serveMux := http.NewServeMux()
	err := EndpointsFromConfig(serveMux, jsonBytes)

	req, err := http.NewRequest("GET", "/test/test.txt", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	serveMux.ServeHTTP(recorder, req)

	expectedBody := "hello!"
	if body := recorder.Body; body.String() != expectedBody {
		t.Errorf("Handler did not return correct body: got %v want %v",
			body, expectedBody)
	}
}

func TestEndpointsFromConfigResourceTypeError(t *testing.T) {
	invalidJsonBytes := []byte(
		`[{"endpoint" : "/test", "resourceType" : "INVALID", "path" : "/test"}]`)
	err := EndpointsFromConfig(nil, invalidJsonBytes)

	if err == nil {
		t.Fatal("Didn't get an error")
	}

	_, ok := err.(*endpointError)

	if !ok {
		t.Errorf("Did not get expected type of error: got %v want *EndpointError", reflect.TypeOf(err).Kind())
	}

	expectedErrorMessage := "No such resourceType: INVALID"
	if err.Error() != expectedErrorMessage {
		t.Errorf("Did not get expected error message: got %v want %v", err.Error(), expectedErrorMessage)
	}
}

func TestEndpointsFromConfigInvalidFileError(t *testing.T) {
	invalidFileJsonBytes := []byte(
		`[{"endpoint" : "/test", "resourceType" : "file", "path" : "./test/nonexistent.txt"}]`)
	err := EndpointsFromConfig(nil, invalidFileJsonBytes)

	if err == nil {
		t.Fatal("Didn't get an error")
	}

	_, ok := err.(*endpointError)

	if !ok {
		t.Errorf("Did not get expected type of error: got %v want *EndpointError", reflect.TypeOf(err).Kind())
	}

	expectedErrorMessage := "CreateFile ./test/nonexistent.txt: The system cannot find the file specified."
	if err.Error() != expectedErrorMessage {
		t.Errorf("Did not get expected error message: got %v want %v", err.Error(), expectedErrorMessage)
	}
}

func TestEndpointsFromConfigInvalidDirectoryError(t *testing.T) {
	invalidDirectoryJsonBytes := []byte(
		`[{"endpoint" : "/test/", "resourceType" : "directory", "path" : "./nonexistent"}]`)
	err := EndpointsFromConfig(nil, invalidDirectoryJsonBytes)

	if err == nil {
		t.Fatal("Didn't get an error")
	}

	_, ok := err.(*endpointError)

	if !ok {
		t.Errorf("Did not get expected type of error: got %v want *EndpointError", reflect.TypeOf(err).Kind())
	}

	expectedErrorMessage := "CreateFile ./nonexistent: The system cannot find the file specified."
	if err.Error() != expectedErrorMessage {
		t.Errorf("Did not get expected error message: got %v want %v", err.Error(), expectedErrorMessage)
	}
}
