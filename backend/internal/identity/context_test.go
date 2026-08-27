package identity

import (
	"context"
	"testing"
)

func TestUserIDFromContext(t *testing.T) {
	ctx := WithUserID(context.Background(), 42)

	userID, ok := UserIDFromContext(ctx)
	if !ok {
		t.Fatal("expected user id in context")
	}

	if userID != 42 {
		t.Fatalf("expected user id 42, got %d", userID)
	}
}

func TestUserIDFromContextRejectsMissingOrInvalidValue(t *testing.T) {
	testCases := []struct {
		name string
		ctx  context.Context
	}{
		{name: "missing", ctx: context.Background()},
		{name: "zero", ctx: WithUserID(context.Background(), 0)},
		{name: "negative", ctx: WithUserID(context.Background(), -1)},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, ok := UserIDFromContext(testCase.ctx); ok {
				t.Fatal("expected user id lookup to fail")
			}
		})
	}
}
