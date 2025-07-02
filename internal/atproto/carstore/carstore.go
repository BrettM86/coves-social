package carstore

import (
	"context"
	"fmt"
	"io"

	"github.com/bluesky-social/indigo/carstore"
	"github.com/bluesky-social/indigo/models"
	"github.com/ipfs/go-cid"
	"gorm.io/gorm"
)

// CarStore wraps Indigo's carstore for managing ATProto repository CAR files
type CarStore struct {
	cs carstore.CarStore
}

// NewCarStore creates a new CarStore instance using Indigo's implementation
func NewCarStore(db *gorm.DB, carDirs []string) (*CarStore, error) {
	// Initialize Indigo's carstore
	cs, err := carstore.NewCarStore(db, carDirs)
	if err != nil {
		return nil, fmt.Errorf("failed to create carstore: %w", err)
	}

	return &CarStore{
		cs: cs,
	}, nil
}

// ImportSlice imports a CAR file slice for a user
func (c *CarStore) ImportSlice(ctx context.Context, uid models.Uid, since *string, carData []byte) (cid.Cid, error) {
	rootCid, _, err := c.cs.ImportSlice(ctx, uid, since, carData)
	return rootCid, err
}

// ReadUserCar reads a user's repository CAR file
func (c *CarStore) ReadUserCar(ctx context.Context, uid models.Uid, sinceRev string, incremental bool, w io.Writer) error {
	return c.cs.ReadUserCar(ctx, uid, sinceRev, incremental, w)
}

// GetUserRepoHead gets the latest repository head CID for a user
func (c *CarStore) GetUserRepoHead(ctx context.Context, uid models.Uid) (cid.Cid, error) {
	return c.cs.GetUserRepoHead(ctx, uid)
}

// CompactUserShards performs garbage collection and compaction for a user's data
func (c *CarStore) CompactUserShards(ctx context.Context, uid models.Uid, aggressive bool) error {
	_, err := c.cs.CompactUserShards(ctx, uid, aggressive)
	return err
}

// WipeUserData removes all data for a user
func (c *CarStore) WipeUserData(ctx context.Context, uid models.Uid) error {
	return c.cs.WipeUserData(ctx, uid)
}

// NewDeltaSession creates a new session for writing deltas
func (c *CarStore) NewDeltaSession(ctx context.Context, uid models.Uid, since *string) (*carstore.DeltaSession, error) {
	return c.cs.NewDeltaSession(ctx, uid, since)
}

// ReadOnlySession creates a read-only session for reading user data
func (c *CarStore) ReadOnlySession(uid models.Uid) (*carstore.DeltaSession, error) {
	return c.cs.ReadOnlySession(uid)
}

// Stat returns statistics about the carstore
func (c *CarStore) Stat(ctx context.Context, uid models.Uid) ([]carstore.UserStat, error) {
	return c.cs.Stat(ctx, uid)
}