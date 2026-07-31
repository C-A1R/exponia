package camera

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/C-A1R/exponia/backend/internal/httpapi"
)

type Handler struct {
	repository *Repository
	logger     *slog.Logger
}

type createCameraRequest struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request createCameraRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid request body")
		return
	}

	request.Manufacturer = strings.TrimSpace(request.Manufacturer)
	request.Model = strings.TrimSpace(request.Model)

	if request.Manufacturer == "" {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "manufacturer is required")
		return
	}

	if request.Model == "" {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "model is required")
		return
	}

	camera, err := h.repository.CreateCamera(
		r.Context(),
		request.Manufacturer,
		request.Model,
	)

	if err != nil {
		h.logger.Error("failed to create camera", slog.Any("error", err))

		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info(
		"camera created:",
		slog.Int64("camera_id", camera.ID),
		slog.String("manufacturer", camera.Manufacturer),
		slog.String("model", camera.Model),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusCreated, camera)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	cameras, err := h.repository.ListCameras(r.Context())
	if err != nil {
		h.logger.Error("failed to list cameras", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info(
		"cameras listed:",
		slog.Int("count", len(cameras)),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, cameras)
}
