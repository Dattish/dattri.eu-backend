package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMonitoring(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()

	handler := Monitoring(time.Now())

	handler.ServeHTTP(recorder, req)

	if code := recorder.Code; code != http.StatusOK {
		t.Errorf("Incorrect status returned: got %v want %v", code, http.StatusOK)
	}

	if body := recorder.Body.String(); body == "<nil>" {
		t.Error("Monitoring returned empty body")
	}
}