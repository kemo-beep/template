package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupProductWebhookRoutes(router *gin.RouterGroup, productWebhookController *controllers.ProductWebhookController) {
	// Webhook endpoints (no auth required for webhooks)
	webhookGroup := router.Group("/webhooks")
	{
		webhookGroup.POST("/stripe/products", productWebhookController.HandleStripeWebhook)
		webhookGroup.POST("/polar/products", productWebhookController.HandlePolarWebhook)
	}

	// Admin endpoints (require auth)
	adminGroup := router.Group("/admin/products")
	adminGroup.Use(middleware.AuthMiddleware())
	{
		// Sync statistics
		adminGroup.GET("/sync/stats", productWebhookController.GetProductSyncStats)

		// Manual sync operations
		adminGroup.POST("/sync/:provider/:external_id", productWebhookController.SyncProductFromProvider)

		// Product operations by external ID
		adminGroup.GET("/:provider/:external_id", productWebhookController.GetProductByExternalID)
		adminGroup.DELETE("/:provider/:external_id", productWebhookController.DeactivateProductByExternalID)

		// Plan operations by external ID
		adminGroup.GET("/plans/:provider/:external_id", productWebhookController.GetPlanByExternalID)
		adminGroup.DELETE("/plans/:provider/:external_id", productWebhookController.DeactivatePlanByExternalID)

		// List operations by provider
		adminGroup.GET("/:provider", productWebhookController.ListProductsByProvider)
		adminGroup.GET("/plans/:provider", productWebhookController.ListPlansByProvider)
	}
}
