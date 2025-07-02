-- +goose Up
-- +goose StatementBegin

-- WARNING: This migration removes blob storage tables. 
-- Ensure all blob data has been migrated to carstore before running this migration.
-- This migration is NOT reversible if blob data exists!

-- Remove the value column from records table since blocks are now stored in filesystem
ALTER TABLE records DROP COLUMN IF EXISTS value;

-- Drop blob-related tables since FileCarStore handles block storage
-- WARNING: This will permanently delete all blob data!
DROP TABLE IF EXISTS blob_refs;
DROP TABLE IF EXISTS blobs;

-- Create block_refs table for garbage collection tracking
CREATE TABLE block_refs (
    cid VARCHAR(256) NOT NULL,
    did VARCHAR(256) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (cid, did),
    FOREIGN KEY (did) REFERENCES repositories(did) ON DELETE CASCADE
);

CREATE INDEX idx_block_refs_did ON block_refs(did);
CREATE INDEX idx_block_refs_created_at ON block_refs(created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Recreate the original schema for rollback
DROP TABLE IF EXISTS block_refs;

-- Add back the value column to records table
ALTER TABLE records ADD COLUMN value BYTEA;

-- Recreate blobs table
CREATE TABLE blobs (
    cid VARCHAR(256) PRIMARY KEY,
    mime_type VARCHAR(256) NOT NULL,
    size BIGINT NOT NULL,
    ref_count INTEGER NOT NULL DEFAULT 0,
    data BYTEA NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_blobs_ref_count ON blobs(ref_count);
CREATE INDEX idx_blobs_created_at ON blobs(created_at);

-- Recreate blob_refs table
CREATE TABLE blob_refs (
    id SERIAL PRIMARY KEY,
    record_id INTEGER NOT NULL,
    blob_cid VARCHAR(256) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (record_id) REFERENCES records(id) ON DELETE CASCADE,
    FOREIGN KEY (blob_cid) REFERENCES blobs(cid) ON DELETE RESTRICT,
    UNIQUE(record_id, blob_cid)
);

CREATE INDEX idx_blob_refs_blob_cid ON blob_refs(blob_cid);

-- +goose StatementEnd