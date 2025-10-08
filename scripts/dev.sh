#!/bin/bash

# Development setup script
echo "🚀 Setting up Mobile Backend Development Environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp env.example .env
    echo "✅ .env file created. Please update the values as needed."
fi

# Create necessary directories
echo "📁 Creating necessary directories..."
mkdir -p backend/uploads
mkdir -p backend/logs
mkdir -p monitoring/prometheus
mkdir -p monitoring/grafana/dashboards
mkdir -p monitoring/grafana/datasources

# Start development environment
echo "🐳 Starting development environment..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if services are running
echo "🔍 Checking service health..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ Backend is running at http://localhost:8080"
else
    echo "❌ Backend is not responding"
fi

if curl -s http://localhost:3000 > /dev/null; then
    echo "✅ Grafana is running at http://localhost:3000 (admin/admin)"
else
    echo "❌ Grafana is not responding"
fi

if curl -s http://localhost:9090 > /dev/null; then
    echo "✅ Prometheus is running at http://localhost:9090"
else
    echo "❌ Prometheus is not responding"
fi

if curl -s http://localhost:16686 > /dev/null; then
    echo "✅ Jaeger is running at http://localhost:16686"
else
    echo "❌ Jaeger is not responding"
fi

echo ""
echo "🎉 Development environment is ready!"
echo ""
echo "📋 Available services:"
echo "  - Backend API: http://localhost:8080"
echo "  - API Docs: http://localhost:8080/swagger/index.html"
echo "  - Database: localhost:5432"
echo "  - Redis: localhost:6379"
echo "  - Prometheus: http://localhost:9090"
echo "  - Grafana: http://localhost:3000 (admin/admin)"
echo "  - Jaeger: http://localhost:16686"
echo ""
echo "🔧 Useful commands:"
echo "  - View logs: make dev-logs"
echo "  - Stop services: make dev-stop"
echo "  - Run tests: make test"
echo "  - Run linter: make lint"
