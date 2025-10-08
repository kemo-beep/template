package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"
	"github.com/gin-gonic/gin"
)

// SetupOrderRoutes sets up Order routes
func SetupOrderRoutes(r *gin.Engine, orderController *controllers.OrderController) {
	order := r.Group("/api/v1/order")
	order.Use(middleware.AuthMiddleware())
	{
		order.GET("/", orderController.GetOrderList)
		order.GET("/:id", orderController.GetOrder)
		order.POST("/", orderController.CreateOrder)
		order.PUT("/:id", orderController.UpdateOrder)
		order.DELETE("/:id", orderController.DeleteOrder)
	}
}
