# 🚀 Quick Start Guide - Production-Ready Backend Template

## 📋 Prerequisites

### Required Software
- **Go 1.21+** - [Download](https://golang.org/dl/)
- **PostgreSQL 13+** - [Download](https://www.postgresql.org/download/)
- **Redis 6+** - [Download](https://redis.io/download)
- **Docker** (optional) - [Download](https://www.docker.com/get-started)
- **Git** - [Download](https://git-scm.com/downloads)

### Development Tools (Recommended)
- **VS Code** with Go extension
- **Postman** or **Insomnia** for API testing
- **pgAdmin** or **DBeaver** for database management
- **RedisInsight** for Redis management

## ⚡ Quick Setup (5 minutes)

### 1. Clone and Setup
```bash
# Clone the template
git clone <your-template-repo>
cd backend-template

# Install dependencies
go mod download

# Copy environment file
cp env.example .env
```

### 2. Configure Environment
```bash
# Edit .env file with your settings
DATABASE_URL=postgres://username:password@localhost:5432/your_db?sslmode=disable
REDIS_URL=localhost:6379
JWT_SECRET=your_super_secret_jwt_key_change_this_in_production
GIN_MODE=debug
LOG_LEVEL=debug
PORT=8081
```

### 3. Setup Database
```bash
# Start PostgreSQL and Redis
# (Use Docker for quick setup)
docker-compose up -d postgres redis

# Run migrations
go run scripts/migrate/main.go

# Seed initial data (optional)
go run scripts/seed/main.go
```

### 4. Start Development Server
```bash
# Start the server
go run main.go

# Or use Air for hot reloading
air
```

### 5. Test the API
```bash
# Health check
curl http://localhost:8081/health

# API documentation
open http://localhost:8081/swagger/index.html
```

## 🏗️ Project Structure Overview

```
backend/
├── cmd/                    # Application entry points
│   └── server/            # Main server application
├── config/                # Configuration management
│   ├── config.go         # Main config struct
│   └── logging.go        # Logging configuration
├── controllers/           # HTTP request handlers
│   ├── auth.go           # Authentication endpoints
│   ├── user.go           # User management
│   ├── payment.go        # Payment processing
│   └── websocket.go      # Real-time features
├── middleware/            # HTTP middleware
│   ├── auth.go           # Authentication middleware
│   ├── cors.go           # CORS handling
│   ├── logging.go        # Request logging
│   └── rate_limit.go     # Rate limiting
├── models/                # Data models
│   ├── base.go           # Base model with common fields
│   ├── user.go           # User model
│   └── payment.go        # Payment models
├── services/              # Business logic
│   ├── auth.go           # Authentication service
│   ├── cache.go          # Caching service
│   ├── stripe.go         # Stripe integration
│   └── websocket.go      # Real-time service
├── routes/                # Route definitions
│   ├── routes.go         # Main routes
│   ├── auth_routes.go    # Auth routes
│   └── websocket_routes.go # WebSocket routes
├── utils/                 # Utility functions
│   ├── response.go       # HTTP response helpers
│   ├── validation.go     # Input validation
│   └── errors.go         # Error handling
├── tests/                 # Test suites
│   ├── unit/             # Unit tests
│   └── integration/      # Integration tests
├── scripts/               # Build and deployment scripts
│   ├── migrate/          # Database migrations
│   └── seed/             # Database seeding
├── docs/                  # API documentation
└── deployments/           # Docker and K8s configs
    ├── docker/           # Docker configurations
    └── k8s/              # Kubernetes manifests
```

## 🔧 Core Features Available

### ✅ Authentication & Authorization
- JWT-based authentication
- OAuth2 integration (Google, GitHub, Microsoft)
- Role-based access control (RBAC)
- Password reset functionality
- Account verification

### ✅ User Management
- User registration and login
- Profile management
- User CRUD operations
- Session management

### ✅ Payment Processing
- Stripe integration
- PayPal integration
- Polar integration
- Subscription management
- Webhook handling
- Invoice generation

### ✅ Real-Time Features
- WebSocket connections
- Live notifications
- Real-time updates
- Room-based messaging
- Typing indicators
- Presence tracking

### ✅ File Management
- File upload/download
- Image processing
- Multiple storage backends
- File validation
- CDN integration

### ✅ Caching & Performance
- Redis caching
- Query optimization
- Connection pooling
- Rate limiting
- Performance monitoring

### ✅ Monitoring & Logging
- Structured logging
- Health checks
- Metrics collection
- Error tracking
- Request tracing

## 🚀 Common Use Cases

### 1. E-commerce Backend
```go
// Features you get out of the box:
- User authentication
- Product management
- Shopping cart
- Payment processing
- Order management
- Real-time notifications
- Admin dashboard APIs
```

### 2. SaaS Application
```go
// Features you get out of the box:
- Multi-tenant architecture
- Subscription management
- User management
- Payment processing
- Real-time collaboration
- File sharing
- Analytics APIs
```

### 3. Social Media Platform
```go
// Features you get out of the box:
- User profiles
- Content management
- Real-time messaging
- Notifications
- File uploads
- Social features
- Moderation tools
```

### 4. IoT Data Platform
```go
// Features you get out of the box:
- Device authentication
- Data ingestion APIs
- Real-time data streaming
- Data visualization APIs
- Alert management
- Historical data storage
```

## 🛠️ Customization Guide

### 1. Adding New Models
```go
// 1. Create model in models/your_model.go
type YourModel struct {
    BaseModel
    Name        string `json:"name" gorm:"not null"`
    Description string `json:"description"`
    UserID      uint   `json:"user_id"`
    User        User   `json:"user" gorm:"foreignKey:UserID"`
}

// 2. Add to migrations
// 3. Create controller in controllers/your_model.go
// 4. Add routes in routes/your_model_routes.go
// 5. Add service in services/your_model.go
```

### 2. Adding New API Endpoints
```go
// 1. Add handler to controller
func (c *YourController) YourHandler(ctx *gin.Context) {
    // Implementation
}

// 2. Add route
r.GET("/api/v1/your-endpoint", c.YourHandler)

// 3. Add middleware if needed
r.GET("/api/v1/your-endpoint", middleware.AuthMiddleware(), c.YourHandler)
```

### 3. Adding New Services
```go
// 1. Create service interface
type YourService interface {
    DoSomething(ctx context.Context, input Input) (*Output, error)
}

// 2. Implement service
type yourService struct {
    db    *gorm.DB
    cache *CacheService
}

func NewYourService(db *gorm.DB, cache *CacheService) YourService {
    return &yourService{db: db, cache: cache}
}

// 3. Add to dependency injection in main.go
```

### 4. Adding New Middleware
```go
// 1. Create middleware function
func YourMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Implementation
        c.Next()
    }
}

// 2. Apply to routes
r.Use(YourMiddleware())
```

## 🧪 Testing Your Implementation

### 1. Unit Tests
```bash
# Run all unit tests
go test ./tests/unit/...

# Run specific test
go test ./tests/unit/user_test.go

# Run with coverage
go test -cover ./tests/unit/...
```

### 2. Integration Tests
```bash
# Run integration tests
go test ./tests/integration/...

# Run with verbose output
go test -v ./tests/integration/...
```

### 3. API Testing
```bash
# Test health endpoint
curl http://localhost:8081/health

# Test authentication
curl -X POST http://localhost:8081/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Test with authentication
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8081/api/v1/profile
```

### 4. WebSocket Testing
```html
<!-- Use the provided HTML client -->
open websocket_client_example.html
```

## 📊 Monitoring Your Application

### 1. Health Checks
```bash
# Basic health check
curl http://localhost:8081/health

# Detailed health check
curl http://localhost:8081/health/ready

# Liveness check
curl http://localhost:8081/health/live
```

### 2. Metrics
```bash
# View metrics (if enabled)
curl http://localhost:8081/metrics

# View cache statistics
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8081/api/v1/cache/stats
```

### 3. Logs
```bash
# View application logs
tail -f logs/app.log

# View error logs
grep "ERROR" logs/app.log
```

## 🚀 Deployment Options

### 1. Docker Deployment
```bash
# Build Docker image
docker build -t your-app .

# Run container
docker run -p 8081:8081 --env-file .env your-app
```

### 2. Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### 3. Kubernetes Deployment
```bash
# Apply Kubernetes manifests
kubectl apply -f deployments/k8s/

# Check deployment status
kubectl get pods

# View logs
kubectl logs -f deployment/your-app
```

## 🔧 Configuration Options

### Environment Variables
```bash
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/db

# Redis
REDIS_URL=localhost:6379

# JWT
JWT_SECRET=your-secret-key

# Server
PORT=8081
GIN_MODE=debug

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m

# File Upload
MAX_UPLOAD_SIZE=10MB
UPLOAD_PATH=./uploads
```

### Feature Flags
```go
// Enable/disable features via environment variables
ENABLE_WEBSOCKET=true
ENABLE_PAYMENTS=true
ENABLE_OAUTH2=true
ENABLE_CACHING=true
ENABLE_METRICS=true
```

## 🆘 Troubleshooting

### Common Issues

#### 1. Database Connection Failed
```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Check database credentials
psql -h localhost -U username -d database_name
```

#### 2. Redis Connection Failed
```bash
# Check if Redis is running
redis-cli ping

# Check Redis configuration
redis-cli config get "*"
```

#### 3. Port Already in Use
```bash
# Find process using port
lsof -i :8081

# Kill process
kill -9 PID
```

#### 4. JWT Token Issues
```bash
# Check JWT secret is set
echo $JWT_SECRET

# Verify token format
# Use jwt.io to decode and verify
```

### Debug Mode
```bash
# Enable debug logging
export LOG_LEVEL=debug
export GIN_MODE=debug

# Run with debug flags
go run -race main.go
```

## 📚 Next Steps

### 1. Explore the Code
- Read through the `PRODUCTION_READY_TEMPLATE_GUIDE.md`
- Examine the example implementations
- Check out the test files

### 2. Customize for Your Project
- Add your specific models
- Implement your business logic
- Customize the API endpoints
- Add your authentication requirements

### 3. Set Up Production Environment
- Configure production database
- Set up monitoring and alerting
- Implement backup strategies
- Configure security settings

### 4. Deploy and Scale
- Deploy to your preferred platform
- Set up CI/CD pipeline
- Configure load balancing
- Implement auto-scaling

## 🎯 Success Metrics

After implementing this template, you should have:
- ✅ **Working API** with authentication
- ✅ **Database** with migrations
- ✅ **Real-time features** via WebSocket
- ✅ **Payment processing** capabilities
- ✅ **File upload** functionality
- ✅ **Monitoring** and logging
- ✅ **Tests** for reliability
- ✅ **Documentation** for maintenance

## 🆘 Getting Help

- **Documentation**: Check the comprehensive guides
- **Issues**: Look at existing issues or create new ones
- **Community**: Join the discussion forums
- **Support**: Contact the maintainers

---

**Happy Coding! 🚀**

This template gives you a solid foundation to build production-ready applications quickly and efficiently. Start with the basic setup, then customize it to fit your specific needs.
