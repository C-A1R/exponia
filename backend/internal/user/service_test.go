package user

import (
	"context"
	"errors"
	"testing"
)

type fakeUserRepository struct {
	user            User
	getErr          error
	findOrCreateErr error
}

func (r *fakeUserRepository) GetByAuthIdentity(
	_ context.Context,
	_ string,
	_ string,
) (User, error) {
	return r.user, r.getErr
}

func (r *fakeUserRepository) FindOrCreate(
	_ context.Context,
	_ string,
	_ string,
	_ string,
	_ string,
) (User, error) {
	return r.user, r.findOrCreateErr
}

func TestGetByAuthIdentityReturnsUser(t *testing.T) {
	expected := User{ID: 42, Email: "alex@example.com"}
	repository := &fakeUserRepository{user: expected}
	service := NewService(repository)

	actual, err := service.GetByAuthIdentity(
		t.Context(),
		"https://auth.example.com",
		"external-user-42",
	)
	if err != nil {
		t.Fatalf("get user by auth identity: %v", err)
	}

	if actual != expected {
		t.Fatalf("got %+v, want %+v", actual, expected)
	}
}

func TestFindOrCreateReturnsUser(t *testing.T) {
	expected := User{ID: 42, Email: "alex@example.com"}
	repository := &fakeUserRepository{user: expected}
	service := NewService(repository)

	actual, err := service.FindOrCreate(
		t.Context(),
		"alex@example.com",
		"Alex",
		"https://auth.example.com",
		"external-user-42",
	)
	if err != nil {
		t.Fatalf("find or create user: %v", err)
	}

	if actual != expected {
		t.Fatalf("got %+v, want %+v", actual, expected)
	}
}

func TestFindOrCreateReturnsRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database unavailable")
	repository := &fakeUserRepository{findOrCreateErr: repositoryErr}
	service := NewService(repository)

	_, err := service.FindOrCreate(
		t.Context(),
		"alex@example.com",
		"Alex",
		"https://auth.example.com",
		"external-user-42",
	)
	if !errors.Is(err, repositoryErr) {
		t.Fatalf("expected repository error, got %v", err)
	}
}
