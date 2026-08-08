package filmstock

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc(
		"GET /api/v1/film-stocks",
		handler.List,
	)

	mux.HandleFunc(
		"GET /api/v1/film-stocks/{id}",
		handler.GetByID,
	)
}
