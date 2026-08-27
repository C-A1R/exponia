package filmroll

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

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

type createFilmRollRequest struct {
	FilmStockID int64 `json:"film_stock_id"`
	FormatID    int64 `json:"format_id"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	var req createFilmRollRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FilmStockID <= 0 || req.FormatID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid film stock or format")
		return
	}

	roll, err := h.service.Create(
		r.Context(),
		userID,
		req.FilmStockID,
		req.FormatID,
	)
	if err != nil {
		h.logger.Error("failed to create film roll", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info(
		"film roll created",
		slog.Int64("film_roll_id", roll.ID),
		slog.Int64("film_stock_id", roll.FilmStock.ID),
		slog.Int64("format_id", roll.Format.ID),
		slog.String("status", string(roll.Status)),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusCreated, roll)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	rolls, err := h.service.List(r.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list film rolls", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info("film rolls listed", slog.Int("count", len(rolls)))

	httpapi.WriteJSON(h.logger, w, http.StatusOK, rolls)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid film roll id")
		return
	}

	roll, err := h.service.GetByID(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "film roll not found")
			return
		}

		h.logger.Error("failed to get film roll", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info(
		"film roll retrieved",
		slog.Int64("film_roll_id", roll.ID),
		slog.Int64("film_stock_id", roll.FilmStock.ID),
		slog.Int64("format_id", roll.Format.ID),
		slog.String("status", string(roll.Status)),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, roll)
}

type updateStatusRequest struct {
	Status FilmRollStatus `json:"status"`
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid film roll id")
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid request body")
		return
	}

	roll, err := h.service.UpdateStatus(
		r.Context(),
		userID,
		id,
		req.Status,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidStatus):
			httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid film roll status")
		case errors.Is(err, ErrNotFound):
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "film roll not found")
		default:
			h.logger.Error("failed to update film roll status", slog.Any("error", err))
			httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.logger.Info(
		"film roll status updated",
		slog.Int64("film_roll_id", roll.ID),
		slog.String("status", string(roll.Status)),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, roll)
}

type updateExposureISORequest struct {
	ExposureISO int32 `json:"exposure_iso"`
}

func (h *Handler) UpdateExposureISO(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid film roll id")
		return
	}

	var req updateExposureISORequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid request body")
		return
	}

	roll, err := h.service.UpdateExposureISO(
		r.Context(),
		userID,
		id,
		req.ExposureISO,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidExposureISO):
			httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid exposure ISO")
		case errors.Is(err, ErrNotFound):
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "film roll not found")
		default:
			h.logger.Error("failed to update film roll exposure ISO", slog.Any("error", err))
			httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	h.logger.Info(
		"film roll exposure ISO updated",
		slog.Int64("film_roll_id", roll.ID),
		slog.Int("exposure_iso", int(roll.ExposureISO)),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, roll)
}

type updateCameraRequest struct {
	CameraID *int64 `json:"camera_id"`
}

func (h *Handler) UpdateCamera(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid film roll id")
		return
	}

	var req updateCameraRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CameraID != nil && *req.CameraID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid camera id")
		return
	}

	roll, err := h.service.UpdateCamera(
		r.Context(),
		userID,
		id,
		req.CameraID,
	)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "film roll not found")
			return
		}

		h.logger.Error("failed to update film roll camera", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	logAttrs := []any{slog.Int64("film_roll_id", roll.ID)}
	if roll.Camera != nil {
		logAttrs = append(logAttrs, slog.Int64("camera_id", roll.Camera.ID))
	}

	h.logger.Info("film roll camera updated", logAttrs...)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, roll)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid film roll id")
		return
	}

	err = h.service.Delete(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "film roll not found")
			return
		}

		h.logger.Error("failed to delete film roll", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info("film roll deleted", slog.Int64("film_roll_id", id))

	w.WriteHeader(http.StatusNoContent)
}
