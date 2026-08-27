package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/C-A1R/exponia/backend/internal/identity"
	"github.com/C-A1R/exponia/backend/internal/user"
)

func TestAuthenticationChain(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	verifier := &fakeTokenVerifier{
		externalIdentity: identity.ExternalIdentity{
			Email:       "alex@example.com",
			DisplayName: "Alex",
			Issuer:      "https://auth.example.com",
			Subject:     "external-user-42",
		},
	}
	users := &fakeUserService{
		user: user.User{ID: 42},
	}

	var actualUserID int64
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := RequireUser(logger, w, r)
		if !ok {
			return
		}

		actualUserID = userID
		w.WriteHeader(http.StatusNoContent)
	})

	handler := AuthenticateBearer(
		verifier,
		logger,
		ResolveUser(users, logger, finalHandler),
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer test-token")
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNoContent,
			recorder.Code,
		)
	}

	if actualUserID != 42 {
		t.Fatalf("expected user id 42, got %d", actualUserID)
	}
}
