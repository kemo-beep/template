#!/bin/bash

# Wait for database to be ready
echo "Waiting for database to be ready..."

# Wait for PostgreSQL to be ready
until pg_isready -h postgres -p 5432 -U appuser -d appdb; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 2
done

echo "PostgreSQL is ready!"

# Wait for Redis to be ready
until redis-cli -h redis -p 6379 ping; do
  echo "Redis is unavailable - sleeping"
  sleep 2
done

echo "Redis is ready!"

# Start the application
echo "Starting mobile backend..."
exec "$@"
