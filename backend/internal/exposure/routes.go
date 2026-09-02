package exposure

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc(
		"POST /api/v1/frames/{frameID}/exposures",
		handler.Create,
	)
	mux.HandleFunc(
		"GET /api/v1/frames/{frameID}/exposures",
		handler.List,
	)
	mux.HandleFunc(
		"GET /api/v1/exposures/{exposureID}",
		handler.GetByID,
	)
	mux.HandleFunc(
		"PUT /api/v1/exposures/{exposureID}",
		handler.Update,
	)
	mux.HandleFunc(
		"DELETE /api/v1/exposures/{exposureID}",
		handler.Delete,
	)
}
