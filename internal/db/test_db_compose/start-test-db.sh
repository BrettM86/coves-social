#!/bin/bash
# Start the test database

echo "Starting Coves test database on port 5434..."
docker-compose -f docker-compose.yml up -d

# Wait for database to be ready
echo "Waiting for database to be ready..."
for i in {1..30}; do
    if docker-compose -f docker-compose.yml exec -T postgres_test pg_isready -U test_user -d coves_test -p 5434 &>/dev/null; then
        echo "Test database is ready!"
        echo ""
        echo "Connection string:"
        echo "TEST_DATABASE_URL=postgres://test_user:test_password@localhost:5434/coves_test?sslmode=disable"
        echo ""
        echo "To run tests:"
        echo "TEST_DATABASE_URL=postgres://test_user:test_password@localhost:5434/coves_test?sslmode=disable go test -v ./..."
        exit 0
    fi
    echo -n "."
    sleep 1
done

echo "Failed to start test database"
exit 1