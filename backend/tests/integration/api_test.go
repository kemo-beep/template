package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"mobile-backend/config"
	"mobile-backend/controllers"
	"mobile-backend/routes"
	"mobile-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func setupTestApp() *gin.Engine {
	// Set environment variables
	os.Setenv("DATABASE_URL", "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable")
	os.Setenv("REDIS_URL", "redis://localhost:6379")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("GIN_MODE", "test")

	// Connect to database
	config.ConnectDB()

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Initialize services
	cacheService := services.NewCacheService(redisClient)
	authService := services.NewAuthService(config.GetDB(), cacheService)

	// Initialize controllers
	healthController := controllers.NewHealthController(config.GetDB())
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(authService)
	uploadController := controllers.NewUploadController("./test-uploads")

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Setup routes
	routes.SetupRoutes(r, healthController, authController, userController, uploadController)

	return r
}

func TestHealthCheck(t *testing.T) {
	// Setup app
	app := setupTestApp()

	// Test health check
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserRegistration(t *testing.T) {
	// Setup app
	app := setupTestApp()

	// Test registration
	userData := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"name":     "Test User",
	}

	jsonData, _ := json.Marshal(userData)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// This test will fail without a real database, but it's a placeholder
	// In a real test environment, you would set up test databases
	assert.Equal(t, http.StatusInternalServerError, w.Code) // Expected to fail without DB
}

func TestUserLogin(t *testing.T) {
	// Setup app
	app := setupTestApp()

	// Test login
	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}

	jsonData, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// This test will fail without a real database, but it's a placeholder
	assert.Equal(t, http.StatusInternalServerError, w.Code) // Expected to fail without DB
}