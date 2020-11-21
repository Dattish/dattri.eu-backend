package handler

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"alive": true}`)
}

func TestCORS(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	allowedMethods := "GET"
	allowedOrigin := "localhost"
	handler := CORS(allowedMethods, allowedOrigin, []string{}, http.HandlerFunc(testHandler))

	handler.ServeHTTP(recorder, req)

	if methods := recorder.Header().Get("Access-Control-Allow-Methods"); methods != allowedMethods {
		t.Errorf("Handler did not return correct Access-Control-Allow-Methods: got %v want %v",
			methods, allowedMethods)
	}

	if origin := recorder.Header().Get("Access-Control-Allow-Origin"); origin != allowedOrigin {
		t.Errorf("Handler did not return correct Access-Control-Allow-Origin: got %v want %v",
			origin, allowedOrigin)
	}
}

func TestCSP(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	expectedPolicy := "default-src 'self'"
	handler := CSP(expectedPolicy, []string{}, http.HandlerFunc(testHandler))

	handler.ServeHTTP(recorder, req)

	if policy := recorder.Header().Get("Content-Security-Policy"); policy != expectedPolicy {
		t.Errorf("Handler did not return correct Access-Control-Allow-Methods: got %v want %v",
			policy, expectedPolicy)
	}

}

func TestLogger(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	var logBuffer bytes.Buffer
	log.SetOutput(&logBuffer)

	recorder := httptest.NewRecorder()

	handler := Logging(http.HandlerFunc(testHandler))

	handler.ServeHTTP(recorder, req)

	expectedLoggingWithoutTimestamp := "[] GET: /test | 0s\n"
	actualLogging := logBuffer.String()
	if !strings.HasSuffix(actualLogging, expectedLoggingWithoutTimestamp) {
		t.Errorf("Handler did not log as expected: got %v want %v, ignoring the timestamp",
			actualLogging, expectedLoggingWithoutTimestamp)
	}
}
