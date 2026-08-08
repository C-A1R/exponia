package lens

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /api/v1/lenses", handler.List)
	mux.HandleFunc("POST /api/v1/lenses", handler.Create)
	mux.HandleFunc("GET /api/v1/lenses/{id}", handler.GetByID)
	mux.HandleFunc("PUT /api/v1/lenses/{id}", handler.Update)
	mux.HandleFunc("DELETE /api/v1/lenses/{id}", handler.Delete)
}
