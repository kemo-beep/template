# Dokploy Deployment Guide

This guide will help you deploy the Mobile Backend Template on Dokploy using Docker Compose.

## 🚀 Quick Start

### Prerequisites

1. **Dokploy Instance**: A running Dokploy instance
2. **Domain**: A domain name pointing to your Dokploy server (optional but recommended)
3. **SSL Certificate**: For production deployment (handled by Dokploy)

### Step 1: Prepare Your Repository

1. **Clone or Fork** this repository
2. **Push to Git**: Ensure your code is in a Git repository (GitHub, GitLab, etc.)

### Step 2: Create New Project in Dokploy

1. **Login** to your Dokploy dashboard
2. **Create New Project**:
   - Project Name: `mobile-backend`
   - Description: `Production-ready Go mobile backend`
   - Repository: Your Git repository URL
   - Branch: `main` (or your preferred branch)

### Step 3: Configure Environment Variables

In the Dokploy project settings, add these environment variables:

#### Required Variables
```env
POSTGRES_PASSWORD=your_secure_database_password_here
JWT_SECRET=your_super_secret_jwt_key_change_this_in_production
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
GRAFANA_ADMIN_PASSWORD=your_grafana_admin_password
```

#### Optional Variables
```env
POSTGRES_DB=appdb
POSTGRES_USER=appuser
GIN_MODE=release
LOG_LEVEL=info
```

### Step 4: Configure Docker Compose

1. **Docker Compose File**: Use `docker-compose.dokploy.yml`
2. **Port Mapping**: 
   - `8080:8080` (Main API)
   - `3001:3000` (Grafana - Optional)
   - `9090:9090` (Prometheus - Optional)

### Step 5: Deploy

1. **Deploy Project**: Click "Deploy" in Dokploy
2. **Monitor Logs**: Watch the deployment logs for any issues
3. **Health Check**: Verify the health endpoint is responding

## 🔧 Configuration Details

### Docker Compose Configuration

The `docker-compose.dokploy.yml` file includes:

- **Mobile Backend**: Main Go application
- **PostgreSQL**: Database with persistent storage
- **Redis**: Caching and session storage
- **Prometheus**: Metrics collection
- **Grafana**: Monitoring dashboards

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `POSTGRES_PASSWORD` | Database password | Yes | - |
| `JWT_SECRET` | JWT signing secret | Yes | - |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins | No | `*` |
| `GRAFANA_ADMIN_PASSWORD` | Grafana admin password | No | `admin` |
| `POSTGRES_DB` | Database name | No | `appdb` |
| `POSTGRES_USER` | Database user | No | `appuser` |

### Health Checks

The application includes comprehensive health checks:

- **Liveness**: `GET /health/live`
- **Readiness**: `GET /health/ready`
- **General Health**: `GET /health`

## 📊 Monitoring Setup

### Accessing Monitoring Tools

After deployment, you can access:

- **API**: `https://yourdomain.com:8080`
- **API Documentation**: `https://yourdomain.com:8080/swagger/index.html`
- **Grafana**: `https://yourdomain.com:3001` (admin/your_password)
- **Prometheus**: `https://yourdomain.com:9090`

### Grafana Dashboards

The deployment includes pre-configured dashboards for:

- HTTP request metrics
- Database performance
- Redis metrics
- Application health

## 🔒 Security Considerations

### Production Security

1. **Change Default Passwords**:
   - Database password
   - Grafana admin password
   - JWT secret

2. **Configure CORS**:
   - Set specific allowed origins
   - Avoid using `*` in production

3. **SSL/TLS**:
   - Dokploy handles SSL certificates
   - Ensure HTTPS is enabled

4. **Firewall Rules**:
   - Only expose necessary ports
   - Consider restricting monitoring ports

### Environment Security

```env
# Production environment variables
POSTGRES_PASSWORD=very_secure_random_password_here
JWT_SECRET=very_long_random_jwt_secret_key_here
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
GRAFANA_ADMIN_PASSWORD=secure_grafana_password
```

## 🚀 Deployment Commands

### Using Dokploy UI

1. **Create Project** in Dokploy dashboard
2. **Connect Repository** to your Git repo
3. **Set Environment Variables** in project settings
4. **Deploy** using the Dokploy UI

### Using Dokploy CLI (if available)

```bash
# Install Dokploy CLI (if available)
# Configure Dokploy CLI
dokploy config set server https://your-dokploy-instance.com

# Deploy project
dokploy deploy mobile-backend
```

## 📝 Post-Deployment

### 1. Verify Deployment

```bash
# Check API health
curl https://yourdomain.com:8080/health

# Check API documentation
open https://yourdomain.com:8080/swagger/index.html
```

### 2. Database Setup

The application will automatically:
- Create database tables
- Run migrations
- Set up initial data (if seeding is enabled)

### 3. Monitor Application

- Check application logs in Dokploy
- Monitor metrics in Grafana
- Verify all services are healthy

## 🔄 Updates and Maintenance

### Updating the Application

1. **Push Changes** to your Git repository
2. **Redeploy** in Dokploy dashboard
3. **Monitor** deployment logs

### Database Backups

```bash
# Create database backup
docker exec mobile-backend-postgres pg_dump -U appuser appdb > backup.sql

# Restore database backup
docker exec -i mobile-backend-postgres psql -U appuser appdb < backup.sql
```

### Log Management

- **Application Logs**: Available in Dokploy dashboard
- **Database Logs**: Check PostgreSQL container logs
- **Monitoring Logs**: Check Prometheus/Grafana logs

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check `POSTGRES_PASSWORD` environment variable
   - Verify database container is running
   - Check network connectivity

2. **JWT Token Issues**
   - Verify `JWT_SECRET` is set
   - Check token expiration settings
   - Ensure consistent secret across deployments

3. **CORS Issues**
   - Check `CORS_ALLOWED_ORIGINS` configuration
   - Verify frontend domain is included
   - Test with browser developer tools

4. **Health Check Failures**
   - Check application logs
   - Verify all dependencies are running
   - Check resource limits

### Debug Commands

```bash
# Check container status
docker ps

# Check application logs
docker logs mobile-backend

# Check database logs
docker logs mobile-backend-postgres

# Check Redis logs
docker logs mobile-backend-redis

# Test database connection
docker exec mobile-backend-postgres psql -U appuser -d appdb -c "SELECT 1;"

# Test Redis connection
docker exec mobile-backend-redis redis-cli ping
```

## 📚 Additional Resources

- [Dokploy Documentation](https://dokploy.com/docs)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Redis Documentation](https://redis.io/docs/)

## 🆘 Support

If you encounter issues:

1. Check the troubleshooting section above
2. Review application logs in Dokploy
3. Check the project's GitHub issues
4. Create a new issue with detailed information

---

**Happy Deploying! 🚀**
