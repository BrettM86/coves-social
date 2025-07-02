package carstore

import (
	"context"
	"fmt"
	"sync"

	"github.com/bluesky-social/indigo/models"
	"gorm.io/gorm"
)

// UserMapping manages the mapping between DIDs and numeric UIDs required by Indigo's carstore
type UserMapping struct {
	db       *gorm.DB
	mu       sync.RWMutex
	didToUID map[string]models.Uid
	uidToDID map[models.Uid]string
	nextUID  models.Uid
}

// UserMap represents the database model for DID to UID mapping
type UserMap struct {
	UID       models.Uid `gorm:"primaryKey;autoIncrement"`
	DID       string     `gorm:"column:did;uniqueIndex;not null"`
	CreatedAt int64
	UpdatedAt int64
}

// NewUserMapping creates a new UserMapping instance
func NewUserMapping(db *gorm.DB) (*UserMapping, error) {
	// Auto-migrate the user mapping table
	if err := db.AutoMigrate(&UserMap{}); err != nil {
		return nil, fmt.Errorf("migrating user mapping table: %w", err)
	}

	um := &UserMapping{
		db:       db,
		didToUID: make(map[string]models.Uid),
		uidToDID: make(map[models.Uid]string),
		nextUID:  1,
	}

	// Load existing mappings
	if err := um.loadMappings(); err != nil {
		return nil, fmt.Errorf("loading user mappings: %w", err)
	}

	return um, nil
}

// loadMappings loads all existing DID to UID mappings from the database
func (um *UserMapping) loadMappings() error {
	var mappings []UserMap
	if err := um.db.Find(&mappings).Error; err != nil {
		return fmt.Errorf("querying user mappings: %w", err)
	}

	um.mu.Lock()
	defer um.mu.Unlock()

	for _, m := range mappings {
		um.didToUID[m.DID] = m.UID
		um.uidToDID[m.UID] = m.DID
		if m.UID >= um.nextUID {
			um.nextUID = m.UID + 1
		}
	}

	return nil
}

// GetOrCreateUID gets or creates a UID for a given DID
func (um *UserMapping) GetOrCreateUID(ctx context.Context, did string) (models.Uid, error) {
	um.mu.RLock()
	if uid, exists := um.didToUID[did]; exists {
		um.mu.RUnlock()
		return uid, nil
	}
	um.mu.RUnlock()

	// Need to create a new mapping
	um.mu.Lock()
	defer um.mu.Unlock()

	// Double-check in case another goroutine created it
	if uid, exists := um.didToUID[did]; exists {
		return uid, nil
	}

	// Create new mapping
	userMap := &UserMap{
		DID: did,
	}

	if err := um.db.Create(userMap).Error; err != nil {
		return 0, fmt.Errorf("creating user mapping for DID %s: %w", did, err)
	}

	um.didToUID[did] = userMap.UID
	um.uidToDID[userMap.UID] = did

	return userMap.UID, nil
}

// GetUID returns the UID for a DID, or an error if not found
func (um *UserMapping) GetUID(did string) (models.Uid, error) {
	um.mu.RLock()
	defer um.mu.RUnlock()

	uid, exists := um.didToUID[did]
	if !exists {
		return 0, fmt.Errorf("UID not found for DID: %s", did)
	}
	return uid, nil
}

// GetDID returns the DID for a UID, or an error if not found
func (um *UserMapping) GetDID(uid models.Uid) (string, error) {
	um.mu.RLock()
	defer um.mu.RUnlock()

	did, exists := um.uidToDID[uid]
	if !exists {
		return "", fmt.Errorf("DID not found for UID: %d", uid)
	}
	return did, nil
}
