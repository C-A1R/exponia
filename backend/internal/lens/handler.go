package lens

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
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

type createLensRequest struct {
	Manufacturer  string  `json:"manufacturer"`
	Model         string  `json:"model"`
	FocalLengthMm int32   `json:"focal_length_mm"`
	MaxAperture   float64 `json:"max_aperture"`
}

type updateLensRequest struct {
	Manufacturer  string  `json:"manufacturer"`
	Model         string  `json:"model"`
	FocalLengthMm int32   `json:"focal_length_mm"`
	MaxAperture   float64 `json:"max_aperture"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request createLensRequest

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

	if request.FocalLengthMm <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "focal_length_mm must be greater than zero")
		return
	}

	if request.MaxAperture <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "max_aperture must be greater than zero")
		return
	}

	lens, err := h.service.Create(
		r.Context(),
		userID,
		request.Manufacturer,
		request.Model,
		request.FocalLengthMm,
		request.MaxAperture,
	)
	if err != nil {
		h.logger.Error("failed to create lens", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info("lens created", slog.Int64("lens_id", lens.ID))

	httpapi.WriteJSON(h.logger, w, http.StatusCreated, lens)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	lenses, err := h.service.List(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list lenses", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpapi.WriteJSON(h.logger, w, http.StatusOK, lenses)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	idValue := r.PathValue("id")

	id, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid lens id")
		return
	}

	lens, err := h.service.GetByID(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "lens not found")
			return
		}

		h.logger.Error("failed to get lens", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	httpapi.WriteJSON(h.logger, w, http.StatusOK, lens)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	idValue := r.PathValue("id")

	id, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid lens id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request updateLensRequest

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

	if request.FocalLengthMm <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "focal_length_mm must be greater than zero")
		return
	}

	if request.MaxAperture <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "max_aperture must be greater than zero")
		return
	}

	lens, err := h.service.Update(
		r.Context(),
		userID,
		id,
		request.Manufacturer,
		request.Model,
		request.FocalLengthMm,
		request.MaxAperture,
	)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "lens not found")
			return
		}

		h.logger.Error("failed to update lens", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info("lens updated", slog.Int64("lens_id", id))

	httpapi.WriteJSON(h.logger, w, http.StatusOK, lens)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	idValue := r.PathValue("id")

	id, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid lens id")
		return
	}

	if err := h.service.Delete(r.Context(), userID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "lens not found")
			return
		}

		h.logger.Error("failed to delete lens", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info("lens deleted", slog.Int64("lens_id", id))

	w.WriteHeader(http.StatusNoContent)
}
