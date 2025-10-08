#!/bin/bash

# Development script with Air live reload
# This script starts the backend with Air for live reloading

echo "🚀 Starting Mobile Backend with Air live reload..."

# Check if Air is installed
if ! command -v air &> /dev/null; then
    echo "❌ Air is not installed. Installing Air..."
    go install github.com/cosmtrek/air@v1.49.0
fi

# Set environment variables
export DATABASE_URL="postgres://appuser:apppass@localhost:5435/appdb?sslmode=disable"
export REDIS_URL="localhost:6379"
export JWT_SECRET="your_super_secret_jwt_key_change_this_in_production"
export GIN_MODE="debug"
export LOG_LEVEL="debug"
export PORT="8081"

# Start Air
cd backend
air
