package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupExampleRoutes(r *gin.Engine, exampleController *controllers.ExampleController) {
	// Example routes with comprehensive features
	examples := r.Group("/api/v1/examples")
	examples.Use(middleware.AuthMiddleware()) // Require authentication
	{
		examples.POST("/", exampleController.CreateExample)
		examples.GET("/", exampleController.GetExamples)
		examples.GET("/stats", exampleController.GetExampleStats)
		examples.GET("/:id", exampleController.GetExample)
		examples.PUT("/:id", exampleController.UpdateExample)
		examples.DELETE("/:id", exampleController.DeleteExample)
	}
}
