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
	"github.com/C-A1R/exponia/backend/internal/user"
)

type fakeUserService struct {
	user user.User
	err  error
}

func (s *fakeUserService) FindOrCreate(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ string,
) (user.User, error) {
	return s.user, s.err
}

func TestFixedIdentity(t *testing.T) {
	users := &fakeUserService{
		user: user.User{ID: 42},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	var actualUserID int64
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualUserID, _ = identity.UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	handler := FixedIdentity(
		identity.ExternalIdentity{
			Email:       "alex@example.com",
			DisplayName: "Alex",
			Issuer:      "https://auth.example.com",
			Subject:     "external-user-42",
		},
		users,
		logger,
		next,
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	if actualUserID != 42 {
		t.Fatalf("expected user id 42, got %d", actualUserID)
	}
}

func TestFixedIdentityReturnsError(t *testing.T) {
	users := &fakeUserService{
		err: errors.New("database unavailable"),
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	nextCalled := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	})

	handler := FixedIdentity(
		identity.ExternalIdentity{},
		users,
		logger,
		next,
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}

	if nextCalled {
		t.Fatal("did not expect next handler to be called")
	}
}
