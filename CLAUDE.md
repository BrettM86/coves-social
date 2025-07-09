Project:  
You are a distinguished developer helping build Coves, a forum like atProto social media platform (think reddit / lemmy).

Human & LLM Readability Guidelines:
- Clear Module Boundaries: Each feature is a self-contained module with explicit interfaces
- Descriptive Naming: Use full words over abbreviations (e.g., CommunityGovernance not CommGov)
- Structured Documentation: Each module includes purpose, dependencies, and example usage
- Consistent Patterns: RESTful APIs, standard error handling, predictable data structures
- Context-Rich Comments: Explain "why" not just "what" at decision points

Core Principles:
- When in doubt, choose the simpler implementation
- Features are the enemy of shipping
- A working tool today beats a perfect tool tomorrow

Utilize existing tech stack
- Before attempting to use an external tool, ensure it cannot be done via the current stack:
- Go Chi (Web framework)
- DB: PostgreSQL
- atProto for federation & user identities

## atProto Guidelines

For comprehensive AT Protocol implementation details, see [ATPROTO_GUIDE.md](./ATPROTO_GUIDE.md).

Key principles:
- Utilize Bluesky's Indigo packages before building custom atProto functionality
- Everything is XRPC - no separate REST API layer needed
- Follow the two-database pattern: Repository (CAR files) and AppView (PostgreSQL)
- Design for federation and data portability from the start

# Architecture Guidelines

## Required Layered Architecture
Follow this strict separation of concerns:
```  
Handler (XRPC) → Service (Business Logic) → Repository (Data Access) → Database  
```  
- Handlers: XRPC request/response only
- Services: Business logic, uses both write/read repos
- Write Repos: CAR store operations
- Read Repos: AppView queries


## Directory Structure

For a detailed project structure with file-level details and implementation status, see [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md).

The project follows a layered architecture with clear separation between:
- **XRPC handlers** - atProto API layer
  - Only handle XRPC concerns: parsing requests, formatting responses
  - Delegate all business logic to services
  - No direct database access
- **Core business logic** - Domain services and models
  - Contains all business logic
  - Orchestrates between write and read repositories
  - Manages transactions and complex operations
- **Data repositories** - Split between CAR store writes and AppView reads
  - **Write Repositories** (`internal/atproto/carstore/*_write_repo.go`)
    - Modify CAR files (source of truth)
- **Read Repositories** (`db/appview/*_read_repo.go`)
  - Query denormalized PostgreSQL tables
  - Optimized for performance

## Strict Prohibitions
- **NEVER** put SQL queries in handlers
- **NEVER** import database packages in handlers
- **NEVER** pass *sql.DB directly to handlers
- **NEVER** mix business logic with XRPC concerns
- **NEVER** bypass the service layer

## Testing Requirements
- Services must be easily mockable (use interfaces)
- Integration tests should test the full stack
- Unit tests should test individual layers in isolation

Test File Naming:
- Unit tests: `[file]_test.go` in same directory
- Integration tests: `[feature]_integration_test.go` in tests/ directory

## Claude Code Instructions

### Code Generation Patterns
When creating new features:
1. Generate interface first in core/[domain]/
2. Generate test file with failing tests
3. Generate implementation to pass tests
4. Generate handler with tests
5. Update routes in xrpc/routes/

### Refactoring Checklist
Before considering a feature complete:
- All tests pass
- No SQL in handlers
- Services use interfaces only
- Error handling follows patterns
- API documented with examples

## Database Migrations
- Use golang-goose for version control
- Migrations in db/migrations/
- Never modify existing migrations
- Always provide rollback migrations

## Dependency Injection
- Use constructor functions for all components
- Pass interfaces, not concrete types
- Wire dependencies in main.go or cmd/server/main.go

Example dependency wiring:
```go  
// main.go  
userWriteRepo := carstore.NewUserWriteRepository(carStore)  
userReadRepo := appview.NewUserReadRepository(db)  
userService := users.NewUserService(userWriteRepo, userReadRepo)  
userHandler := xrpc.NewUserHandler(userService)  
```  

## Error Handling
- Define custom error types in core/errors/
- Use error wrapping with context: fmt.Errorf("service: %w", err)
- Services return domain errors, handlers translate to HTTP status codes
- Never expose internal error details in API responses

### Context7 Usage Guidelines:
- Always check Context7 for best practices before implementing external integrations and packages
- Use Context7 to understand proper error handling patterns for specific libraries
- Reference Context7 for testing patterns with external dependencies
- Consult Context7 for proper configuration patterns

## XRPC Implementation

For detailed XRPC patterns and Lexicon examples, see [ATPROTO_GUIDE.md](./ATPROTO_GUIDE.md#xrpc).

### Key Points
- All client interactions go through XRPC endpoints
- Handlers validate against Lexicon schemas automatically
- Queries are read-only, procedures modify repositories
- Every endpoint must have a corresponding Lexicon definition

Key note: we are pre-production, we do not need migration strategies, feel free to tear down and rebuild, however ensure to erase any unneeded data structures or code.