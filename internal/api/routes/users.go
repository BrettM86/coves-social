package routes

import (
	"github.com/go-chi/chi/v5"

	"Coves/internal/api/handlers"
	"Coves/internal/core/users"
)

func UserRoutes(userService users.UserServiceInterface) chi.Router {
	r := chi.NewRouter()
	userHandler := handlers.NewUserHandler(userService)

	r.Post("/", userHandler.CreateUser)
	r.Get("/{id}", userHandler.GetUser)

	return r
}