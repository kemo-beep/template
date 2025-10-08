# 🚀 API Generator System - Developer Experience Guide

## Overview

The Mobile Backend Template now includes a powerful **API Generator System** that provides an amazing developer experience (DX) for rapidly creating CRUD APIs from database schemas and migrations.

## ✨ Features

### 🎯 Schema-Based API Generation
- **JSON Schema Definition**: Define your data models using JSON schemas
- **Auto CRUD Generation**: Automatically generates full CRUD APIs
- **Code Generation**: Creates models, controllers, and routes
- **Swagger Integration**: Auto-generates API documentation

### 🔄 Migration-Based API Management
- **Auto-Migration Processing**: Automatically generates APIs when running migrations
- **Migration Cleanup**: Removes APIs when dropping tables
- **Database Introspection**: Generate APIs for existing tables

### 🛠 Developer Tools
- **CLI Generator**: Command-line tools for development
- **Makefile Integration**: Easy commands for common tasks
- **Hot Reload**: Live development with Air
- **Template System**: Pre-built templates for common patterns

## 🚀 Quick Start

### 1. Setup Generator
```bash
# Setup the generator system
make generator-setup

# Or manually
cd backend && go run cmd/generator/main.go -action setup
```

### 2. Generate APIs from Schema
```bash
# Generate APIs from JSON schema
make generator-schema
# Enter: examples/product.json

# Or directly
curl -X POST http://localhost:8081/api/v1/schema/generate \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d @backend/examples/product.json
```

### 3. Generate APIs for All Tables
```bash
# Generate APIs for all existing database tables
make generator-all

# Or via API
curl -X POST http://localhost:8081/api/v1/schema/generate-all \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 📋 API Endpoints

All generator endpoints are protected and require authentication:

### Schema Management
- `GET /api/v1/schema/template` - Get schema template
- `POST /api/v1/schema/generate` - Generate APIs from schema
- `POST /api/v1/schema/upload` - Generate from JSON file upload
- `POST /api/v1/schema/validate` - Validate schema definition

### Migration Management
- `POST /api/v1/schema/migration` - Generate APIs from migration file
- `POST /api/v1/schema/generate-all` - Generate APIs for all tables
- `GET /api/v1/schema/migrations` - Get migration status

### Model Management
- `GET /api/v1/schema/models` - List generated models
- `DELETE /api/v1/schema/cleanup/:model` - Cleanup model files

## 📝 Schema Definition

### Basic Schema Structure
```json
{
  "name": "Product",
  "package": "models",
  "table_name": "products",
  "has_timestamps": true,
  "has_soft_delete": true,
  "comment": "Product catalog for e-commerce",
  "fields": [
    {
      "name": "Name",
      "type": "VARCHAR(255)",
      "go_type": "string",
      "json_tag": "json:\"name\"",
      "gorm_tag": "gorm:\"not null;uniqueIndex\"",
      "validate_tag": "validate:\"required,min=3,max=255\"",
      "required": true,
      "unique": true,
      "comment": "Product name"
    }
  ]
}
```

### Field Properties
- **name**: Field name (Go struct field)
- **type**: SQL data type
- **go_type**: Go data type
- **json_tag**: JSON serialization tag
- **gorm_tag**: GORM database tag
- **validate_tag**: Validation tag
- **required**: Whether field is required
- **unique**: Whether field is unique
- **index**: Whether field is indexed
- **comment**: Field description

## 🔧 CLI Commands

### Generator CLI
```bash
# Generate from schema file
go run cmd/generator/main.go -action generate -schema product.json

# Generate from migration
go run cmd/generator/main.go -action migrate -migration 001_create_products.sql

# Generate for all tables
go run cmd/generator/main.go -action generate-all

# Cleanup model
go run cmd/generator/main.go -action cleanup -model product

# Generate migration template
go run cmd/generator/main.go -action template -table products -op create

# Watch for migrations
go run cmd/generator/main.go -action watch
```

### Makefile Commands
```bash
# Setup
make generator-setup

# Generate APIs
make generator-schema
make generator-migrate
make generator-migrate-all
make generator-all

# Management
make generator-cleanup
make generator-template

# Examples
make generator-examples

# Watch mode
make generator-watch
```

## 🗄️ Migration Integration

### Auto-Migration Processing
When you run a migration that creates a table, the system automatically:
1. Parses the SQL migration
2. Extracts table structure
3. Generates Go model
4. Creates CRUD controller
5. Sets up API routes
6. Updates Swagger documentation

### Migration Cleanup
When you drop a table, the system automatically:
1. Removes generated model file
2. Removes controller file
3. Removes route file
4. Updates documentation

### Migration Templates
```bash
# Create migration template
make generator-template
# Enter table name: products
# Enter operation: create
```

This creates a migration file like:
```sql
-- Migration: Create products table
-- Generated: 2025-10-08 15:04:05

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Add your columns here
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Add indexes if needed
    INDEX idx_products_name (name),
    INDEX idx_products_is_active (is_active)
);
```

## 🎨 Generated Code Examples

### Generated Model
```go
package models

import (
    "time"
    "gorm.io/gorm"
)

// Product represents a Product catalog for e-commerce
type Product struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
    
    Name        string  `json:"name" gorm:"not null;uniqueIndex"`
    Description string  `json:"description" gorm:"type:text"`
    Price       float64 `json:"price" gorm:"type:decimal(10,2);not null"`
    SKU         string  `json:"sku" gorm:"not null;uniqueIndex"`
    Stock       int     `json:"stock" gorm:"not null;default:0"`
    IsActive    bool    `json:"is_active" gorm:"default:true"`
    CategoryID  uint    `json:"category_id" gorm:"not null;index"`
}

func (Product) TableName() string {
    return "products"
}
```

### Generated Controller
```go
package controllers

type ProductController struct {
    db *gorm.DB
}

// GetProductList retrieves all Product records
func (c *ProductController) GetProductList(ctx *gin.Context) { ... }

// GetProduct retrieves a Product by ID
func (c *ProductController) GetProduct(ctx *gin.Context) { ... }

// CreateProduct creates a new Product
func (c *ProductController) CreateProduct(ctx *gin.Context) { ... }

// UpdateProduct updates a Product by ID
func (c *ProductController) UpdateProduct(ctx *gin.Context) { ... }

// DeleteProduct deletes a Product by ID
func (c *ProductController) DeleteProduct(ctx *gin.Context) { ... }
```

### Generated Routes
```go
func SetupProductRoutes(r *gin.Engine, productController *controllers.ProductController) {
    products := r.Group("/api/v1/products")
    products.Use(middleware.AuthMiddleware())
    {
        products.GET("", productController.GetProductList)
        products.GET("/:id", productController.GetProduct)
        products.POST("", productController.CreateProduct)
        products.PUT("/:id", productController.UpdateProduct)
        products.DELETE("/:id", productController.DeleteProduct)
    }
}
```

## 🔄 Development Workflow

### 1. Define Your Schema
Create a JSON schema file:
```json
{
  "name": "User",
  "table_name": "users",
  "has_timestamps": true,
  "fields": [
    {
      "name": "Email",
      "type": "VARCHAR(255)",
      "go_type": "string",
      "json_tag": "json:\"email\"",
      "gorm_tag": "gorm:\"not null;uniqueIndex\"",
      "required": true,
      "unique": true
    }
  ]
}
```

### 2. Generate APIs
```bash
# Via API
curl -X POST http://localhost:8081/api/v1/schema/generate \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d @user.json

# Via CLI
go run cmd/generator/main.go -action generate -schema user.json
```

### 3. Test Your APIs
```bash
# Create a user
curl -X POST http://localhost:8081/api/v1/users \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "name": "Test User"}'

# Get all users
curl -X GET http://localhost:8081/api/v1/users \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. View Documentation
Visit: http://localhost:8081/swagger/index.html

## 🎯 Best Practices

### Schema Design
1. **Use descriptive names** for models and fields
2. **Include proper validation tags** for data integrity
3. **Set appropriate GORM tags** for database constraints
4. **Add meaningful comments** for documentation

### Migration Management
1. **Use timestamped migration files** for proper ordering
2. **Include proper indexes** for performance
3. **Add foreign key constraints** where appropriate
4. **Test migrations** before deploying

### Code Organization
1. **Keep generated files** in version control
2. **Customize generated code** as needed
3. **Add business logic** to controllers
4. **Extend models** with additional methods

## 🚨 Troubleshooting

### Common Issues

#### Permission Denied
```bash
# Fix directory permissions
docker-compose exec --user root backend sh -c "mkdir -p models controllers routes && chown -R appuser:appuser models controllers routes"
```

#### Schema Validation Errors
- Check JSON syntax
- Ensure required fields are present
- Validate field types match Go types

#### Migration Parsing Errors
- Ensure SQL syntax is correct
- Check table name extraction
- Verify column definitions

### Debug Mode
```bash
# Enable debug logging
export LOG_LEVEL=debug
docker-compose up -d --build backend
```

## 🔮 Advanced Features

### Custom Templates
You can customize the generated code by modifying the templates in:
- `backend/generators/schema.go` - Model templates
- `backend/generators/migration.go` - Migration templates

### Integration with CI/CD
```yaml
# GitHub Actions example
- name: Generate APIs
  run: |
    make generator-setup
    make generator-all
    git add models/ controllers/ routes/
    git commit -m "Auto-generate APIs" || true
```

### Database Seeding
```bash
# Generate seed data
go run cmd/generator/main.go -action generate-all
# Then add seed data to generated models
```

## 📚 Examples

Check out the `backend/examples/` directory for:
- `product.json` - E-commerce product schema
- `category.json` - Category management schema
- Migration examples
- Template examples

## 🎉 Conclusion

The API Generator System provides an incredible developer experience by:

1. **Eliminating boilerplate code** - No more writing CRUD APIs by hand
2. **Ensuring consistency** - All APIs follow the same patterns
3. **Saving time** - Generate full APIs in seconds
4. **Maintaining quality** - Generated code includes proper validation, error handling, and documentation
5. **Enabling rapid prototyping** - Quickly test ideas with full API support

This system transforms your mobile backend development from a tedious, repetitive process into a fast, enjoyable experience focused on building great features rather than writing boilerplate code.

**Happy coding! 🚀**
