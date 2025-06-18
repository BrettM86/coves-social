Project:
You are a distinguished developer helping build Coves, a forum like atProto social media platform (think reddit).

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

# Architecture Guidelines

## Required Layered Architecture
Follow this strict separation of concerns:
```
Handler (HTTP) → Service (Business Logic) → Repository (Data Access) → Database
```

## Directory Structure
```
internal/
├── api/
│   ├── handlers/     # HTTP request/response handling ONLY
│   └── routes/       # Route definitions
├── core/
│   └── [domain]/     # Business logic, domain models, service interfaces
│       ├── service.go     # Business logic implementation
│       ├── repository.go  # Data access interface
│       └── [domain].go    # Domain models
└── db/
    └── postgres/     # Database implementation details
        └── [domain]_repo.go  # Repository implementations
```

## Strict Prohibitions
- **NEVER** put SQL queries in handlers
- **NEVER** import database packages in handlers
- **NEVER** pass *sql.DB directly to handlers
- **NEVER** mix business logic with HTTP concerns
- **NEVER** bypass the service layer

## Required Patterns

### Handlers (HTTP Layer)
- Only handle HTTP concerns: parsing requests, formatting responses
- Delegate all business logic to services
- No direct database access

Example:
```go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    json.NewDecoder(r.Body).Decode(&req)
    
    user, err := h.userService.CreateUser(req)  // Delegate to service
    // Handle response formatting only
}
```

### Services (Business Layer)
- Contain all business logic and validation
- Use repository interfaces, never concrete implementations
- Handle transactions and complex operations

Example:
```go
type UserService struct {
    userRepo UserRepository  // Interface, not concrete type
}
```

### Repositories (Data Layer)
- Define interfaces in core/[domain]/
- Implement in db/postgres/
- Handle all SQL queries and database operations

Example:
```go
// Interface in core/users/repository.go
type UserRepository interface {
    Create(user User) (*User, error)
    GetByID(id int) (*User, error)
}

// Implementation in db/postgres/user_repo.go
type PostgresUserRepo struct {
    db *sql.DB
}
```

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
5. Update routes in api/routes/

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
userRepo := postgres.NewUserRepository(db)
userService := users.NewUserService(userRepo)
userHandler := handlers.NewUserHandler(userService)
```

## Error Handling
- Define custom error types in core/errors/
- Use error wrapping with context: fmt.Errorf("service: %w", err)
- Services return domain errors, handlers translate to HTTP status codes
- Never expose internal error details in API responses

### Context7 Usage Guidelines:
- Always check Context7 for best practices before implementing external integrations
- Use Context7 to understand proper error handling patterns for specific libraries
- Reference Context7 for testing patterns with external dependencies
- Consult Context7 for proper configuration patterns
