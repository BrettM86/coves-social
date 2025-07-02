package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"Coves/internal/api/routes"
	"Coves/internal/atproto/carstore"
	"Coves/internal/core/repository"
	"Coves/internal/core/users"
	postgresRepo "Coves/internal/db/postgres"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/coves?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("Failed to set goose dialect:", err)
	}

	if err := goose.Up(db, "internal/db/migrations"); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Initialize GORM
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt:                              true, // Enable prepared statements for better performance
	})
	if err != nil {
		log.Fatal("Failed to initialize GORM:", err)
	}

	// Initialize repositories
	userRepo := postgresRepo.NewUserRepository(db)
	_ = users.NewUserService(userRepo) // TODO: Use when UserRoutes is fixed

	// Initialize carstore for ATProto repository storage
	carDirs := []string{"./data/carstore"}
	repoStore, err := carstore.NewRepoStore(gormDB, carDirs)
	if err != nil {
		log.Fatal("Failed to initialize repo store:", err)
	}

	repositoryRepo := postgresRepo.NewRepositoryRepo(db)
	repositoryService := repository.NewService(repositoryRepo, repoStore)

	// Mount routes
	// TODO: Fix UserRoutes to accept *UserService
	// r.Mount("/api/users", routes.UserRoutes(userService))
	r.Mount("/", routes.RepositoryRoutes(repositoryService))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
