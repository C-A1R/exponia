package camera

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
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

func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	idValue := r.PathValue("id")
	id, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(
			h.logger,
			w,
			http.StatusBadRequest,
			"invalid camera id",
		)
		return
	}

	camera, err := h.repository.GetCameraById(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(
				h.logger,
				w,
				http.StatusNotFound,
				"camera not found",
			)
			return
		}

		h.logger.Error("failed to get camera by id", slog.Any("error", err))
		httpapi.WriteError(
			h.logger,
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	h.logger.Info(
		"camera retrieved:",
		slog.Int64("camera_id", camera.ID),
		slog.String("manufacturer", camera.Manufacturer),
		slog.String("model", camera.Model),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, camera)
}

type updateCameraRequest struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idValue := r.PathValue("id")
	id, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(
			h.logger,
			w,
			http.StatusBadRequest,
			"invalid camera id",
		)
		return
	}

	var request updateCameraRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid request body")
		return
	}

	request.Manufacturer = strings.TrimSpace(request.Manufacturer)
	request.Model = strings.TrimSpace(request.Model)

	if request.Manufacturer == "" {
		httpapi.WriteError(
			h.logger,
			w,
			http.StatusBadRequest,
			"manufacturer is required",
		)
		return
	}

	if request.Model == "" {
		httpapi.WriteError(
			h.logger,
			w,
			http.StatusBadRequest,
			"model is required",
		)
		return
	}

	camera, err := h.repository.UpdateCamera(
		r.Context(),
		id,
		request.Manufacturer,
		request.Model,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(
				h.logger,
				w,
				http.StatusNotFound,
				"camera not found",
			)
			return
		}

		h.logger.Error("failed to update camera", slog.Any("error", err))

		httpapi.WriteError(
			h.logger,
			w,
			http.StatusInternalServerError,
			"internal server error",
		)

		return
	}

	h.logger.Info(
		"camera updated",
		slog.Int64("camera_id", camera.ID),
		slog.String("manufacturer", camera.Manufacturer),
		slog.String("model", camera.Model),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, camera)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idValue := r.PathValue("id")
	id, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(
			h.logger,
			w,
			http.StatusBadRequest,
			"invalid camera id",
		)
		return
	}

	err = h.repository.DeleteCamera(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(
				h.logger,
				w,
				http.StatusNotFound,
				"camera not found",
			)
			return
		}

		h.logger.Error("failed to delete camera", slog.Any("error", err))
		httpapi.WriteError(
			h.logger,
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	h.logger.Info(
		"camera deleted",
		slog.Int64("camera_id", id),
	)

	w.WriteHeader(http.StatusNoContent)
}
