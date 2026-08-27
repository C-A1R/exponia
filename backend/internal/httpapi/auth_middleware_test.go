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
	user               user.User
	getErr             error
	findOrCreateErr    error
	findOrCreateCalled bool
}

func (s *fakeUserService) GetByAuthIdentity(
	_ context.Context,
	_ string,
	_ string,
) (user.User, error) {
	return s.user, s.getErr
}

func (s *fakeUserService) FindOrCreate(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ string,
) (user.User, error) {
	s.findOrCreateCalled = true
	return s.user, s.findOrCreateErr
}

func TestResolveExistingUser(t *testing.T) {
	users := &fakeUserService{
		user: user.User{ID: 42},
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	var actualUserID int64
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualUserID, _ = identity.UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	handler := ResolveUser(users, logger, next)

	externalIdentity := identity.ExternalIdentity{
		Email:       "alex@example.com",
		DisplayName: "Alex",
		Issuer:      "https://auth.example.com",
		Subject:     "external-user-42",
	}
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request = request.WithContext(
		identity.WithExternalIdentity(
			request.Context(),
			externalIdentity,
		),
	)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	if actualUserID != 42 {
		t.Fatalf("expected user id 42, got %d", actualUserID)
	}

	if users.findOrCreateCalled {
		t.Fatal("did not expect existing user to be created")
	}
}

func TestResolveUserCreatesMissingUser(t *testing.T) {
	users := &fakeUserService{
		user:   user.User{ID: 42},
		getErr: user.ErrNotFound,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	var actualUserID int64
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualUserID, _ = identity.UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})
	handler := ResolveUser(users, logger, next)

	request := requestWithExternalIdentity()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, recorder.Code)
	}

	if !users.findOrCreateCalled {
		t.Fatal("expected missing user to be created")
	}

	if actualUserID != 42 {
		t.Fatalf("expected user id 42, got %d", actualUserID)
	}
}

func TestResolveUserReturnsError(t *testing.T) {
	users := &fakeUserService{
		getErr: errors.New("database unavailable"),
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	nextCalled := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	})

	handler := ResolveUser(users, logger, next)

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request = request.WithContext(
		identity.WithExternalIdentity(
			request.Context(),
			identity.ExternalIdentity{
				Email:       "alex@example.com",
				DisplayName: "Alex",
				Issuer:      "https://auth.example.com",
				Subject:     "external-user-42",
			},
		),
	)

	recorder := httptest.NewRecorder()
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

	if users.findOrCreateCalled {
		t.Fatal("did not expect user creation after database error")
	}
}

func requestWithExternalIdentity() *http.Request {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	return request.WithContext(
		identity.WithExternalIdentity(
			request.Context(),
			identity.ExternalIdentity{
				Email:       "alex@example.com",
				DisplayName: "Alex",
				Issuer:      "https://auth.example.com",
				Subject:     "external-user-42",
			},
		),
	)
}

func TestResolveUserWithoutExternalIdentity(t *testing.T) {
	users := &fakeUserService{user: user.User{ID: 42}}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	nextCalled := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	})
	handler := ResolveUser(users, logger, next)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
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
}
