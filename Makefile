# Makefile
.PHONY: help build test clean dev prod lint security-scan

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
dev: ## Start development environment
	docker-compose up -d
	@echo "Development environment started!"
	@echo "Backend: http://localhost:8080"
	@echo "Database: localhost:5432"
	@echo "Redis: localhost:6379"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3000 (admin/admin)"
	@echo "Jaeger: http://localhost:16686"

dev-build: ## Build and start development environment
	docker-compose up -d --build

dev-logs: ## Show development logs
	docker-compose logs -f backend

dev-stop: ## Stop development environment
	docker-compose down

dev-air: ## Start development with Air live reload (requires local Go and services)
	@echo "Starting development with Air live reload..."
	@echo "Make sure PostgreSQL and Redis are running on ports 5435 and 6379"
	@echo "Starting Air..."
	cd backend && air

dev-air-setup: ## Install Air for live reload
	@echo "Installing Air for live reload..."
	go install github.com/cosmtrek/air@v1.49.0

# Production
prod: ## Start production environment
	docker-compose -f docker-compose.prod.yml up -d

prod-build: ## Build and start production environment
	docker-compose -f docker-compose.prod.yml up -d --build

prod-logs: ## Show production logs
	docker-compose -f docker-compose.prod.yml logs -f backend

prod-stop: ## Stop production environment
	docker-compose -f docker-compose.prod.yml down

# Testing
test: ## Run all tests
	cd backend && go test -v -race -coverprofile=coverage.out ./...

test-unit: ## Run unit tests only
	cd backend && go test -v -short ./...

test-integration: ## Run integration tests
	cd backend && go test -v -tags=integration ./...

test-coverage: ## Generate test coverage report
	cd backend && go test -v -race -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: backend/coverage.html"

# Code Quality
lint: ## Run linter
	cd backend && golangci-lint run

lint-fix: ## Fix linting issues
	cd backend && golangci-lint run --fix

format: ## Format code
	cd backend && go fmt ./...
	cd backend && goimports -w .

# Security
security-scan: ## Run security scan
	cd backend && gosec ./...
	cd backend && govulncheck ./...

# Database
db-migrate: ## Run database migrations (GORM AutoMigrate)
	cd backend && go run scripts/migrate.go

db-migrate-sql: ## Run SQL migrations (recommended for production)
	cd backend && go run scripts/run_migrations.go

db-seed: ## Seed database with test data
	cd backend && go run scripts/seed.go

db-reset: ## Reset database (drop, create, migrate, seed)
	docker-compose exec db psql -U appuser -d appdb -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	$(MAKE) db-migrate-sql
	$(MAKE) db-seed

db-status: ## Check migration status
	@echo "Checking database migration status..."
	@echo "SQL migrations available:"
	@ls -la backend/migrations/*.sql 2>/dev/null || echo "No SQL migrations found"
	@echo ""
	@echo "To run SQL migrations: make db-migrate-sql"
	@echo "To run GORM migrations: make db-migrate"

# Build
build: ## Build the application
	cd backend && go build -o bin/main .

build-linux: ## Build for Linux
	cd backend && GOOS=linux GOARCH=amd64 go build -o bin/main-linux .

# Docker
docker-build: ## Build Docker image
	docker build -t mobile-backend:latest ./backend

docker-build-prod: ## Build production Docker image
	docker build -f ./backend/Dockerfile.prod -t mobile-backend:prod ./backend

# Cleanup
clean: ## Clean up build artifacts
	cd backend && go clean
	cd backend && rm -rf bin/
	cd backend && rm -f coverage.out coverage.html
	docker-compose down -v
	docker system prune -f

# Dependencies
deps: ## Install dependencies
	cd backend && go mod download
	cd backend && go mod tidy

deps-update: ## Update dependencies
	cd backend && go get -u ./...
	cd backend && go mod tidy

# Documentation
docs: ## Generate API documentation
	cd backend && swag init -g main.go -o docs/

docs-serve: ## Serve API documentation
	@echo "API Documentation: http://localhost:8080/swagger/index.html"
	@echo "Make sure the backend is running first!"

# Monitoring
monitor: ## Start monitoring stack only
	docker-compose up -d prometheus grafana jaeger

monitor-stop: ## Stop monitoring stack
	docker-compose stop prometheus grafana jaeger

# Load Testing
load-test: ## Run load tests
	k6 run k6-load-test.js

# Backup
backup: ## Backup database
	docker-compose exec db pg_dump -U appuser appdb > backup_$(shell date +%Y%m%d_%H%M%S).sql

# Health Check
health: ## Check service health
	@echo "Checking service health..."
	@curl -s http://localhost:8081/health | jq .
	@curl -s http://localhost:8081/health/ready | jq .
	@curl -s http://localhost:8081/health/live | jq .

# API Generator
generator-setup: ## Setup API generator (create migrations directory)
	@echo "Setting up API generator..."
	cd backend && DATABASE_URL="postgres://appuser:apppass@localhost:5435/appdb?sslmode=disable" go run cmd/generator/main.go -action setup

generator-schema: ## Generate APIs from schema file
	@echo "Generating APIs from schema..."
	@read -p "Enter schema file path: " file; \
	cd backend && DATABASE_URL="postgres://appuser:apppass@localhost:5435/appdb?sslmode=disable" go run cmd/generator/main.go -action generate -schema $$file

generator-migrate: ## Run migration and generate APIs
	@echo "Running migration..."
	@read -p "Enter migration file path: " file; \
	cd backend && DATABASE_URL="postgres://appuser:apppass@localhost:5435/appdb?sslmode=disable" go run cmd/generator/main.go -action migrate -migration $$file

generator-migrate-all: ## Run all migrations
	@echo "Running all migrations..."
	cd backend && DATABASE_URL="postgres://appuser:apppass@localhost:5435/appdb?sslmode=disable" go run cmd/generator/main.go -action migrate-all

generator-all: ## Generate APIs for all existing tables
	@echo "Generating APIs for all tables..."
	cd backend && DATABASE_URL="postgres://appuser:apppass@localhost:5435/appdb?sslmode=disable" go run cmd/generator/main.go -action generate-all

generator-cleanup: ## Cleanup generated files for a model
	@echo "Cleaning up model files..."
	@read -p "Enter model name: " model; \
	cd backend && DATABASE_URL="postgres://appuser:apppass@localhost:5435/appdb?sslmode=disable" go run cmd/generator/main.go -action cleanup -model $$model

generator-template: ## Generate migration template
	@echo "Generating migration template..."
	@read -p "Enter table name: " table; \
	@read -p "Enter operation (create/drop/alter): " op; \
	cd backend && DATABASE_URL="postgres://appuser:apppass@localhost:5435/appdb?sslmode=disable" go run cmd/generator/main.go -action template -table $$table -op $$op

generator-watch: ## Watch for migration files
	@echo "Watching for migration files..."
	cd backend && DATABASE_URL="postgres://appuser:apppass@localhost:5435/appdb?sslmode=disable" go run cmd/generator/main.go -action watch

generator-examples: ## Generate example APIs
	@echo "Generating example APIs..."
	cd backend && DATABASE_URL="postgres://appuser:apppass@localhost:5435/appdb?sslmode=disable" go run cmd/generator/main.go -action generate -schema examples/product.json
	cd backend && DATABASE_URL="postgres://appuser:apppass@localhost:5435/appdb?sslmode=disable" go run cmd/generator/main.go -action generate -schema examples/category.json
	@echo "✅ Example APIs generated!"

# Dokploy Deployment
dokploy-prepare: ## Prepare for Dokploy deployment
	@echo "Preparing for Dokploy deployment..."
	@./scripts/dokploy-deploy.sh

dokploy-env: ## Generate environment variables for Dokploy
	@echo "Generating Dokploy environment variables..."
	@./scripts/dokploy-env-setup.sh

dokploy-deploy: ## Deploy to Dokploy (requires manual setup in Dokploy UI)
	@echo "Deploy to Dokploy:"
	@echo "1. Push code to Git repository"
	@echo "2. Create project in Dokploy dashboard"
	@echo "3. Connect Git repository"
	@echo "4. Set environment variables"
	@echo "5. Use docker-compose.dokploy.yml for deployment"
