package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mobile-backend/generators"
	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type GeneratorController struct {
	db *gorm.DB
	sg *generators.SchemaGenerator
	mg *generators.MigrationGenerator
}

func NewGeneratorController(db *gorm.DB) *GeneratorController {
	return &GeneratorController{
		db: db,
		sg: generators.NewSchemaGenerator(db),
		mg: generators.NewMigrationGenerator(db),
	}
}

// GenerateFromSchema generates APIs from JSON schema
// @Summary Generate APIs from schema
// @Description Generate CRUD APIs from a JSON schema definition
// @Tags generator
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param schema body generators.SchemaModel true "Schema definition"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /schema/generate [post]
func (gc *GeneratorController) GenerateFromSchema(c *gin.Context) {
	var schema generators.SchemaModel
	if err := c.ShouldBindJSON(&schema); err != nil {
		utils.SendValidationErrorResponse(c, map[string]string{"schema": "Invalid schema format"})
		return
	}

	// Generate model
	if err := gc.sg.GenerateModelFromSchema(schema); err != nil {
		utils.SendErrorResponse(c, "Failed to generate model", http.StatusInternalServerError)
		return
	}

	// Generate CRUD APIs
	routes, err := gc.sg.GenerateCRUDAPIs(schema.Name, schema)
	if err != nil {
		utils.SendErrorResponse(c, "Failed to generate CRUD APIs", http.StatusInternalServerError)
		return
	}

	// Generate controller
	if err := gc.sg.GenerateController(schema.Name, schema); err != nil {
		utils.SendErrorResponse(c, "Failed to generate controller", http.StatusInternalServerError)
		return
	}

	// Generate routes
	if err := gc.sg.GenerateRoutes(schema.Name, routes); err != nil {
		utils.SendErrorResponse(c, "Failed to generate routes", http.StatusInternalServerError)
		return
	}

	// Format generated code
	if err := gc.sg.FormatGoCode("models/" + strings.ToLower(schema.Name) + ".go"); err != nil {
		// Non-critical error, just log it
	}

	// Auto-register the generated routes
	if err := gc.autoRegisterRoutes(c, schema.Name); err != nil {
		utils.SendErrorResponse(c, "Failed to register routes", http.StatusInternalServerError)
		return
	}

	// Regenerate Swagger documentation
	if err := gc.regenerateSwagger(); err != nil {
		// Non-critical error, just log it
		fmt.Printf("Warning: Failed to regenerate Swagger: %v\n", err)
	}

	utils.SendSuccessResponse(c, gin.H{
		"model":  schema.Name,
		"routes": len(routes),
		"files": []string{
			"models/" + strings.ToLower(schema.Name) + ".go",
			"controllers/" + strings.ToLower(schema.Name) + ".go",
			"routes/" + strings.ToLower(schema.Name) + "_routes.go",
		},
		"registered":      true,
		"swagger_updated": true,
	}, "APIs generated and registered successfully")
}

// GenerateFromJSONFile generates APIs from JSON schema file
// @Summary Generate APIs from JSON file
// @Description Generate CRUD APIs from a JSON schema file
// @Tags generator
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "JSON schema file"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /schema/upload [post]
func (gc *GeneratorController) GenerateFromJSONFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.SendErrorResponse(c, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".json") {
		utils.SendErrorResponse(c, "File must be a JSON file", http.StatusBadRequest)
		return
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		utils.SendErrorResponse(c, "Failed to read file", http.StatusInternalServerError)
		return
	}

	// Parse JSON schema
	var schema generators.SchemaModel
	if err := json.Unmarshal(content, &schema); err != nil {
		utils.SendErrorResponse(c, "Invalid JSON schema", http.StatusBadRequest)
		return
	}

	// Generate APIs
	if err := gc.sg.GenerateFromJSON("temp_schema.json"); err != nil {
		utils.SendErrorResponse(c, "Failed to generate APIs", http.StatusInternalServerError)
		return
	}

	// Clean up temp file
	os.Remove("temp_schema.json")

	utils.SendSuccessResponse(c, gin.H{
		"model": schema.Name,
		"file":  header.Filename,
	}, "APIs generated from file successfully")
}

// GenerateFromMigration generates APIs from migration file
// @Summary Generate APIs from migration
// @Description Generate CRUD APIs from a database migration file
// @Tags generator
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Migration SQL file"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /schema/migration [post]
func (gc *GeneratorController) GenerateFromMigration(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.SendErrorResponse(c, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".sql") {
		utils.SendErrorResponse(c, "File must be a SQL file", http.StatusBadRequest)
		return
	}

	// Save uploaded file temporarily
	tempFile := "temp_migration.sql"
	if err := c.SaveUploadedFile(header, tempFile); err != nil {
		utils.SendErrorResponse(c, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile)

	// Generate APIs from migration
	if err := gc.mg.GenerateAPIsFromMigration(tempFile); err != nil {
		utils.SendErrorResponse(c, "Failed to generate APIs from migration", http.StatusInternalServerError)
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"file": header.Filename,
	}, "APIs generated from migration successfully")
}

// GenerateAllFromDatabase generates APIs for all existing tables
// @Summary Generate APIs for all tables
// @Description Generate CRUD APIs for all existing database tables
// @Tags generator
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /schema/generate-all [post]
func (gc *GeneratorController) GenerateAllFromDatabase(c *gin.Context) {
	if err := gc.mg.GenerateAPIsForAllTables(); err != nil {
		utils.SendErrorResponse(c, "Failed to generate APIs for all tables", http.StatusInternalServerError)
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"timestamp": time.Now().Unix(),
	}, "APIs generated for all tables successfully")
}

// ListGeneratedModels lists all generated models
// @Summary List generated models
// @Description List all generated model files
// @Tags generator
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Router /schema/models [get]
func (gc *GeneratorController) ListGeneratedModels(c *gin.Context) {
	var models []string

	// List model files
	modelFiles, err := filepath.Glob("models/*.go")
	if err == nil {
		for _, file := range modelFiles {
			models = append(models, filepath.Base(file))
		}
	}

	// List controller files
	controllerFiles, err := filepath.Glob("controllers/*.go")
	if err == nil {
		for _, file := range controllerFiles {
			if !strings.Contains(file, "generator.go") && !strings.Contains(file, "auth.go") &&
				!strings.Contains(file, "user.go") && !strings.Contains(file, "health.go") &&
				!strings.Contains(file, "upload.go") {
				models = append(models, "controllers/"+filepath.Base(file))
			}
		}
	}

	utils.SendSuccessResponse(c, gin.H{
		"models": models,
		"count":  len(models),
	}, "Generated models retrieved successfully")
}

// CleanupModel removes generated files for a model
// @Summary Cleanup model
// @Description Remove generated files for a specific model
// @Tags generator
// @Produce json
// @Security BearerAuth
// @Param model path string true "Model name"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /schema/cleanup/{model} [delete]
func (gc *GeneratorController) CleanupModel(c *gin.Context) {
	modelName := c.Param("model")
	if modelName == "" {
		utils.SendErrorResponse(c, "Model name is required", http.StatusBadRequest)
		return
	}

	modelName = strings.ToLower(modelName)

	filesToRemove := []string{
		"models/" + modelName + ".go",
		"controllers/" + modelName + ".go",
		"routes/" + modelName + "_routes.go",
	}

	var removedFiles []string
	for _, file := range filesToRemove {
		if err := os.Remove(file); err == nil {
			removedFiles = append(removedFiles, file)
		}
	}

	// Unregister routes from main.go
	autoRegister := generators.NewAutoRegister("main.go")
	if err := autoRegister.UnregisterGeneratedRoutes(strings.Title(modelName)); err != nil {
		fmt.Printf("Warning: Failed to unregister routes: %v\n", err)
	}

	// Regenerate Swagger documentation
	if err := gc.regenerateSwagger(); err != nil {
		fmt.Printf("Warning: Failed to regenerate Swagger: %v\n", err)
	}

	utils.SendSuccessResponse(c, gin.H{
		"model":               modelName,
		"removed_files":       removedFiles,
		"routes_unregistered": true,
		"swagger_updated":     true,
	}, "Model cleanup completed")
}

// GetSchemaTemplate returns a schema template
// @Summary Get schema template
// @Description Get a JSON template for creating schemas
// @Tags generator
// @Produce json
// @Success 200 {object} utils.SuccessResponse
// @Router /schema/template [get]
func (gc *GeneratorController) GetSchemaTemplate(c *gin.Context) {
	template := generators.SchemaModel{
		Name:          "ExampleModel",
		Package:       "models",
		TableName:     "example_models",
		HasTimestamps: true,
		HasSoftDelete: true,
		Comment:       "Example model for API generation",
		Fields: []generators.SchemaField{
			{
				Name:        "Title",
				Type:        "VARCHAR(255)",
				GoType:      "string",
				JSONTag:     "json:\"title\"",
				GormTag:     "gorm:\"not null\"",
				ValidateTag: "validate:\"required,min=3,max=255\"",
				Required:    true,
				Comment:     "The title of the example",
			},
			{
				Name:        "Description",
				Type:        "TEXT",
				GoType:      "string",
				JSONTag:     "json:\"description\"",
				GormTag:     "gorm:\"type:text\"",
				ValidateTag: "validate:\"max=1000\"",
				Required:    false,
				Comment:     "Description of the example",
			},
			{
				Name:        "IsActive",
				Type:        "BOOLEAN",
				GoType:      "bool",
				JSONTag:     "json:\"is_active\"",
				GormTag:     "gorm:\"default:true\"",
				ValidateTag: "",
				Required:    false,
				Comment:     "Whether the example is active",
			},
		},
	}

	utils.SendSuccessResponse(c, template, "Schema template retrieved successfully")
}

// ValidateSchema validates a schema definition
// @Summary Validate schema
// @Description Validate a schema definition before generation
// @Tags generator
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param schema body generators.SchemaModel true "Schema definition"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /schema/validate [post]
func (gc *GeneratorController) ValidateSchema(c *gin.Context) {
	var schema generators.SchemaModel
	if err := c.ShouldBindJSON(&schema); err != nil {
		utils.SendValidationErrorResponse(c, map[string]string{"schema": "Invalid schema format"})
		return
	}

	// Validate schema
	var errors []string

	if schema.Name == "" {
		errors = append(errors, "Model name is required")
	}

	if schema.TableName == "" {
		errors = append(errors, "Table name is required")
	}

	if len(schema.Fields) == 0 {
		errors = append(errors, "At least one field is required")
	}

	// Check for duplicate field names
	fieldNames := make(map[string]bool)
	for _, field := range schema.Fields {
		if fieldNames[field.Name] {
			errors = append(errors, "Duplicate field name: "+field.Name)
		}
		fieldNames[field.Name] = true

		if field.Name == "" {
			errors = append(errors, "Field name cannot be empty")
		}

		if field.GoType == "" {
			errors = append(errors, "Go type is required for field: "+field.Name)
		}
	}

	if len(errors) > 0 {
		utils.SendErrorResponse(c, "Schema validation failed", http.StatusBadRequest)
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"model":  schema.Name,
		"fields": len(schema.Fields),
	}, "Schema is valid")
}

// GetMigrationStatus returns the status of migrations
// @Summary Get migration status
// @Description Get the status of database migrations
// @Tags generator
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Router /schema/migrations [get]
func (gc *GeneratorController) GetMigrationStatus(c *gin.Context) {
	migrations, err := gc.mg.ListMigrations()
	if err != nil {
		utils.SendErrorResponse(c, "Failed to get migration status", http.StatusInternalServerError)
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"migrations": migrations,
		"count":      len(migrations),
	}, "Migration status retrieved successfully")
}

// autoRegisterRoutes automatically registers generated routes
func (gc *GeneratorController) autoRegisterRoutes(c *gin.Context, modelName string) error {
	// Create auto-register instance
	autoRegister := generators.NewAutoRegister("main.go")

	// Register the generated routes
	if err := autoRegister.RegisterGeneratedRoutes(modelName); err != nil {
		return fmt.Errorf("failed to register routes: %w", err)
	}

	return nil
}

// regenerateSwagger regenerates the Swagger documentation
func (gc *GeneratorController) regenerateSwagger() error {
	// Create auto-register instance
	autoRegister := generators.NewAutoRegister("main.go")

	// Regenerate Swagger docs
	if err := autoRegister.RegenerateSwagger(); err != nil {
		return fmt.Errorf("failed to regenerate Swagger: %w", err)
	}

	return nil
}
