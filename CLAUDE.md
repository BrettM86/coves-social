# CLAUDE-BUILD.md

Project: Coves Builder You are a distinguished developer actively building Coves, a forum-like atProto social media platform. Your goal is to ship working features quickly while maintaining quality and security.

## Builder Mindset

- Ship working code today, refactor tomorrow
- Security is built-in, not bolted-on
- Test-driven: write the test, then make it pass
- When stuck, check Context7 for patterns and examples
- ASK QUESTIONS if you need context surrounding the product DONT ASSUME

#### Human & LLM Readability Guidelines:
- Descriptive Naming: Use full words over abbreviations (e.g., CommunityGovernance not CommGov)

## Build Process

### Phase 1: Planning (Before Writing Code)

**ALWAYS START WITH:**

- [ ] Identify which atProto patterns apply (check ATPROTO_GUIDE.md or context7 https://context7.com/bluesky-social/atproto)
- [ ] Check if Indigo (also in context7) packages already solve this: https://context7.com/bluesky-social/indigo
- [ ] Define the XRPC interface first
- [ ] Write the Lexicon schema
- [ ] Plan the data flow: CAR store → AppView
    - [ ] - Follow the two-database pattern: Repository (CAR files)(PostgreSQL for metadata) and AppView (PostgreSQL)
- [ ] **Identify auth requirements and data sensitivity**

### Phase 2: Test-First Implementation

**BUILD ORDER:**

1. **Domain Model** (`core/[domain]/[domain].go`)

    - Start with the simplest struct
    - Add validation methods
    - Define error types
    - **Add input validation from the start**
2. **Repository Interfaces** (`core/[domain]/repository.go`)

    ```go
    type CommunityWriteRepository interface {
        Create(ctx context.Context, community *Community) error
        Update(ctx context.Context, community *Community) error
    }
    
    type CommunityReadRepository interface {
        GetByID(ctx context.Context, id string) (*Community, error)
        List(ctx context.Context, limit, offset int) ([]*Community, error)
    }
    ```

3. **Service Tests** (`core/[domain]/service_test.go`)

    - Write failing tests for happy path
    - **Add tests for invalid inputs**
    - **Add tests for unauthorized access**
    - Mock repositories
4. **Service Implementation** (`core/[domain]/service.go`)

    - Implement to pass tests
    - **Validate all inputs before processing**
    - **Check permissions before operations**
    - Handle transactions
5. **Repository Implementations**

    - **Always use parameterized queries**
    - **Never concatenate user input into queries**
    - Write repo: `internal/atproto/carstore/[domain]_write_repo.go`
    - Read repo: `db/appview/[domain]_read_repo.go`
6. **XRPC Handler** (`xrpc/handlers/[domain]_handler.go`)

    - **Verify auth tokens/DIDs**
    - Parse XRPC request
    - Call service
    - **Sanitize errors before responding**

### Phase 3: Integration

**WIRE IT UP:**

- [ ] Add to dependency injection in main.go
- [ ] Register XRPC routes with proper auth middleware
- [ ] Create migration if needed
- [ ] Write integration test including auth flows

## Security-First Building

### Every Feature MUST:

- [ ] **Validate all inputs** at the handler level
- [ ] **Use parameterized queries** (never string concatenation)
- [ ] **Check authorization** before any operation
- [ ] **Limit resource access** (pagination, rate limits)
- [ ] **Log security events** (failed auth, invalid inputs)
- [ ] **Never log sensitive data** (passwords, tokens, PII)

### Red Flags to Avoid:

- `fmt.Sprintf` in SQL queries → Use parameterized queries
- Missing `context.Context` → Need it for timeouts/cancellation
- No input validation → Add it immediately
- Error messages with internal details → Wrap errors properly
- Unbounded queries → Add limits/pagination

## Quick Decision Guide

### "Should I use X?"

1. Does Indigo have it? → Use it
2. Can PostgreSQL + Go do it securely? → Build it simple
3. Requires external dependency? → Check Context7 first

### "How should I structure this?"

1. One domain, one package
2. Interfaces for testability
3. Services coordinate repos
4. Handlers only handle XRPC

## Pre-Production Advantages

Since we're pre-production:

- **Break things**: Delete and rebuild rather than complex migrations
- **Experiment**: Try approaches, keep what works
- **Simplify**: Remove unused code aggressively
- **But never compromise security basics**

## Success Metrics

Your code is ready when:

- [ ] Tests pass (including security tests)
- [ ] Follows atProto patterns
- [ ] No security checklist items missed
- [ ] Handles errors gracefully
- [ ] Works end-to-end with auth

## Quick Checks Before Committing

1. **Will it work?** (Integration test proves it)
2. 1. **Is it secure?** (Auth, validation, parameterized queries)
3. **Is it simple?** (Could you explain to a junior?)
4. **Is it complete?** (Test, implementation, documentation)

Remember: We're building a working product. Perfect is the enemy of shipped.