package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/C-A1R/exponia/backend/internal/identity"
)

type TokenVerifier interface {
	Verify(
		ctx context.Context,
		rawToken string,
	) (identity.ExternalIdentity, error)
}

func AuthenticateBearer(
	verifier TokenVerifier,
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawToken, ok := RequireBearerToken(logger, w, r)
		if !ok {
			return
		}

		externalIdentity, err := verifier.Verify(r.Context(), rawToken)
		if err != nil {
			logger.Warn(
				"failed to verify bearer token",
				slog.Any("error", err),
			)
			WriteError(
				logger,
				w,
				http.StatusUnauthorized,
				"invalid bearer token",
			)
			return
		}

		ctx := identity.WithExternalIdentity(
			r.Context(),
			externalIdentity,
		)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
