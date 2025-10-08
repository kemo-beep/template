#!/bin/bash

# Test runner script
echo "🧪 Running Mobile Backend Tests..."

# Change to backend directory
cd backend

# Run unit tests
echo "📋 Running unit tests..."
go test -v -short ./...

# Run integration tests
echo "📋 Running integration tests..."
go test -v -tags=integration ./...

# Run all tests with coverage
echo "📊 Running tests with coverage..."
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
echo "📈 Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo "✅ Tests completed!"
echo "📊 Coverage report generated: backend/coverage.html"
