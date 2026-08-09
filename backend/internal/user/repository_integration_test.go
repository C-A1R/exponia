package user_test

import (
	"errors"
	"testing"

	"github.com/C-A1R/exponia/backend/internal/testutil"
	"github.com/C-A1R/exponia/backend/internal/user"
)

func TestRepository(t *testing.T) {
	pool := testutil.StartPostgres(t)
	repository := user.NewRepository(pool)
	ctx := t.Context()

	created, err := repository.Create(
		ctx,
		"alex@example.com",
		"Alex",
		"https://auth.example.com",
		"external-user-42",
	)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if created.ID <= 0 {
		t.Fatalf("expected positive user id, got %d", created.ID)
	}

	if created.Email != "alex@example.com" {
		t.Fatalf("expected email %q, got %q", "alex@example.com", created.Email)
	}

	if created.CreatedAt.IsZero() {
		t.Fatal("expected created_at to be set")
	}

	byID, err := repository.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}

	if byID != created {
		t.Fatalf("user fetched by id differs: got %+v, want %+v", byID, created)
	}

	byAuthIdentity, err := repository.GetByAuthIdentity(
		ctx,
		created.AuthIssuer,
		created.AuthSubject,
	)
	if err != nil {
		t.Fatalf("get user by auth identity: %v", err)
	}

	if byAuthIdentity != created {
		t.Fatalf(
			"user fetched by auth identity differs: got %+v, want %+v",
			byAuthIdentity,
			created,
		)
	}
}

func TestRepositoryNotFound(t *testing.T) {
	pool := testutil.StartPostgres(t)
	repository := user.NewRepository(pool)
	ctx := t.Context()

	if _, err := repository.GetByID(ctx, 999999); !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("expected ErrNotFound by id, got %v", err)
	}

	if _, err := repository.GetByAuthIdentity(
		ctx,
		"https://auth.example.com",
		"missing-user",
	); !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("expected ErrNotFound by auth identity, got %v", err)
	}
}
