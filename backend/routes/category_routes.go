package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"
	"github.com/gin-gonic/gin"
)

// SetupCategoryRoutes sets up Category routes
func SetupCategoryRoutes(r *gin.Engine, categoryController *controllers.CategoryController) {
	category := r.Group("/api/v1/category")
	category.Use(middleware.AuthMiddleware())
	{
		category.GET("/", categoryController.GetCategoryList)
		category.GET("/:id", categoryController.GetCategory)
		category.POST("/", categoryController.CreateCategory)
		category.PUT("/:id", categoryController.UpdateCategory)
		category.DELETE("/:id", categoryController.DeleteCategory)
	}
}
