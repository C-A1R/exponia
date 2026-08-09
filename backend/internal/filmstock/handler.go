package filmstock

import (
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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	stocks, err := h.service.List(r.Context())
	if err != nil {
		h.logger.Error(
			"failed to list film stocks",
			slog.Any("error", err),
		)

		httpapi.WriteError(
			h.logger,
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	h.logger.Info(
		"film stocks listed",
		slog.Int("count", len(stocks)),
	)

	httpapi.WriteJSON(
		h.logger,
		w,
		http.StatusOK,
		stocks,
	)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	idValue := r.PathValue("id")

	id, err := strconv.ParseInt(idValue, 10, 64)
	if err != nil || id <= 0 {
		httpapi.WriteError(
			h.logger,
			w,
			http.StatusBadRequest,
			"invalid film stock id",
		)
		return
	}

	stock, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(
				h.logger,
				w,
				http.StatusNotFound,
				"film stock not found",
			)
			return
		}

		h.logger.Error(
			"failed to get film stock",
			slog.Any("error", err),
		)

		httpapi.WriteError(
			h.logger,
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	h.logger.Info(
		"film stock retrieved",
		slog.Int64("film_stock_id", stock.ID),
		slog.String("manufacturer", stock.Manufacturer),
		slog.String("name", stock.Name),
		slog.Int("iso", int(stock.ISO)),
	)

	httpapi.WriteJSON(
		h.logger,
		w,
		http.StatusOK,
		stock,
	)
}
