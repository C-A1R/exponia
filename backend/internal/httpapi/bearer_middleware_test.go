package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/C-A1R/exponia/backend/internal/identity"
)

type fakeTokenVerifier struct {
	externalIdentity identity.ExternalIdentity
	err              error
	rawToken         string
}

func (v *fakeTokenVerifier) Verify(
	_ context.Context,
	rawToken string,
) (identity.ExternalIdentity, error) {
	v.rawToken = rawToken
	return v.externalIdentity, v.err
}

func TestAuthenticateBearer(t *testing.T) {
	expected := identity.ExternalIdentity{
		Email:       "alex@example.com",
		DisplayName: "Alex",
		Issuer:      "https://auth.example.com",
		Subject:     "external-user-42",
	}
	verifier := &fakeTokenVerifier{externalIdentity: expected}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	var actual identity.ExternalIdentity
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actual, _ = identity.ExternalIdentityFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})
	handler := AuthenticateBearer(verifier, logger, next)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	if verifier.rawToken != "test-token" {
		t.Fatalf("expected token %q, got %q", "test-token", verifier.rawToken)
	}

	if actual != expected {
		t.Fatalf("expected identity %#v, got %#v", expected, actual)
	}
}

func TestAuthenticateBearerRejectsInvalidToken(t *testing.T) {
	verifier := &fakeTokenVerifier{err: errors.New("invalid token")}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	nextCalled := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	})
	handler := AuthenticateBearer(verifier, logger, next)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer invalid-token")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}

	if nextCalled {
		t.Fatal("did not expect next handler to be called")
	}

	if recorder.Body.String() != "{\"error\":\"invalid bearer token\"}\n" {
		t.Fatalf("unexpected response body: %q", recorder.Body.String())
	}
}
