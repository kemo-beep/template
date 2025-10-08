#!/bin/bash

# Dokploy Environment Variables Setup Script
# This script helps you set up the required environment variables for Dokploy deployment

echo "🚀 Dokploy Environment Variables Setup"
echo "======================================"
echo ""

# Generate a random JWT secret
JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "your_super_secret_jwt_key_$(date +%s)")

# Generate a random database password
DB_PASSWORD=$(openssl rand -base64 16 2>/dev/null || echo "secure_password_$(date +%s)")

echo "📋 Required Environment Variables for Dokploy:"
echo ""
echo "Copy and paste these into your Dokploy project environment variables:"
echo ""
echo "=========================================="
echo "# Database Configuration (REQUIRED)"
echo "DATABASE_URL=postgres://appuser:apppass@postgres:5432/appdb?sslmode=disable"
echo "POSTGRES_PASSWORD=$DB_PASSWORD"
echo "JWT_SECRET=$JWT_SECRET"
echo ""
echo "# Redis Configuration (REQUIRED)"
echo "REDIS_URL=redis:6379"
echo ""
echo "# Optional Variables"
echo "CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com"
echo "GRAFANA_ADMIN_PASSWORD=admin"
echo "POSTGRES_DB=appdb"
echo "POSTGRES_USER=appuser"
echo "GIN_MODE=release"
echo "LOG_LEVEL=info"
echo "=========================================="
echo ""

echo "⚠️  Important Notes:"
echo "- DATABASE_URL uses 'postgres' as hostname (not localhost)"
echo "- REDIS_URL uses 'redis' as hostname (not localhost)"
echo "- Replace 'yourdomain.com' with your actual domain"
echo "- Save these credentials securely!"
echo ""

echo "✅ After setting these variables in Dokploy, redeploy your project."
echo ""

# Create a .env file for reference
cat > .env.dokploy << EOF
# Dokploy Environment Variables
# Copy these to your Dokploy project settings

# Database Configuration (REQUIRED)
DATABASE_URL=postgres://appuser:apppass@postgres:5432/appdb?sslmode=disable
POSTGRES_PASSWORD=$DB_PASSWORD
JWT_SECRET=$JWT_SECRET

# Redis Configuration (REQUIRED)
REDIS_URL=redis:6379

# Optional Variables
CORS_ALLOWED_ORIGINS=https://yourdomain.com,https://www.yourdomain.com
GRAFANA_ADMIN_PASSWORD=admin
POSTGRES_DB=appdb
POSTGRES_USER=appuser
GIN_MODE=release
LOG_LEVEL=info
EOF

echo "📄 Environment variables saved to .env.dokploy for reference"
echo ""
echo "🔧 Next Steps:"
echo "1. Copy the variables above to your Dokploy project settings"
echo "2. Update CORS_ALLOWED_ORIGINS with your actual domain"
echo "3. Redeploy your project in Dokploy"
echo "4. Check the logs to ensure the application starts successfully"
echo ""
echo "🐛 If you still get DATABASE_URL errors:"
echo "- Verify the environment variables are set correctly in Dokploy"
echo "- Check that the postgres service is running"
echo "- Ensure the DATABASE_URL uses 'postgres' as the hostname"
