package repository

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"Coves/internal/atproto/carstore"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// Service implements the RepositoryService interface using Indigo's carstore
type Service struct {
	repo         RepositoryRepository
	repoStore    *carstore.RepoStore
	signingKeys  map[string]interface{} // DID -> signing key
}

// NewService creates a new repository service using carstore
func NewService(repo RepositoryRepository, repoStore *carstore.RepoStore) *Service {
	return &Service{
		repo:        repo,
		repoStore:   repoStore,
		signingKeys: make(map[string]interface{}),
	}
}

// SetSigningKey sets the signing key for a DID
func (s *Service) SetSigningKey(did string, signingKey interface{}) {
	s.signingKeys[did] = signingKey
}

// CreateRepository creates a new repository
func (s *Service) CreateRepository(did string) (*Repository, error) {
	// Check if repository already exists
	existing, err := s.repo.GetByDID(did)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing repository: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("repository already exists for DID: %s", did)
	}

	// For now, just create the user mapping without importing CAR data
	// The actual repository data will be created when records are added
	ctx := context.Background()
	
	// Ensure user mapping exists
	_, err = s.repoStore.GetOrCreateUID(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("failed to create user mapping: %w", err)
	}
	

	// Create a placeholder CID for the empty repository
	emptyData := []byte("empty")
	mh, _ := multihash.Sum(emptyData, multihash.SHA2_256, -1)
	placeholderCID := cid.NewCidV1(cid.Raw, mh)

	// Create repository record
	repository := &Repository{
		DID:         did,
		HeadCID:     placeholderCID,
		Revision:    "rev-0",
		RecordCount: 0,
		StorageSize: 0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database
	if err := s.repo.Create(repository); err != nil {
		return nil, fmt.Errorf("failed to save repository: %w", err)
	}

	return repository, nil
}

// GetRepository retrieves a repository by DID
func (s *Service) GetRepository(did string) (*Repository, error) {
	repo, err := s.repo.GetByDID(did)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	if repo == nil {
		return nil, fmt.Errorf("repository not found for DID: %s", did)
	}

	// Update head CID from carstore
	headCID, err := s.repoStore.GetRepoHead(context.Background(), did)
	if err == nil && headCID.Defined() {
		repo.HeadCID = headCID
	}

	return repo, nil
}

// DeleteRepository deletes a repository
func (s *Service) DeleteRepository(did string) error {
	// Delete from carstore
	if err := s.repoStore.DeleteRepo(context.Background(), did); err != nil {
		return fmt.Errorf("failed to delete repo from carstore: %w", err)
	}

	// Delete from database
	if err := s.repo.Delete(did); err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}

	return nil
}

// ExportRepository exports a repository as a CAR file
func (s *Service) ExportRepository(did string) ([]byte, error) {
	// First check if repository exists in our database
	repo, err := s.repo.GetByDID(did)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	if repo == nil {
		return nil, fmt.Errorf("repository not found for DID: %s", did)
	}

	// Try to read from carstore
	carData, err := s.repoStore.ReadRepo(context.Background(), did, "")
	if err != nil {
		// If no data in carstore yet, return empty CAR
		// This happens when a repo is created but no records added yet
		// Check for the specific error pattern from Indigo's carstore
		errMsg := err.Error()
		if strings.Contains(errMsg, "no data found for user") ||
		   strings.Contains(errMsg, "user not found") {
			return []byte{}, nil
		}
		return nil, fmt.Errorf("failed to export repository: %w", err)
	}

	return carData, nil
}

// ImportRepository imports a repository from a CAR file
func (s *Service) ImportRepository(did string, carData []byte) error {
	ctx := context.Background()
	
	// If empty CAR data, just create user mapping
	if len(carData) == 0 {
		_, err := s.repoStore.GetOrCreateUID(ctx, did)
		if err != nil {
			return fmt.Errorf("failed to create user mapping: %w", err)
		}
		
		// Create placeholder CID
		emptyData := []byte("empty")
		mh, _ := multihash.Sum(emptyData, multihash.SHA2_256, -1)
		headCID := cid.NewCidV1(cid.Raw, mh)
		
		// Create repository record
		repo := &Repository{
			DID:         did,
			HeadCID:     headCID,
			Revision:    "imported-empty",
			RecordCount: 0,
			StorageSize: 0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := s.repo.Create(repo); err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}
		return nil
	}
	
	// Import non-empty CAR into carstore
	headCID, err := s.repoStore.ImportRepo(ctx, did, bytes.NewReader(carData))
	if err != nil {
		return fmt.Errorf("failed to import repository: %w", err)
	}

	// Create or update repository record
	repo, err := s.repo.GetByDID(did)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	if repo == nil {
		// Create new repository
		repo = &Repository{
			DID:         did,
			HeadCID:     headCID,
			Revision:    "imported",
			RecordCount: 0, // TODO: Count records in CAR
			StorageSize: int64(len(carData)),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := s.repo.Create(repo); err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}
	} else {
		// Update existing repository
		repo.HeadCID = headCID
		repo.UpdatedAt = time.Now()
		if err := s.repo.Update(repo); err != nil {
			return fmt.Errorf("failed to update repository: %w", err)
		}
	}

	return nil
}

// CompactRepository runs garbage collection on a repository
func (s *Service) CompactRepository(did string) error {
	return s.repoStore.CompactRepo(context.Background(), did)
}

// Note: Record-level operations would require more complex implementation
// to work with the carstore. For now, these are placeholder implementations
// that would need to be expanded to properly handle record CRUD operations
// by reading the CAR, modifying the repo structure, and writing back.

func (s *Service) CreateRecord(input CreateRecordInput) (*Record, error) {
	return nil, fmt.Errorf("record operations not yet implemented for carstore")
}

func (s *Service) GetRecord(input GetRecordInput) (*Record, error) {
	return nil, fmt.Errorf("record operations not yet implemented for carstore")
}

func (s *Service) UpdateRecord(input UpdateRecordInput) (*Record, error) {
	return nil, fmt.Errorf("record operations not yet implemented for carstore")
}

func (s *Service) DeleteRecord(input DeleteRecordInput) error {
	return fmt.Errorf("record operations not yet implemented for carstore")
}

func (s *Service) ListRecords(did string, collection string, limit int, cursor string) ([]*Record, string, error) {
	return nil, "", fmt.Errorf("record operations not yet implemented for carstore")
}

func (s *Service) GetCommit(did string, commitCID cid.Cid) (*Commit, error) {
	return nil, fmt.Errorf("commit operations not yet implemented for carstore")
}

func (s *Service) ListCommits(did string, limit int, cursor string) ([]*Commit, string, error) {
	return nil, "", fmt.Errorf("commit operations not yet implemented for carstore")
}