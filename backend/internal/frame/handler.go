package frame

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

type createFrameRequest struct {
	FrameLabel *string `json:"frame_label"`
	Note       *string `json:"note"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	filmRollID, err := strconv.ParseInt(
		r.PathValue("filmRollID"),
		10,
		64,
	)
	if err != nil || filmRollID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid film roll id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request createFrameRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid request body")
		return
	}

	createdFrame, err := h.service.Create(
		r.Context(),
		userID,
		filmRollID,
		request.FrameLabel,
		request.Note,
	)
	if err != nil {
		if errors.Is(err, ErrFilmRollNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "film roll not found")
			return
		}

		h.logger.Error("failed to create frame", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info(
		"frame created",
		slog.Int64("frame_id", createdFrame.ID),
		slog.Int64("film_roll_id", createdFrame.FilmRollID),
		slog.Int("frame_index", int(createdFrame.FrameIndex)),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusCreated, createdFrame)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	filmRollID, err := strconv.ParseInt(
		r.PathValue("filmRollID"),
		10,
		64,
	)
	if err != nil || filmRollID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid film roll id")
		return
	}

	frames, err := h.service.List(r.Context(), userID, filmRollID)
	if err != nil {
		if errors.Is(err, ErrFilmRollNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "film roll not found")
			return
		}

		h.logger.Error("failed to list frames", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info(
		"frames listed",
		slog.Int64("film_roll_id", filmRollID),
		slog.Int("count", len(frames)),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, frames)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	frameID, err := strconv.ParseInt(
		r.PathValue("frameID"),
		10,
		64,
	)
	if err != nil || frameID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid frame id")
		return
	}

	foundFrame, err := h.service.GetByID(r.Context(), userID, frameID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "frame not found")
			return
		}

		h.logger.Error("failed to get frame", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info(
		"frame retrieved",
		slog.Int64("frame_id", foundFrame.ID),
		slog.Int64("film_roll_id", foundFrame.FilmRollID),
		slog.Int("frame_index", int(foundFrame.FrameIndex)),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, foundFrame)
}

type updateFrameRequest struct {
	FrameLabel *string `json:"frame_label"`
	Note       *string `json:"note"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	frameID, err := strconv.ParseInt(
		r.PathValue("frameID"),
		10,
		64,
	)
	if err != nil || frameID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid frame id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request updateFrameRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid request body")
		return
	}

	updatedFrame, err := h.service.Update(
		r.Context(),
		userID,
		frameID,
		request.FrameLabel,
		request.Note,
	)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "frame not found")
			return
		}

		h.logger.Error("failed to update frame", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info(
		"frame updated",
		slog.Int64("frame_id", updatedFrame.ID),
		slog.Int64("film_roll_id", updatedFrame.FilmRollID),
		slog.Int("frame_index", int(updatedFrame.FrameIndex)),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, updatedFrame)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	frameID, err := strconv.ParseInt(
		r.PathValue("frameID"),
		10,
		64,
	)
	if err != nil || frameID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid frame id")
		return
	}

	if err := h.service.Delete(r.Context(), userID, frameID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "frame not found")
			return
		}

		h.logger.Error("failed to delete frame", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info("frame deleted", slog.Int64("frame_id", frameID))

	w.WriteHeader(http.StatusNoContent)
}
