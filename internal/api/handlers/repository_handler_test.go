package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Coves/internal/core/repository"
	"github.com/ipfs/go-cid"
)

// MockRepositoryService is a mock implementation for testing
type MockRepositoryService struct {
	repositories map[string]*repository.Repository
	records      map[string]*repository.Record
}

func NewMockRepositoryService() *MockRepositoryService {
	return &MockRepositoryService{
		repositories: make(map[string]*repository.Repository),
		records:      make(map[string]*repository.Record),
	}
}

func (m *MockRepositoryService) CreateRepository(did string) (*repository.Repository, error) {
	repo := &repository.Repository{
		DID:     did,
		HeadCID: cid.Undef,
	}
	m.repositories[did] = repo
	return repo, nil
}

func (m *MockRepositoryService) GetRepository(did string) (*repository.Repository, error) {
	repo, exists := m.repositories[did]
	if !exists {
		return nil, nil
	}
	return repo, nil
}

func (m *MockRepositoryService) DeleteRepository(did string) error {
	delete(m.repositories, did)
	return nil
}

func (m *MockRepositoryService) CreateRecord(input repository.CreateRecordInput) (*repository.Record, error) {
	uri := "at://" + input.DID + "/" + input.Collection + "/" + input.RecordKey
	record := &repository.Record{
		URI:        uri,
		CID:        cid.Undef,
		Collection: input.Collection,
		RecordKey:  input.RecordKey,
		Value:      []byte(`{"test": "data"}`),
	}
	m.records[uri] = record
	return record, nil
}

func (m *MockRepositoryService) GetRecord(input repository.GetRecordInput) (*repository.Record, error) {
	uri := "at://" + input.DID + "/" + input.Collection + "/" + input.RecordKey
	record, exists := m.records[uri]
	if !exists {
		return nil, nil
	}
	return record, nil
}

func (m *MockRepositoryService) UpdateRecord(input repository.UpdateRecordInput) (*repository.Record, error) {
	uri := "at://" + input.DID + "/" + input.Collection + "/" + input.RecordKey
	record := &repository.Record{
		URI:        uri,
		CID:        cid.Undef,
		Collection: input.Collection,
		RecordKey:  input.RecordKey,
		Value:      []byte(`{"test": "updated"}`),
	}
	m.records[uri] = record
	return record, nil
}

func (m *MockRepositoryService) DeleteRecord(input repository.DeleteRecordInput) error {
	uri := "at://" + input.DID + "/" + input.Collection + "/" + input.RecordKey
	delete(m.records, uri)
	return nil
}

func (m *MockRepositoryService) ListRecords(did string, collection string, limit int, cursor string) ([]*repository.Record, string, error) {
	var records []*repository.Record
	for _, record := range m.records {
		if record.Collection == collection {
			records = append(records, record)
		}
	}
	return records, "", nil
}

func (m *MockRepositoryService) GetCommit(did string, cid cid.Cid) (*repository.Commit, error) {
	return nil, nil
}

func (m *MockRepositoryService) ListCommits(did string, limit int, cursor string) ([]*repository.Commit, string, error) {
	return []*repository.Commit{}, "", nil
}

func (m *MockRepositoryService) ExportRepository(did string) ([]byte, error) {
	return []byte("mock-car-data"), nil
}

func (m *MockRepositoryService) ImportRepository(did string, carData []byte) error {
	return nil
}

func TestCreateRecordHandler(t *testing.T) {
	mockService := NewMockRepositoryService()
	handler := NewRepositoryHandler(mockService)

	// Create test request
	reqData := CreateRecordRequest{
		Repo:       "did:plc:test123",
		Collection: "app.bsky.feed.post",
		RKey:       "testkey",
		Record:     json.RawMessage(`{"text": "Hello, world!"}`),
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest("POST", "/xrpc/com.atproto.repo.createRecord", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	handler.CreateRecord(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp CreateRecordResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	expectedURI := "at://did:plc:test123/app.bsky.feed.post/testkey"
	if resp.URI != expectedURI {
		t.Errorf("Expected URI %s, got %s", expectedURI, resp.URI)
	}
}

func TestGetRecordHandler(t *testing.T) {
	mockService := NewMockRepositoryService()
	handler := NewRepositoryHandler(mockService)

	// Create a test record first
	uri := "at://did:plc:test123/app.bsky.feed.post/testkey"
	testRecord := &repository.Record{
		URI:        uri,
		CID:        cid.Undef,
		Collection: "app.bsky.feed.post",
		RecordKey:  "testkey",
		Value:      []byte(`{"text": "Hello, world!"}`),
	}
	mockService.records[uri] = testRecord

	// Create test request
	req := httptest.NewRequest("GET", "/xrpc/com.atproto.repo.getRecord?repo=did:plc:test123&collection=app.bsky.feed.post&rkey=testkey", nil)
	w := httptest.NewRecorder()

	// Call handler
	handler.GetRecord(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp GetRecordResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.URI != uri {
		t.Errorf("Expected URI %s, got %s", uri, resp.URI)
	}
}