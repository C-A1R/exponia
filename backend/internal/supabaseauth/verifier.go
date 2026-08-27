package supabaseauth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/C-A1R/exponia/backend/internal/identity"
	"github.com/coreos/go-oidc/v3/oidc"
)

const authenticatedAudience = "authenticated"

var ErrInvalidToken = errors.New("invalid access token")

type Verifier struct {
	verifier *oidc.IDTokenVerifier
}

func NewVerifier(ctx context.Context, issuer string) *Verifier {
	issuer = strings.TrimRight(issuer, "/")
	jwksURL := issuer + "/.well-known/jwks.json"
	keySet := oidc.NewRemoteKeySet(ctx, jwksURL)

	return newVerifier(issuer, keySet)
}

func newVerifier(issuer string, keySet oidc.KeySet) *Verifier {
	return &Verifier{
		verifier: oidc.NewVerifier(
			issuer,
			keySet,
			&oidc.Config{
				ClientID: authenticatedAudience,
				SupportedSigningAlgs: []string{
					oidc.ES256,
					oidc.RS256,
				},
			},
		),
	}
}

type tokenClaims struct {
	Email        string `json:"email"`
	UserMetadata struct {
		Name     string `json:"name"`
		FullName string `json:"full_name"`
	} `json:"user_metadata"`
}

func (v *Verifier) Verify(
	ctx context.Context,
	rawToken string,
) (identity.ExternalIdentity, error) {
	token, err := v.verifier.Verify(ctx, rawToken)
	if err != nil {
		return identity.ExternalIdentity{}, fmt.Errorf(
			"%w: %v",
			ErrInvalidToken,
			err,
		)
	}

	var claims tokenClaims
	if err := token.Claims(&claims); err != nil {
		return identity.ExternalIdentity{}, fmt.Errorf(
			"%w: decode claims: %v",
			ErrInvalidToken,
			err,
		)
	}

	if token.Subject == "" || strings.TrimSpace(claims.Email) == "" {
		return identity.ExternalIdentity{}, fmt.Errorf(
			"%w: subject and email are required",
			ErrInvalidToken,
		)
	}

	displayName := strings.TrimSpace(claims.UserMetadata.FullName)
	if displayName == "" {
		displayName = strings.TrimSpace(claims.UserMetadata.Name)
	}
	if displayName == "" {
		displayName = claims.Email
	}

	return identity.ExternalIdentity{
		Email:       claims.Email,
		DisplayName: displayName,
		Issuer:      token.Issuer,
		Subject:     token.Subject,
	}, nil
}
