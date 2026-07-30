package camera

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

type Handler struct {
	repository *Repository
	logger     *slog.Logger
}

type createCameraRequest struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHandler(
	repository *Repository,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		repository: repository,
		logger:     logger,
	}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		h.logger.Error(
			"failed to encode JSON response",
			slog.Any("error", err),
		)
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request createCameraRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		h.writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "invalid request body",
		})
		return
	}

	request.Manufacturer = strings.TrimSpace(request.Manufacturer)
	request.Model = strings.TrimSpace(request.Model)

	if request.Manufacturer == "" {
		h.writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "manufacturer is required",
		})
		return
	}

	if request.Model == "" {
		h.writeJSON(w, http.StatusBadRequest, errorResponse{
			Error: "model is required",
		})
		return
	}

	camera, err := h.repository.CreateCamera(
		r.Context(),
		request.Manufacturer,
		request.Model,
	)

	if err != nil {
		h.logger.Error("failed to create camera", slog.Any("error", err))

		h.writeJSON(w, http.StatusInternalServerError, errorResponse{
			Error: "internal server error",
		})
		return
	}

	h.writeJSON(w, http.StatusCreated, camera)
}
