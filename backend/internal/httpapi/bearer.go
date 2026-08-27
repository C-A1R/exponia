package httpapi

import (
	"log/slog"
	"net/http"
	"strings"
)

func RequireBearerToken(
	logger *slog.Logger,
	w http.ResponseWriter,
	r *http.Request,
) (string, bool) {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		WriteError(
			logger,
			w,
			http.StatusUnauthorized,
			"bearer token required",
		)
		return "", false
	}

	return parts[1], true
}
