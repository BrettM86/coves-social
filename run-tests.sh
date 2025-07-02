#!/bin/bash
# Helper script to run tests with the test database

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🧪 Coves Test Runner"
echo "==================="
echo ""

# Check if test database is running
if ! nc -z localhost 5434 2>/dev/null; then
    echo -e "${RED}❌ Test database is not running${NC}"
    echo ""
    echo "Starting test database..."
    cd internal/db/test_db_compose && ./start-test-db.sh
    cd ../../..
    echo ""
fi

# Load test environment
if [ -f .env.test ]; then
    export $(cat .env.test | grep -v '^#' | xargs)
fi

# Run tests
echo "Running tests..."
echo ""

if [ $# -eq 0 ]; then
    # No arguments, run all tests
    go test -v ./...
else
    # Pass arguments to go test
    go test -v "$@"
fi

TEST_RESULT=$?

if [ $TEST_RESULT -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ All tests passed!${NC}"
else
    echo ""
    echo -e "${RED}❌ Some tests failed${NC}"
fi

exit $TEST_RESULT