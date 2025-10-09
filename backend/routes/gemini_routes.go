package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SetupGeminiRoutes sets up all Gemini AI related routes
func SetupGeminiRoutes(r *gin.Engine, geminiController *controllers.GeminiController, logger *zap.Logger) {
	// Public routes (no authentication required)
	gemini := r.Group("/api/v1/gemini")
	{
		// Health check and models (public)
		gemini.GET("/health", geminiController.HealthCheck)
		gemini.GET("/models", geminiController.GetAvailableModels)
	}

	// Protected routes (authentication required)
	protectedGemini := r.Group("/api/v1/gemini")
	protectedGemini.Use(middleware.AuthMiddleware())
	{
		// Text generation
		protectedGemini.POST("/generate", geminiController.GenerateText)

		// Conversation management
		conversations := protectedGemini.Group("/conversations")
		{
			conversations.POST("", geminiController.CreateConversation)
			conversations.GET("", geminiController.ListConversations)
			conversations.GET("/:conversation_id", geminiController.GetConversation)
			conversations.DELETE("/:conversation_id", geminiController.DeleteConversation)

			// Context-aware text generation
			conversations.POST("/:conversation_id/generate", geminiController.GenerateTextWithContext)

			// Message management
			conversations.POST("/:conversation_id/messages", geminiController.AddMessage)
		}

		// Service statistics (admin/authenticated users)
		protectedGemini.GET("/stats", geminiController.GetServiceStats)
	}
}

// SetupGeminiRoutesWithRateLimit sets up Gemini routes with rate limiting
func SetupGeminiRoutesWithRateLimit(r *gin.Engine, geminiController *controllers.GeminiController, rateLimiter *middleware.RateLimiter, logger *zap.Logger) {
	// Public routes (no authentication required)
	gemini := r.Group("/api/v1/gemini")
	{
		// Health check and models (public)
		gemini.GET("/health", geminiController.HealthCheck)
		gemini.GET("/models", geminiController.GetAvailableModels)
	}

	// Protected routes with rate limiting
	protectedGemini := r.Group("/api/v1/gemini")
	protectedGemini.Use(middleware.AuthMiddleware())

	// Apply different rate limits for different operations
	{
		// Text generation with stricter rate limiting
		textGen := protectedGemini.Group("/")
		textGen.Use(rateLimiter.APIRateLimit()) // Use existing API rate limit
		{
			textGen.POST("/generate", geminiController.GenerateText)
		}

		// Conversation management with standard rate limiting
		conversations := protectedGemini.Group("/conversations")
		conversations.Use(rateLimiter.APIRateLimit())
		{
			conversations.POST("", geminiController.CreateConversation)
			conversations.GET("", geminiController.ListConversations)
			conversations.GET("/:conversation_id", geminiController.GetConversation)
			conversations.DELETE("/:conversation_id", geminiController.DeleteConversation)

			// Context-aware text generation
			conversations.POST("/:conversation_id/generate", geminiController.GenerateTextWithContext)

			// Message management
			conversations.POST("/:conversation_id/messages", geminiController.AddMessage)
		}

		// Service statistics (admin/authenticated users)
		stats := protectedGemini.Group("/")
		stats.Use(rateLimiter.APIRateLimit())
		{
			stats.GET("/stats", geminiController.GetServiceStats)
		}
	}
}
