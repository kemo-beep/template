package generators

import (
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gorm.io/gorm"
)

// SchemaField represents a field in a database schema
type SchemaField struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	GoType      string `json:"go_type"`
	JSONTag     string `json:"json_tag"`
	GormTag     string `json:"gorm_tag"`
	ValidateTag string `json:"validate_tag"`
	Required    bool   `json:"required"`
	Unique      bool   `json:"unique"`
	Index       bool   `json:"index"`
	Comment     string `json:"comment"`
}

// SchemaModel represents a complete model schema
type SchemaModel struct {
	Name          string        `json:"name"`
	Package       string        `json:"package"`
	TableName     string        `json:"table_name"`
	Fields        []SchemaField `json:"fields"`
	HasTimestamps bool          `json:"has_timestamps"`
	HasSoftDelete bool          `json:"has_soft_delete"`
	Comment       string        `json:"comment"`
}

// APIRoute represents a generated API route
type APIRoute struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Handler     string `json:"handler"`
	Description string `json:"description"`
	Auth        bool   `json:"auth"`
	Validation  bool   `json:"validation"`
}

// SchemaGenerator handles schema-based code generation
type SchemaGenerator struct {
	db *gorm.DB
}

// NewSchemaGenerator creates a new schema generator
func NewSchemaGenerator(db *gorm.DB) *SchemaGenerator {
	return &SchemaGenerator{db: db}
}

// GenerateModelFromSchema generates a Go model from schema definition
func (sg *SchemaGenerator) GenerateModelFromSchema(schema SchemaModel) error {
	// Create models directory if it doesn't exist
	modelsDir := "models"
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return err
	}

	// Generate model file
	modelFile := filepath.Join(modelsDir, strings.ToLower(schema.Name)+".go")
	return sg.generateModelFile(modelFile, schema)
}

// GenerateCRUDAPIs generates CRUD API routes for a model
func (sg *SchemaGenerator) GenerateCRUDAPIs(modelName string, schema SchemaModel) ([]APIRoute, error) {
	routes := []APIRoute{
		{
			Method:      "GET",
			Path:        fmt.Sprintf("/api/v1/%s", strings.ToLower(modelName)),
			Handler:     fmt.Sprintf("Get%sList", modelName),
			Description: fmt.Sprintf("Get list of %s", strings.ToLower(modelName)),
			Auth:        true,
			Validation:  false,
		},
		{
			Method:      "GET",
			Path:        fmt.Sprintf("/api/v1/%s/:id", strings.ToLower(modelName)),
			Handler:     fmt.Sprintf("Get%s", modelName),
			Description: fmt.Sprintf("Get %s by ID", strings.ToLower(modelName)),
			Auth:        true,
			Validation:  true,
		},
		{
			Method:      "POST",
			Path:        fmt.Sprintf("/api/v1/%s", strings.ToLower(modelName)),
			Handler:     fmt.Sprintf("Create%s", modelName),
			Description: fmt.Sprintf("Create new %s", strings.ToLower(modelName)),
			Auth:        true,
			Validation:  true,
		},
		{
			Method:      "PUT",
			Path:        fmt.Sprintf("/api/v1/%s/:id", strings.ToLower(modelName)),
			Handler:     fmt.Sprintf("Update%s", modelName),
			Description: fmt.Sprintf("Update %s by ID", strings.ToLower(modelName)),
			Auth:        true,
			Validation:  true,
		},
		{
			Method:      "DELETE",
			Path:        fmt.Sprintf("/api/v1/%s/:id", strings.ToLower(modelName)),
			Handler:     fmt.Sprintf("Delete%s", modelName),
			Description: fmt.Sprintf("Delete %s by ID", strings.ToLower(modelName)),
			Auth:        true,
			Validation:  true,
		},
	}

	return routes, nil
}

// GenerateController generates a controller for a model
func (sg *SchemaGenerator) GenerateController(modelName string, schema SchemaModel) error {
	controllersDir := "controllers"
	if err := os.MkdirAll(controllersDir, 0755); err != nil {
		return err
	}

	controllerFile := filepath.Join(controllersDir, strings.ToLower(modelName)+".go")
	return sg.generateControllerFile(controllerFile, modelName, schema)
}

// GenerateRoutes generates route definitions
func (sg *SchemaGenerator) GenerateRoutes(modelName string, routes []APIRoute) error {
	routesDir := "routes"
	if err := os.MkdirAll(routesDir, 0755); err != nil {
		return err
	}

	routesFile := filepath.Join(routesDir, strings.ToLower(modelName)+"_routes.go")
	return sg.generateRoutesFile(routesFile, modelName, routes)
}

// generateModelFile generates the model file
func (sg *SchemaGenerator) generateModelFile(filename string, schema SchemaModel) error {
	tmpl := `package models

import (
	"time"
	"gorm.io/gorm"
)

// {{.Name}} represents a {{.Comment}}
type {{.Name}} struct {
	{{if .HasTimestamps}}ID        uint           ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	CreatedAt time.Time      ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time      ` + "`json:\"updated_at\"`" + `{{end}}
	{{if .HasSoftDelete}}DeletedAt gorm.DeletedAt ` + "`json:\"-\" gorm:\"index\"`" + `{{end}}
	{{range .Fields}}
	{{.Name}} {{.GoType}} ` + "`{{.JSONTag}} {{.GormTag}}`" + `{{if .Comment}} // {{.Comment}}{{end}}{{end}}
}

// TableName returns the table name for {{.Name}}
func ({{.Name}}) TableName() string {
	return "{{.TableName}}"
}
`

	t, err := template.New("model").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, schema)
}

// generateControllerFile generates the controller file
func (sg *SchemaGenerator) generateControllerFile(filename string, modelName string, schema SchemaModel) error {
	tmpl := `package controllers

import (
	"net/http"
	"strconv"
	"mobile-backend/models"
	"mobile-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type {{.ModelName}}Controller struct {
	db *gorm.DB
}

func New{{.ModelName}}Controller(db *gorm.DB) *{{.ModelName}}Controller {
	return &{{.ModelName}}Controller{db: db}
}

// Get{{.ModelName}}List retrieves all {{.ModelName}} records
// @Summary Get {{.ModelName}} list
// @Description Get list of all {{.ModelName}} records
// @Tags {{.ModelName}}
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /{{.ModelNameLower}} [get]
func (c *{{.ModelName}}Controller) Get{{.ModelName}}List(ctx *gin.Context) {
	var {{.ModelNameLower}}s []models.{{.ModelName}}
	
	if err := c.db.Find(&{{.ModelNameLower}}s).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to fetch {{.ModelNameLower}}s", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, "{{.ModelName}}s retrieved successfully", {{.ModelNameLower}}s)
}

// Get{{.ModelName}} retrieves a {{.ModelName}} by ID
// @Summary Get {{.ModelName}} by ID
// @Description Get {{.ModelName}} record by ID
// @Tags {{.ModelName}}
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "{{.ModelName}} ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /{{.ModelNameLower}}/{id} [get]
func (c *{{.ModelName}}Controller) Get{{.ModelName}}(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var {{.ModelNameLower}} models.{{.ModelName}}
	if err := c.db.First(&{{.ModelNameLower}}, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFoundResponse(ctx, "{{.ModelName}} not found")
			return
		}
		utils.SendErrorResponse(ctx, "Failed to fetch {{.ModelNameLower}}", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, "{{.ModelName}} retrieved successfully", {{.ModelNameLower}})
}

// Create{{.ModelName}} creates a new {{.ModelName}}
// @Summary Create {{.ModelName}}
// @Description Create a new {{.ModelName}} record
// @Tags {{.ModelName}}
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param {{.ModelNameLower}} body models.{{.ModelName}} true "{{.ModelName}} data"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /{{.ModelNameLower}} [post]
func (c *{{.ModelName}}Controller) Create{{.ModelName}}(ctx *gin.Context) {
	var {{.ModelNameLower}} models.{{.ModelName}}
	
	if err := ctx.ShouldBindJSON(&{{.ModelNameLower}}); err != nil {
		utils.SendValidationErrorResponse(ctx, "Invalid input", err)
		return
	}
	
	if err := c.db.Create(&{{.ModelNameLower}}).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to create {{.ModelNameLower}}", http.StatusInternalServerError)
		return
	}
	
	utils.SendCreatedResponse(ctx, "{{.ModelName}} created successfully", {{.ModelNameLower}})
}

// Update{{.ModelName}} updates a {{.ModelName}} by ID
// @Summary Update {{.ModelName}}
// @Description Update {{.ModelName}} record by ID
// @Tags {{.ModelName}}
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "{{.ModelName}} ID"
// @Param {{.ModelNameLower}} body models.{{.ModelName}} true "{{.ModelName}} data"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /{{.ModelNameLower}}/{id} [put]
func (c *{{.ModelName}}Controller) Update{{.ModelName}}(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var {{.ModelNameLower}} models.{{.ModelName}}
	if err := c.db.First(&{{.ModelNameLower}}, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFoundResponse(ctx, "{{.ModelName}} not found")
			return
		}
		utils.SendErrorResponse(ctx, "Failed to fetch {{.ModelNameLower}}", http.StatusInternalServerError)
		return
	}
	
	if err := ctx.ShouldBindJSON(&{{.ModelNameLower}}); err != nil {
		utils.SendValidationErrorResponse(ctx, "Invalid input", err)
		return
	}
	
	if err := c.db.Save(&{{.ModelNameLower}}).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to update {{.ModelNameLower}}", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, "{{.ModelName}} updated successfully", {{.ModelNameLower}})
}

// Delete{{.ModelName}} deletes a {{.ModelName}} by ID
// @Summary Delete {{.ModelName}}
// @Description Delete {{.ModelName}} record by ID
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /{{.ModelNameLower}}/{id} [delete]
func (c *{{.ModelName}}Controller) Delete{{.ModelName}}(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	if err := c.db.Delete(&models.{{.ModelName}}{}, uint(id)).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to delete {{.ModelNameLower}}", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, "{{.ModelName}} deleted successfully", nil)
}
`

	data := struct {
		ModelName      string
		ModelNameLower string
	}{
		ModelName:      modelName,
		ModelNameLower: strings.ToLower(modelName),
	}

	t, err := template.New("controller").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

// generateRoutesFile generates the routes file
func (sg *SchemaGenerator) generateRoutesFile(filename string, modelName string, routes []APIRoute) error {
	tmpl := `package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"
	"github.com/gin-gonic/gin"
)

// Setup{{.ModelName}}Routes sets up {{.ModelName}} routes
func Setup{{.ModelName}}Routes(r *gin.Engine, {{.ModelNameLower}}Controller *controllers.{{.ModelName}}Controller) {
	{{.ModelNameLower}} := r.Group("/api/v1/{{.ModelNameLower}}")
	{{.ModelNameLower}}.Use(middleware.AuthMiddleware())
	{
		{{range .Routes}}
		{{.Method}}("{{.Path}}", {{.Handler}})
		{{end}}
	}
}
`

	data := struct {
		ModelName      string
		ModelNameLower string
		Routes         []APIRoute
	}{
		ModelName:      modelName,
		ModelNameLower: strings.ToLower(modelName),
		Routes:         routes,
	}

	t, err := template.New("routes").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, data)
}

// GenerateFromJSON generates code from a JSON schema file
func (sg *SchemaGenerator) GenerateFromJSON(schemaFile string) error {
	data, err := os.ReadFile(schemaFile)
	if err != nil {
		return err
	}

	var schema SchemaModel
	if err := json.Unmarshal(data, &schema); err != nil {
		return err
	}

	// Generate model
	if err := sg.GenerateModelFromSchema(schema); err != nil {
		return err
	}

	// Generate CRUD APIs
	routes, err := sg.GenerateCRUDAPIs(schema.Name, schema)
	if err != nil {
		return err
	}

	// Generate controller
	if err := sg.GenerateController(schema.Name, schema); err != nil {
		return err
	}

	// Generate routes
	if err := sg.GenerateRoutes(schema.Name, routes); err != nil {
		return err
	}

	return nil
}

// GenerateFromSchema generates code from a schema model
func (sg *SchemaGenerator) GenerateFromSchema(schema SchemaModel) error {
	// Generate model
	if err := sg.GenerateModelFromSchema(schema); err != nil {
		return err
	}

	// Generate CRUD APIs
	routes, err := sg.GenerateCRUDAPIs(schema.Name, schema)
	if err != nil {
		return err
	}

	// Generate controller
	if err := sg.GenerateController(schema.Name, schema); err != nil {
		return err
	}

	// Generate routes
	if err := sg.GenerateRoutes(schema.Name, routes); err != nil {
		return err
	}

	return nil
}

// FormatGoCode formats Go code using gofmt
func (sg *SchemaGenerator) FormatGoCode(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	formatted, err := format.Source(data)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, formatted, 0644)
}
