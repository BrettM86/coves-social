package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"Coves/internal/core/repository"
	"github.com/ipfs/go-cid"
	"github.com/lib/pq"
)

// RepositoryRepo implements repository.RepositoryRepository using PostgreSQL
type RepositoryRepo struct {
	db *sql.DB
}

// NewRepositoryRepo creates a new PostgreSQL repository implementation
func NewRepositoryRepo(db *sql.DB) *RepositoryRepo {
	return &RepositoryRepo{db: db}
}

// Repository operations

func (r *RepositoryRepo) Create(repo *repository.Repository) error {
	query := `
		INSERT INTO repositories (did, head_cid, revision, record_count, storage_size, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err := r.db.Exec(query,
		repo.DID,
		repo.HeadCID.String(),
		repo.Revision,
		repo.RecordCount,
		repo.StorageSize,
		repo.CreatedAt,
		repo.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}
	
	return nil
}

func (r *RepositoryRepo) GetByDID(did string) (*repository.Repository, error) {
	query := `
		SELECT did, head_cid, revision, record_count, storage_size, created_at, updated_at
		FROM repositories
		WHERE did = $1`
	
	var repo repository.Repository
	var headCIDStr string
	
	err := r.db.QueryRow(query, did).Scan(
		&repo.DID,
		&headCIDStr,
		&repo.Revision,
		&repo.RecordCount,
		&repo.StorageSize,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	
	repo.HeadCID, err = cid.Parse(headCIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse head CID: %w", err)
	}
	
	return &repo, nil
}

func (r *RepositoryRepo) Update(repo *repository.Repository) error {
	query := `
		UPDATE repositories
		SET head_cid = $2, revision = $3, record_count = $4, storage_size = $5, updated_at = $6
		WHERE did = $1`
	
	result, err := r.db.Exec(query,
		repo.DID,
		repo.HeadCID.String(),
		repo.Revision,
		repo.RecordCount,
		repo.StorageSize,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("repository not found: %s", repo.DID)
	}
	
	return nil
}

func (r *RepositoryRepo) Delete(did string) error {
	query := `DELETE FROM repositories WHERE did = $1`
	
	result, err := r.db.Exec(query, did)
	if err != nil {
		return fmt.Errorf("failed to delete repository: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("repository not found: %s", did)
	}
	
	return nil
}

// Commit operations

func (r *RepositoryRepo) CreateCommit(commit *repository.Commit) error {
	query := `
		INSERT INTO commits (cid, did, version, prev_cid, data_cid, revision, signature, signing_key_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	
	var prevCID *string
	if commit.PrevCID != nil {
		s := commit.PrevCID.String()
		prevCID = &s
	}
	
	_, err := r.db.Exec(query,
		commit.CID.String(),
		commit.DID,
		commit.Version,
		prevCID,
		commit.DataCID.String(),
		commit.Revision,
		commit.Signature,
		commit.SigningKeyID,
		commit.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}
	
	return nil
}

func (r *RepositoryRepo) GetCommit(did string, commitCID cid.Cid) (*repository.Commit, error) {
	query := `
		SELECT cid, did, version, prev_cid, data_cid, revision, signature, signing_key_id, created_at
		FROM commits
		WHERE did = $1 AND cid = $2`
	
	var commit repository.Commit
	var cidStr, dataCIDStr string
	var prevCIDStr sql.NullString
	
	err := r.db.QueryRow(query, did, commitCID.String()).Scan(
		&cidStr,
		&commit.DID,
		&commit.Version,
		&prevCIDStr,
		&dataCIDStr,
		&commit.Revision,
		&commit.Signature,
		&commit.SigningKeyID,
		&commit.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}
	
	commit.CID, err = cid.Parse(cidStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse commit CID: %w", err)
	}
	
	commit.DataCID, err = cid.Parse(dataCIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data CID: %w", err)
	}
	
	if prevCIDStr.Valid {
		prevCID, err := cid.Parse(prevCIDStr.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse prev CID: %w", err)
		}
		commit.PrevCID = &prevCID
	}
	
	return &commit, nil
}

func (r *RepositoryRepo) GetLatestCommit(did string) (*repository.Commit, error) {
	query := `
		SELECT cid, did, version, prev_cid, data_cid, revision, signature, signing_key_id, created_at
		FROM commits
		WHERE did = $1
		ORDER BY created_at DESC
		LIMIT 1`
	
	var commit repository.Commit
	var cidStr, dataCIDStr string
	var prevCIDStr sql.NullString
	
	err := r.db.QueryRow(query, did).Scan(
		&cidStr,
		&commit.DID,
		&commit.Version,
		&prevCIDStr,
		&dataCIDStr,
		&commit.Revision,
		&commit.Signature,
		&commit.SigningKeyID,
		&commit.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest commit: %w", err)
	}
	
	commit.CID, err = cid.Parse(cidStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse commit CID: %w", err)
	}
	
	commit.DataCID, err = cid.Parse(dataCIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data CID: %w", err)
	}
	
	if prevCIDStr.Valid {
		prevCID, err := cid.Parse(prevCIDStr.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse prev CID: %w", err)
		}
		commit.PrevCID = &prevCID
	}
	
	return &commit, nil
}

func (r *RepositoryRepo) ListCommits(did string, limit int, offset int) ([]*repository.Commit, error) {
	query := `
		SELECT cid, did, version, prev_cid, data_cid, revision, signature, signing_key_id, created_at
		FROM commits
		WHERE did = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`
	
	rows, err := r.db.Query(query, did, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %w", err)
	}
	defer rows.Close()
	
	var commits []*repository.Commit
	for rows.Next() {
		var commit repository.Commit
		var cidStr, dataCIDStr string
		var prevCIDStr sql.NullString
		
		err := rows.Scan(
			&cidStr,
			&commit.DID,
			&commit.Version,
			&prevCIDStr,
			&dataCIDStr,
			&commit.Revision,
			&commit.Signature,
			&commit.SigningKeyID,
			&commit.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan commit: %w", err)
		}
		
		commit.CID, err = cid.Parse(cidStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse commit CID: %w", err)
		}
		
		commit.DataCID, err = cid.Parse(dataCIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse data CID: %w", err)
		}
		
		if prevCIDStr.Valid {
			prevCID, err := cid.Parse(prevCIDStr.String)
			if err != nil {
				return nil, fmt.Errorf("failed to parse prev CID: %w", err)
			}
			commit.PrevCID = &prevCID
		}
		
		commits = append(commits, &commit)
	}
	
	return commits, nil
}

// Record operations

func (r *RepositoryRepo) CreateRecord(record *repository.Record) error {
	query := `
		INSERT INTO records (did, uri, cid, collection, record_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err := r.db.Exec(query,
		record.URI[:len("at://")+len(record.URI[len("at://"):])-len(record.Collection)-len(record.RecordKey)-2], // Extract DID from URI
		record.URI,
		record.CID.String(),
		record.Collection,
		record.RecordKey,
		record.CreatedAt,
		record.UpdatedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // unique_violation
			return fmt.Errorf("record already exists: %s", record.URI)
		}
		return fmt.Errorf("failed to create record: %w", err)
	}
	
	return nil
}

func (r *RepositoryRepo) GetRecord(did string, collection string, recordKey string) (*repository.Record, error) {
	query := `
		SELECT uri, cid, collection, record_key, created_at, updated_at
		FROM records
		WHERE did = $1 AND collection = $2 AND record_key = $3`
	
	var record repository.Record
	var cidStr string
	
	err := r.db.QueryRow(query, did, collection, recordKey).Scan(
		&record.URI,
		&cidStr,
		&record.Collection,
		&record.RecordKey,
		&record.CreatedAt,
		&record.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get record: %w", err)
	}
	
	record.CID, err = cid.Parse(cidStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse record CID: %w", err)
	}
	
	return &record, nil
}

func (r *RepositoryRepo) UpdateRecord(record *repository.Record) error {
	did := record.URI[:len("at://")+len(record.URI[len("at://"):])-len(record.Collection)-len(record.RecordKey)-2]
	
	query := `
		UPDATE records
		SET cid = $4, updated_at = $5
		WHERE did = $1 AND collection = $2 AND record_key = $3`
	
	result, err := r.db.Exec(query,
		did,
		record.Collection,
		record.RecordKey,
		record.CID.String(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("record not found: %s", record.URI)
	}
	
	return nil
}

func (r *RepositoryRepo) DeleteRecord(did string, collection string, recordKey string) error {
	query := `DELETE FROM records WHERE did = $1 AND collection = $2 AND record_key = $3`
	
	result, err := r.db.Exec(query, did, collection, recordKey)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("record not found")
	}
	
	return nil
}

func (r *RepositoryRepo) ListRecords(did string, collection string, limit int, offset int) ([]*repository.Record, error) {
	query := `
		SELECT uri, cid, collection, record_key, created_at, updated_at
		FROM records
		WHERE did = $1 AND collection = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`
	
	rows, err := r.db.Query(query, did, collection, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}
	defer rows.Close()
	
	var records []*repository.Record
	for rows.Next() {
		var record repository.Record
		var cidStr string
		
		err := rows.Scan(
			&record.URI,
			&cidStr,
			&record.Collection,
			&record.RecordKey,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}
		
		record.CID, err = cid.Parse(cidStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse record CID: %w", err)
		}
		
		records = append(records, &record)
	}
	
	return records, nil
}

