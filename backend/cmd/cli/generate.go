package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code, models, controllers, and other components",
	Long: `Generate various components for your mobile backend:

Examples:
  mobile-backend-cli generate model User
  mobile-backend-cli generate controller UserController
  mobile-backend-cli generate service PaymentService
  mobile-backend-cli generate middleware AuthMiddleware
  mobile-backend-cli generate test UserControllerTest
  mobile-backend-cli generate migration AddUserTable
  mobile-backend-cli generate route UserRoutes
  mobile-backend-cli generate api UserAPI`,
}

var generateModelCmd = &cobra.Command{
	Use:   "model [name]",
	Short: "Generate a new model",
	Long:  `Generate a new model with basic CRUD operations and validation.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modelName := args[0]
		generateModel(modelName)
	},
}

var generateControllerCmd = &cobra.Command{
	Use:   "controller [name]",
	Short: "Generate a new controller",
	Long:  `Generate a new controller with CRUD operations and proper error handling.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		controllerName := args[0]
		generateController(controllerName)
	},
}

var generateServiceCmd = &cobra.Command{
	Use:   "service [name]",
	Short: "Generate a new service",
	Long:  `Generate a new service with business logic and database operations.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceName := args[0]
		generateService(serviceName)
	},
}

var generateMiddlewareCmd = &cobra.Command{
	Use:   "middleware [name]",
	Short: "Generate a new middleware",
	Long:  `Generate a new middleware with proper request/response handling.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		middlewareName := args[0]
		generateMiddleware(middlewareName)
	},
}

var generateTestCmd = &cobra.Command{
	Use:   "test [name]",
	Short: "Generate a new test file",
	Long:  `Generate a new test file with unit tests and integration tests.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		testName := args[0]
		generateTest(testName)
	},
}

var generateMigrationCmd = &cobra.Command{
	Use:   "migration [name]",
	Short: "Generate a new database migration",
	Long:  `Generate a new database migration file with proper SQL structure.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		migrationName := args[0]
		generateMigration(migrationName)
	},
}

var generateRouteCmd = &cobra.Command{
	Use:   "route [name]",
	Short: "Generate a new route file",
	Long:  `Generate a new route file with RESTful endpoints.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		routeName := args[0]
		generateRoute(routeName)
	},
}

var generateAPICmd = &cobra.Command{
	Use:   "api [name]",
	Short: "Generate a complete API module",
	Long:  `Generate a complete API module with model, controller, service, routes, and tests.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		apiName := args[0]
		generateAPI(apiName)
	},
}

func init() {
	generateCmd.AddCommand(generateModelCmd)
	generateCmd.AddCommand(generateControllerCmd)
	generateCmd.AddCommand(generateServiceCmd)
	generateCmd.AddCommand(generateMiddlewareCmd)
	generateCmd.AddCommand(generateTestCmd)
	generateCmd.AddCommand(generateMigrationCmd)
	generateCmd.AddCommand(generateRouteCmd)
	generateCmd.AddCommand(generateAPICmd)
}

// Model template
const modelTemplate = `package models

import (
	"time"
	"gorm.io/gorm"
)

// {{.Name}} represents a {{.LowerName}} entity
type {{.Name}} struct {
	ID        uint           ` + "`" + `json:"id" gorm:"primaryKey"` + "`" + `
	CreatedAt time.Time      ` + "`" + `json:"created_at"` + "`" + `
	UpdatedAt time.Time      ` + "`" + `json:"updated_at"` + "`" + `
	DeletedAt gorm.DeletedAt ` + "`" + `json:"deleted_at,omitempty" gorm:"index"` + "`" + `
	
	// Add your fields here
	// Name string ` + "`" + `json:"name" gorm:"not null"` + "`" + `
	// Email string ` + "`" + `json:"email" gorm:"uniqueIndex"` + "`" + `
}

// TableName returns the table name for {{.Name}}
func ({{.Name}}) TableName() string {
	return "{{.SnakeName}}"
}

// BeforeCreate hook
func ({{.ShortName}} *{{.Name}}) BeforeCreate(tx *gorm.DB) error {
	// Add any pre-creation logic here
	return nil
}

// BeforeUpdate hook
func ({{.ShortName}} *{{.Name}}) BeforeUpdate(tx *gorm.DB) error {
	// Add any pre-update logic here
	return nil
}

// Validate validates the {{.Name}} model
func ({{.ShortName}} *{{.Name}}) Validate() error {
	// Add validation logic here
	return nil
}
`

// Controller template
const controllerTemplate = `package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"{{.ModulePath}}/models"
	"{{.ModulePath}}/services"
	"{{.ModulePath}}/utils"
)

// {{.Name}}Controller handles {{.LowerName}} related requests
type {{.Name}}Controller struct {
	{{.LowerName}}Service *services.{{.Name}}Service
}

// New{{.Name}}Controller creates a new {{.Name}}Controller
func New{{.Name}}Controller({{.LowerName}}Service *services.{{.Name}}Service) *{{.Name}}Controller {
	return &{{.Name}}Controller{
		{{.LowerName}}Service: {{.LowerName}}Service,
	}
}

// Create{{.Name}} creates a new {{.LowerName}}
// @Summary Create {{.LowerName}}
// @Description Create a new {{.LowerName}}
// @Tags {{.LowerName}}
// @Accept json
// @Produce json
// @Param {{.LowerName}} body models.{{.Name}} true "{{.Name}} data"
// @Success 201 {object} utils.Response{data=models.{{.Name}}}
// @Failure 400 {object} utils.Response
// @Failure 500 {object} utils.Response
// @Router /{{.KebabName}} [post]
func (c *{{.Name}}Controller) Create{{.Name}}(ctx *gin.Context) {
	var {{.LowerName}} models.{{.Name}}
	if err := ctx.ShouldBindJSON(&{{.LowerName}}); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := {{.LowerName}}.Validate(); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err)
		return
	}

	created{{.Name}}, err := c.{{.LowerName}}Service.Create{{.Name}}(&{{.LowerName}})
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create {{.LowerName}}", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusCreated, "{{.Name}} created successfully", created{{.Name}})
}

// Get{{.Name}} retrieves a {{.LowerName}} by ID
// @Summary Get {{.LowerName}}
// @Description Get a {{.LowerName}} by ID
// @Tags {{.LowerName}}
// @Produce json
// @Param id path int true "{{.Name}} ID"
// @Success 200 {object} utils.Response{data=models.{{.Name}}}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /{{.KebabName}}/{id} [get]
func (c *{{.Name}}Controller) Get{{.Name}}(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	{{.LowerName}}, err := c.{{.LowerName}}Service.Get{{.Name}}ByID(uint(id))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, "{{.Name}} not found", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "{{.Name}} retrieved successfully", {{.LowerName}})
}

// Update{{.Name}} updates a {{.LowerName}}
// @Summary Update {{.LowerName}}
// @Description Update a {{.LowerName}} by ID
// @Tags {{.LowerName}}
// @Accept json
// @Produce json
// @Param id path int true "{{.Name}} ID"
// @Param {{.LowerName}} body models.{{.Name}} true "{{.Name}} data"
// @Success 200 {object} utils.Response{data=models.{{.Name}}}
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /{{.KebabName}}/{id} [put]
func (c *{{.Name}}Controller) Update{{.Name}}(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	var {{.LowerName}} models.{{.Name}}
	if err := ctx.ShouldBindJSON(&{{.LowerName}}); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	{{.LowerName}}.ID = uint(id)
	if err := {{.LowerName}}.Validate(); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err)
		return
	}

	updated{{.Name}}, err := c.{{.LowerName}}Service.Update{{.Name}}(&{{.LowerName}})
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update {{.LowerName}}", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "{{.Name}} updated successfully", updated{{.Name}})
}

// Delete{{.Name}} deletes a {{.LowerName}}
// @Summary Delete {{.LowerName}}
// @Description Delete a {{.LowerName}} by ID
// @Tags {{.LowerName}}
// @Produce json
// @Param id path int true "{{.Name}} ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /{{.KebabName}}/{id} [delete]
func (c *{{.Name}}Controller) Delete{{.Name}}(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	err = c.{{.LowerName}}Service.Delete{{.Name}}(uint(id))
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete {{.LowerName}}", err)
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "{{.Name}} deleted successfully", nil)
}

// List{{.Name}}s retrieves all {{.LowerName}}s with pagination
// @Summary List {{.LowerName}}s
// @Description Get a list of {{.LowerName}}s with pagination
// @Tags {{.LowerName}}
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.Response{data=[]models.{{.Name}}}
// @Failure 400 {object} utils.Response
// @Router /{{.KebabName}} [get]
func (c *{{.Name}}Controller) List{{.Name}}s(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	{{.LowerName}}s, total, err := c.{{.LowerName}}Service.List{{.Name}}s(page, limit)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve {{.LowerName}}s", err)
		return
	}

	response := map[string]interface{}{
		"data":  {{.LowerName}}s,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	utils.SuccessResponse(ctx, http.StatusOK, "{{.Name}}s retrieved successfully", response)
}
`

// Service template
const serviceTemplate = `package services

import (
	"errors"
	"{{.ModulePath}}/models"
	"gorm.io/gorm"
)

// {{.Name}}Service handles {{.LowerName}} business logic
type {{.Name}}Service struct {
	db *gorm.DB
}

// New{{.Name}}Service creates a new {{.Name}}Service
func New{{.Name}}Service(db *gorm.DB) *{{.Name}}Service {
	return &{{.Name}}Service{
		db: db,
	}
}

// Create{{.Name}} creates a new {{.LowerName}}
func (s *{{.Name}}Service) Create{{.Name}}({{.LowerName}} *models.{{.Name}}) (*models.{{.Name}}, error) {
	if err := s.db.Create({{.LowerName}}).Error; err != nil {
		return nil, err
	}
	return {{.LowerName}}, nil
}

// Get{{.Name}}ByID retrieves a {{.LowerName}} by ID
func (s *{{.Name}}Service) Get{{.Name}}ByID(id uint) (*models.{{.Name}}, error) {
	var {{.LowerName}} models.{{.Name}}
	if err := s.db.First(&{{.LowerName}}, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("{{.LowerName}} not found")
		}
		return nil, err
	}
	return &{{.LowerName}}, nil
}

// Update{{.Name}} updates a {{.LowerName}}
func (s *{{.Name}}Service) Update{{.Name}}({{.LowerName}} *models.{{.Name}}) (*models.{{.Name}}, error) {
	if err := s.db.Save({{.LowerName}}).Error; err != nil {
		return nil, err
	}
	return {{.LowerName}}, nil
}

// Delete{{.Name}} deletes a {{.LowerName}}
func (s *{{.Name}}Service) Delete{{.Name}}(id uint) error {
	if err := s.db.Delete(&models.{{.Name}}{}, id).Error; err != nil {
		return err
	}
	return nil
}

// List{{.Name}}s retrieves all {{.LowerName}}s with pagination
func (s *{{.Name}}Service) List{{.Name}}s(page, limit int) ([]models.{{.Name}}, int64, error) {
	var {{.LowerName}}s []models.{{.Name}}
	var total int64

	offset := (page - 1) * limit

	if err := s.db.Model(&models.{{.Name}}{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.Offset(offset).Limit(limit).Find(&{{.LowerName}}s).Error; err != nil {
		return nil, 0, err
	}

	return {{.LowerName}}s, total, nil
}
`

// Test template
const testTemplate = `package {{.PackageName}}

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"{{.ModulePath}}/models"
	"{{.ModulePath}}/services"
)

// Mock{{.Name}}Service is a mock implementation of {{.Name}}Service
type Mock{{.Name}}Service struct {
	mock.Mock
}

func (m *Mock{{.Name}}Service) Create{{.Name}}({{.LowerName}} *models.{{.Name}}) (*models.{{.Name}}, error) {
	args := m.Called({{.LowerName}})
	return args.Get(0).(*models.{{.Name}}), args.Error(1)
}

func (m *Mock{{.Name}}Service) Get{{.Name}}ByID(id uint) (*models.{{.Name}}, error) {
	args := m.Called(id)
	return args.Get(0).(*models.{{.Name}}), args.Error(1)
}

func (m *Mock{{.Name}}Service) Update{{.Name}}({{.LowerName}} *models.{{.Name}}) (*models.{{.Name}}, error) {
	args := m.Called({{.LowerName}})
	return args.Get(0).(*models.{{.Name}}), args.Error(1)
}

func (m *Mock{{.Name}}Service) Delete{{.Name}}(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *Mock{{.Name}}Service) List{{.Name}}s(page, limit int) ([]models.{{.Name}}, int64, error) {
	args := m.Called(page, limit)
	return args.Get(0).([]models.{{.Name}}), args.Get(1).(int64), args.Error(2)
}

func Test{{.Name}}Controller_Create{{.Name}}(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*Mock{{.Name}}Service)
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful creation",
			requestBody: models.{{.Name}}{
				// Add test fields here
			},
			mockSetup: func(m *Mock{{.Name}}Service) {
				m.On("Create{{.Name}}", mock.AnythingOfType("*models.{{.Name}}")).Return(&models.{{.Name}}{ID: 1}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		// Add more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(Mock{{.Name}}Service)
			tt.mockSetup(mockService)

			controller := New{{.Name}}Controller(mockService)

			router := gin.New()
			router.POST("/{{.KebabName}}", controller.Create{{.Name}})

			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/{{.KebabName}}", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

// Add more test functions for other controller methods
`

// Migration template
const migrationTemplate = `-- Migration: {{.Name}}
-- Created: {{.Timestamp}}

-- Up migration
CREATE TABLE IF NOT EXISTS {{.SnakeName}} (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    
    -- Add your columns here
    -- name VARCHAR(255) NOT NULL,
    -- email VARCHAR(255) UNIQUE NOT NULL,
    -- description TEXT,
    -- is_active BOOLEAN DEFAULT true
);

-- Create indexes
-- CREATE INDEX IF NOT EXISTS idx_{{.SnakeName}}_email ON {{.SnakeName}}(email);
-- CREATE INDEX IF NOT EXISTS idx_{{.SnakeName}}_created_at ON {{.SnakeName}}(created_at);

-- Down migration
-- DROP TABLE IF EXISTS {{.SnakeName}};
`

// Route template
const routeTemplate = `package routes

import (
	"{{.ModulePath}}/controllers"
	"{{.ModulePath}}/middleware"
	"{{.ModulePath}}/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Setup{{.Name}}Routes sets up {{.LowerName}} routes
func Setup{{.Name}}Routes(router *gin.RouterGroup, db *gorm.DB) {
	{{.LowerName}}Service := services.New{{.Name}}Service(db)
	{{.LowerName}}Controller := controllers.New{{.Name}}Controller({{.LowerName}}Service)

	{{.LowerName}}Routes := router.Group("/{{.KebabName}}")
	{
		{{.LowerName}}Routes.POST("", {{.LowerName}}Controller.Create{{.Name}})
		{{.LowerName}}Routes.GET("", {{.LowerName}}Controller.List{{.Name}}s)
		{{.LowerName}}Routes.GET("/:id", {{.LowerName}}Controller.Get{{.Name}})
		{{.LowerName}}Routes.PUT("/:id", {{.LowerName}}Controller.Update{{.Name}})
		{{.LowerName}}Routes.DELETE("/:id", {{.LowerName}}Controller.Delete{{.Name}})
	}
}
`

// Template data structure
type TemplateData struct {
	Name        string
	LowerName   string
	SnakeName   string
	KebabName   string
	ShortName   string
	ModulePath  string
	PackageName string
	Timestamp   string
}

// Helper functions
func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func toKebabCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func getModulePath() string {
	// Try to read go.mod to get module path
	if data, err := os.ReadFile("go.mod"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "module ") {
				return strings.TrimSpace(strings.TrimPrefix(line, "module "))
			}
		}
	}
	return "github.com/your-org/your-project"
}

func createTemplateData(name string) TemplateData {
	return TemplateData{
		Name:        name,
		LowerName:   strings.ToLower(name),
		SnakeName:   toSnakeCase(name),
		KebabName:   toKebabCase(name),
		ShortName:   strings.ToLower(name[:1]),
		ModulePath:  getModulePath(),
		PackageName: strings.ToLower(name),
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}
}

func generateModel(name string) {
	data := createTemplateData(name)

	// Create models directory if it doesn't exist
	os.MkdirAll("models", 0755)

	// Generate model file
	generateFile("models/"+strings.ToLower(name)+".go", modelTemplate, data)
	fmt.Printf("✅ Generated model: models/%s.go\n", strings.ToLower(name))
}

func generateController(name string) {
	data := createTemplateData(name)

	// Create controllers directory if it doesn't exist
	os.MkdirAll("controllers", 0755)

	// Generate controller file
	generateFile("controllers/"+strings.ToLower(name)+".go", controllerTemplate, data)
	fmt.Printf("✅ Generated controller: controllers/%s.go\n", strings.ToLower(name))
}

func generateService(name string) {
	data := createTemplateData(name)

	// Create services directory if it doesn't exist
	os.MkdirAll("services", 0755)

	// Generate service file
	generateFile("services/"+strings.ToLower(name)+".go", serviceTemplate, data)
	fmt.Printf("✅ Generated service: services/%s.go\n", strings.ToLower(name))
}

func generateMiddleware(name string) {
	// Create middleware directory if it doesn't exist
	os.MkdirAll("middleware", 0755)

	// Generate middleware file
	middlewareTemplate := `package middleware

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// {{.Name}} middleware
func {{.Name}}() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add your middleware logic here
		
		// Example: Log request
		// log.Printf("Request: %s %s", c.Request.Method, c.Request.URL.Path)
		
		// Continue to next handler
		c.Next()
	}
}
`

	data := createTemplateData(name)
	generateFile("middleware/"+strings.ToLower(name)+".go", middlewareTemplate, data)
	fmt.Printf("✅ Generated middleware: middleware/%s.go\n", strings.ToLower(name))
}

func generateTest(name string) {
	data := createTemplateData(name)

	// Create tests directory if it doesn't exist
	os.MkdirAll("tests/unit", 0755)

	// Generate test file
	generateFile("tests/unit/"+strings.ToLower(name)+"_test.go", testTemplate, data)
	fmt.Printf("✅ Generated test: tests/unit/%s_test.go\n", strings.ToLower(name))
}

func generateMigration(name string) {
	data := createTemplateData(name)

	// Create migrations directory if it doesn't exist
	os.MkdirAll("migrations", 0755)

	// Generate migration file with timestamp
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("migrations/%s_%s.sql", timestamp, data.SnakeName)
	generateFile(filename, migrationTemplate, data)
	fmt.Printf("✅ Generated migration: %s\n", filename)
}

func generateRoute(name string) {
	data := createTemplateData(name)

	// Create routes directory if it doesn't exist
	os.MkdirAll("routes", 0755)

	// Generate route file
	generateFile("routes/"+strings.ToLower(name)+"_routes.go", routeTemplate, data)
	fmt.Printf("✅ Generated route: routes/%s_routes.go\n", strings.ToLower(name))
}

func generateAPI(name string) {
	// Generate all components for a complete API
	generateModel(name)
	generateService(name)
	generateController(name)
	generateRoute(name)
	generateTest(name)
	generateMigration(name)

	fmt.Printf("🎉 Generated complete API module for: %s\n", name)
	fmt.Println("📝 Next steps:")
	fmt.Println("   1. Update the model with your specific fields")
	fmt.Println("   2. Add the route to your main router")
	fmt.Println("   3. Run database migrations")
	fmt.Println("   4. Test your API endpoints")
}

func generateFile(filename, templateStr string, data TemplateData) {
	tmpl, err := template.New("").Parse(templateStr)
	if err != nil {
		log.Fatalf("Error parsing template: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		log.Fatalf("Error executing template: %v", err)
	}
}
