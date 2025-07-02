package routes

import (
	"Coves/internal/core/users"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// UserRoutes returns user-related routes
func UserRoutes(service users.UserService) chi.Router {
	r := chi.NewRouter()
	
	// TODO: Implement user handlers
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User routes not yet implemented"))
	})
	
	return r
}