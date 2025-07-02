-- +goose Up
-- +goose StatementBegin

-- Repositories table stores metadata about each user's repository
CREATE TABLE repositories (
    did VARCHAR(256) PRIMARY KEY,
    head_cid VARCHAR(256) NOT NULL,
    revision VARCHAR(64) NOT NULL,
    record_count INTEGER NOT NULL DEFAULT 0,
    storage_size BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_repositories_updated_at ON repositories(updated_at);

-- Commits table stores the commit history
CREATE TABLE commits (
    cid VARCHAR(256) PRIMARY KEY,
    did VARCHAR(256) NOT NULL,
    version INTEGER NOT NULL,
    prev_cid VARCHAR(256),
    data_cid VARCHAR(256) NOT NULL,
    revision VARCHAR(64) NOT NULL,
    signature BYTEA NOT NULL,
    signing_key_id VARCHAR(256) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (did) REFERENCES repositories(did) ON DELETE CASCADE
);

CREATE INDEX idx_commits_did ON commits(did);
CREATE INDEX idx_commits_created_at ON commits(created_at);

-- Records table stores record metadata (actual data is in MST)
CREATE TABLE records (
    id SERIAL PRIMARY KEY,
    did VARCHAR(256) NOT NULL,
    uri VARCHAR(512) NOT NULL,
    cid VARCHAR(256) NOT NULL,
    collection VARCHAR(256) NOT NULL,
    record_key VARCHAR(256) NOT NULL,
    value BYTEA NOT NULL, -- CBOR-encoded record data
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(did, collection, record_key),
    FOREIGN KEY (did) REFERENCES repositories(did) ON DELETE CASCADE
);

CREATE INDEX idx_records_did_collection ON records(did, collection);
CREATE INDEX idx_records_uri ON records(uri);
CREATE INDEX idx_records_updated_at ON records(updated_at);

-- Blobs table stores binary large objects
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

-- Blob references table tracks which records reference which blobs
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

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS blob_refs;
DROP TABLE IF EXISTS blobs;
DROP TABLE IF EXISTS records;
DROP TABLE IF EXISTS commits;
DROP TABLE IF EXISTS repositories;
-- +goose StatementEnd