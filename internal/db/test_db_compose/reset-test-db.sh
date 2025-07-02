#!/bin/bash
# Reset the test database by removing all data

echo "WARNING: This will delete all test database data!"
echo "Press Ctrl+C to cancel, or Enter to continue..."
read

echo "Stopping test database..."
docker-compose -f docker-compose.yml down

echo "Removing test data volume..."
rm -rf ~/Code/Coves/test_db_data

echo "Starting fresh test database..."
./start-test-db.sh