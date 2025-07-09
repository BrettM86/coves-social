# Repository Testing Summary

## Lexicon Schema Validation

We use Indigo's lexicon validator to ensure all AT Protocol schemas are valid and properly structured.

### Validation Components:
1. **Schema Validation** (`cmd/validate-lexicon/main.go`)
   - Validates all 57 lexicon schema files are valid JSON
   - Checks cross-references between schemas
   - Validates test data files against their schemas

2. **Test Data** (`tests/lexicon-test-data/`)
   - Contains example records for validation testing
   - Files use naming convention:
     - `*-valid*.json` - Should pass validation
     - `*-invalid-*.json` - Should fail validation (tests error detection)
   - Currently covers 5 record types:
     - social.coves.actor.profile
     - social.coves.community.profile  
     - social.coves.post.record
     - social.coves.interaction.vote
     - social.coves.moderation.ban

3. **Validation Library** (`internal/validation/lexicon.go`)
   - Wrapper around Indigo's ValidateRecord
   - Provides type-specific validation methods
   - Supports multiple input formats

### Running Lexicon Validation
```bash
# Full validation (schemas + test data) - DEFAULT
go run cmd/validate-lexicon/main.go

# Schemas only (skip test data validation)
go run cmd/validate-lexicon/main.go --schemas-only

# Verbose output
go run cmd/validate-lexicon/main.go -v

# Strict validation mode
go run cmd/validate-lexicon/main.go --strict
```

### Test Coverage Warning
The validator explicitly outputs which record types have test data and which don't. This prevents false confidence from passing tests when schemas lack test coverage.

Example output:
```
📋 Validation Summary:
  Valid test files:   5/5 passed
  Invalid test files: 2/2 correctly rejected

  ✅ All test files behaved as expected!

📊 Test Data Coverage Summary:
  - Records with test data: 5 types
  - Valid test files: 5
  - Invalid test files: 2 (for error validation)

  Tested record types:
    ✓ social.coves.actor.profile
    ✓ social.coves.community.profile
    ✓ social.coves.post.record
    ✓ social.coves.interaction.vote
    ✓ social.coves.moderation.ban

  ⚠️  Record types without test data:
    - social.coves.actor.membership
    - social.coves.actor.subscription
    - social.coves.community.rules
    ... (8 more)

  Coverage: 5/16 record types have test data (31.2%)
```

### Running Tests
```bash
# Run lexicon validation tests
go test -v ./tests/lexicon_validation_test.go

# Run validation library tests  
go test -v ./internal/validation/...
```

## Test Infrastructure Setup
- Created Docker Compose configuration for isolated test database on port 5434
- Test database is completely separate from development (5433) and production (5432)
- Configuration location: `/internal/db/test_db_compose/docker-compose.yml`

## Repository Service Implementation
Successfully integrated Indigo's carstore for ATProto repository management:

### Key Components:
1. **CarStore Wrapper** (`/internal/atproto/carstore/carstore.go`)
   - Wraps Indigo's carstore implementation
   - Manages CAR file storage with PostgreSQL metadata

2. **RepoStore** (`/internal/atproto/carstore/repo_store.go`)
   - Combines CarStore with UserMapping for DID-based access
   - Handles DID to UID conversions transparently

3. **UserMapping** (`/internal/atproto/carstore/user_mapping.go`)
   - Maps ATProto DIDs to numeric UIDs (required by Indigo)
   - Auto-creates user_maps table via GORM

4. **Repository Service** (`/internal/core/repository/service.go`)
   - Updated to use Indigo's carstore instead of custom implementation
   - Handles empty repositories gracefully
   - Placeholder CID for empty repos until records are added

## Test Results
All repository tests passing:
- ✅ CreateRepository - Creates user mapping and repository record
- ✅ ImportExport - Handles empty CAR data correctly
- ✅ DeleteRepository - Removes repository and carstore data
- ✅ CompactRepository - Runs garbage collection
- ✅ UserMapping - DID to UID mapping works correctly

## Implementation Notes
1. **Empty Repositories**: Since Indigo's carstore expects actual CAR data, we handle empty repositories by:
   - Creating user mapping only
   - Using placeholder CID
   - Returning empty byte array on export
   - Actual CAR data will be created when records are added

2. **Database Tables**: Indigo's carstore auto-creates:
   - `user_maps` (DID ↔ UID mapping)
   - `car_shards` (CAR file metadata)
   - `block_refs` (IPLD block references)

3. **Migration**: Created migration to drop our custom block_refs table to avoid conflicts

## Next Steps
To fully utilize the carstore, implement:
1. Record CRUD operations using carstore's DeltaSession
2. Proper CAR file generation when adding records
3. Commit tracking with proper signatures
4. Repository versioning and history

## Running Tests
```bash
# Start test database
cd internal/db/test_db_compose
docker-compose up -d

# Run repository tests
TEST_DATABASE_URL="postgres://test_user:test_password@localhost:5434/coves_test?sslmode=disable" \
  go test -v ./internal/core/repository/...

# Stop test database
docker-compose down
```