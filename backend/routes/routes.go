package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func SetupRoutes(r *gin.Engine, healthController *controllers.HealthController,
	authController *controllers.AuthController, userController *controllers.UserController,
	uploadController *controllers.UploadController, generatorController *controllers.GeneratorController,
	oauth2Controller *controllers.OAuth2Controller, cacheController *controllers.CacheController,
	paymentController *controllers.PaymentController, websocketController *controllers.WebSocketController) {

	// Health check routes
	r.GET("/health", healthController.HealthCheck)
	r.GET("/health/ready", healthController.ReadinessCheck)
	r.GET("/health/live", healthController.LivenessCheck)

	// API v1 routes
	api := r.Group("/api/v1")
	{
		// Public routes with rate limiting
		auth := api.Group("/auth")
		// Note: Rate limiting will be applied in main.go
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
			auth.POST("/refresh", authController.RefreshToken)

			// OAuth2 routes
			oauth2 := auth.Group("/oauth2")
			{
				oauth2.GET("/providers", oauth2Controller.OAuth2Providers)
				oauth2.GET("/:provider", oauth2Controller.OAuth2Login)
				oauth2.GET("/callback", oauth2Controller.OAuth2Callback)
			}
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Auth routes
			protected.POST("/auth/logout", authController.Logout)

			// User routes
			protected.GET("/profile", userController.GetProfile)
			protected.PUT("/profile", userController.UpdateProfile)
			protected.DELETE("/profile", userController.DeleteProfile)
			protected.GET("/users/:id", userController.GetUserByID)

			// Upload routes
			protected.POST("/upload", uploadController.UploadFile)
			protected.POST("/upload/multiple", uploadController.UploadMultipleFiles)
			protected.GET("/uploads/:filename", uploadController.GetFile)
			protected.DELETE("/uploads/:filename", uploadController.DeleteFile)

			// Cache management routes
			cache := protected.Group("/cache")
			{
				cache.GET("/stats", cacheController.GetCacheStats)
				cache.GET("/metrics", cacheController.GetCacheMetrics)
				cache.GET("/health", cacheController.GetCacheHealth)
				cache.GET("/recommendations", cacheController.GetCacheRecommendations)
				cache.POST("/metrics/reset", cacheController.ResetCacheMetrics)
				cache.GET("/key", cacheController.GetCacheKey)
				cache.POST("/key", cacheController.SetCacheKey)
				cache.POST("/invalidate", cacheController.InvalidateCache)
				cache.POST("/warm", cacheController.WarmCache)
				cache.POST("/clear", cacheController.ClearCache)
			}
		}

		// Payment routes (separate group for better organization)
		SetupPaymentRoutes(api, paymentController)
	}

	// Serve uploaded files
	r.Static("/uploads", "./uploads")
}

func SetupRoutesWithRateLimit(r *gin.Engine, healthController *controllers.HealthController,
	authController *controllers.AuthController, userController *controllers.UserController,
	uploadController *controllers.UploadController, redisClient *redis.Client) {

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// Health check routes (no rate limit)
	r.GET("/health", healthController.HealthCheck)
	r.GET("/health/ready", healthController.ReadinessCheck)
	r.GET("/health/live", healthController.LivenessCheck)

	// API v1 routes
	api := r.Group("/api/v1")
	{
		// Public routes with rate limiting
		auth := api.Group("/auth")
		auth.Use(rateLimiter.RateLimit(10, 1)) // 10 requests per minute
		{
			auth.POST("/register", authController.Register)
			auth.POST("/login", authController.Login)
		}

		// Protected routes with higher rate limit
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		protected.Use(rateLimiter.RateLimit(100, 1)) // 100 requests per minute
		{
			// Auth routes
			protected.POST("/auth/logout", authController.Logout)

			// User routes
			protected.GET("/profile", userController.GetProfile)
			protected.PUT("/profile", userController.UpdateProfile)
			protected.DELETE("/profile", userController.DeleteProfile)
			protected.GET("/users/:id", userController.GetUserByID)

			// Upload routes
			protected.POST("/upload", uploadController.UploadFile)
			protected.POST("/upload/multiple", uploadController.UploadMultipleFiles)
			protected.GET("/uploads/:filename", uploadController.GetFile)
			protected.DELETE("/uploads/:filename", uploadController.DeleteFile)
		}
	}

	// Serve uploaded files
	r.Static("/uploads", "./uploads")
}
