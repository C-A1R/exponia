package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/C-A1R/exponia/backend/internal/identity"
)

func TestRequireUser(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request = request.WithContext(
		identity.WithUserID(request.Context(), 42),
	)

	userID, ok := RequireUser(logger, recorder, request)
	if !ok {
		t.Fatal("expected authenticated request")
	}

	if userID != 42 {
		t.Fatalf("expected user id 42, got %d", userID)
	}
}

func TestRequireUserWithoutUser(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	_, ok := RequireUser(logger, recorder, request)
	if ok {
		t.Fatal("expected unauthenticated request")
	}

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}

	if recorder.Body.String() != "{\"error\":\"authentication required\"}\n" {
		t.Fatalf("unexpected response body: %q", recorder.Body.String())
	}
}
