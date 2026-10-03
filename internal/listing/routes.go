package listing

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("GET /listings", handler.List)
	mux.HandleFunc("GET /listings/{id}", handler.FindByID)
	mux.HandleFunc("POST /listings", handler.Create)
	mux.HandleFunc("PUT /listings/{id}", handler.Update)
	mux.HandleFunc("DELETE /listings/{id}", handler.Delete)
}
