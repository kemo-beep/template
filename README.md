# Mobile Backend Template

A comprehensive, production-ready Go (Gin) + PostgreSQL mobile backend template with excellent developer experience. This template includes authentication, database management, monitoring, testing, and all the tools needed for modern mobile app development.

## 🚀 Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Git

### 1. Clone and Setup

```bash
git clone <your-repo-url>
cd mobile-backend
cp env.example .env
# Edit .env with your configuration
```

### 2. Start Development Environment

```bash
# Option 1: Using Make (recommended)
make dev

# Option 2: Using Docker Compose directly
docker-compose up -d

# Option 3: Using the development script
./scripts/dev.sh
```

### 3. Access Services

- **Backend API**: http://localhost:8081
- **API Documentation**: http://localhost:8081/swagger/index.html
   - **Database**: localhost:5434
- **Redis**: localhost:6379
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)
- **Jaeger**: http://localhost:16686

## 🛠 Features

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

## 📁 Project Structure

```
mobile-backend/
├── backend/                    # Go backend application
│   ├── config/                # Configuration files
│   ├── controllers/           # HTTP handlers
│   ├── models/                # Database models
│   ├── routes/                # API routes
│   ├── middleware/            # HTTP middleware
│   ├── services/              # Business logic
│   ├── utils/                 # Utility functions
│   ├── tests/                 # Test files
│   ├── scripts/               # Utility scripts
│   └── migrations/            # Database migrations
├── infrastructure/            # Infrastructure as Code
│   ├── kubernetes/           # K8s manifests
│   ├── terraform/            # Terraform configs
│   └── helm/                 # Helm charts
├── monitoring/               # Monitoring configurations
│   ├── prometheus/           # Prometheus config
│   ├── grafana/              # Grafana dashboards
│   └── jaeger/               # Jaeger tracing
├── scripts/                  # Development scripts
├── .github/workflows/        # CI/CD pipelines
├── docker-compose.yml        # Development environment
├── docker-compose.prod.yml   # Production environment
├── Makefile                  # Build automation
└── README.md                 # This file
```

## 🔧 Development Commands

### Using Make (Recommended)

```bash
# Development
make dev              # Start development environment
make dev-build        # Build and start development
make dev-logs         # View development logs
make dev-stop         # Stop development environment

# Production
make prod             # Start production environment
make prod-build       # Build and start production
make prod-logs        # View production logs
make prod-stop        # Stop production environment

# Testing
make test             # Run all tests
make test-unit        # Run unit tests only
make test-integration # Run integration tests
make test-coverage    # Generate coverage report

# Code Quality
make lint             # Run linter
make lint-fix         # Fix linting issues
make format           # Format code
make security-scan    # Run security scan

# Database
make db-migrate       # Run database migrations
make db-seed          # Seed database with test data
make db-reset         # Reset database

# Build
make build            # Build the application
make build-linux      # Build for Linux
make docker-build     # Build Docker image

# Monitoring
make monitor          # Start monitoring stack
make monitor-stop     # Stop monitoring stack
make health           # Check service health

# Utilities
make clean            # Clean up build artifacts
make deps             # Install dependencies
make docs             # Generate API documentation
make backup           # Backup database
```

### Using Docker Compose

```bash
# Development
docker-compose up -d                    # Start all services
docker-compose up -d --build           # Build and start
docker-compose logs -f backend         # View backend logs
docker-compose down                    # Stop all services

# Production
docker-compose -f docker-compose.prod.yml up -d
```

## 🧪 Testing

### Run Tests

```bash
# All tests
make test

# Unit tests only
make test-unit

# Integration tests only
make test-integration

# With coverage
make test-coverage
```

### Load Testing

```bash
# Install k6
brew install k6  # macOS
# or
curl https://github.com/grafana/k6/releases/download/v0.45.0/k6-v0.45.0-linux-amd64.tar.gz -L | tar xvz --strip-components 1

# Run load test
make load-test
# or
k6 run k6-load-test.js
```

## 📊 Monitoring

### Prometheus Metrics

Access Prometheus at http://localhost:9090 to view:
- HTTP request metrics
- Database connection metrics
- Redis metrics
- Custom application metrics

### Grafana Dashboards

Access Grafana at http://localhost:3001 (admin/admin) to view:
- Application performance dashboards
- Infrastructure monitoring
- Custom business metrics

### Jaeger Tracing

Access Jaeger at http://localhost:16686 to view:
- Distributed request tracing
- Performance analysis
- Error tracking

## 🔒 Security Features

- **Rate Limiting**: Redis-based rate limiting
- **CORS**: Configurable CORS policies
- **Input Validation**: Comprehensive input validation
- **JWT Authentication**: Secure token-based auth
- **Password Hashing**: bcrypt password hashing
- **Security Headers**: XSS, CSRF, and other security headers
- **SQL Injection Prevention**: GORM ORM protection

## 📱 Mobile-Specific Features

- **Push Notifications**: Firebase Cloud Messaging integration
- **File Upload**: Secure file upload with validation
- **Image Processing**: Built-in image handling
- **Real-time Features**: WebSocket support ready
- **Offline Sync**: Database design for offline capabilities

## 🚀 Deployment

### Dokploy Deployment (Recommended)

```bash
# Prepare for Dokploy deployment
make dokploy-prepare

# Follow the Dokploy deployment guide
# See DOKPLOY_DEPLOYMENT.md for detailed instructions
```

**Quick Dokploy Setup:**
1. Push code to Git repository
2. Create project in Dokploy dashboard
3. Connect Git repository
4. Set environment variables in Dokploy
5. Deploy using `docker-compose.dokploy.yml`

### Docker Deployment

```bash
# Build production image
make docker-build-prod

# Run production container
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL="your-db-url" \
  -e REDIS_URL="your-redis-url" \
  -e JWT_SECRET="your-secret" \
  mobile-backend:prod
```

### Kubernetes Deployment

```bash
# Apply Kubernetes manifests
kubectl apply -f infrastructure/kubernetes/

# Check deployment status
kubectl get pods -n mobile-backend
kubectl get services -n mobile-backend
```

### Environment Variables

Create a `.env` file with the following variables:

```env
# Database
DATABASE_URL=postgres://user:pass@localhost:5432/dbname?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-super-secret-jwt-key

# Server
GIN_MODE=debug
PORT=8080

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3001,http://localhost:8080

# Email (optional)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
FROM_EMAIL=noreply@yourapp.com

# Firebase (optional)
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_PRIVATE_KEY=your-private-key
FIREBASE_CLIENT_EMAIL=your-client-email
```

## 📚 API Documentation

### Swagger Documentation

Once the server is running, access the interactive API documentation at:
http://localhost:8080/swagger/index.html

### API Endpoints

#### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/logout` - Logout user (protected)

#### User Management
- `GET /api/v1/profile` - Get user profile (protected)
- `PUT /api/v1/profile` - Update user profile (protected)
- `DELETE /api/v1/profile` - Delete user profile (protected)
- `GET /api/v1/users/:id` - Get user by ID (protected)

#### File Upload
- `POST /api/v1/upload` - Upload single file (protected)
- `POST /api/v1/upload/multiple` - Upload multiple files (protected)
- `GET /api/v1/uploads/:filename` - Get uploaded file (protected)
- `DELETE /api/v1/uploads/:filename` - Delete uploaded file (protected)

#### Health Checks
- `GET /health` - Basic health check
- `GET /health/ready` - Readiness check
- `GET /health/live` - Liveness check

## 🔧 Configuration

### Database Configuration

The application uses GORM with PostgreSQL. Database configuration is handled through environment variables:

- `DATABASE_URL`: PostgreSQL connection string
- Auto-migration is enabled by default
- Connection pooling is configured for optimal performance

### Redis Configuration

Redis is used for caching and session storage:

- `REDIS_URL`: Redis connection string
- Session tokens are stored with 72-hour expiration
- Cache keys are namespaced for organization

### Logging Configuration

Structured logging with Zap:

- Development mode: Human-readable logs
- Production mode: JSON logs
- Log levels: debug, info, warn, error, fatal
- Request logging middleware included

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and idioms
- Write tests for new features
- Update documentation as needed
- Run `make lint` before committing
- Ensure all tests pass with `make test`

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

If you encounter any issues or have questions:

1. Check the [troubleshooting section](#troubleshooting) below
2. Search existing [GitHub Issues](https://github.com/your-org/mobile-backend/issues)
3. Create a new issue with detailed information

## 🐛 Troubleshooting

### Common Issues

1. **Database connection failed**
   ```bash
   # Check if PostgreSQL is running
   docker-compose ps
   
   # Check database logs
   docker-compose logs db
   
   # Verify DATABASE_URL in .env
   ```

2. **Redis connection failed**
   ```bash
   # Check if Redis is running
   docker-compose ps
   
   # Check Redis logs
   docker-compose logs redis
   
   # Test Redis connection
   docker-compose exec redis redis-cli ping
   ```

3. **Port conflicts**
   ```bash
   # Check if ports are in use
   lsof -i :8080
   lsof -i :5432
   lsof -i :6379
   
   # Change ports in docker-compose.yml if needed
   ```

4. **Build failures**
   ```bash
   # Clean and rebuild
   make clean
   make dev-build
   
   # Check Go version
   go version
   ```

5. **Test failures**
   ```bash
   # Run tests with verbose output
   make test
   
   # Check test database connection
   # Ensure test containers are working
   ```

## 🎯 Roadmap

- [ ] GraphQL API support
- [ ] WebSocket real-time features
- [ ] Advanced caching strategies
- [ ] Multi-tenant support
- [ ] Advanced monitoring dashboards
- [ ] API versioning
- [ ] Advanced security features
- [ ] Performance optimization tools

---

**Happy Coding! 🚀**
