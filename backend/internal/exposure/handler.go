package exposure

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

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

type createExposureRequest struct {
	LensID         *int64     `json:"lens_id"`
	Aperture       *float64   `json:"aperture"`
	ShutterSpeedUS *int64     `json:"shutter_speed_us"`
	ShotAt         *time.Time `json:"shot_at"`
	Note           *string    `json:"note"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	frameID, err := strconv.ParseInt(r.PathValue("frameID"), 10, 64)
	if err != nil || frameID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid frame id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request createExposureRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid request body")
		return
	}

	createdExposure, err := h.service.Create(
		r.Context(),
		userID,
		frameID,
		CreateInput{
			LensID:         request.LensID,
			Aperture:       request.Aperture,
			ShutterSpeedUS: request.ShutterSpeedUS,
			ShotAt:         request.ShotAt,
			Note:           request.Note,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidLensID),
			errors.Is(err, ErrInvalidAperture),
			errors.Is(err, ErrInvalidShutterSpeed):
			httpapi.WriteError(h.logger, w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrFrameNotFound):
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "frame not found")
		case errors.Is(err, ErrLensNotFound):
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "lens not found")
		case errors.Is(err, ErrCameraNotSelected):
			httpapi.WriteError(
				h.logger,
				w,
				http.StatusConflict,
				"camera is not selected for film roll",
			)
		case errors.Is(err, ErrCreateRejected):
			httpapi.WriteError(h.logger, w, http.StatusConflict, "exposure creation rejected")
		default:
			h.logger.Error("failed to create exposure", slog.Any("error", err))
			httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		}

		return
	}

	h.logger.Info(
		"exposure created",
		slog.Int64("exposure_id", createdExposure.ID),
		slog.Int64("frame_id", createdExposure.FrameID),
		slog.Int("exposure_index", int(createdExposure.ExposureIndex)),
		slog.Int64("camera_id", createdExposure.CameraID),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusCreated, createdExposure)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	frameID, err := strconv.ParseInt(r.PathValue("frameID"), 10, 64)
	if err != nil || frameID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid frame id")
		return
	}

	exposures, err := h.service.List(r.Context(), userID, frameID)
	if err != nil {
		if errors.Is(err, ErrFrameNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "frame not found")
			return
		}

		h.logger.Error("failed to list exposures", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info(
		"exposures listed",
		slog.Int64("frame_id", frameID),
		slog.Int("count", len(exposures)),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, exposures)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	exposureID, err := strconv.ParseInt(r.PathValue("exposureID"), 10, 64)
	if err != nil || exposureID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid exposure id")
		return
	}

	foundExposure, err := h.service.GetByID(r.Context(), userID, exposureID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "exposure not found")
			return
		}

		h.logger.Error("failed to get exposure", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info(
		"exposure retrieved",
		slog.Int64("exposure_id", foundExposure.ID),
		slog.Int64("frame_id", foundExposure.FrameID),
		slog.Int("exposure_index", int(foundExposure.ExposureIndex)),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, foundExposure)
}

type updateExposureRequest struct {
	CameraID       int64      `json:"camera_id"`
	LensID         *int64     `json:"lens_id"`
	Aperture       *float64   `json:"aperture"`
	ShutterSpeedUS *int64     `json:"shutter_speed_us"`
	ShotAt         *time.Time `json:"shot_at"`
	Note           *string    `json:"note"`
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	exposureID, err := strconv.ParseInt(r.PathValue("exposureID"), 10, 64)
	if err != nil || exposureID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid exposure id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var request updateExposureRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid request body")
		return
	}

	updatedExposure, err := h.service.Update(
		r.Context(),
		userID,
		exposureID,
		UpdateInput{
			CameraID:       request.CameraID,
			LensID:         request.LensID,
			Aperture:       request.Aperture,
			ShutterSpeedUS: request.ShutterSpeedUS,
			ShotAt:         request.ShotAt,
			Note:           request.Note,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCameraID),
			errors.Is(err, ErrInvalidLensID),
			errors.Is(err, ErrInvalidAperture),
			errors.Is(err, ErrInvalidShutterSpeed):
			httpapi.WriteError(h.logger, w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrNotFound):
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "exposure not found")
		case errors.Is(err, ErrCameraNotFound):
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "camera not found")
		case errors.Is(err, ErrLensNotFound):
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "lens not found")
		case errors.Is(err, ErrUpdateRejected):
			httpapi.WriteError(h.logger, w, http.StatusConflict, "exposure update rejected")
		default:
			h.logger.Error("failed to update exposure", slog.Any("error", err))
			httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		}

		return
	}

	h.logger.Info(
		"exposure updated",
		slog.Int64("exposure_id", updatedExposure.ID),
		slog.Int64("frame_id", updatedExposure.FrameID),
		slog.Int64("camera_id", updatedExposure.CameraID),
	)

	httpapi.WriteJSON(h.logger, w, http.StatusOK, updatedExposure)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := httpapi.RequireUser(h.logger, w, r)
	if !ok {
		return
	}

	exposureID, err := strconv.ParseInt(r.PathValue("exposureID"), 10, 64)
	if err != nil || exposureID <= 0 {
		httpapi.WriteError(h.logger, w, http.StatusBadRequest, "invalid exposure id")
		return
	}

	if err := h.service.Delete(r.Context(), userID, exposureID); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(h.logger, w, http.StatusNotFound, "exposure not found")
			return
		}

		h.logger.Error("failed to delete exposure", slog.Any("error", err))
		httpapi.WriteError(h.logger, w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info("exposure deleted", slog.Int64("exposure_id", exposureID))

	w.WriteHeader(http.StatusNoContent)
}
