-- +goose Up
-- +goose StatementBegin
-- Drop our block_refs table since Indigo's carstore will create its own
DROP TABLE IF EXISTS block_refs;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Recreate block_refs table
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