package testutil

import (
	"context"
	"testing"
)

func TestStartPostgres(t *testing.T) {
	pool := StartPostgres(t)

	var exists bool

	err := pool.QueryRow(
		context.Background(),
		`
			SELECT EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = 'cameras'
			)
		`,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("check cameras table: %v", err)
	}

	if !exists {
		t.Fatal("expected cameras table to exist")
	}
}
