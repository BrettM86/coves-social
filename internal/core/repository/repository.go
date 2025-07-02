package repository

import (
	"time"

	"github.com/ipfs/go-cid"
)

// Repository represents an AT Protocol data repository
type Repository struct {
	DID            string    // Decentralized identifier of the repository owner
	HeadCID        cid.Cid   // CID of the latest commit
	Revision       string    // Current revision identifier
	RecordCount    int       // Number of records in the repository
	StorageSize    int64     // Total storage size in bytes
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Commit represents a signed repository commit
type Commit struct {
	CID            cid.Cid   // Content identifier of this commit
	DID            string    // DID of the committer
	Version        int       // Repository version
	PrevCID        *cid.Cid  // CID of the previous commit (nil for first commit)
	DataCID        cid.Cid   // CID of the MST root
	Revision       string    // Revision identifier
	Signature      []byte    // Cryptographic signature
	SigningKeyID   string    // Key ID used for signing
	CreatedAt      time.Time
}

// Record represents a record in the repository
type Record struct {
	URI            string    // AT-URI of the record (e.g., at://did:plc:123/app.bsky.feed.post/abc)
	CID            cid.Cid   // Content identifier
	Collection     string    // Collection name (e.g., app.bsky.feed.post)
	RecordKey      string    // Record key within collection
	Value          []byte    // The actual record data (typically CBOR)
	CreatedAt      time.Time
	UpdatedAt      time.Time
}


// CreateRecordInput represents input for creating a record
type CreateRecordInput struct {
	DID            string
	Collection     string
	RecordKey      string    // Optional - will be generated if not provided
	Record         interface{}
	Validate       bool      // Whether to validate against lexicon
}

// UpdateRecordInput represents input for updating a record
type UpdateRecordInput struct {
	DID            string
	Collection     string
	RecordKey      string
	Record         interface{}
	Validate       bool
}

// GetRecordInput represents input for retrieving a record
type GetRecordInput struct {
	DID            string
	Collection     string
	RecordKey      string
}

// DeleteRecordInput represents input for deleting a record
type DeleteRecordInput struct {
	DID            string
	Collection     string
	RecordKey      string
}

// RepositoryService defines the business logic for repository operations
type RepositoryService interface {
	// Repository operations
	CreateRepository(did string) (*Repository, error)
	GetRepository(did string) (*Repository, error)
	DeleteRepository(did string) error
	
	// Record operations
	CreateRecord(input CreateRecordInput) (*Record, error)
	GetRecord(input GetRecordInput) (*Record, error)
	UpdateRecord(input UpdateRecordInput) (*Record, error)
	DeleteRecord(input DeleteRecordInput) error
	
	// Collection operations
	ListRecords(did string, collection string, limit int, cursor string) ([]*Record, string, error)
	
	// Commit operations
	GetCommit(did string, cid cid.Cid) (*Commit, error)
	ListCommits(did string, limit int, cursor string) ([]*Commit, string, error)
	
	// Export operations
	ExportRepository(did string) ([]byte, error) // Returns CAR file
	ImportRepository(did string, carData []byte) error
}

// RepositoryRepository defines the data access interface for repositories
type RepositoryRepository interface {
	// Repository operations
	Create(repo *Repository) error
	GetByDID(did string) (*Repository, error)
	Update(repo *Repository) error
	Delete(did string) error
	
	// Commit operations
	CreateCommit(commit *Commit) error
	GetCommit(did string, cid cid.Cid) (*Commit, error)
	GetLatestCommit(did string) (*Commit, error)
	ListCommits(did string, limit int, offset int) ([]*Commit, error)
	
	// Record operations
	CreateRecord(record *Record) error
	GetRecord(did string, collection string, recordKey string) (*Record, error)
	UpdateRecord(record *Record) error
	DeleteRecord(did string, collection string, recordKey string) error
	ListRecords(did string, collection string, limit int, offset int) ([]*Record, error)
	
}