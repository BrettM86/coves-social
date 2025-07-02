package routes

import (
	"Coves/internal/api/handlers"
	"Coves/internal/core/repository"
	"github.com/go-chi/chi/v5"
)

// RepositoryRoutes returns repository-related routes
func RepositoryRoutes(service repository.RepositoryService) chi.Router {
	handler := handlers.NewRepositoryHandler(service)
	
	r := chi.NewRouter()
	
	// AT Protocol XRPC endpoints for repository operations
	r.Route("/xrpc", func(r chi.Router) {
		// Record operations
		r.Post("/com.atproto.repo.createRecord", handler.CreateRecord)
		r.Get("/com.atproto.repo.getRecord", handler.GetRecord)
		r.Post("/com.atproto.repo.putRecord", handler.PutRecord)
		r.Post("/com.atproto.repo.deleteRecord", handler.DeleteRecord)
		r.Get("/com.atproto.repo.listRecords", handler.ListRecords)
		
		// Repository operations
		r.Post("/com.atproto.repo.createRepo", handler.CreateRepository)
		
		// Sync operations
		r.Get("/com.atproto.sync.getRepo", handler.GetRepo)
		r.Get("/com.atproto.sync.getCommit", handler.GetCommit)
	})
	
	return r
}