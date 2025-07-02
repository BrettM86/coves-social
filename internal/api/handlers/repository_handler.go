package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"Coves/internal/core/repository"
	"github.com/ipfs/go-cid"
	cbornode "github.com/ipfs/go-ipld-cbor"
)

// RepositoryHandler handles HTTP requests for repository operations
type RepositoryHandler struct {
	service repository.RepositoryService
}

// NewRepositoryHandler creates a new repository handler
func NewRepositoryHandler(service repository.RepositoryService) *RepositoryHandler {
	return &RepositoryHandler{
		service: service,
	}
}

// AT Protocol XRPC request/response types

// CreateRecordRequest represents a request to create a record
type CreateRecordRequest struct {
	Repo       string          `json:"repo"`       // DID of the repository
	Collection string          `json:"collection"` // NSID of the collection
	RKey       string          `json:"rkey,omitempty"` // Optional record key
	Validate   bool            `json:"validate"`   // Whether to validate against lexicon
	Record     json.RawMessage `json:"record"`     // The record data
}

// CreateRecordResponse represents the response after creating a record
type CreateRecordResponse struct {
	URI string `json:"uri"` // AT-URI of the created record
	CID string `json:"cid"` // CID of the record
}

// GetRecordRequest represents a request to get a record
type GetRecordRequest struct {
	Repo       string `json:"repo"`       // DID of the repository
	Collection string `json:"collection"` // NSID of the collection
	RKey       string `json:"rkey"`       // Record key
}

// GetRecordResponse represents the response when getting a record
type GetRecordResponse struct {
	URI   string          `json:"uri"`   // AT-URI of the record
	CID   string          `json:"cid"`   // CID of the record
	Value json.RawMessage `json:"value"` // The record data
}

// PutRecordRequest represents a request to update a record
type PutRecordRequest struct {
	Repo       string          `json:"repo"`       // DID of the repository
	Collection string          `json:"collection"` // NSID of the collection
	RKey       string          `json:"rkey"`       // Record key
	Validate   bool            `json:"validate"`   // Whether to validate against lexicon
	Record     json.RawMessage `json:"record"`     // The record data
}

// PutRecordResponse represents the response after updating a record
type PutRecordResponse struct {
	URI string `json:"uri"` // AT-URI of the updated record
	CID string `json:"cid"` // CID of the record
}

// DeleteRecordRequest represents a request to delete a record
type DeleteRecordRequest struct {
	Repo       string `json:"repo"`       // DID of the repository
	Collection string `json:"collection"` // NSID of the collection
	RKey       string `json:"rkey"`       // Record key
}

// ListRecordsRequest represents a request to list records
type ListRecordsRequest struct {
	Repo       string `json:"repo"`       // DID of the repository
	Collection string `json:"collection"` // NSID of the collection
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
}

// ListRecordsResponse represents the response when listing records
type ListRecordsResponse struct {
	Cursor  string         `json:"cursor,omitempty"`
	Records []RecordOutput `json:"records"`
}

// RecordOutput represents a record in list responses
type RecordOutput struct {
	URI   string          `json:"uri"`
	CID   string          `json:"cid"`
	Value json.RawMessage `json:"value"`
}

// Handler methods

// CreateRecord handles POST /xrpc/com.atproto.repo.createRecord
func (h *RepositoryHandler) CreateRecord(w http.ResponseWriter, r *http.Request) {
	var req CreateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	// Validate required fields
	if req.Repo == "" || req.Collection == "" || len(req.Record) == 0 {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}

	// Create a generic record structure for CBOR encoding
	// In a real implementation, you would unmarshal to the specific lexicon type
	recordData := &GenericRecord{
		Data: req.Record,
	}

	input := repository.CreateRecordInput{
		DID:        req.Repo,
		Collection: req.Collection,
		RecordKey:  req.RKey,
		Record:     recordData,
		Validate:   req.Validate,
	}

	record, err := h.service.CreateRecord(input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create record: %v", err))
		return
	}

	resp := CreateRecordResponse{
		URI: record.URI,
		CID: record.CID.String(),
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetRecord handles GET /xrpc/com.atproto.repo.getRecord
func (h *RepositoryHandler) GetRecord(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	repo := r.URL.Query().Get("repo")
	collection := r.URL.Query().Get("collection")
	rkey := r.URL.Query().Get("rkey")

	if repo == "" || collection == "" || rkey == "" {
		writeError(w, http.StatusBadRequest, "missing required parameters")
		return
	}

	input := repository.GetRecordInput{
		DID:        repo,
		Collection: collection,
		RecordKey:  rkey,
	}

	record, err := h.service.GetRecord(input)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "record not found")
			return
		}
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get record: %v", err))
		return
	}

	resp := GetRecordResponse{
		URI:   record.URI,
		CID:   record.CID.String(),
		Value: json.RawMessage(record.Value),
	}

	writeJSON(w, http.StatusOK, resp)
}

// PutRecord handles POST /xrpc/com.atproto.repo.putRecord
func (h *RepositoryHandler) PutRecord(w http.ResponseWriter, r *http.Request) {
	var req PutRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	// Validate required fields
	if req.Repo == "" || req.Collection == "" || req.RKey == "" || len(req.Record) == 0 {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}

	// Create a generic record structure for CBOR encoding
	recordData := &GenericRecord{
		Data: req.Record,
	}

	input := repository.UpdateRecordInput{
		DID:        req.Repo,
		Collection: req.Collection,
		RecordKey:  req.RKey,
		Record:     recordData,
		Validate:   req.Validate,
	}

	record, err := h.service.UpdateRecord(input)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "record not found")
			return
		}
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to update record: %v", err))
		return
	}

	resp := PutRecordResponse{
		URI: record.URI,
		CID: record.CID.String(),
	}

	writeJSON(w, http.StatusOK, resp)
}

// DeleteRecord handles POST /xrpc/com.atproto.repo.deleteRecord
func (h *RepositoryHandler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	var req DeleteRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	// Validate required fields
	if req.Repo == "" || req.Collection == "" || req.RKey == "" {
		writeError(w, http.StatusBadRequest, "missing required fields")
		return
	}

	input := repository.DeleteRecordInput{
		DID:        req.Repo,
		Collection: req.Collection,
		RecordKey:  req.RKey,
	}

	err := h.service.DeleteRecord(input)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "record not found")
			return
		}
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete record: %v", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

// ListRecords handles GET /xrpc/com.atproto.repo.listRecords
func (h *RepositoryHandler) ListRecords(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	repo := r.URL.Query().Get("repo")
	collection := r.URL.Query().Get("collection")
	limit := 50 // Default limit
	cursor := r.URL.Query().Get("cursor")

	if repo == "" || collection == "" {
		writeError(w, http.StatusBadRequest, "missing required parameters")
		return
	}

	// Parse limit if provided
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
		if limit > 100 {
			limit = 100 // Max limit
		}
	}

	records, nextCursor, err := h.service.ListRecords(repo, collection, limit, cursor)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to list records: %v", err))
		return
	}

	// Convert to output format
	recordOutputs := make([]RecordOutput, len(records))
	for i, record := range records {
		recordOutputs[i] = RecordOutput{
			URI:   record.URI,
			CID:   record.CID.String(),
			Value: json.RawMessage(record.Value),
		}
	}

	resp := ListRecordsResponse{
		Cursor:  nextCursor,
		Records: recordOutputs,
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetRepo handles GET /xrpc/com.atproto.sync.getRepo
func (h *RepositoryHandler) GetRepo(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	did := r.URL.Query().Get("did")
	if did == "" {
		writeError(w, http.StatusBadRequest, "missing did parameter")
		return
	}

	// Export repository as CAR file
	carData, err := h.service.ExportRepository(did)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "repository not found")
			return
		}
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to export repository: %v", err))
		return
	}

	// Set appropriate headers for CAR file
	w.Header().Set("Content-Type", "application/vnd.ipld.car")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(carData)))
	w.WriteHeader(http.StatusOK)
	w.Write(carData)
}

// Additional repository management endpoints

// CreateRepository handles POST /xrpc/com.atproto.repo.createRepo
func (h *RepositoryHandler) CreateRepository(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DID string `json:"did"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid request: %v", err))
		return
	}

	if req.DID == "" {
		writeError(w, http.StatusBadRequest, "missing did")
		return
	}

	repo, err := h.service.CreateRepository(req.DID)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			writeError(w, http.StatusConflict, "repository already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create repository: %v", err))
		return
	}

	resp := struct {
		DID     string `json:"did"`
		HeadCID string `json:"head"`
	}{
		DID:     repo.DID,
		HeadCID: repo.HeadCID.String(),
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetCommit handles GET /xrpc/com.atproto.sync.getCommit
func (h *RepositoryHandler) GetCommit(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	did := r.URL.Query().Get("did")
	commitCIDStr := r.URL.Query().Get("cid")

	if did == "" || commitCIDStr == "" {
		writeError(w, http.StatusBadRequest, "missing required parameters")
		return
	}

	// Parse CID
	commitCID, err := cid.Parse(commitCIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid cid")
		return
	}

	commit, err := h.service.GetCommit(did, commitCID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeError(w, http.StatusNotFound, "commit not found")
			return
		}
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get commit: %v", err))
		return
	}

	resp := struct {
		CID       string  `json:"cid"`
		DID       string  `json:"did"`
		Version   int     `json:"version"`
		PrevCID   *string `json:"prev,omitempty"`
		DataCID   string  `json:"data"`
		Revision  string  `json:"rev"`
		Signature string  `json:"sig"`
		CreatedAt string  `json:"createdAt"`
	}{
		CID:       commit.CID.String(),
		DID:       commit.DID,
		Version:   commit.Version,
		DataCID:   commit.DataCID.String(),
		Revision:  commit.Revision,
		Signature: fmt.Sprintf("%x", commit.Signature),
		CreatedAt: commit.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if commit.PrevCID != nil {
		prev := commit.PrevCID.String()
		resp.PrevCID = &prev
	}

	writeJSON(w, http.StatusOK, resp)
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   http.StatusText(status),
		"message": message,
	})
}

// GenericRecord is a temporary structure for CBOR encoding
// In a real implementation, you would have specific types for each lexicon
type GenericRecord struct {
	Data json.RawMessage
}

// MarshalCBOR implements the CBORMarshaler interface
func (g *GenericRecord) MarshalCBOR(w io.Writer) error {
	// Parse JSON data into a generic map for proper CBOR encoding
	var data map[string]interface{}
	if err := json.Unmarshal(g.Data, &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}
	
	// Use IPFS CBOR encoding to properly encode the data
	cborData, err := cbornode.DumpObject(data)
	if err != nil {
		return fmt.Errorf("failed to marshal as CBOR: %w", err)
	}
	
	_, err = w.Write(cborData)
	if err != nil {
		return fmt.Errorf("failed to write CBOR data: %w", err)
	}
	
	return nil
}