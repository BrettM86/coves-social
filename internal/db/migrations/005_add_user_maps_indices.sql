-- +goose Up
-- +goose StatementBegin

-- Note: The user_maps table is created by GORM's AutoMigrate in the carstore package
-- Only add indices if the table exists
DO $$
BEGIN
    -- Check if user_maps table exists
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_maps') THEN
        -- Check if column exists before creating index
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_maps' AND column_name = 'did') THEN
            -- Explicit column name specified in GORM tag
            CREATE INDEX IF NOT EXISTS idx_user_maps_did ON user_maps(did);
        END IF;
        
        -- Add index on created_at if column exists
        IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'user_maps' AND column_name = 'created_at') THEN
            CREATE INDEX IF NOT EXISTS idx_user_maps_created_at ON user_maps(created_at);
        END IF;
    END IF;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove indices if they exist
DROP INDEX IF EXISTS idx_user_maps_did;
DROP INDEX IF EXISTS idx_user_maps_created_at;

-- +goose StatementEnd