package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/C-A1R/exponia/backend/internal/buildinfo"
)

func TestHealthHandlerIncludesBuildInfo(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	healthHandler(logger).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	var body healthResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("expected status %q, got %q", "ok", body.Status)
	}

	if body.BuildInfo != buildinfo.Current() {
		t.Fatalf("got build info %+v, want %+v", body.BuildInfo, buildinfo.Current())
	}
}
