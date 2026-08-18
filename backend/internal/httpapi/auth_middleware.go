package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/C-A1R/exponia/backend/internal/identity"
	"github.com/C-A1R/exponia/backend/internal/user"
)

type UserService interface {
	GetByAuthIdentity(
		ctx context.Context,
		authIssuer string,
		authSubject string,
	) (user.User, error)

	FindOrCreate(
		ctx context.Context,
		email string,
		displayName string,
		authIssuer string,
		authSubject string,
	) (user.User, error)
}

func ResolveUser(
	users UserService,
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		externalIdentity, ok := identity.ExternalIdentityFromContext(r.Context())
		if !ok {
			WriteError(logger, w, http.StatusUnauthorized, "authentication required")
			return
		}

		currentUser, err := users.GetByAuthIdentity(
			r.Context(),
			externalIdentity.Issuer,
			externalIdentity.Subject,
		)
		if errors.Is(err, user.ErrNotFound) {
			currentUser, err = users.FindOrCreate(
				r.Context(),
				externalIdentity.Email,
				externalIdentity.DisplayName,
				externalIdentity.Issuer,
				externalIdentity.Subject,
			)
		}
		if err != nil {
			logger.Error("failed to resolve authenticated user", slog.Any("error", err))
			WriteError(logger, w, http.StatusInternalServerError, "internal server error")
			return
		}

		ctx := identity.WithUserID(r.Context(), currentUser.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
