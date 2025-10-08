# Mobile Backend Template

A comprehensive Go (Gin) + PostgreSQL Docker-based mobile backend template with a focus on excellent developer experience (DX). This template includes authentication, database management, migrations, and streamlined local development setup.

## 🚀 Features

### Core Features
- **Backend Framework**: Go with Gin web framework
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT-based authentication system
- **Development**: Live reload with Air
- **Containerization**: Docker Compose for easy setup
- **Database Migrations**: GORM AutoMigrate support
- **API Testing**: Ready for Postman/Insomnia integration

### Production Features
- **Monitoring & Observability**: Structured logging, metrics, tracing
- **Security**: Rate limiting, CORS, input validation, OAuth2
- **Performance**: Redis caching, connection pooling, compression
- **Testing**: Unit, integration, and load testing
- **CI/CD**: Automated testing, security scanning, deployment
- **API Documentation**: Swagger/OpenAPI with code generation
- **Mobile-Specific**: Push notifications, file upload, real-time features

## 🛠 Tech Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Backend** | Go + Gin | High-performance web framework |
| **ORM** | GORM | Database abstraction layer |
| **Database** | PostgreSQL | Robust relational database |
| **Cache** | Redis | Session storage and caching |
| **Auth** | JWT + OAuth2 | Stateless authentication |
| **Monitoring** | Prometheus + Grafana | Metrics and monitoring |
| **Logging** | Zap | Structured logging |
| **Tracing** | Jaeger | Distributed tracing |
| **Testing** | Testify + Testcontainers | Unit and integration testing |
| **CI/CD** | GitHub Actions | Automated testing and deployment |
| **Documentation** | Swagger/OpenAPI | API documentation |
| **Live Reload** | Air | Development hot reloading |
| **Containerization** | Docker + K8s | Environment management |
| **API Testing** | Postman/Insomnia | API development and testing |

## 📁 Project Structure

```
mobile-backend/
├── backend/
│   ├── main.go                 # Application entry point
│   ├── go.mod                  # Go module dependencies
│   ├── go.sum                  # Dependency checksums
│   ├── .air.toml              # Air configuration
│   ├── Dockerfile             # Backend container config
│   ├── Dockerfile.prod        # Production Dockerfile
│   ├── config/
│   │   ├── config.go          # Environment configuration
│   │   └── logging.go         # Logging configuration
│   ├── controllers/
│   │   ├── user.go            # User API handlers
│   │   ├── auth.go            # Authentication handlers
│   │   └── health.go          # Health check handlers
│   ├── models/
│   │   ├── user.go            # User database model
│   │   ├── session.go         # Session model
│   │   └── base.go            # Base model with common fields
│   ├── routes/
│   │   ├── routes.go          # API route definitions
│   │   └── middleware.go      # Route middleware
│   ├── middleware/
│   │   ├── auth.go            # Authentication middleware
│   │   ├── cors.go            # CORS middleware
│   │   ├── rate_limit.go      # Rate limiting middleware
│   │   ├── logging.go         # Request logging middleware
│   │   └── security.go        # Security headers middleware
│   ├── services/
│   │   ├── auth.go            # Authentication service
│   │   ├── cache.go           # Cache service
│   │   ├── email.go           # Email service
│   │   └── notification.go    # Push notification service
│   ├── utils/
│   │   ├── jwt.go             # JWT utilities
│   │   ├── validator.go       # Input validation
│   │   ├── crypto.go          # Cryptographic utilities
│   │   └── response.go        # API response helpers
│   ├── tests/
│   │   ├── integration/       # Integration tests
│   │   ├── unit/              # Unit tests
│   │   └── fixtures/          # Test data fixtures
│   ├── migrations/
│   │   └── *.sql             # Database migration files
│   ├── docs/
│   │   └── swagger.yaml       # API documentation
│   └── scripts/
│       ├── seed.go            # Database seeding
│       └── migrate.go         # Migration runner
├── infrastructure/
│   ├── kubernetes/            # K8s manifests
│   │   ├── namespace.yaml
│   │   ├── configmap.yaml
│   │   ├── secret.yaml
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── terraform/             # Infrastructure as Code
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── helm/                  # Helm charts
│       └── mobile-backend/
├── monitoring/
│   ├── prometheus/            # Prometheus config
│   ├── grafana/               # Grafana dashboards
│   └── jaeger/                # Jaeger tracing config
├── scripts/
│   ├── dev.sh                 # Development setup
│   ├── test.sh                # Test runner
│   ├── deploy.sh              # Deployment script
│   └── backup.sh              # Database backup
├── .github/
│   └── workflows/             # CI/CD pipelines
│       ├── ci.yml
│       ├── security.yml
│       └── deploy.yml
├── docker-compose.yml         # Development environment
├── docker-compose.prod.yml    # Production environment
├── .env                       # Environment variables
├── .env.example              # Environment template
├── .gitignore
├── .golangci.yml             # Linting configuration
├── Makefile                  # Build automation
└── README.md                 # Project documentation
```

## 🐳 Docker Compose Setup

### Development Environment

```yaml
version: "3.8"

services:
  db:
    image: postgres:15
    container_name: mobile_backend_db
    environment:
      POSTGRES_USER: appuser
      POSTGRES_PASSWORD: apppass
      POSTGRES_DB: appdb
    ports:
      - "5432:5432"
    volumes:
      - db_data:/var/lib/postgresql/data
      - ./database/init:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U appuser -d appdb"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: mobile_backend_redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  backend:
    build: 
      context: ./backend
      dockerfile: Dockerfile
    container_name: mobile_backend_api
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://appuser:apppass@db:5432/appdb?sslmode=disable
      REDIS_URL: redis://redis:6379
      JWT_SECRET: your_super_secret_jwt_key_change_this_in_production
      GIN_MODE: debug
      LOG_LEVEL: debug
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./backend:/app
      - /app/vendor
    command: air
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    container_name: mobile_backend_prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=200h'
      - '--web.enable-lifecycle'

  grafana:
    image: grafana/grafana:latest
    container_name: mobile_backend_grafana
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./monitoring/grafana/datasources:/etc/grafana/provisioning/datasources

  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: mobile_backend_jaeger
    ports:
      - "16686:16686"
      - "14268:14268"
    environment:
      COLLECTOR_OTLP_ENABLED: true

volumes:
  db_data:
    driver: local
  redis_data:
    driver: local
  prometheus_data:
    driver: local
  grafana_data:
    driver: local
```

### Production Environment

```yaml
version: "3.8"

services:
  db:
    image: postgres:15
    container_name: mobile_backend_db_prod
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - db_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${DB_USER} -d ${DB_NAME}"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: mobile_backend_redis_prod
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped

  backend:
    build: 
      context: ./backend
      dockerfile: Dockerfile.prod
    container_name: mobile_backend_api_prod
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: ${DATABASE_URL}
      REDIS_URL: ${REDIS_URL}
      JWT_SECRET: ${JWT_SECRET}
      GIN_MODE: release
      LOG_LEVEL: info
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'

volumes:
  db_data:
    driver: local
  redis_data:
    driver: local
```

## ⚡ Live Reload Development

### Install Air

```bash
go install github.com/cosmtrek/air@latest
```

### Air Configuration (`.air.toml`)

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

## 🗄 Database Configuration

### GORM + PostgreSQL Setup

```go
// backend/config/config.go
package config

import (
    "fmt"
    "log"
    "os"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() error {
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        return fmt.Errorf("DATABASE_URL environment variable is required")
    }

    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }

    // Configure connection pool
    sqlDB, err := DB.DB()
    if err != nil {
        return fmt.Errorf("failed to get underlying sql.DB: %w", err)
    }

    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)

    log.Println("Database connected successfully")
    return nil
}

func GetDB() *gorm.DB {
    return DB
}
```

## 👤 User Model & Authentication

### User Model

```go
// backend/models/user.go
package models

import (
    "time"
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
    Email     string         `json:"email" gorm:"uniqueIndex;not null"`
    Password  string         `json:"-" gorm:"not null"`
    Name      string         `json:"name"`
    IsActive  bool           `json:"is_active" gorm:"default:true"`
}

// HashPassword hashes the user's password
func (u *User) HashPassword(password string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedPassword)
    return nil
}

// CheckPassword verifies the user's password
func (u *User) CheckPassword(password string) error {
    return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}
```

### JWT Utilities

```go
// backend/utils/jwt.go
package utils

import (
    "errors"
    "os"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID uint   `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func GenerateToken(userID uint, email string) (string, error) {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        return "", errors.New("JWT_SECRET environment variable is required")
    }

    claims := Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString string) (*Claims, error) {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        return nil, errors.New("JWT_SECRET environment variable is required")
    }

    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}
```

## 🛡 Authentication Middleware

```go
// backend/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"

    "mobile-backend/utils"
    "github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Bearer token required"})
            c.Abort()
            return
        }

        claims, err := utils.ValidateToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        c.Set("user_id", claims.UserID)
        c.Set("user_email", claims.Email)
        c.Next()
    }
}
```

## 🛣 API Routes

```go
// backend/routes/routes.go
package routes

import (
    "mobile-backend/controllers"
    "mobile-backend/middleware"
    "github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // API v1 routes
    api := r.Group("/api/v1")
    {
        // Public routes
        auth := api.Group("/auth")
        {
            auth.POST("/register", controllers.RegisterUser)
            auth.POST("/login", controllers.LoginUser)
        }

        // Protected routes
        protected := api.Group("/")
        protected.Use(middleware.AuthMiddleware())
        {
            protected.GET("/profile", controllers.GetProfile)
            protected.PUT("/profile", controllers.UpdateProfile)
            protected.DELETE("/profile", controllers.DeleteProfile)
        }
    }
}
```

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.19+ (for local development)
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd mobile-backend
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start the services**
   ```bash
docker-compose up -d
   ```

4. **Verify the setup**
   ```bash
   # Check if services are running
   docker-compose ps
   
   # Test the API
   curl http://localhost:8080/health
   ```

### Development

1. **Start live reload development**
   ```bash
   cd backend
   air
   ```

2. **Access the services**
   - **Backend API**: http://localhost:8081
   - **Database**: localhost:5434
   - **API Documentation**: http://localhost:8081/api/v1/docs (if Swagger is configured)

## 📝 Environment Variables

Create a `.env` file in the root directory:

```env
# Database
DATABASE_URL=postgres://appuser:apppass@db:5432/appdb?sslmode=disable

# JWT
JWT_SECRET=your_super_secret_jwt_key_change_this_in_production

# Server
GIN_MODE=debug
PORT=8080

# CORS (optional)
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

## 🧪 API Testing

### Example API Calls

**Register a new user:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword",
    "name": "John Doe"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

**Get profile (authenticated):**
```bash
curl -X GET http://localhost:8080/api/v1/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 🔧 Development Tools

### Database Management

```bash
# Connect to PostgreSQL
docker-compose exec db psql -U appuser -d appdb

# Run migrations (if using GORM AutoMigrate)
docker-compose exec backend go run main.go migrate

# View logs
docker-compose logs -f backend
docker-compose logs -f db
```

### Useful Commands

```bash
# Rebuild and restart services
docker-compose up -d --build

# Stop all services
docker-compose down

# Stop and remove volumes (⚠️ This will delete all data)
docker-compose down -v

# View service logs
docker-compose logs -f [service_name]
```

## 🚀 Production Deployment

### Environment Setup

1. **Update environment variables for production**
2. **Use a proper JWT secret**
3. **Configure CORS for your frontend domain**
4. **Set up SSL/TLS certificates**
5. **Configure database backups**

### Docker Production Build

```dockerfile
# backend/Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/.env .

EXPOSE 8080
CMD ["./main"]
```

## 🐛 Troubleshooting

### Common Issues

1. **Database connection failed**
   - Check if PostgreSQL container is running: `docker-compose ps`
   - Verify DATABASE_URL in .env file
   - Check database logs: `docker-compose logs db`

2. **JWT token issues**
   - Ensure JWT_SECRET is set in environment
   - Check token expiration time
   - Verify Authorization header format

3. **Live reload not working**
   - Check if Air is installed: `air -v`
   - Verify .air.toml configuration
   - Check file permissions

4. **Port conflicts**
   - Change ports in docker-compose.yml
   - Check if ports are already in use: `lsof -i :8080`

## 📚 Additional Resources

- [Gin Web Framework Documentation](https://gin-gonic.com/docs/)
- [GORM Documentation](https://gorm.io/docs/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [JWT.io](https://jwt.io/) - JWT token debugging

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📊 Monitoring & Observability

### Structured Logging

```go
// backend/config/logging.go
package config

import (
    "os"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func SetupLogger() *zap.Logger {
    config := zap.NewProductionConfig()
    
    if os.Getenv("GIN_MODE") == "debug" {
        config = zap.NewDevelopmentConfig()
    }
    
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    config.EncoderConfig.LevelKey = "level"
    config.EncoderConfig.MessageKey = "message"
    config.EncoderConfig.CallerKey = "caller"
    
    logger, _ := config.Build()
    return logger
}
```

### Prometheus Metrics

```go
// backend/middleware/metrics.go
package middleware

import (
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status_code"},
    )
    
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "Duration of HTTP requests in seconds",
        },
        []string{"method", "endpoint"},
    )
)

func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        status := strconv.Itoa(c.Writer.Status())
        
        httpRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            status,
        ).Inc()
        
        httpRequestDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
        ).Observe(duration)
    }
}
```

### Health Checks

```go
// backend/controllers/health.go
package controllers

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type HealthController struct {
    db *gorm.DB
}

func NewHealthController(db *gorm.DB) *HealthController {
    return &HealthController{db: db}
}

func (h *HealthController) HealthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
        "version": "1.0.0",
    })
}

func (h *HealthController) ReadinessCheck(c *gin.Context) {
    sqlDB, err := h.db.DB()
    if err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "unhealthy",
            "error": "database connection failed",
        })
        return
    }
    
    if err := sqlDB.Ping(); err != nil {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "status": "unhealthy",
            "error": "database ping failed",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "status": "ready",
        "timestamp": time.Now().Unix(),
    })
}

func (h *HealthController) LivenessCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status": "alive",
        "timestamp": time.Now().Unix(),
    })
}
```

## 🔒 Security Features

### Rate Limiting

```go
// backend/middleware/rate_limit.go
package middleware

import (
    "fmt"
    "net/http"
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
)

type RateLimiter struct {
    redis *redis.Client
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
    return &RateLimiter{redis: redis}
}

func (rl *RateLimiter) RateLimit(limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := fmt.Sprintf("rate_limit:%s", c.ClientIP())
        
        current, err := rl.redis.Incr(c.Request.Context(), key).Result()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter error"})
            c.Abort()
            return
        }
        
        if current == 1 {
            rl.redis.Expire(c.Request.Context(), key, window)
        }
        
        if current > int64(limit) {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": window.Seconds(),
            })
            c.Abort()
            return
        }
        
        c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
        c.Header("X-RateLimit-Remaining", strconv.Itoa(limit-int(current)))
        c.Next()
    }
}
```

### CORS Configuration

```go
// backend/middleware/cors.go
package middleware

import (
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    "os"
    "strings"
)

func CORS() gin.HandlerFunc {
    config := cors.DefaultConfig()
    
    allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
    if allowedOrigins != "" {
        config.AllowOrigins = strings.Split(allowedOrigins, ",")
    } else {
        config.AllowAllOrigins = true
    }
    
    config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
    config.AllowHeaders = []string{
        "Origin", "Content-Length", "Content-Type", "Authorization",
        "X-Requested-With", "X-API-Key", "X-Client-Version",
    }
    config.ExposeHeaders = []string{"Content-Length", "X-Total-Count"}
    config.AllowCredentials = true
    config.MaxAge = 12 * time.Hour
    
    return cors.New(config)
}
```

### Input Validation

```go
// backend/utils/validator.go
package utils

import (
    "github.com/go-playground/validator/v10"
    "reflect"
    "strings"
)

var Validator *validator.Validate

func InitValidator() {
    Validator = validator.New()
    
    // Register custom tag name function
    Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
        if name == "-" {
            return ""
        }
        return name
    })
}

func ValidateStruct(s interface{}) error {
    return Validator.Struct(s)
}

// Custom validation tags
func RegisterCustomValidators() {
    Validator.RegisterValidation("password", validatePassword)
    Validator.RegisterValidation("email", validateEmail)
}

func validatePassword(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    if len(password) < 8 {
        return false
    }
    // Add more password complexity rules
    return true
}

func validateEmail(fl validator.FieldLevel) bool {
    email := fl.Field().String()
    return strings.Contains(email, "@") && strings.Contains(email, ".")
}
```

## 🧪 Testing Framework

### Unit Tests

```go
// backend/tests/unit/user_test.go
package unit

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "mobile-backend/models"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) Create(user *models.User) error {
    args := m.Called(user)
    return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(email string) (*models.User, error) {
    args := m.Called(email)
    return args.Get(0).(*models.User), args.Error(1)
}

func TestUserService_CreateUser(t *testing.T) {
    mockRepo := new(MockUserRepository)
    userService := NewUserService(mockRepo)
    
    user := &models.User{
        Email: "test@example.com",
        Name:  "Test User",
    }
    
    mockRepo.On("Create", user).Return(nil)
    
    err := userService.CreateUser(user)
    
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### Integration Tests

```go
// backend/tests/integration/api_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "mobile-backend/config"
    "mobile-backend/routes"
)

func TestUserRegistration(t *testing.T) {
    // Setup test database
    ctx := context.Background()
    
    postgresContainer, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("testuser"),
        postgres.WithPassword("testpass"),
    )
    assert.NoError(t, err)
    defer postgresContainer.Terminate(ctx)
    
    // Get connection string
    connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
    assert.NoError(t, err)
    
    // Setup app
    app := setupTestApp(connStr)
    
    // Test registration
    userData := map[string]string{
        "email":    "test@example.com",
        "password": "password123",
        "name":     "Test User",
    }
    
    jsonData, _ := json.Marshal(userData)
    req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    app.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### Load Testing

```javascript
// k6-load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },
    { duration: '5m', target: 100 },
    { duration: '2m', target: 200 },
    { duration: '5m', target: 200 },
    { duration: '2m', target: 0 },
  ],
};

export default function() {
  let response = http.get('http://localhost:8080/api/v1/health');
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  sleep(1);
}
```

## ⚡ Performance & Caching

### Redis Caching

```go
// backend/services/cache.go
package services

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type CacheService struct {
    redis *redis.Client
}

func NewCacheService(redis *redis.Client) *CacheService {
    return &CacheService{redis: redis}
}

func (c *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
    jsonData, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.redis.Set(ctx, key, jsonData, expiration).Err()
}

func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.redis.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(val), dest)
}

func (c *CacheService) Delete(ctx context.Context, key string) error {
    return c.redis.Del(ctx, key).Err()
}

func (c *CacheService) GetOrSet(ctx context.Context, key string, dest interface{}, 
    fetchFunc func() (interface{}, error), expiration time.Duration) error {
    
    err := c.Get(ctx, key, dest)
    if err == nil {
        return nil
    }
    
    if err != redis.Nil {
        return err
    }
    
    data, err := fetchFunc()
    if err != nil {
        return err
    }
    
    err = c.Set(ctx, key, data, expiration)
    if err != nil {
        return err
    }
    
    return json.Unmarshal([]byte(data.(string)), dest)
}
```

### Response Compression

```go
// backend/middleware/compression.go
package middleware

import (
    "github.com/gin-contrib/gzip"
    "github.com/gin-gonic/gin"
)

func Compression() gin.HandlerFunc {
    return gzip.Gzip(gzip.DefaultCompression)
}
```

## 📱 Mobile-Specific Features

### Push Notifications

```go
// backend/services/notification.go
package services

import (
    "context"
    "firebase.google.com/go/v4/messaging"
    "github.com/go-redis/redis/v8"
)

type NotificationService struct {
    fcmClient *messaging.Client
    redis     *redis.Client
}

func NewNotificationService(fcmClient *messaging.Client, redis *redis.Client) *NotificationService {
    return &NotificationService{
        fcmClient: fcmClient,
        redis:     redis,
    }
}

func (n *NotificationService) SendPushNotification(ctx context.Context, userID uint, title, body string) error {
    // Get user's FCM token from cache
    token, err := n.redis.Get(ctx, fmt.Sprintf("fcm_token:%d", userID)).Result()
    if err != nil {
        return err
    }
    
    message := &messaging.Message{
        Token: token,
        Notification: &messaging.Notification{
            Title: title,
            Body:  body,
        },
        Data: map[string]string{
            "user_id": fmt.Sprintf("%d", userID),
        },
    }
    
    _, err = n.fcmClient.Send(ctx, message)
    return err
}

func (n *NotificationService) RegisterToken(ctx context.Context, userID uint, token string) error {
    return n.redis.Set(ctx, fmt.Sprintf("fcm_token:%d", userID), token, 0).Err()
}
```

### File Upload

```go
// backend/controllers/upload.go
package controllers

import (
    "fmt"
    "io"
    "mime/multipart"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type UploadController struct {
    uploadPath string
}

func NewUploadController(uploadPath string) *UploadController {
    return &UploadController{uploadPath: uploadPath}
}

func (u *UploadController) UploadFile(c *gin.Context) {
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(400, gin.H{"error": "No file uploaded"})
        return
    }
    defer file.Close()
    
    // Validate file type
    if !u.isValidFileType(header) {
        c.JSON(400, gin.H{"error": "Invalid file type"})
        return
    }
    
    // Generate unique filename
    ext := filepath.Ext(header.Filename)
    filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
    
    // Create upload directory if not exists
    if err := os.MkdirAll(u.uploadPath, 0755); err != nil {
        c.JSON(500, gin.H{"error": "Failed to create upload directory"})
        return
    }
    
    // Save file
    filepath := filepath.Join(u.uploadPath, filename)
    if err := u.saveFile(file, filepath); err != nil {
        c.JSON(500, gin.H{"error": "Failed to save file"})
        return
    }
    
    c.JSON(200, gin.H{
        "filename": filename,
        "url":      fmt.Sprintf("/uploads/%s", filename),
        "size":     header.Size,
    })
}

func (u *UploadController) isValidFileType(header *multipart.FileHeader) bool {
    allowedTypes := []string{"image/jpeg", "image/png", "image/gif", "application/pdf"}
    contentType := header.Header.Get("Content-Type")
    
    for _, allowedType := range allowedTypes {
        if contentType == allowedType {
            return true
        }
    }
    return false
}

func (u *UploadController) saveFile(src multipart.File, dst string) error {
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()
    
    _, err = io.Copy(out, src)
    return err
}
```

## 🚀 CI/CD Pipeline

### GitHub Actions

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: testdb
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run linter
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
    
    - name: Run tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
    
    - name: Run security scan
      uses: securecodewarrior/github-action-add-sarif@v1
      with:
        sarif-file: security-scan-results.sarif

  security:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'
        format: 'sarif'
        output: 'security-scan-results.sarif'
    
    - name: Upload Trivy scan results
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'security-scan-results.sarif'

  build:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Login to Docker Hub
      uses: docker/login-action@v2
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}
    
    - name: Build and push
      uses: docker/build-push-action@v3
      with:
        context: ./backend
        push: true
        tags: |
          mobile-backend:latest
          mobile-backend:${{ github.sha }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Deploy to Kubernetes
      run: |
        echo "${{ secrets.KUBE_CONFIG }}" | base64 -d > kubeconfig
        export KUBECONFIG=kubeconfig
        kubectl apply -f infrastructure/kubernetes/
```

## 📚 API Documentation

### Swagger Configuration

```go
// backend/docs/swagger.go
package docs

import (
    "mobile-backend/docs"
    "github.com/gin-gonic/gin"
    ginSwagger "github.com/swaggo/gin-swagger"
    "github.com/swaggo/gin-swagger/swaggerFiles"
)

func SetupSwagger(r *gin.Engine) {
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// @title Mobile Backend API
// @version 1.0
// @description A comprehensive mobile backend API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name auth
// @tag.description Authentication endpoints

// @tag.name users
// @tag.description User management endpoints

// @tag.name health
// @tag.description Health check endpoints
```

### API Examples

```go
// User registration endpoint
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.User true "User registration data"
// @Success 201 {object} response.UserResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /auth/register [post]
func RegisterUser(c *gin.Context) {
    // Implementation
}

// Get user profile
// @Summary Get user profile
// @Description Get the authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.UserResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /profile [get]
func GetProfile(c *gin.Context) {
    // Implementation
}
```

## 🏗 Infrastructure as Code

### Kubernetes Manifests

```yaml
# infrastructure/kubernetes/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mobile-backend
  namespace: mobile-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mobile-backend
  template:
    metadata:
      labels:
        app: mobile-backend
    spec:
      containers:
      - name: mobile-backend
        image: mobile-backend:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: mobile-backend-secrets
              key: database-url
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: mobile-backend-secrets
              key: jwt-secret
        - name: REDIS_URL
          valueFrom:
            configMapKeyRef:
              name: mobile-backend-config
              key: redis-url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### Terraform Configuration

```hcl
# infrastructure/terraform/main.tf
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

resource "aws_ecs_cluster" "mobile_backend" {
  name = "mobile-backend-cluster"
}

resource "aws_ecs_service" "mobile_backend" {
  name            = "mobile-backend-service"
  cluster         = aws_ecs_cluster.mobile_backend.id
  task_definition = aws_ecs_task_definition.mobile_backend.arn
  desired_count   = 3

  load_balancer {
    target_group_arn = aws_lb_target_group.mobile_backend.arn
    container_name   = "mobile-backend"
    container_port   = 8080
  }
}

resource "aws_rds_instance" "postgres" {
  identifier = "mobile-backend-db"
  engine     = "postgres"
  engine_version = "15.4"
  instance_class = "db.t3.micro"
  allocated_storage = 20
  storage_type = "gp2"
  
  db_name  = "appdb"
  username = "appuser"
  password = var.db_password
  
  vpc_security_group_ids = [aws_security_group.rds.id]
  db_subnet_group_name   = aws_db_subnet_group.mobile_backend.name
  
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"
  
  skip_final_snapshot = true
}
```

## 🛠 Development Tools

### Makefile

```makefile
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
db-migrate: ## Run database migrations
	cd backend && go run scripts/migrate.go up

db-rollback: ## Rollback database migrations
	cd backend && go run scripts/migrate.go down

db-seed: ## Seed database with test data
	cd backend && go run scripts/seed.go

db-reset: ## Reset database (drop, create, migrate, seed)
	docker-compose exec db psql -U appuser -d appdb -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	$(MAKE) db-migrate
	$(MAKE) db-seed

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
	@curl -s http://localhost:8080/health | jq .
	@curl -s http://localhost:8080/health/ready | jq .
	@curl -s http://localhost:8080/health/live | jq .
```

### Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
      - id: check-merge-conflict

  - repo: https://github.com/golangci/golangci-lint
    rev: v1.54.2
    hooks:
      - id: golangci-lint
        args: [--fix]

  - repo: https://github.com/psf/black
    rev: 23.7.0
    hooks:
      - id: black
        language_version: python3

  - repo: https://github.com/pre-commit/mirrors-eslint
    rev: v8.45.0
    hooks:
      - id: eslint
        files: \.(js|ts)$
```

### Go Linting Configuration

```yaml
# .golangci.yml
run:
  timeout: 5m
  modules-download-mode: readonly

linters-settings:
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  funlen:
    lines: 100
    statements: 50
  gocritic:
    enabled-tags:
      - diagnostic
      - experimental
      - opinionated
      - performance
      - style
  goconst:
    min-len: 2
    min-occurrences: 2
  gocognit:
    min-complexity: 20
  gomnd:
    settings:
      mnd:
        checks: argument,case,condition,operation,return,assign
  govet:
    check-shadowing: true
  lll:
    line-length: 140
  maligned:
    suggest-new: true
  misspell:
    locale: US
  nolintlint:
    allow-leading-space: true
    allow-unused: false
    require-explanation: false
    require-specific: false
  revive:
    min-confidence: 0
    rules:
      - name: blank-imports
      - name: context-as-argument
      - name: context-keys-type
      - name: dot-imports
      - name: error-return
      - name: error-strings
      - name: error-naming
      - name: exported
      - name: if-return
      - name: increment-decrement
      - name: var-naming
      - name: var-declaration
      - name: package-comments
      - name: range
      - name: receiver-naming
      - name: time-naming
      - name: unexported-return
      - name: indent-error-flow
      - name: errorf
  unused:
    go: "1.19"

linters:
  enable:
    - bodyclose
    - deadcode
    - depguard
    - dogsled
    - dupl
    - errcheck
    - exportloopref
    - exhaustive
    - funlen
    - gochecknoinits
    - goconst
    - gocritic
    - gocyclo
    - gofmt
    - goimports
    - golint
    - gomnd
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - lll
    - misspell
    - nakedret
    - noctx
    - nolintlint
    - revive
    - staticcheck
    - structcheck
    - stylecheck
    - typecheck
    - unconvert
    - unparam
    - unused
    - varcheck
    - whitespace

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gomnd
        - funlen
        - gocyclo
        - goconst
    - path: config/
      linters:
        - gomnd
    - text: "weak cryptographic primitive"
      linters:
        - gosec
    - text: "G204"
      linters:
        - gosec
```

## 📊 Performance Monitoring

### Prometheus Configuration

```yaml
# monitoring/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "rules/*.yml"

scrape_configs:
  - job_name: 'mobile-backend'
    static_configs:
      - targets: ['backend:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Mobile Backend Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status_code=~\"5..\"}[5m])",
            "legendFormat": "5xx errors"
          }
        ]
      }
    ]
  }
}
```

## 🔧 Advanced Development Features

### Database Seeding

```go
// backend/scripts/seed.go
package main

import (
    "log"
    "mobile-backend/config"
    "mobile-backend/models"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    db := config.GetDB()
    
    // Create admin user
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
    admin := models.User{
        Email:    "admin@example.com",
        Password: string(hashedPassword),
        Name:     "Admin User",
        IsActive: true,
    }
    
    if err := db.Create(&admin).Error; err != nil {
        log.Printf("Admin user already exists or error: %v", err)
    }
    
    // Create test users
    for i := 1; i <= 10; i++ {
        hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
        user := models.User{
            Email:    fmt.Sprintf("user%d@example.com", i),
            Password: string(hashedPassword),
            Name:     fmt.Sprintf("Test User %d", i),
            IsActive: true,
        }
        
        if err := db.Create(&user).Error; err != nil {
            log.Printf("User %d already exists or error: %v", i, err)
        }
    }
    
    log.Println("Database seeded successfully!")
}
```

### Code Generation

```go
// backend/scripts/generate.go
package main

import (
    "fmt"
    "os"
    "text/template"
)

type ModelInfo struct {
    Name    string
    Fields  []Field
    Package string
}

type Field struct {
    Name string
    Type string
    Tag  string
}

func main() {
    models := []ModelInfo{
        {
            Name:    "Product",
            Package: "models",
            Fields: []Field{
                {Name: "Name", Type: "string", Tag: `json:"name" gorm:"not null"`},
                {Name: "Price", Type: "float64", Tag: `json:"price" gorm:"type:decimal(10,2)"`},
                {Name: "Description", Type: "string", Tag: `json:"description"`},
            },
        },
    }
    
    tmpl := `package {{.Package}}

import "gorm.io/gorm"

type {{.Name}} struct {
    gorm.Model
    {{range .Fields}}
    {{.Name}} {{.Type}} ` + "`{{.Tag}}`" + `{{end}}
}
`
    
    t := template.Must(template.New("model").Parse(tmpl))
    
    for _, model := range models {
        file, err := os.Create(fmt.Sprintf("models/%s.go", model.Name))
        if err != nil {
            panic(err)
        }
        
        if err := t.Execute(file, model); err != nil {
            panic(err)
        }
        
        file.Close()
        fmt.Printf("Generated %s model\n", model.Name)
    }
}
```

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.