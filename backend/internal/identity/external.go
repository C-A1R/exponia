package identity

import "context"

type ExternalIdentity struct {
	Email       string
	DisplayName string
	Issuer      string
	Subject     string
}

type externalIdentityContextKey struct{}

func WithExternalIdentity(
	ctx context.Context,
	externalIdentity ExternalIdentity,
) context.Context {
	return context.WithValue(
		ctx,
		externalIdentityContextKey{},
		externalIdentity,
	)
}

func ExternalIdentityFromContext(
	ctx context.Context,
) (ExternalIdentity, bool) {
	externalIdentity, ok := ctx.Value(
		externalIdentityContextKey{},
	).(ExternalIdentity)

	return externalIdentity, ok
}
