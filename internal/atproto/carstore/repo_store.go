package carstore

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/bluesky-social/indigo/models"
	"github.com/ipfs/go-cid"
	"gorm.io/gorm"
)

// RepoStore combines CarStore with UserMapping to provide DID-based repository storage
type RepoStore struct {
	cs      *CarStore
	mapping *UserMapping
}

// NewRepoStore creates a new RepoStore instance
func NewRepoStore(db *gorm.DB, carDirs []string) (*RepoStore, error) {
	// Create carstore
	cs, err := NewCarStore(db, carDirs)
	if err != nil {
		return nil, fmt.Errorf("creating carstore: %w", err)
	}

	// Create user mapping
	mapping, err := NewUserMapping(db)
	if err != nil {
		return nil, fmt.Errorf("creating user mapping: %w", err)
	}

	return &RepoStore{
		cs:      cs,
		mapping: mapping,
	}, nil
}

// ImportRepo imports a repository CAR file for a DID
func (rs *RepoStore) ImportRepo(ctx context.Context, did string, carData io.Reader) (cid.Cid, error) {
	uid, err := rs.mapping.GetOrCreateUID(ctx, did)
	if err != nil {
		return cid.Undef, fmt.Errorf("getting UID for DID %s: %w", did, err)
	}

	// Read all data from the reader
	data, err := io.ReadAll(carData)
	if err != nil {
		return cid.Undef, fmt.Errorf("reading CAR data: %w", err)
	}

	return rs.cs.ImportSlice(ctx, uid, nil, data)
}

// ReadRepo reads a repository CAR file for a DID
func (rs *RepoStore) ReadRepo(ctx context.Context, did string, sinceRev string) ([]byte, error) {
	uid, err := rs.mapping.GetUID(did)
	if err != nil {
		return nil, fmt.Errorf("getting UID for DID %s: %w", did, err)
	}

	var buf bytes.Buffer
	err = rs.cs.ReadUserCar(ctx, uid, sinceRev, false, &buf)
	if err != nil {
		return nil, fmt.Errorf("reading repo for DID %s: %w", did, err)
	}

	return buf.Bytes(), nil
}

// GetRepoHead gets the latest repository head CID for a DID
func (rs *RepoStore) GetRepoHead(ctx context.Context, did string) (cid.Cid, error) {
	uid, err := rs.mapping.GetUID(did)
	if err != nil {
		return cid.Undef, fmt.Errorf("getting UID for DID %s: %w", did, err)
	}

	return rs.cs.GetUserRepoHead(ctx, uid)
}

// CompactRepo performs garbage collection for a DID's repository
func (rs *RepoStore) CompactRepo(ctx context.Context, did string) error {
	uid, err := rs.mapping.GetUID(did)
	if err != nil {
		return fmt.Errorf("getting UID for DID %s: %w", did, err)
	}

	return rs.cs.CompactUserShards(ctx, uid, false)
}

// DeleteRepo removes all data for a DID's repository
func (rs *RepoStore) DeleteRepo(ctx context.Context, did string) error {
	uid, err := rs.mapping.GetUID(did)
	if err != nil {
		return fmt.Errorf("getting UID for DID %s: %w", did, err)
	}

	return rs.cs.WipeUserData(ctx, uid)
}

// HasRepo checks if a repository exists for a DID
func (rs *RepoStore) HasRepo(ctx context.Context, did string) (bool, error) {
	uid, err := rs.mapping.GetUID(did)
	if err != nil {
		// If no UID mapping exists, repo doesn't exist
		return false, nil
	}

	// Try to get the repo head
	head, err := rs.cs.GetUserRepoHead(ctx, uid)
	if err != nil {
		return false, nil
	}

	return head.Defined(), nil
}

// GetOrCreateUID gets or creates a UID for a DID
func (rs *RepoStore) GetOrCreateUID(ctx context.Context, did string) (models.Uid, error) {
	return rs.mapping.GetOrCreateUID(ctx, did)
}
