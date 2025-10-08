#!/bin/bash

# Dokploy Deployment Script
# This script helps prepare and deploy the mobile backend to Dokploy

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_NAME="mobile-backend"
DOCKER_COMPOSE_FILE="docker-compose.dokploy.yml"
ENV_FILE="env.dokploy"

echo -e "${BLUE}🚀 Mobile Backend Dokploy Deployment Script${NC}"
echo "=================================================="

# Check if required files exist
check_files() {
    echo -e "${YELLOW}📋 Checking required files...${NC}"
    
    if [ ! -f "$DOCKER_COMPOSE_FILE" ]; then
        echo -e "${RED}❌ Error: $DOCKER_COMPOSE_FILE not found${NC}"
        exit 1
    fi
    
    if [ ! -f "$ENV_FILE" ]; then
        echo -e "${RED}❌ Error: $ENV_FILE not found${NC}"
        exit 1
    fi
    
    if [ ! -f "backend/Dockerfile.prod" ]; then
        echo -e "${RED}❌ Error: backend/Dockerfile.prod not found${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ All required files found${NC}"
}

# Validate environment variables
validate_env() {
    echo -e "${YELLOW}🔍 Validating environment variables...${NC}"
    
    # Check if .env file exists
    if [ ! -f ".env" ]; then
        echo -e "${YELLOW}⚠️  .env file not found. Creating from template...${NC}"
        cp "$ENV_FILE" ".env"
        echo -e "${YELLOW}📝 Please edit .env file with your actual values before deploying${NC}"
        echo -e "${YELLOW}   Required variables: POSTGRES_PASSWORD, JWT_SECRET${NC}"
        return 1
    fi
    
    # Check required variables
    source .env
    
    if [ -z "$POSTGRES_PASSWORD" ] || [ "$POSTGRES_PASSWORD" = "your_secure_database_password_here" ]; then
        echo -e "${RED}❌ Error: POSTGRES_PASSWORD must be set in .env file${NC}"
        return 1
    fi
    
    if [ -z "$JWT_SECRET" ] || [ "$JWT_SECRET" = "your_super_secret_jwt_key_change_this_in_production" ]; then
        echo -e "${RED}❌ Error: JWT_SECRET must be set in .env file${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ Environment variables validated${NC}"
    return 0
}

# Build and test Docker image
build_and_test() {
    echo -e "${YELLOW}🔨 Building Docker image...${NC}"
    
    cd backend
    docker build -f Dockerfile.prod -t mobile-backend:latest .
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Docker image built successfully${NC}"
    else
        echo -e "${RED}❌ Error: Docker image build failed${NC}"
        exit 1
    fi
    
    cd ..
}

# Test Docker Compose configuration
test_compose() {
    echo -e "${YELLOW}🧪 Testing Docker Compose configuration...${NC}"
    
    docker-compose -f "$DOCKER_COMPOSE_FILE" config > /dev/null
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Docker Compose configuration is valid${NC}"
    else
        echo -e "${RED}❌ Error: Docker Compose configuration is invalid${NC}"
        exit 1
    fi
}

# Generate deployment summary
generate_summary() {
    echo -e "${BLUE}📊 Deployment Summary${NC}"
    echo "===================="
    echo "Project Name: $PROJECT_NAME"
    echo "Docker Compose: $DOCKER_COMPOSE_FILE"
    echo "Environment: $(pwd)/.env"
    echo ""
    echo -e "${YELLOW}📋 Next Steps:${NC}"
    echo "1. Push your code to Git repository"
    echo "2. Create new project in Dokploy dashboard"
    echo "3. Connect your Git repository"
    echo "4. Set environment variables in Dokploy"
    echo "5. Deploy using the Docker Compose file"
    echo ""
    echo -e "${YELLOW}🔧 Required Environment Variables in Dokploy:${NC}"
    echo "- POSTGRES_PASSWORD"
    echo "- JWT_SECRET"
    echo "- CORS_ALLOWED_ORIGINS (optional)"
    echo "- GRAFANA_ADMIN_PASSWORD (optional)"
    echo ""
    echo -e "${YELLOW}🌐 Services will be available at:${NC}"
    echo "- API: https://yourdomain.com:8080"
    echo "- API Docs: https://yourdomain.com:8080/swagger/index.html"
    echo "- Grafana: https://yourdomain.com:3000"
    echo "- Prometheus: https://yourdomain.com:9090"
}

# Main execution
main() {
    check_files
    
    if validate_env; then
        build_and_test
        test_compose
    else
        echo -e "${YELLOW}⚠️  Please configure .env file and run the script again${NC}"
        exit 1
    fi
    
    generate_summary
    
    echo -e "${GREEN}🎉 Preparation complete! Ready for Dokploy deployment.${NC}"
}

# Run main function
main "$@"
