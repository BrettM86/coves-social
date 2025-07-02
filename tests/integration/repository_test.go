package integration_test

import (
	"os"
	"testing"

	"Coves/internal/atproto/carstore"
	"Coves/internal/core/repository"
	"Coves/internal/db/postgres"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestRepositoryIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Use test database URL from environment or default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://test_user:test_password@localhost:5434/coves_test?sslmode=disable"
	}

	// Connect to test database with sql.DB for migrations
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer sqlDB.Close()

	// Run migrations
	if err := goose.Up(sqlDB, "../../internal/db/migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Connect with GORM for carstore
	gormDB, err := gorm.Open(postgresDriver.Open(dbURL), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt: false,
	})
	if err != nil {
		t.Fatalf("Failed to create GORM connection: %v", err)
	}

	// Create temporary directory for carstore
	tempDir, err := os.MkdirTemp("", "carstore_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize carstore
	carDirs := []string{tempDir}
	repoStore, err := carstore.NewRepoStore(gormDB, carDirs)
	if err != nil {
		t.Fatalf("Failed to create repo store: %v", err)
	}

	// Create repository repo
	repoRepo := postgres.NewRepositoryRepo(sqlDB)
	
	// Create service with both repo and repoStore
	service := repository.NewService(repoRepo, repoStore)
	
	// Test creating a repository
	did := "did:plc:testuser123"
	service.SetSigningKey(did, "mock-signing-key")
	
	repo, err := service.CreateRepository(did)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	
	if repo.DID != did {
		t.Errorf("Expected DID %s, got %s", did, repo.DID)
	}
	
	// Test getting the repository
	fetchedRepo, err := service.GetRepository(did)
	if err != nil {
		t.Fatalf("Failed to get repository: %v", err)
	}
	
	if fetchedRepo.DID != did {
		t.Errorf("Expected DID %s, got %s", did, fetchedRepo.DID)
	}
	
	// Clean up
	err = service.DeleteRepository(did)
	if err != nil {
		t.Fatalf("Failed to delete repository: %v", err)
	}

	// Clean up test data
	gormDB.Exec("DELETE FROM repositories")
	gormDB.Exec("DELETE FROM user_maps")
	gormDB.Exec("DELETE FROM car_shards")
}