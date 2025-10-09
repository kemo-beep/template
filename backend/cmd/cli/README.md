# 🚀 Mobile Backend CLI

A comprehensive command-line tool for mobile backend development, testing, and deployment. This CLI provides everything you need to scaffold, test, and manage your mobile backend with an exceptional developer experience.

## ✨ Features

### 🔧 **Code Generation & Scaffolding**
- Generate models, controllers, services, middleware, and tests
- Create complete API modules with one command
- Database migration generation
- Route scaffolding with proper structure

### 🌐 **API Testing & Exploration**
- Interactive web-based API explorer
- Command-line API testing
- Load testing and benchmarking
- Health checks and monitoring

### 🗄️ **Database Management**
- Migration management (up/down)
- Database seeding
- Backup and restore
- Interactive database shell
- Query execution

### 🚀 **Deployment & DevOps**
- Multi-environment deployment
- Blue-green deployments
- Scaling and health monitoring
- Configuration management
- Log streaming

### 🧪 **Testing & Quality Assurance**
- Unit, integration, and E2E testing
- Coverage analysis
- Performance benchmarking
- Security scanning
- Code linting

## 📦 Installation

### Prerequisites
- Go 1.21 or later
- Git

### Install from Source
```bash
# Clone the repository
git clone <repository-url>
cd backend/cmd/cli

# Install dependencies
go mod tidy

# Build and install
go build -o mobile-backend-cli .
sudo mv mobile-backend-cli /usr/local/bin/

# Or install directly
go install .
```

### Verify Installation
```bash
mobile-backend-cli --version
mobile-backend-cli --help
```

## 🚀 Quick Start

### 1. Initialize a New Project
```bash
# Navigate to your project directory
cd /path/to/your/project

# Initialize CLI configuration
mobile-backend-cli init

# This creates a .mobile-backend-cli.yaml config file
```

### 2. Generate Your First API
```bash
# Generate a complete User API module
mobile-backend-cli generate api User

# This creates:
# - models/user.go
# - controllers/user.go
# - services/user.go
# - routes/user_routes.go
# - tests/unit/user_test.go
# - migrations/001_create_users_table.sql
```

### 3. Set Up Database
```bash
# Run migrations
mobile-backend-cli db migrate

# Seed with sample data
mobile-backend-cli db seed

# Check status
mobile-backend-cli db status
```

### 4. Test Your API
```bash
# Start the API explorer
mobile-backend-cli explorer --open

# Or test via command line
mobile-backend-cli api test GET /users
mobile-backend-cli api test POST /users -d '{"name":"John","email":"john@example.com"}'
```

### 5. Deploy Your Application
```bash
# Deploy to staging
mobile-backend-cli deploy --env staging --build

# Deploy to production
mobile-backend-cli deploy --env production --build
```

## 📚 Command Reference

### Code Generation Commands

#### Generate Models
```bash
mobile-backend-cli generate model User
mobile-backend-cli generate model Product
mobile-backend-cli generate model Order
```

#### Generate Controllers
```bash
mobile-backend-cli generate controller UserController
mobile-backend-cli generate controller ProductController
```

#### Generate Services
```bash
mobile-backend-cli generate service UserService
mobile-backend-cli generate service PaymentService
```

#### Generate Complete API
```bash
mobile-backend-cli generate api User
mobile-backend-cli generate api Product
mobile-backend-cli generate api Order
```

### API Testing Commands

#### Test API Endpoints
```bash
# Basic GET request
mobile-backend-cli api test GET /users

# POST with data
mobile-backend-cli api test POST /users -d '{"name":"John","email":"john@example.com"}'

# With custom headers
mobile-backend-cli api test GET /users -H '{"Authorization":"Bearer token"}'

# With query parameters
mobile-backend-cli api test GET /users -q "page=1&limit=10"
```

#### API Health & Status
```bash
# Check API health
mobile-backend-cli api health

# List all endpoints
mobile-backend-cli api list

# Generate documentation
mobile-backend-cli api docs --open
```

#### Interactive API Explorer
```bash
# Start web-based explorer
mobile-backend-cli explorer

# With custom port
mobile-backend-cli explorer --port 3000

# Open in browser automatically
mobile-backend-cli explorer --open
```

### Database Commands

#### Migration Management
```bash
# Run all pending migrations
mobile-backend-cli db migrate

# Rollback last migration
mobile-backend-cli db rollback

# Rollback multiple migrations
mobile-backend-cli db rollback 3

# Check migration status
mobile-backend-cli db status
```

#### Database Operations
```bash
# Seed database
mobile-backend-cli db seed

# Backup database
mobile-backend-cli db backup

# Restore from backup
mobile-backend-cli db restore backup.sql

# Open database shell
mobile-backend-cli db shell

# Execute SQL query
mobile-backend-cli db query "SELECT * FROM users"
```

### Testing Commands

#### Run Tests
```bash
# Run unit tests
mobile-backend-cli test unit

# Run integration tests
mobile-backend-cli test integration

# Run E2E tests
mobile-backend-cli test e2e

# Run all tests
mobile-backend-cli test unit integration e2e
```

#### Coverage Analysis
```bash
# Generate coverage report
mobile-backend-cli test coverage

# HTML coverage report
mobile-backend-cli test coverage --format html --open

# JSON coverage report
mobile-backend-cli test coverage --format json
```

#### Performance Testing
```bash
# Run benchmarks
mobile-backend-cli test benchmark

# Load testing
mobile-backend-cli test load --users 100 --duration 60s

# Security scanning
mobile-backend-cli test security
```

### Deployment Commands

#### Deploy Application
```bash
# Deploy to development
mobile-backend-cli deploy --env development

# Deploy to staging with build
mobile-backend-cli deploy --env staging --build

# Deploy to production
mobile-backend-cli deploy --env production --build --force
```

#### Deployment Management
```bash
# Check deployment status
mobile-backend-cli deploy status

# View deployment logs
mobile-backend-cli deploy logs --follow

# Rollback deployment
mobile-backend-cli deploy rollback

# Scale deployment
mobile-backend-cli deploy scale --replicas 3
```

## ⚙️ Configuration

### Configuration File
The CLI uses a YAML configuration file (`.mobile-backend-cli.yaml`) in your project root:

```yaml
# API Configuration
base_url: "http://localhost:8080"
api_key: "your-api-key-here"

# Database Configuration
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "password"
  name: "mobile_backend"
  driver: "postgres"

# Deployment Configuration
deployment:
  environments:
    development:
      region: "us-east-1"
      replicas: 1
    staging:
      region: "us-east-1"
      replicas: 2
    production:
      region: "us-west-2"
      replicas: 3

# Testing Configuration
testing:
  timeout: "30s"
  coverage_threshold: 80
  load_test_users: 100
```

### Environment Variables
You can also use environment variables:

```bash
export MOBILE_BACKEND_BASE_URL="https://api.example.com"
export MOBILE_BACKEND_API_KEY="your-api-key"
export MOBILE_BACKEND_DB_HOST="localhost"
export MOBILE_BACKEND_DB_PORT="5432"
```

## 🎨 API Explorer

The interactive API explorer provides a web-based interface for testing your APIs:

### Features
- **Visual Interface**: Clean, modern web interface
- **Endpoint Discovery**: Automatically loads available endpoints
- **Request Builder**: Easy-to-use form for building requests
- **Response Viewer**: Pretty-printed JSON responses
- **Header Management**: Custom headers and authentication
- **Theme Support**: Light and dark themes
- **Real-time Testing**: Live API testing without external tools

### Usage
```bash
# Start the explorer
mobile-backend-cli explorer

# With custom configuration
mobile-backend-cli explorer --port 3000 --theme dark --open
```

### Screenshots
The explorer provides:
- 📋 **Endpoint List**: Browse all available API endpoints
- 🚀 **Request Panel**: Build and send API requests
- 📥 **Response Panel**: View formatted responses
- ⚙️ **Configuration**: Manage API settings

## 🔧 Advanced Usage

### Custom Code Generation
You can customize the generated code by modifying the templates in the CLI source code.

### Custom Test Scenarios
Create custom load test scenarios by modifying the k6 scripts.

### Custom Deployment Scripts
Add custom deployment logic for your specific infrastructure.

### Plugin System
Extend the CLI with custom commands and functionality.

## 🐛 Troubleshooting

### Common Issues

#### CLI Not Found
```bash
# Make sure the CLI is in your PATH
which mobile-backend-cli

# If not found, add to PATH
export PATH=$PATH:/usr/local/bin
```

#### Database Connection Issues
```bash
# Check database configuration
mobile-backend-cli db status

# Test connection
mobile-backend-cli db query "SELECT 1"
```

#### API Connection Issues
```bash
# Check API health
mobile-backend-cli api health

# Test with verbose output
mobile-backend-cli api test GET /health -v
```

### Debug Mode
Enable debug mode for detailed logging:

```bash
mobile-backend-cli --verbose <command>
```

## 🤝 Contributing

We welcome contributions! Please see our contributing guidelines for details.

### Development Setup
```bash
# Fork and clone the repository
git clone <your-fork-url>
cd backend/cmd/cli

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build
go build -o mobile-backend-cli .
```

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI framework
- Uses [Viper](https://github.com/spf13/viper) for configuration management
- Inspired by modern CLI tools like `kubectl`, `docker`, and `git`

---

**Happy coding! 🚀**

For more information, visit our [documentation](https://docs.example.com) or join our [community](https://discord.gg/example).
