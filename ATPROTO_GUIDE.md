# AT Protocol Implementation Guide

This guide provides comprehensive information about implementing AT Protocol (atproto) in the Coves platform.

## Table of Contents
- [Core Concepts](#core-concepts)
- [Architecture Overview](#architecture-overview)
- [Lexicons](#lexicons)
- [XRPC](#xrpc)
- [Data Storage](#data-storage)
- [Identity & Authentication](#identity--authentication)
- [Firehose & Sync](#firehose--sync)
- [Go Implementation Patterns](#go-implementation-patterns)
- [Best Practices](#best-practices)

## Core Concepts

### What is AT Protocol?
AT Protocol is a federated social networking protocol that enables:
- **Decentralized identity** - Users own their identity (DID) and can move between providers
- **Data portability** - Users can export and migrate their full social graph and content
- **Interoperability** - Different apps can interact with the same underlying data

### Key Components
1. **DIDs (Decentralized Identifiers)** - Persistent user identifiers (e.g., `did:plc:xyz123`)
2. **Handles** - Human-readable names that resolve to DIDs (e.g., `alice.bsky.social`)
3. **Repositories** - User data stored as signed Merkle trees in CAR files
4. **Lexicons** - Schema definitions for data types and API methods
5. **XRPC** - The RPC protocol for client-server communication
6. **Firehose** - Real-time event stream of repository changes

## Architecture Overview

### Two-Database Pattern
AT Protocol requires two distinct data stores:

#### 1. Repository Database (Source of Truth)
- **Purpose**: Stores user-generated content as immutable, signed records
- **Storage**: CAR files containing Merkle trees + PostgreSQL metadata
- **Access**: Through XRPC procedures that modify repositories
- **Properties**:
  - Append-only (soft deletes via tombstones)
  - Cryptographically verifiable
  - User-controlled and portable

#### 2. AppView Database (Query Layer)
- **Purpose**: Denormalized, indexed data optimized for queries
- **Storage**: PostgreSQL with application-specific schema
- **Access**: Through XRPC queries (read-only)
- **Properties**:
  - Eventually consistent with repositories
  - Can be rebuilt from repository data
  - Application-specific aggregations

### Data Flow

```
Write Path:
Client → XRPC Procedure → Service → Write Repo → CAR Store
                                           ↓
                                    Firehose Event
                                           ↓
                                    AppView Indexer
                                           ↓
                                    AppView Database

Read Path:
Client → XRPC Query → Service → Read Repo → AppView Database
```

## Lexicons

### What are Lexicons?
Lexicons are JSON schema files that define:
- Data types (records stored in repositories)
- API methods (queries and procedures)
- Input/output schemas for API calls

### Lexicon Structure
```json
{
  "lexicon": 1,
  "id": "social.coves.community.profile",
  "defs": {
    "main": {
      "type": "record",
      "key": "self",
      "record": {
        "type": "object",
        "required": ["name", "createdAt"],
        "properties": {
          "name": {"type": "string", "maxLength": 64},
          "description": {"type": "string", "maxLength": 256},
          "rules": {"type": "array", "items": {"type": "string"}},
          "createdAt": {"type": "string", "format": "datetime"}
        }
      }
    }
  }
}
```

### Lexicon Types

#### 1. Record Types
Define data structures stored in user repositories:
```json
{
  "type": "record",
  "key": "tid|rkey|literal",
  "record": { /* schema */ }
}
```

#### 2. Query Types (Read-only)
Define read operations that don't modify state:
```json
{
  "type": "query",
  "parameters": { /* input schema */ },
  "output": { /* response schema */ }
}
```

#### 3. Procedure Types (Write)
Define operations that modify repositories:
```json
{
  "type": "procedure", 
  "input": { /* request body schema */ },
  "output": { /* response schema */ }
}
```

### Naming Conventions
- Use reverse-DNS format: `social.coves.community.profile`
- Queries often start with `get`, `list`, or `search`
- Procedures often start with `create`, `update`, or `delete`
- Keep names descriptive but concise

## XRPC

### What is XRPC?
XRPC (Cross-Protocol RPC) is AT Protocol's HTTP-based RPC system:
- All methods live under `/xrpc/` path
- Method names map directly to Lexicon IDs
- Supports both JSON and binary data

### Request Format
```
# Query (GET)
GET /xrpc/social.coves.community.getCommunity?id=123

# Procedure (POST)
POST /xrpc/social.coves.community.createPost
Content-Type: application/json
Authorization: Bearer <token>

{"text": "Hello, Coves!"}
```

### Authentication
- Uses Bearer tokens in Authorization header
- Tokens are JWTs signed by the user's signing key
- Service auth for server-to-server calls

## Data Storage

### CAR Files
Content Addressable archive files store repository data:
- Contains IPLD blocks forming a Merkle tree
- Each block identified by CID (Content IDentifier)
- Enables cryptographic verification and efficient sync

### Record Keys (rkeys)
- Unique identifiers for records within a collection
- Can be TIDs (timestamp-based) or custom strings
- Must match pattern: `[a-zA-Z0-9._~-]{1,512}`

### Repository Structure
```
Repository (did:plc:user123)
├── social.coves.post
│   ├── 3kkreaz3amd27 (TID)
│   └── 3kkreaz3amd28 (TID)
├── social.coves.community.member
│   ├── community123
│   └── community456
└── app.bsky.actor.profile
    └── self
```

## Identity & Authentication

### DIDs (Decentralized Identifiers)
- Permanent, unique identifiers for users
- Two types supported:
  - `did:plc:*` - Hosted by PLC Directory
  - `did:web:*` - Self-hosted

### Handle Resolution
Handles resolve to DIDs via:
1. DNS TXT record: `_atproto.alice.com → did:plc:xyz`
2. HTTPS well-known: `https://alice.com/.well-known/atproto-did`

### Authentication Flow
1. Client creates session with identifier/password
2. Server returns access/refresh tokens
3. Client uses access token for API requests
4. Refresh when access token expires

## Firehose & Sync

### Firehose Events
Real-time stream of repository changes:
- Commit events (creates, updates, deletes)
- Identity events (handle changes)
- Account events (status changes)

### Subscribing to Firehose
Connect via WebSocket to `com.atproto.sync.subscribeRepos`:
```
wss://bsky.network/xrpc/com.atproto.sync.subscribeRepos
```

### Processing Events
- Events include full record data and operation type
- Process events to update AppView database
- Handle out-of-order events with sequence numbers

## Go Implementation Patterns

### Using Indigo Library
Bluesky's official Go implementation provides:
- Lexicon code generation
- CAR file handling
- XRPC client/server
- Firehose subscription

### Code Generation
Generate Go types from Lexicons:
```bash
go run github.com/bluesky-social/indigo/cmd/lexgen \
  --package coves \
  --prefix social.coves \
  --outdir api/coves \
  lexicons/social/coves/*.json
```

### Repository Operations
```go
// Write to repository
rkey := models.GenerateTID()
err := repoStore.CreateRecord(ctx, userDID, "social.coves.post", rkey, &Post{
    Text:      "Hello",
    CreatedAt: time.Now().Format(time.RFC3339),
})

// Read from repository  
records, err := repoStore.ListRecords(ctx, userDID, "social.coves.post", limit, cursor)
```

### XRPC Handler Pattern
```go
func (s *Server) HandleGetCommunity(ctx context.Context) error {
    // 1. Parse and validate input
    id := xrpc.QueryParam(ctx, "id")
    
    // 2. Call service layer
    community, err := s.communityService.GetByID(ctx, id)
    if err != nil {
        return err
    }
    
    // 3. Return response
    return xrpc.WriteJSONResponse(ctx, community)
}
```

## Best Practices

### 1. Lexicon Design
- Keep schemas focused and single-purpose
- Use references (`$ref`) for shared types
- Version carefully - Lexicons are contracts
- Document thoroughly with descriptions

### 2. Data Modeling
- Store minimal data in repositories
- Denormalize extensively in AppView
- Use record keys that are meaningful
- Plan for data portability

### 3. Performance
- Batch firehose processing
- Use database transactions wisely
- Index AppView tables appropriately
- Cache frequently accessed data

### 4. Error Handling
- Use standard XRPC error codes
- Provide meaningful error messages
- Handle network failures gracefully
- Implement proper retry logic

### 5. Security
- Validate all inputs against Lexicons
- Verify signatures on repository data
- Rate limit API endpoints
- Sanitize user-generated content

### 6. Federation
- Design for multi-instance deployment
- Handle remote user identities
- Respect instance-specific policies
- Plan for cross-instance data sync

## Common Patterns

### Handling User Content
- Always validate against Lexicon schemas
- Store in user's repository via CAR files
- Index in AppView for efficient queries
- Emit firehose events for subscribers

## Resources

### Official Documentation
- [ATProto Specifications](https://atproto.com/specs)
- [Lexicon Documentation](https://atproto.com/specs/lexicon)
- [XRPC Specification](https://atproto.com/specs/xrpc)

### Reference Implementations
- [Indigo (Go)](https://github.com/bluesky-social/indigo)
- [ATProto SDK (TypeScript)](https://github.com/bluesky-social/atproto)

### Tools
- [Lexicon CLI](https://github.com/bluesky-social/atproto/tree/main/packages/lex-cli)
- [goat CLI](https://github.com/bluesky-social/indigo/tree/main/cmd/goat)