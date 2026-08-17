package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/C-A1R/exponia/backend/internal/identity"
)

func RequireUser(
	logger *slog.Logger,
	w http.ResponseWriter,
	r *http.Request,
) (int64, bool) {
	userID, ok := identity.UserIDFromContext(r.Context())
	if !ok {
		WriteError(
			logger,
			w,
			http.StatusUnauthorized,
			"authentication required",
		)
		return 0, false
	}

	return userID, true
}
