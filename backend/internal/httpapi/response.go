package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(
	logger *slog.Logger,
	w http.ResponseWriter,
	status int,
	value any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		logger.Error(
			"failed to encode JSON response",
			slog.Any("error", err),
		)
	}

}

func WriteError(
	logger *slog.Logger,
	w http.ResponseWriter,
	status int,
	message string,
) {
	WriteJSON(logger, w, status, ErrorResponse{
		Error: message,
	})
}
