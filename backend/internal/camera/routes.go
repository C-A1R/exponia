package camera

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /api/v1/cameras", handler.List)
	mux.HandleFunc("POST /api/v1/cameras", handler.Create)
	mux.HandleFunc("GET /api/v1/cameras/{id}", handler.GetByID)
	mux.HandleFunc("PUT /api/v1/cameras/{id}", handler.Update)
	mux.HandleFunc("DELETE /api/v1/cameras/{id}", handler.Delete)
}
