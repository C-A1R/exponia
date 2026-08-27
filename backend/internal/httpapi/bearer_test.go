package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireBearerToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer test-token")

	token, ok := RequireBearerToken(logger, recorder, request)
	if !ok {
		t.Fatal("expected bearer token")
	}

	if token != "test-token" {
		t.Fatalf("expected token %q, got %q", "test-token", token)
	}
}

func TestRequireBearerTokenRejectsInvalidHeader(t *testing.T) {
	tests := []struct {
		name   string
		header string
	}{
		{name: "missing header"},
		{name: "missing token", header: "Bearer"},
		{name: "wrong scheme", header: "Basic test-token"},
		{name: "too many parts", header: "Bearer first second"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set("Authorization", tt.header)

			_, ok := RequireBearerToken(logger, recorder, request)
			if ok {
				t.Fatal("expected invalid authorization header")
			}

			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf(
					"expected status %d, got %d",
					http.StatusUnauthorized,
					recorder.Code,
				)
			}

			if recorder.Body.String() != "{\"error\":\"bearer token required\"}\n" {
				t.Fatalf("unexpected response body: %q", recorder.Body.String())
			}
		})
	}
}
