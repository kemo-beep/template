package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupGeneratorRoutes sets up generator-related routes
func SetupGeneratorRoutes(r *gin.Engine, generatorController *controllers.GeneratorController) {
	// Generator routes (API generation) - protected routes
	api := r.Group("/api/v1")
	generator := api.Group("/schema")
	generator.Use(middleware.AuthMiddleware())
	{
		// Schema-based generation
		generator.POST("/generate", generatorController.GenerateFromSchema)
		generator.POST("/upload", generatorController.GenerateFromJSONFile)
		generator.POST("/migration", generatorController.GenerateFromMigration)
		generator.POST("/generate-all", generatorController.GenerateAllFromDatabase)

		// Management
		generator.GET("/models", generatorController.ListGeneratedModels)
		generator.DELETE("/cleanup/:model", generatorController.CleanupModel)

		// Utilities
		generator.GET("/template", generatorController.GetSchemaTemplate)
		generator.POST("/validate", generatorController.ValidateSchema)
		generator.GET("/migrations", generatorController.GetMigrationStatus)
	}
}
