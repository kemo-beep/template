package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPaymentRoutes(router *gin.RouterGroup, paymentController *controllers.PaymentController) {
	// Product routes
	products := router.Group("/products")
	{
		products.POST("", middleware.AuthMiddleware(), paymentController.CreateProduct)
		products.GET("", middleware.AuthMiddleware(), paymentController.GetProducts)
		products.GET("/:id", middleware.AuthMiddleware(), paymentController.GetProduct)
		products.PUT("/:id", middleware.AuthMiddleware(), paymentController.UpdateProduct)
	}

	// Plan routes
	plans := router.Group("/plans")
	{
		plans.POST("", middleware.AuthMiddleware(), paymentController.CreatePlan)
		plans.GET("", middleware.AuthMiddleware(), paymentController.GetPlans)
	}

	// Payment routes
	payments := router.Group("/payments")
	{
		payments.POST("", middleware.AuthMiddleware(), paymentController.CreatePayment)
		payments.POST("/checkout", middleware.AuthMiddleware(), paymentController.CreateCheckoutSession)
		payments.GET("", middleware.AuthMiddleware(), paymentController.GetPayments)
	}

	// Subscription routes
	subscriptions := router.Group("/subscriptions")
	{
		subscriptions.POST("", middleware.AuthMiddleware(), paymentController.CreateSubscription)
		subscriptions.GET("", middleware.AuthMiddleware(), paymentController.GetSubscriptions)
		subscriptions.POST("/:id/cancel", middleware.AuthMiddleware(), paymentController.CancelSubscription)
	}

	// Webhook routes (no auth required)
	webhooks := router.Group("/webhooks")
	{
		webhooks.POST("/stripe", paymentController.StripeWebhook)
		webhooks.POST("/polar", paymentController.PolarWebhook)
	}
}
