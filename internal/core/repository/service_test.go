package repository_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"Coves/internal/atproto/carstore"
	"Coves/internal/core/repository"
	"Coves/internal/db/postgres"
	
	"github.com/ipfs/go-cid"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Mock signing key for testing
type mockSigningKey struct{}

// Test database connection
func setupTestDB(t *testing.T) (*sql.DB, *gorm.DB, func()) {
	// Use test database URL from environment or default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		// Skip test if no database configured
		t.Skip("TEST_DATABASE_URL not set, skipping database tests")
	}

	// Connect with sql.DB for migrations
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := goose.Up(sqlDB, "../../db/migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Connect with GORM using a fresh connection
	gormDB, err := gorm.Open(postgresDriver.Open(dbURL), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		PrepareStmt: false,
	})
	if err != nil {
		t.Fatalf("Failed to create GORM connection: %v", err)
	}

	// Cleanup function
	cleanup := func() {
		// Clean up test data
		gormDB.Exec("DELETE FROM repositories")
		gormDB.Exec("DELETE FROM commits")
		gormDB.Exec("DELETE FROM records")
		gormDB.Exec("DELETE FROM user_maps")
		gormDB.Exec("DELETE FROM car_shards")
		sqlDB.Close()
	}

	return sqlDB, gormDB, cleanup
}

func TestRepositoryService_CreateRepository(t *testing.T) {
	sqlDB, gormDB, cleanup := setupTestDB(t)
	defer cleanup()

	// Create temporary directory for carstore
	tempDir, err := os.MkdirTemp("", "carstore_test")
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

	// Initialize repository service
	repoRepo := postgres.NewRepositoryRepo(sqlDB)
	service := repository.NewService(repoRepo, repoStore)

	// Test DID
	testDID := "did:plc:testuser123"
	
	// Set signing key
	service.SetSigningKey(testDID, &mockSigningKey{})

	// Create repository
	repo, err := service.CreateRepository(testDID)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Verify repository was created
	if repo.DID != testDID {
		t.Errorf("Expected DID %s, got %s", testDID, repo.DID)
	}
	if !repo.HeadCID.Defined() {
		t.Error("Expected HeadCID to be defined")
	}
	if repo.RecordCount != 0 {
		t.Errorf("Expected RecordCount 0, got %d", repo.RecordCount)
	}

	// Verify repository exists in database
	fetchedRepo, err := service.GetRepository(testDID)
	if err != nil {
		t.Fatalf("Failed to get repository: %v", err)
	}
	if fetchedRepo.DID != testDID {
		t.Errorf("Expected fetched DID %s, got %s", testDID, fetchedRepo.DID)
	}

	// Test duplicate creation should fail
	_, err = service.CreateRepository(testDID)
	if err == nil {
		t.Error("Expected error creating duplicate repository")
	}
}

func TestRepositoryService_ImportExport(t *testing.T) {
	sqlDB, gormDB, cleanup := setupTestDB(t)
	defer cleanup()

	// Create temporary directory for carstore
	tempDir, err := os.MkdirTemp("", "carstore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Log the temp directory for debugging
	t.Logf("Using carstore directory: %s", tempDir)

	// Initialize carstore
	carDirs := []string{tempDir}
	repoStore, err := carstore.NewRepoStore(gormDB, carDirs)
	if err != nil {
		t.Fatalf("Failed to create repo store: %v", err)
	}

	// Initialize repository service
	repoRepo := postgres.NewRepositoryRepo(sqlDB)
	service := repository.NewService(repoRepo, repoStore)

	// Create first repository
	did1 := "did:plc:user1"
	service.SetSigningKey(did1, &mockSigningKey{})
	repo1, err := service.CreateRepository(did1)
	if err != nil {
		t.Fatalf("Failed to create repository 1: %v", err)
	}
	t.Logf("Created repository with HeadCID: %s", repo1.HeadCID)
	
	// Check what's in the database
	var userMapCount int
	gormDB.Raw("SELECT COUNT(*) FROM user_maps").Scan(&userMapCount)
	t.Logf("User maps count: %d", userMapCount)
	
	var carShardCount int
	gormDB.Raw("SELECT COUNT(*) FROM car_shards").Scan(&carShardCount)
	t.Logf("Car shards count: %d", carShardCount)
	
	// Check block_refs too
	var blockRefCount int
	gormDB.Raw("SELECT COUNT(*) FROM block_refs").Scan(&blockRefCount)
	t.Logf("Block refs count: %d", blockRefCount)

	// Export repository
	carData, err := service.ExportRepository(did1)
	if err != nil {
		t.Fatalf("Failed to export repository: %v", err)
	}
	// For now, empty repositories return empty CAR data
	t.Logf("Exported CAR data size: %d bytes", len(carData))

	// Import to new DID
	did2 := "did:plc:user2"
	err = service.ImportRepository(did2, carData)
	if err != nil {
		t.Fatalf("Failed to import repository: %v", err)
	}

	// Verify imported repository
	repo2, err := service.GetRepository(did2)
	if err != nil {
		t.Fatalf("Failed to get imported repository: %v", err)
	}
	if repo2.DID != did2 {
		t.Errorf("Expected DID %s, got %s", did2, repo2.DID)
	}
	// Note: HeadCID might differ due to new import
}

func TestRepositoryService_DeleteRepository(t *testing.T) {
	sqlDB, gormDB, cleanup := setupTestDB(t)
	defer cleanup()

	// Create temporary directory for carstore
	tempDir, err := os.MkdirTemp("", "carstore_test")
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

	// Initialize repository service
	repoRepo := postgres.NewRepositoryRepo(sqlDB)
	service := repository.NewService(repoRepo, repoStore)

	// Create repository
	testDID := "did:plc:deletetest"
	service.SetSigningKey(testDID, &mockSigningKey{})
	_, err = service.CreateRepository(testDID)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Delete repository
	err = service.DeleteRepository(testDID)
	if err != nil {
		t.Fatalf("Failed to delete repository: %v", err)
	}

	// Verify repository is deleted
	_, err = service.GetRepository(testDID)
	if err == nil {
		t.Error("Expected error getting deleted repository")
	}
}

func TestRepositoryService_CompactRepository(t *testing.T) {
	sqlDB, gormDB, cleanup := setupTestDB(t)
	defer cleanup()

	// Create temporary directory for carstore
	tempDir, err := os.MkdirTemp("", "carstore_test")
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

	// Initialize repository service
	repoRepo := postgres.NewRepositoryRepo(sqlDB)
	service := repository.NewService(repoRepo, repoStore)

	// Create repository
	testDID := "did:plc:compacttest"
	service.SetSigningKey(testDID, &mockSigningKey{})
	_, err = service.CreateRepository(testDID)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Run compaction (should not error even with minimal data)
	err = service.CompactRepository(testDID)
	if err != nil {
		t.Errorf("Failed to compact repository: %v", err)
	}
}

// Test UserMapping functionality
func TestUserMapping(t *testing.T) {
	_, gormDB, cleanup := setupTestDB(t)
	defer cleanup()

	// Create user mapping
	mapping, err := carstore.NewUserMapping(gormDB)
	if err != nil {
		t.Fatalf("Failed to create user mapping: %v", err)
	}

	// Test creating new mapping
	did1 := "did:plc:mapping1"
	uid1, err := mapping.GetOrCreateUID(context.Background(), did1)
	if err != nil {
		t.Fatalf("Failed to create UID for %s: %v", did1, err)
	}
	if uid1 == 0 {
		t.Error("Expected non-zero UID")
	}

	// Test getting existing mapping
	uid1Again, err := mapping.GetOrCreateUID(context.Background(), did1)
	if err != nil {
		t.Fatalf("Failed to get UID for %s: %v", did1, err)
	}
	if uid1 != uid1Again {
		t.Errorf("Expected same UID, got %d and %d", uid1, uid1Again)
	}

	// Test reverse lookup
	didLookup, err := mapping.GetDID(uid1)
	if err != nil {
		t.Fatalf("Failed to get DID for UID %d: %v", uid1, err)
	}
	if didLookup != did1 {
		t.Errorf("Expected DID %s, got %s", did1, didLookup)
	}

	// Test second user gets different UID
	did2 := "did:plc:mapping2"
	uid2, err := mapping.GetOrCreateUID(context.Background(), did2)
	if err != nil {
		t.Fatalf("Failed to create UID for %s: %v", did2, err)
	}
	if uid2 == uid1 {
		t.Error("Expected different UIDs for different DIDs")
	}
}

// Test with mock repository and carstore
func TestRepositoryService_MockedComponents(t *testing.T) {
	// Use the existing mock repository from the old test file
	_ = NewMockRepositoryRepository()
	
	// For unit testing without real carstore, we would need to mock RepoStore
	// For now, this demonstrates the structure
	t.Skip("Mocked carstore tests would require creating mock RepoStore interface")
}

// Benchmark repository creation
func BenchmarkRepositoryCreation(b *testing.B) {
	sqlDB, gormDB, cleanup := setupTestDB(&testing.T{})
	defer cleanup()

	tempDir, _ := os.MkdirTemp("", "carstore_bench")
	defer os.RemoveAll(tempDir)

	carDirs := []string{tempDir}
	repoStore, _ := carstore.NewRepoStore(gormDB, carDirs)
	repoRepo := postgres.NewRepositoryRepo(sqlDB)
	service := repository.NewService(repoRepo, repoStore)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		did := fmt.Sprintf("did:plc:bench%d", i)
		service.SetSigningKey(did, &mockSigningKey{})
		_, _ = service.CreateRepository(did)
	}
}

// MockRepositoryRepository is a mock implementation of repository.RepositoryRepository
type MockRepositoryRepository struct {
	repositories map[string]*repository.Repository
	commits      map[string][]*repository.Commit
	records      map[string]*repository.Record
}

func NewMockRepositoryRepository() *MockRepositoryRepository {
	return &MockRepositoryRepository{
		repositories: make(map[string]*repository.Repository),
		commits:      make(map[string][]*repository.Commit),
		records:      make(map[string]*repository.Record),
	}
}

// Repository operations
func (m *MockRepositoryRepository) Create(repo *repository.Repository) error {
	m.repositories[repo.DID] = repo
	return nil
}

func (m *MockRepositoryRepository) GetByDID(did string) (*repository.Repository, error) {
	repo, exists := m.repositories[did]
	if !exists {
		return nil, nil
	}
	return repo, nil
}

func (m *MockRepositoryRepository) Update(repo *repository.Repository) error {
	if _, exists := m.repositories[repo.DID]; !exists {
		return nil
	}
	m.repositories[repo.DID] = repo
	return nil
}

func (m *MockRepositoryRepository) Delete(did string) error {
	delete(m.repositories, did)
	return nil
}

// Commit operations
func (m *MockRepositoryRepository) CreateCommit(commit *repository.Commit) error {
	m.commits[commit.DID] = append(m.commits[commit.DID], commit)
	return nil
}

func (m *MockRepositoryRepository) GetCommit(did string, commitCID cid.Cid) (*repository.Commit, error) {
	commits, exists := m.commits[did]
	if !exists {
		return nil, nil
	}

	for _, c := range commits {
		if c.CID.Equals(commitCID) {
			return c, nil
		}
	}
	return nil, nil
}

func (m *MockRepositoryRepository) GetLatestCommit(did string) (*repository.Commit, error) {
	commits, exists := m.commits[did]
	if !exists || len(commits) == 0 {
		return nil, nil
	}
	return commits[len(commits)-1], nil
}

func (m *MockRepositoryRepository) ListCommits(did string, limit int, offset int) ([]*repository.Commit, error) {
	commits, exists := m.commits[did]
	if !exists {
		return []*repository.Commit{}, nil
	}

	start := offset
	if start >= len(commits) {
		return []*repository.Commit{}, nil
	}

	end := start + limit
	if end > len(commits) {
		end = len(commits)
	}

	return commits[start:end], nil
}

// Record operations
func (m *MockRepositoryRepository) CreateRecord(record *repository.Record) error {
	key := record.URI
	m.records[key] = record
	return nil
}

func (m *MockRepositoryRepository) GetRecord(did string, collection string, recordKey string) (*repository.Record, error) {
	uri := "at://" + did + "/" + collection + "/" + recordKey
	record, exists := m.records[uri]
	if !exists {
		return nil, nil
	}
	return record, nil
}

func (m *MockRepositoryRepository) UpdateRecord(record *repository.Record) error {
	key := record.URI
	if _, exists := m.records[key]; !exists {
		return nil
	}
	m.records[key] = record
	return nil
}

func (m *MockRepositoryRepository) DeleteRecord(did string, collection string, recordKey string) error {
	uri := "at://" + did + "/" + collection + "/" + recordKey
	delete(m.records, uri)
	return nil
}

func (m *MockRepositoryRepository) ListRecords(did string, collection string, limit int, offset int) ([]*repository.Record, error) {
	var records []*repository.Record
	prefix := "at://" + did + "/" + collection + "/"

	for uri, record := range m.records {
		if len(uri) > len(prefix) && uri[:len(prefix)] == prefix {
			records = append(records, record)
		}
	}

	// Simple pagination
	start := offset
	if start >= len(records) {
		return []*repository.Record{}, nil
	}

	end := start + limit
	if end > len(records) {
		end = len(records)
	}

	return records[start:end], nil
}