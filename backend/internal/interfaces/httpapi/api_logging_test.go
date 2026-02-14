package httpapi

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler_LogsIncomingRequests(t *testing.T) {
	var buf bytes.Buffer
	a := New(nil, nil, nil, nil, nil, nil, Config{})
	a.logger = log.New(&buf, "", 0)

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rec := httptest.NewRecorder()
	a.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	logged := buf.String()
	for _, want := range []string{"http request", "method=GET", "path=/api/ping", "status=200"} {
		if !strings.Contains(logged, want) {
			t.Fatalf("log output %q does not contain %q", logged, want)
		}
	}
}

func TestWriteErr_LogsError(t *testing.T) {
	var buf bytes.Buffer
	a := New(nil, nil, nil, nil, nil, nil, Config{})
	a.logger = log.New(&buf, "", 0)

	rec := httptest.NewRecorder()
	a.writeErr(rec, errors.New("boom"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(buf.String(), "http error") || !strings.Contains(buf.String(), "boom") {
		t.Fatalf("unexpected log output: %q", buf.String())
	}
}
