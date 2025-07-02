package repo

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bluesky-social/indigo/mst"
	"github.com/bluesky-social/indigo/repo"
	"github.com/ipfs/go-cid"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	cbornode "github.com/ipfs/go-ipld-cbor"
	cbg "github.com/whyrusleeping/cbor-gen"
)

// Wrapper provides a thin wrapper around Indigo's repo package
type Wrapper struct {
	repo       *repo.Repo
	blockstore blockstore.Blockstore
}

// NewWrapper creates a new wrapper for a repository with the provided blockstore
func NewWrapper(did string, signingKey interface{}, bs blockstore.Blockstore) (*Wrapper, error) {
	// Create new repository with the provided blockstore
	r := repo.NewRepo(context.Background(), did, bs)
	
	return &Wrapper{
		repo:       r,
		blockstore: bs,
	}, nil
}

// OpenWrapper opens an existing repository from CAR data with the provided blockstore
func OpenWrapper(carData []byte, signingKey interface{}, bs blockstore.Blockstore) (*Wrapper, error) {
	r, err := repo.ReadRepoFromCar(context.Background(), bytes.NewReader(carData))
	if err != nil {
		return nil, fmt.Errorf("failed to read repo from CAR: %w", err)
	}
	
	return &Wrapper{
		repo:       r,
		blockstore: bs,
	}, nil
}

// CreateRecord adds a new record to the repository
func (w *Wrapper) CreateRecord(collection string, recordKey string, record cbg.CBORMarshaler) (cid.Cid, string, error) {
	// The repo.CreateRecord generates its own key, so we'll use that
	recordCID, rkey, err := w.repo.CreateRecord(context.Background(), collection, record)
	if err != nil {
		return cid.Undef, "", fmt.Errorf("failed to create record: %w", err)
	}
	
	// If a specific key was requested, we'd need to use PutRecord instead
	if recordKey != "" {
		// Use PutRecord for specific keys
		path := fmt.Sprintf("%s/%s", collection, recordKey)
		recordCID, err = w.repo.PutRecord(context.Background(), path, record)
		if err != nil {
			return cid.Undef, "", fmt.Errorf("failed to put record with key: %w", err)
		}
		return recordCID, recordKey, nil
	}
	
	return recordCID, rkey, nil
}

// GetRecord retrieves a record from the repository
func (w *Wrapper) GetRecord(collection string, recordKey string) (cid.Cid, []byte, error) {
	path := fmt.Sprintf("%s/%s", collection, recordKey)
	
	recordCID, rec, err := w.repo.GetRecord(context.Background(), path)
	if err != nil {
		return cid.Undef, nil, fmt.Errorf("failed to get record: %w", err)
	}
	
	// Encode record to CBOR
	buf := new(bytes.Buffer)
	if err := rec.(cbg.CBORMarshaler).MarshalCBOR(buf); err != nil {
		return cid.Undef, nil, fmt.Errorf("failed to encode record: %w", err)
	}
	
	return recordCID, buf.Bytes(), nil
}

// UpdateRecord updates an existing record in the repository
func (w *Wrapper) UpdateRecord(collection string, recordKey string, record cbg.CBORMarshaler) (cid.Cid, error) {
	path := fmt.Sprintf("%s/%s", collection, recordKey)
	
	// Check if record exists
	_, _, err := w.repo.GetRecord(context.Background(), path)
	if err != nil {
		return cid.Undef, fmt.Errorf("record not found: %w", err)
	}
	
	// Update the record
	recordCID, err := w.repo.UpdateRecord(context.Background(), path, record)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to update record: %w", err)
	}
	
	return recordCID, nil
}

// DeleteRecord removes a record from the repository
func (w *Wrapper) DeleteRecord(collection string, recordKey string) error {
	path := fmt.Sprintf("%s/%s", collection, recordKey)
	
	if err := w.repo.DeleteRecord(context.Background(), path); err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	
	return nil
}

// ListRecords returns all records in a collection
func (w *Wrapper) ListRecords(collection string) ([]RecordInfo, error) {
	var records []RecordInfo
	
	err := w.repo.ForEach(context.Background(), collection, func(k string, v cid.Cid) error {
		// Skip if not in the requested collection
		if len(k) <= len(collection)+1 || k[:len(collection)] != collection || k[len(collection)] != '/' {
			return nil
		}
		
		recordKey := k[len(collection)+1:]
		records = append(records, RecordInfo{
			Collection: collection,
			RecordKey:  recordKey,
			CID:        v,
		})
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}
	
	return records, nil
}

// Commit creates a new signed commit
func (w *Wrapper) Commit(did string, signingKey interface{}) (*repo.SignedCommit, error) {
	// The commit function expects a signing function with context
	signingFunc := func(ctx context.Context, did string, data []byte) ([]byte, error) {
		// TODO: Implement proper signing based on signingKey type
		return []byte("mock-signature"), nil
	}
	
	_, _, err := w.repo.Commit(context.Background(), signingFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}
	
	// Return the signed commit from the repo
	sc := w.repo.SignedCommit()
	
	return &sc, nil
}

// GetHeadCID returns the CID of the current repository head
func (w *Wrapper) GetHeadCID() (cid.Cid, error) {
	// TODO: Implement this properly
	// The repo package doesn't expose a direct way to get the head CID
	return cid.Undef, fmt.Errorf("not implemented")
}

// Export exports the repository as a CAR file
func (w *Wrapper) Export() ([]byte, error) {
	// TODO: Implement proper CAR export using Indigo's carstore functionality
	// For now, return a placeholder
	return nil, fmt.Errorf("CAR export not yet implemented")
}

// GetMST returns the underlying Merkle Search Tree
func (w *Wrapper) GetMST() (*mst.MerkleSearchTree, error) {
	// TODO: Implement MST access
	return nil, fmt.Errorf("not implemented")
}

// RecordInfo contains information about a record
type RecordInfo struct {
	Collection string
	RecordKey  string
	CID        cid.Cid
}

// DecodeRecord decodes CBOR data into a record structure
func DecodeRecord(data []byte, v interface{}) error {
	return cbornode.DecodeInto(data, v)
}

// EncodeRecord encodes a record structure into CBOR data
func EncodeRecord(v cbg.CBORMarshaler) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := v.MarshalCBOR(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}