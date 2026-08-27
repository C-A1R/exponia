package identity

import (
	"context"
	"testing"
)

func TestExternalIdentityFromContext(t *testing.T) {
	expected := ExternalIdentity{
		Email:       "alex@example.com",
		DisplayName: "Alex",
		Issuer:      "https://auth.example.com",
		Subject:     "external-user-42",
	}
	ctx := WithExternalIdentity(context.Background(), expected)

	actual, ok := ExternalIdentityFromContext(ctx)
	if !ok {
		t.Fatal("expected external identity in context")
	}

	if actual != expected {
		t.Fatalf("expected identity %#v, got %#v", expected, actual)
	}
}

func TestExternalIdentityFromContextWithoutIdentity(t *testing.T) {
	if _, ok := ExternalIdentityFromContext(context.Background()); ok {
		t.Fatal("did not expect external identity in context")
	}
}
