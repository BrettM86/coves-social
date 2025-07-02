package integration

import (
	"Coves/internal/core/users"
	"Coves/internal/db/postgres"
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	"Coves/internal/api/routes"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://test_user:test_password@localhost:5434/coves_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("Failed to set goose dialect: %v", err)
	}

	if err := goose.Up(db, "../../internal/db/migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Clean up any existing test data
	_, err = db.Exec("DELETE FROM users WHERE email LIKE '%@example.com'")
	if err != nil {
		t.Logf("Warning: Failed to clean up test data: %v", err)
	}

	return db
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Wire up dependencies according to architecture
	userRepo := postgres.NewUserRepository(db)
	userService := users.NewUserService(userRepo)

	r := chi.NewRouter()
	r.Mount("/api/users", routes.UserRoutes(userService))

	user := users.CreateUserRequest{
		Email:    "test@example.com",
		Username: "testuser",
	}

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Response: %s", http.StatusCreated, w.Code, w.Body.String())
		return
	}

	var createdUser users.User
	if err := json.NewDecoder(w.Body).Decode(&createdUser); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if createdUser.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, createdUser.Email)
	}

	if createdUser.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, createdUser.Username)
	}
}

