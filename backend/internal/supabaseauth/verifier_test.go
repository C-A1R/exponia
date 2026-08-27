package supabaseauth

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/coreos/go-oidc/v3/oidc/oidctest"
)

const testIssuer = "https://project.supabase.co/auth/v1"

func TestVerify(t *testing.T) {
	privateKey := generatePrivateKey(t)
	verifier := newVerifier(
		testIssuer,
		&oidc.StaticKeySet{
			PublicKeys: []crypto.PublicKey{privateKey.Public()},
		},
	)

	rawToken := signToken(t, privateKey, map[string]any{
		"iss":   testIssuer,
		"aud":   authenticatedAudience,
		"sub":   "external-user-42",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"email": "alex@example.com",
		"user_metadata": map[string]any{
			"full_name": "Alex",
		},
	})

	actual, err := verifier.Verify(context.Background(), rawToken)
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}

	if actual.Issuer != testIssuer {
		t.Fatalf("expected issuer %q, got %q", testIssuer, actual.Issuer)
	}

	if actual.Subject != "external-user-42" {
		t.Fatalf("expected subject %q, got %q", "external-user-42", actual.Subject)
	}

	if actual.Email != "alex@example.com" {
		t.Fatalf("expected email %q, got %q", "alex@example.com", actual.Email)
	}

	if actual.DisplayName != "Alex" {
		t.Fatalf("expected display name %q, got %q", "Alex", actual.DisplayName)
	}
}

func TestVerifyRejectsExpiredToken(t *testing.T) {
	privateKey := generatePrivateKey(t)
	verifier := newVerifier(
		testIssuer,
		&oidc.StaticKeySet{
			PublicKeys: []crypto.PublicKey{privateKey.Public()},
		},
	)

	rawToken := signToken(t, privateKey, map[string]any{
		"iss":   testIssuer,
		"aud":   authenticatedAudience,
		"sub":   "external-user-42",
		"exp":   time.Now().Add(-time.Hour).Unix(),
		"email": "alex@example.com",
	})

	if _, err := verifier.Verify(
		context.Background(),
		rawToken,
	); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestVerifyRejectsWrongAudience(t *testing.T) {
	privateKey := generatePrivateKey(t)
	verifier := newVerifier(
		testIssuer,
		&oidc.StaticKeySet{
			PublicKeys: []crypto.PublicKey{privateKey.Public()},
		},
	)

	rawToken := signToken(t, privateKey, map[string]any{
		"iss":   testIssuer,
		"aud":   "anon",
		"sub":   "external-user-42",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"email": "alex@example.com",
	})

	if _, err := verifier.Verify(
		context.Background(),
		rawToken,
	); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func generatePrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	return privateKey
}

func signToken(
	t *testing.T,
	privateKey *rsa.PrivateKey,
	claims map[string]any,
) string {
	t.Helper()

	rawClaims, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal claims: %v", err)
	}

	return oidctest.SignIDToken(
		privateKey,
		"test-key",
		oidc.RS256,
		string(rawClaims),
	)
}
