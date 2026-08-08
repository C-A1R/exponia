package filmroll

import "net/http"

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /api/v1/film-rolls", h.Create)
	mux.HandleFunc("GET /api/v1/film-rolls", h.List)
	mux.HandleFunc("GET /api/v1/film-rolls/{id}", h.GetByID)

	mux.HandleFunc("PATCH /api/v1/film-rolls/{id}/status", h.UpdateStatus)
	mux.HandleFunc("PATCH /api/v1/film-rolls/{id}/exposure-iso", h.UpdateExposureISO)
	mux.HandleFunc("PATCH /api/v1/film-rolls/{id}/camera", h.UpdateCamera)

	mux.HandleFunc("DELETE /api/v1/film-rolls/{id}", h.Delete)
}
