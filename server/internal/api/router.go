package api

import (
	"github.com/gorilla/mux"
)

func NewRouter(handler *Handler, middleware *Middleware) *mux.Router {
	r := mux.NewRouter()

	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.CORSMiddleware)

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/auth/login", handler.Login).Methods("POST")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	protected.HandleFunc("/profiles", handler.ListProfiles).Methods("GET")
	protected.HandleFunc("/profiles/{name}", handler.GetProfile).Methods("GET")
	protected.HandleFunc("/profiles/{name}", handler.SaveProfile).Methods("POST")
	protected.HandleFunc("/profiles/{name}", handler.DeleteProfile).Methods("DELETE")

	return r
}
