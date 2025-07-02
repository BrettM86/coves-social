# Test Database Setup

This directory contains the Docker Compose configuration for the Coves test database.

## Overview

The test database is a PostgreSQL instance specifically for running automated tests. It's completely isolated from development and production databases.

### Configuration

- **Port**: 5434 (different from dev: 5433, prod: 5432)
- **Database**: coves_test
- **User**: test_user
- **Password**: test_password
- **Data Volume**: ~/Code/Coves/test_db_data

## Usage

### Starting the Test Database

```bash
cd internal/db/test_db_compose
./start-test-db.sh
```

This will:
1. Start the PostgreSQL container
2. Wait for it to be ready
3. Display the connection string

### Running Tests

Once the database is running, you can run tests with:

```bash
TEST_DATABASE_URL=postgres://test_user:test_password@localhost:5434/coves_test?sslmode=disable go test -v ./...
```

Or set the environment variable:

```bash
export TEST_DATABASE_URL=postgres://test_user:test_password@localhost:5434/coves_test?sslmode=disable
go test -v ./...
```

### Stopping the Test Database

```bash
./stop-test-db.sh
```

### Resetting Test Data

To completely reset the test database (removes all data):

```bash
./reset-test-db.sh
```

## Test Isolation

The test database is isolated from other environments:

| Environment | Port | Database Name | User |
|------------|------|--------------|------|
| Test | 5434 | coves_test | test_user |
| Development | 5433 | coves_dev | dev_user |
| Production | 5432 | coves | (varies) |

## What Gets Tested

When tests run against this database, they will:

1. Run all migrations from `internal/db/migrations/`
2. Create Indigo carstore tables (via GORM auto-migration)
3. Test the full integration including:
   - Repository CRUD operations
   - CAR file metadata storage
   - User DID to UID mapping
   - Carstore operations

## CI/CD Integration

For CI/CD pipelines, you can use the same Docker Compose setup or connect to a dedicated test database instance.