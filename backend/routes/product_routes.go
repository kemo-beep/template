package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"
	"github.com/gin-gonic/gin"
)

// SetupProductRoutes sets up Product routes
func SetupProductRoutes(r *gin.Engine, productController *controllers.ProductController) {
	product := r.Group("/api/v1/product")
	product.Use(middleware.AuthMiddleware())
	{
		product.GET("/", productController.GetProductList)
		product.GET("/:id", productController.GetProduct)
		product.POST("/", productController.CreateProduct)
		product.PUT("/:id", productController.UpdateProduct)
		product.DELETE("/:id", productController.DeleteProduct)
	}
}
