package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(handler *Handler, middleware *Middleware) *mux.Router {
	r := mux.NewRouter()

	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.BodyLimitMiddleware)

	api := r.PathPrefix("/api/v1").Subrouter()

	api.Handle("/auth/login", middleware.LoginRateLimitMiddleware(http.HandlerFunc(handler.Login))).Methods("POST")

	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	protected.HandleFunc("/profiles", handler.ListProfiles).Methods("GET")
	protected.HandleFunc("/profiles/{name}", handler.GetProfile).Methods("GET")
	protected.HandleFunc("/profiles/{name}", handler.SaveProfile).Methods("POST")
	protected.HandleFunc("/profiles/{name}", handler.DeleteProfile).Methods("DELETE")

	protected.HandleFunc("/env/scopes", handler.ListEnvScopes).Methods("GET")
	protected.HandleFunc("/env/scopes/{scope}", handler.GetEnvScope).Methods("GET")
	protected.HandleFunc("/env/scopes/{scope}", handler.SaveEnvScope).Methods("POST")
	protected.HandleFunc("/env/scopes/{scope}", handler.DeleteEnvScope).Methods("DELETE")
	protected.HandleFunc("/env/scopes/{scope}/{key}", handler.GetEnvVariable).Methods("GET")
	protected.HandleFunc("/env/scopes/{scope}/{key}", handler.SetEnvVariable).Methods("PUT")
	protected.HandleFunc("/env/scopes/{scope}/{key}", handler.DeleteEnvVariable).Methods("DELETE")

	return r
}
