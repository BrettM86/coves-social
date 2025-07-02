# CarStore Package

This package provides integration with Indigo's carstore for managing ATProto repository CAR files in the Coves platform.

## Overview

The carstore package wraps Indigo's carstore implementation to provide:
- Filesystem-based storage of CAR (Content Addressable aRchive) files
- PostgreSQL metadata tracking via GORM
- DID to UID mapping for user repositories
- Automatic garbage collection and compaction

## Architecture

```
[Repository Service]
         ↓
    [RepoStore]     ← Provides DID-based interface
         ↓
    [CarStore]      ← Wraps Indigo's carstore
         ↓
[Indigo CarStore]   ← Actual implementation
         ↓
[PostgreSQL + Filesystem]
```

## Components

### CarStore (`carstore.go`)
Wraps Indigo's carstore implementation, providing methods for:
- `ImportSlice`: Import CAR data for a user
- `ReadUserCar`: Export user's repository as CAR
- `GetUserRepoHead`: Get latest repository state
- `CompactUserShards`: Run garbage collection
- `WipeUserData`: Delete all user data

### UserMapping (`user_mapping.go`)
Maps DIDs (Decentralized Identifiers) to numeric UIDs required by Indigo's carstore:
- DIDs are strings like `did:plc:abc123xyz`
- UIDs are numeric identifiers (models.Uid)
- Maintains bidirectional mapping in PostgreSQL

### RepoStore (`repo_store.go`)
Combines CarStore with UserMapping to provide DID-based operations:
- `ImportRepo`: Import repository for a DID
- `ReadRepo`: Export repository for a DID
- `GetRepoHead`: Get latest state for a DID
- `CompactRepo`: Run garbage collection for a DID
- `DeleteRepo`: Remove all data for a DID

## Data Flow

### Creating a New Repository
1. Service calls `RepoStore.ImportRepo(did, carData)`
2. RepoStore maps DID to UID via UserMapping
3. CarStore imports the CAR slice
4. Indigo's carstore:
   - Stores CAR data as file on disk
   - Records metadata in PostgreSQL

### Reading a Repository
1. Service calls `RepoStore.ReadRepo(did)`
2. RepoStore maps DID to UID
3. CarStore reads user's CAR data
4. Returns complete CAR file

## Database Schema

### user_maps table
```sql
CREATE TABLE user_maps (
    uid SERIAL PRIMARY KEY,
    did VARCHAR UNIQUE NOT NULL,
    created_at BIGINT,
    updated_at BIGINT
);
```

### Indigo's tables (auto-created)
- `car_shards`: Metadata about CAR file shards
- `block_refs`: Block reference tracking

## Storage

CAR files are stored on the filesystem at the path specified during initialization (e.g., `./data/carstore/`). The storage is organized by Indigo's carstore implementation, typically with sharding for performance.

## Configuration

Initialize the carstore with:
```go
carDirs := []string{"./data/carstore"}
repoStore, err := carstore.NewRepoStore(gormDB, carDirs)
```

## Future Enhancements

Current implementation supports repository-level operations. Record-level CRUD operations would require:
1. Reading the CAR file
2. Parsing into a repository structure
3. Modifying records
4. Re-serializing as CAR
5. Writing back to carstore

This is planned for future XRPC implementation.