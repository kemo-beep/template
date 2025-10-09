package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"mobile-backend/controllers"
	"mobile-backend/middleware"
	"mobile-backend/services"
)

// SetupSubscriptionManagementRoutes sets up subscription management routes
func SetupSubscriptionManagementRoutes(
	router *gin.RouterGroup,
	db *gorm.DB,
	subscriptionStatusService *services.SubscriptionStatusService,
	stripeService *services.StripeService,
	polarService *services.PolarService,
	subscriptionMiddleware *middleware.SubscriptionMiddleware,
	logger *zap.Logger,
) {
	// Initialize controller
	subscriptionController := controllers.NewSubscriptionManagementController(
		db,
		subscriptionStatusService,
		stripeService,
		polarService,
		logger,
	)

	// Subscription management routes
	subscription := router.Group("/subscription")
	subscription.Use(middleware.AuthMiddleware())
	{
		// Get user subscription status
		subscription.GET("/status", subscriptionController.GetUserSubscriptionStatus)

		// Get available subscription plans
		subscription.GET("/plans", subscriptionController.GetSubscriptionPlans)

		// Create new subscription
		subscription.POST("", subscriptionController.CreateSubscription)

		// Cancel subscription
		subscription.POST("/cancel", subscriptionController.CancelSubscription)

		// Get subscription history
		subscription.GET("/history", subscriptionController.GetSubscriptionHistory)

		// Admin routes
		admin := subscription.Group("/admin")
		// Note: Role-based middleware would need to be implemented
		{
			// Get subscription statistics
			admin.GET("/stats", subscriptionController.GetSubscriptionStats)
		}
	}

	// Pro-only routes (examples)
	pro := router.Group("/pro")
	pro.Use(middleware.AuthMiddleware())
	pro.Use(subscriptionMiddleware.RequireProSubscription())
	{
		// Example pro-only endpoint
		pro.GET("/features", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Welcome to Pro features!",
				"features": []string{
					"Advanced analytics",
					"Priority support",
					"Unlimited usage",
					"Custom integrations",
				},
			})
		})
	}

	// Active subscription routes (pro or trial)
	active := router.Group("/active")
	active.Use(middleware.AuthMiddleware())
	active.Use(subscriptionMiddleware.RequireActiveSubscription())
	{
		// Example active subscription endpoint
		active.GET("/dashboard", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message":             "Welcome to your dashboard!",
				"subscription_status": c.GetString("subscription_status"),
				"is_pro":              c.GetBool("is_pro"),
			})
		})
	}

	// Optional subscription routes (adds subscription info but doesn't require it)
	optional := router.Group("/optional")
	optional.Use(middleware.AuthMiddleware())
	optional.Use(subscriptionMiddleware.OptionalSubscription())
	{
		// Example optional subscription endpoint
		optional.GET("/info", func(c *gin.Context) {
			subscriptionStatus, _ := c.Get("subscription_status")
			isPro, _ := c.Get("is_pro")

			c.JSON(200, gin.H{
				"message":             "Optional subscription info",
				"subscription_status": subscriptionStatus,
				"is_pro":              isPro,
			})
		})
	}

	// Subscription limits example
	limits := router.Group("/limits")
	limits.Use(middleware.AuthMiddleware())
	limits.Use(subscriptionMiddleware.CheckSubscriptionLimits("api_calls", 1000))
	{
		// Example endpoint with subscription limits
		limits.GET("/usage", func(c *gin.Context) {
			limit, _ := c.Get("subscription_limit")
			subscriptionStatus, _ := c.Get("subscription_status")

			c.JSON(200, gin.H{
				"message":             "Usage information",
				"limit":               limit,
				"subscription_status": subscriptionStatus,
			})
		})
	}
}
