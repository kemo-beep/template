# 📚 Mobile Backend CLI Usage Guide

This comprehensive guide covers all aspects of using the Mobile Backend CLI tool for development, testing, and deployment.

## 🚀 Getting Started

### Installation
```bash
# Install from source
cd backend/cmd/cli
make install

# Or use the install script
./install.sh

# Verify installation
mobile-backend-cli --version
```

### Configuration
```bash
# Initialize configuration
mobile-backend-cli init

# Edit configuration
nano .mobile-backend-cli.yaml
```

## 🔧 Code Generation

### Generate Complete API Module
```bash
# Generate a complete User API with all components
mobile-backend-cli generate api User

# This creates:
# - models/user.go
# - controllers/user.go  
# - services/user.go
# - routes/user_routes.go
# - tests/unit/user_test.go
# - migrations/001_create_users_table.sql
```

### Generate Individual Components
```bash
# Generate model only
mobile-backend-cli generate model Product

# Generate controller only
mobile-backend-cli generate controller ProductController

# Generate service only
mobile-backend-cli generate service ProductService

# Generate middleware only
mobile-backend-cli generate middleware AuthMiddleware

# Generate test only
mobile-backend-cli generate test ProductControllerTest

# Generate migration only
mobile-backend-cli generate migration AddProductTable

# Generate routes only
mobile-backend-cli generate route ProductRoutes
```

### Generated Code Structure

#### Model Example
```go
// models/user.go
type User struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
    
    // Add your fields here
    Name  string `json:"name" gorm:"not null"`
    Email string `json:"email" gorm:"uniqueIndex"`
}
```

#### Controller Example
```go
// controllers/user.go
func (c *UserController) CreateUser(ctx *gin.Context) {
    var user models.User
    if err := ctx.ShouldBindJSON(&user); err != nil {
        utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err)
        return
    }
    
    createdUser, err := c.userService.CreateUser(&user)
    if err != nil {
        utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create user", err)
        return
    }
    
    utils.SuccessResponse(ctx, http.StatusCreated, "User created successfully", createdUser)
}
```

## 🌐 API Testing & Exploration

### Interactive Web Explorer
```bash
# Start the web-based API explorer
mobile-backend-cli explorer

# With custom port and auto-open
mobile-backend-cli explorer --port 3000 --open

# With dark theme
mobile-backend-cli explorer --theme dark
```

The API Explorer provides:
- **Visual Interface**: Clean, modern web interface
- **Endpoint Discovery**: Automatically loads available endpoints
- **Request Builder**: Easy-to-use form for building requests
- **Response Viewer**: Pretty-printed JSON responses
- **Header Management**: Custom headers and authentication
- **Real-time Testing**: Live API testing without external tools

### Command Line API Testing
```bash
# Test GET endpoint
mobile-backend-cli api test GET /users

# Test POST with data
mobile-backend-cli api test POST /users -d '{"name":"John","email":"john@example.com"}'

# Test with custom headers
mobile-backend-cli api test GET /users -H '{"Authorization":"Bearer token"}'

# Test with query parameters
mobile-backend-cli api test GET /users -q "page=1&limit=10"

# Test with verbose output
mobile-backend-cli api test GET /users -v

# Test with custom timeout
mobile-backend-cli api test GET /users -t 60
```

### API Health & Documentation
```bash
# Check API health
mobile-backend-cli api health

# List all available endpoints
mobile-backend-cli api list

# Generate and view API documentation
mobile-backend-cli api docs --open

# Generate documentation in different formats
mobile-backend-cli api docs --format json
mobile-backend-cli api docs --format yaml
```

### Load Testing & Benchmarking
```bash
# Run load tests
mobile-backend-cli api load-test /users --users 100 --duration 60s

# Run benchmarks
mobile-backend-cli api benchmark /users --requests 1000 --concurrency 10
```

## 🗄️ Database Management

### Migration Management
```bash
# Run all pending migrations
mobile-backend-cli db migrate

# Check migration status
mobile-backend-cli db status

# Rollback last migration
mobile-backend-cli db rollback

# Rollback multiple migrations
mobile-backend-cli db rollback 3

# Reset database (drop all tables and recreate)
mobile-backend-cli db reset
```

### Database Operations
```bash
# Seed database with sample data
mobile-backend-cli db seed

# Create database backup
mobile-backend-cli db backup

# Create backup with custom filename
mobile-backend-cli db backup my-backup.sql

# Restore from backup
mobile-backend-cli db restore my-backup.sql

# Open interactive database shell
mobile-backend-cli db shell

# Execute SQL query
mobile-backend-cli db query "SELECT * FROM users"
```

### Database Configuration
```bash
# Use custom database settings
mobile-backend-cli db migrate --host localhost --port 5432 --user postgres --database mydb

# Use different database driver
mobile-backend-cli db migrate --driver mysql
```

## 🧪 Testing & Quality Assurance

### Run Tests
```bash
# Run unit tests
mobile-backend-cli test unit

# Run integration tests
mobile-backend-cli test integration

# Run end-to-end tests
mobile-backend-cli test e2e

# Run all tests
mobile-backend-cli test unit integration e2e
```

### Coverage Analysis
```bash
# Generate coverage report
mobile-backend-cli test coverage

# HTML coverage report with auto-open
mobile-backend-cli test coverage --format html --open

# JSON coverage report
mobile-backend-cli test coverage --format json

# Set coverage threshold
mobile-backend-cli test coverage --threshold 90
```

### Performance Testing
```bash
# Run benchmark tests
mobile-backend-cli test benchmark

# Run load tests
mobile-backend-cli test load --users 100 --duration 60s

# Run load tests with custom scenario
mobile-backend-cli test load --scenario api-stress --users 500 --duration 300s
```

### Code Quality
```bash
# Run code linting
mobile-backend-cli test lint

# Run security scan
mobile-backend-cli test security

# Run all quality checks
mobile-backend-cli test lint security
```

### Debugging
```bash
# Start debug session
mobile-backend-cli test debug

# Debug with custom port
mobile-backend-cli test debug --port 2345

# Debug in headless mode
mobile-backend-cli test debug --headless
```

## 🚀 Deployment & DevOps

### Deploy Application
```bash
# Deploy to development
mobile-backend-cli deploy --env development

# Deploy to staging with build
mobile-backend-cli deploy --env staging --build

# Deploy to production
mobile-backend-cli deploy --env production --build --force

# Deploy with specific version
mobile-backend-cli deploy --env production --version v1.2.3
```

### Deployment Management
```bash
# Check deployment status
mobile-backend-cli deploy status

# View deployment logs
mobile-backend-cli deploy logs

# Follow logs in real-time
mobile-backend-cli deploy logs --follow

# Show last 100 lines
mobile-backend-cli deploy logs --lines 100

# Show logs since specific time
mobile-backend-cli deploy logs --since 1h
```

### Scaling & Health
```bash
# Scale deployment
mobile-backend-cli deploy scale --replicas 5

# Scale with rolling strategy
mobile-backend-cli deploy scale --replicas 3 --strategy rolling

# Check deployment health
mobile-backend-cli deploy health

# Rollback deployment
mobile-backend-cli deploy rollback
```

### Configuration Management
```bash
# View current configuration
mobile-backend-cli deploy config

# Set environment variable
mobile-backend-cli deploy config --set LOG_LEVEL=debug

# Unset environment variable
mobile-backend-cli deploy config --unset DEBUG_MODE

# Load configuration from file
mobile-backend-cli deploy config --file production.yaml
```

## ⚙️ Configuration

### Configuration File
The CLI uses a YAML configuration file (`.mobile-backend-cli.yaml`):

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
```bash
export MOBILE_BACKEND_BASE_URL="https://api.example.com"
export MOBILE_BACKEND_API_KEY="your-api-key"
export MOBILE_BACKEND_DB_HOST="localhost"
export MOBILE_BACKEND_DB_PORT="5432"
```

### Command Line Flags
```bash
# Global flags
mobile-backend-cli --config config.yaml --verbose --base-url https://api.example.com --api-key token

# Per-command flags
mobile-backend-cli api test GET /users -v -t 30
mobile-backend-cli db migrate --host localhost --port 5432
mobile-backend-cli deploy --env production --build --force
```

## 🎨 API Explorer Features

### Visual Interface
The API Explorer provides a modern, responsive web interface with:
- **Dark/Light Themes**: Switch between themes
- **Responsive Design**: Works on desktop and mobile
- **Real-time Updates**: Live API testing
- **Syntax Highlighting**: Pretty-printed JSON

### Endpoint Management
- **Auto-discovery**: Automatically loads available endpoints
- **Categorization**: Groups endpoints by functionality
- **Search**: Find endpoints quickly
- **Favorites**: Mark frequently used endpoints

### Request Building
- **Method Selection**: GET, POST, PUT, DELETE, PATCH
- **URL Builder**: Easy endpoint selection
- **Header Management**: Custom headers and authentication
- **Body Editor**: JSON request body with validation
- **Query Parameters**: Easy parameter management

### Response Analysis
- **Status Codes**: Color-coded response status
- **Response Time**: Performance metrics
- **Headers**: Complete response headers
- **Body Formatting**: Pretty-printed JSON responses
- **Error Handling**: Clear error messages

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
# Check if CLI is in PATH
which mobile-backend-cli

# Add to PATH if needed
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

#### Permission Issues
```bash
# Fix permissions
sudo chmod +x /usr/local/bin/mobile-backend-cli

# Or reinstall
sudo make uninstall && sudo make install
```

### Debug Mode
Enable debug mode for detailed logging:

```bash
mobile-backend-cli --verbose <command>
```

### Log Files
Check log files for detailed error information:

```bash
# CLI logs
tail -f ~/.mobile-backend-cli/cli.log

# Application logs
mobile-backend-cli deploy logs --follow
```

## 📚 Examples

### Complete Development Workflow
```bash
# 1. Initialize project
mobile-backend-cli init

# 2. Generate API
mobile-backend-cli generate api User

# 3. Set up database
mobile-backend-cli db migrate
mobile-backend-cli db seed

# 4. Test API
mobile-backend-cli api explorer

# 5. Run tests
mobile-backend-cli test unit
mobile-backend-cli test coverage

# 6. Deploy
mobile-backend-cli deploy --env staging --build
```

### API Testing Workflow
```bash
# 1. Check API health
mobile-backend-cli api health

# 2. List available endpoints
mobile-backend-cli api list

# 3. Test specific endpoints
mobile-backend-cli api test GET /users
mobile-backend-cli api test POST /users -d '{"name":"John","email":"john@example.com"}'

# 4. Run load tests
mobile-backend-cli api load-test /users --users 100 --duration 60s

# 5. Generate documentation
mobile-backend-cli api docs --open
```

### Database Management Workflow
```bash
# 1. Check migration status
mobile-backend-cli db status

# 2. Run migrations
mobile-backend-cli db migrate

# 3. Seed database
mobile-backend-cli db seed

# 4. Create backup
mobile-backend-cli db backup

# 5. Test queries
mobile-backend-cli db query "SELECT COUNT(*) FROM users"
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
make test

# Build
make build

# Install locally
make install
```

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**Happy coding! 🚀**

For more information, visit our [documentation](https://docs.example.com) or join our [community](https://discord.gg/example).
