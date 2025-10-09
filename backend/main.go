// @title Mobile Backend API
// @version 1.0
// @description A comprehensive mobile backend API with authentication, file upload, and more
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"mobile-backend/config"
	"mobile-backend/controllers"
	"mobile-backend/docs"
	"mobile-backend/middleware"
	"mobile-backend/models"
	"mobile-backend/routes"
	"mobile-backend/services"
	"mobile-backend/utils"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger := config.SetupLogger()
	defer logger.Sync()

	// Connect to database
	if err := config.ConnectDB(); err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run database migrations
	if err := config.GetDB().AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Product{},
		&models.Plan{},
		&models.Subscription{},
		&models.Payment{},
		&models.PaymentMethod{},
		&models.WebhookEvent{},
		&models.GeminiConversation{},
		&models.OfflineOperation{},
		&models.SyncConflict{},
		&models.DataVersion{},
		&models.SyncStatus{},
		&models.SyncHistory{},
		&models.PushNotification{},
		&models.NotificationTemplate{},
		&models.NotificationSegment{},
		&models.DeviceToken{},
		&models.NotificationAnalytics{},
	); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}
	logger.Info("Database migrations completed successfully")

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	// Initialize services
	cacheService := services.NewCacheService(redisClient)
	cacheMetricsService := services.NewCacheMetricsService(redisClient)
	authService := services.NewAuthService(config.GetDB(), cacheService)
	oauth2Service := services.NewOAuth2Service(config.GetDB(), redisClient)

	// Initialize subscription status service
	subscriptionStatusService := services.NewSubscriptionStatusService(config.GetDB(), logger.Logger)

	// Initialize WebSocket services first
	websocketHub := services.NewHub(logger.Logger)
	websocketService := services.NewWebSocketService(websocketHub, config.GetDB(), redisClient, cacheService, logger.Logger)

	// Initialize payment services with WebSocket integration and subscription status service
	stripeService := services.NewStripeService(config.GetDB(), cacheService, websocketService, subscriptionStatusService)
	polarService := services.NewPolarService(config.GetDB(), cacheService, subscriptionStatusService)

	// Initialize job queue and background processing services
	jobQueueService := services.NewJobQueueService(os.Getenv("REDIS_URL"), config.GetDB(), logger.Logger)
	cronScheduler := services.NewCronScheduler(jobQueueService, config.GetDB(), logger.Logger)
	workerManager := services.NewWorkerManager(jobQueueService, cronScheduler, config.GetDB(), logger.Logger)
	jobQueueMetrics := services.NewJobQueueMetrics(os.Getenv("REDIS_URL"), logger.Logger)

	// Initialize Gemini AI service
	geminiService, err := services.NewGeminiService(config.GetDB(), cacheService, logger.Logger)
	if err != nil {
		logger.Fatal("Failed to initialize Gemini service", zap.Error(err))
	}

	// Initialize offline sync service
	offlineSyncService := services.NewOfflineSyncService(config.GetDB(), redisClient, cacheService, websocketService, logger.Logger)

	// Start background retry service for offline sync
	go offlineSyncService.StartRetryService(context.Background())

	// Initialize push notification service
	pushNotificationService := services.NewPushNotificationService(config.GetDB(), redisClient, cacheService, websocketService, logger.Logger)

	// Start WebSocket hub in a goroutine
	go websocketHub.Run()

	// Start background workers and scheduler
	go func() {
		if err := workerManager.Start(); err != nil {
			logger.Fatal("Failed to start worker manager", zap.Error(err))
		}
	}()

	// Start job queue metrics collection
	jobQueueMetrics.Start()

	// Start subscription expiry checker
	go subscriptionStatusService.StartSubscriptionExpiryChecker(ctx)

	// Initialize product sync service
	productSyncService := services.NewProductSyncService(config.GetDB())

	// Initialize controllers
	healthController := controllers.NewHealthController(config.GetDB())
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(authService)
	uploadController := controllers.NewUploadController("./uploads")
	generatorController := controllers.NewGeneratorController(config.GetDB())
	oauth2Controller := controllers.NewOAuth2Controller(oauth2Service, authService)
	cacheController := controllers.NewCacheController(cacheService, cacheMetricsService)
	paymentController := controllers.NewPaymentController(stripeService, polarService, config.GetDB())
	websocketController := controllers.NewWebSocketController(websocketService, websocketHub, logger.Logger)
	jobQueueController := controllers.NewJobQueueController(jobQueueService, workerManager, logger.Logger)
	jobQueueMetricsController := controllers.NewJobQueueMetricsController(jobQueueMetrics, logger.Logger)
	productWebhookController := controllers.NewProductWebhookController(stripeService, polarService, productSyncService)
	geminiController := controllers.NewGeminiController(geminiService, logger.Logger)
	offlineSyncController := controllers.NewOfflineSyncController(offlineSyncService, logger.Logger)
	pushNotificationController := controllers.NewPushNotificationController(pushNotificationService, logger.Logger)
	// Subscription management controller is initialized in routes

	// Setup Gin router
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Initialize rate limiter with different strategies
	rateLimiter := middleware.NewRateLimiter(redisClient)

	// Initialize subscription middleware
	subscriptionMiddleware := middleware.NewSubscriptionMiddleware(config.GetDB(), logger.Logger)

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.CORS())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(middleware.PrometheusMiddleware())
	r.Use(utils.ErrorHandler()) // Add error handler

	// Apply rate limiting to specific route groups
	authGroup := r.Group("/api/v1/auth")
	authGroup.Use(rateLimiter.AuthRateLimit())

	apiGroup := r.Group("/api/v1")
	apiGroup.Use(rateLimiter.APIRateLimit())

	// Setup routes
	routes.SetupRoutes(r, healthController, authController, userController, uploadController, generatorController, oauth2Controller, cacheController, paymentController, websocketController, offlineSyncController)

	// Setup WebSocket routes with logger
	routes.SetupWebSocketRoutes(r, websocketController, logger.Logger)

	// Setup generator routes
	routes.SetupGeneratorRoutes(r, generatorController)

	// Setup generated API routes
	productController := controllers.NewProductController(config.GetDB())
	routes.SetupProductRoutes(r, productController)

	// Setup example routes (demonstrates all features)
	exampleController := controllers.NewExampleController()
	routes.SetupExampleRoutes(r, exampleController)

	// Setup job queue routes
	routes.SetupJobQueueRoutes(r, jobQueueController, jobQueueMetricsController)

	// Setup subscription management routes
	routes.SetupSubscriptionManagementRoutes(
		apiGroup,
		config.GetDB(),
		subscriptionStatusService,
		stripeService,
		polarService,
		subscriptionMiddleware,
		logger.Logger,
	)

	// Setup product webhook routes
	routes.SetupProductWebhookRoutes(apiGroup, productWebhookController)

	// Setup Gemini AI routes with rate limiting
	routes.SetupGeminiRoutesWithRateLimit(r, geminiController, rateLimiter, logger.Logger)

	// Setup push notification routes
	routes.SetupPushNotificationRoutes(r, pushNotificationController)

	// Regenerate Swagger documentation on startup
	logger.Info("Regenerating Swagger documentation...")
	if err := regenerateSwaggerDocs(); err != nil {
		logger.Warn("Failed to regenerate Swagger docs", zap.Error(err))
	} else {
		logger.Info("Swagger documentation regenerated successfully")
	}

	// Setup Swagger
	docs.SwaggerInfo.Title = "Mobile Backend API"
	docs.SwaggerInfo.Description = "A comprehensive mobile backend API with authentication, file upload, and more"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Stop background workers and scheduler
	logger.Info("Stopping background workers...")
	workerManager.Stop()

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// regenerateSwaggerDocs runs swag init to regenerate Swagger documentation
func regenerateSwaggerDocs() error {
	cmd := exec.Command("swag", "init")
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to regenerate Swagger docs: %v, output: %s", err, string(output))
	}
	return nil
}
