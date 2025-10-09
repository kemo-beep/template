package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupPushNotificationRoutes sets up push notification routes
func SetupPushNotificationRoutes(router *gin.Engine, pushController *controllers.PushNotificationController) {
	// Create a group for push notification routes with authentication middleware
	pushGroup := router.Group("/api/v1/notifications")
	pushGroup.Use(middleware.AuthMiddleware())

	// Notification management
	pushGroup.POST("/send", pushController.SendNotification)
	pushGroup.GET("/", pushController.GetNotifications)
	pushGroup.GET("/:id/analytics", pushController.GetNotificationAnalytics)
	pushGroup.POST("/:id/opened", pushController.MarkAsOpened)

	// Template management
	templates := pushGroup.Group("/templates")
	{
		templates.POST("/", pushController.CreateTemplate)
		templates.GET("/", pushController.GetTemplates)
	}

	// Segment management
	segments := pushGroup.Group("/segments")
	{
		segments.POST("/", pushController.CreateSegment)
		segments.GET("/", pushController.GetSegments)
	}

	// Device management
	devices := pushGroup.Group("/devices")
	{
		devices.POST("/register", pushController.RegisterDevice)
		devices.POST("/unregister", pushController.UnregisterDevice)
		devices.GET("/", pushController.GetDeviceTokens)
	}
}
