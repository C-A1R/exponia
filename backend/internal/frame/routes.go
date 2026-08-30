package frame

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc(
		"POST /api/v1/film-rolls/{filmRollID}/frames",
		handler.Create,
	)
	mux.HandleFunc(
		"GET /api/v1/film-rolls/{filmRollID}/frames",
		handler.List,
	)
	mux.HandleFunc("GET /api/v1/frames/{frameID}", handler.GetByID)
	mux.HandleFunc("PUT /api/v1/frames/{frameID}", handler.Update)
	mux.HandleFunc("DELETE /api/v1/frames/{frameID}", handler.Delete)
}
