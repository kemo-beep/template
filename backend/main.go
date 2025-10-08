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
	if err := config.GetDB().AutoMigrate(&models.User{}, &models.Session{}); err != nil {
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
	authService := services.NewAuthService(config.GetDB(), cacheService)
	oauth2Service := services.NewOAuth2Service(config.GetDB(), redisClient)

	// Initialize controllers
	healthController := controllers.NewHealthController(config.GetDB())
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(authService)
	uploadController := controllers.NewUploadController("./uploads")
	generatorController := controllers.NewGeneratorController(config.GetDB())
	oauth2Controller := controllers.NewOAuth2Controller(oauth2Service, authService)

	// Setup Gin router
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(middleware.CORS())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(middleware.PrometheusMiddleware())

	// Setup routes
	routes.SetupRoutes(r, healthController, authController, userController, uploadController, generatorController, oauth2Controller)

	// Setup generator routes
	routes.SetupGeneratorRoutes(r, generatorController)

	// Setup generated API routes
	productController := controllers.NewProductController(config.GetDB())
	routes.SetupProductRoutes(r, productController)

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
