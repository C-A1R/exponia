package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/C-A1R/exponia/backend/internal/identity"
	"github.com/C-A1R/exponia/backend/internal/user"
)

type ExternalIdentity struct {
	Email       string
	DisplayName string
	Issuer      string
	Subject     string
}

type UserService interface {
	FindOrCreate(
		ctx context.Context,
		email string,
		displayName string,
		authIssuer string,
		authSubject string,
	) (user.User, error)
}

func FixedIdentity(
	externalIdentity ExternalIdentity,
	users UserService,
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		currentUser, err := users.FindOrCreate(
			r.Context(),
			externalIdentity.Email,
			externalIdentity.DisplayName,
			externalIdentity.Issuer,
			externalIdentity.Subject,
		)
		if err != nil {
			logger.Error(
				"failed to resolve authenticated user",
				slog.Any("error", err),
			)
			WriteError(
				logger,
				w,
				http.StatusInternalServerError,
				"internal server error",
			)
			return
		}

		ctx := identity.WithUserID(r.Context(), currentUser.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
