#!/bin/bash
# Stop the test database

echo "Stopping Coves test database..."
docker-compose -f docker-compose.yml down

echo "Test database stopped."