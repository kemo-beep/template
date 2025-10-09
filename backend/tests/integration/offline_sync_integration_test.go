package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"mobile-backend-template/controllers"
	"mobile-backend-template/middleware"
	"mobile-backend-template/models"
	"mobile-backend-template/routes"
	"mobile-backend-template/services"
	"mobile-backend-template/utils"
)

func setupIntegrationTest() (*gin.Engine, *gorm.DB, *redis.Client, func()) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database")
	}

	// Auto-migrate the models
	db.AutoMigrate(
		&models.User{},
		&models.OfflineOperation{},
		&models.SyncConflict{},
		&models.DataVersion{},
		&models.SyncStatus{},
		&models.SyncHistory{},
	)

	// Setup test Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   2, // Use different DB for integration tests
	})

	// Setup logger
	logger := zap.NewNop()

	// Initialize services
	cacheService := services.NewCacheService(redisClient)
	authService := services.NewAuthService(db, cacheService)
	websocketHub := services.NewHub(logger)
	websocketService := services.NewWebSocketService(websocketHub, db, redisClient, cacheService, logger)
	offlineSyncService := services.NewOfflineSyncService(db, redisClient, cacheService, websocketService, logger)

	// Initialize controllers
	authController := controllers.NewAuthController(authService)
	offlineSyncController := controllers.NewOfflineSyncController(offlineSyncService, logger)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add middleware
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(utils.ErrorHandler())

	// Setup routes
	routes.SetupOfflineSyncRoutes(r, offlineSyncController)

	// Add auth routes for testing
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
	}

	// Start WebSocket hub
	go websocketHub.Run()

	cleanup := func() {
		redisClient.Close()
	}

	return r, db, redisClient, cleanup
}

func createTestUser(db *gorm.DB) (*models.User, string) {
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	user.HashPassword("password123")
	db.Create(user)

	// Create JWT token
	token, _ := utils.GenerateJWT(user.ID, "test@example.com")

	return user, token
}

func TestOfflineSyncAPI_QueueOperation(t *testing.T) {
	router, db, redisClient, cleanup := setupIntegrationTest()
	defer cleanup()

	// Create test user
	user, token := createTestUser(db)

	// Test queue operation
	operationData := map[string]interface{}{
		"operation_type": "create",
		"table_name":     "users",
		"record_id":      "123",
		"data": map[string]interface{}{
			"name":  "Test User",
			"email": "test@example.com",
		},
	}

	jsonData, _ := json.Marshal(operationData)
	req, _ := http.NewRequest("POST", "/api/v1/sync/queue", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "data")

	// Verify operation was saved to database
	var operation models.OfflineOperation
	err = db.Where("user_id = ?", user.ID).First(&operation).Error
	assert.NoError(t, err)
	assert.Equal(t, "create", operation.OperationType)
	assert.Equal(t, "users", operation.TableName)
}

func TestOfflineSyncAPI_GetSyncStatus(t *testing.T) {
	router, db, redisClient, cleanup := setupIntegrationTest()
	defer cleanup()

	// Create test user
	user, token := createTestUser(db)

	// Create sync status
	syncStatus := &models.SyncStatus{
		UserID:                 user.ID,
		PendingOperationsCount: 3,
		ConflictsCount:         1,
		IsOnline:               true,
		LastOnlineAt:           time.Now(),
	}
	db.Create(syncStatus)

	// Test get sync status
	req, _ := http.NewRequest("GET", "/api/v1/sync/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "data")

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(user.ID), data["user_id"])
	assert.Equal(t, float64(3), data["pending_operations_count"])
	assert.Equal(t, float64(1), data["conflicts_count"])
	assert.True(t, data["is_online"].(bool))
}

func TestOfflineSyncAPI_SyncUserData(t *testing.T) {
	router, db, redisClient, cleanup := setupIntegrationTest()
	defer cleanup()

	// Create test user
	user, token := createTestUser(db)

	// Create test operations
	operations := []models.OfflineOperation{
		{
			UserID:        user.ID,
			OperationID:   "op1",
			OperationType: models.OperationTypeCreate,
			TableName:     "users",
			RecordID:      "123",
			Data:          map[string]interface{}{"name": "Test User"},
			Status:        models.OperationStatusPending,
		},
		{
			UserID:        user.ID,
			OperationID:   "op2",
			OperationType: models.OperationTypeUpdate,
			TableName:     "users",
			RecordID:      "456",
			Data:          map[string]interface{}{"name": "Updated User"},
			Status:        models.OperationStatusPending,
		},
	}

	for _, op := range operations {
		db.Create(&op)
	}

	// Test sync user data
	req, _ := http.NewRequest("POST", "/api/v1/sync/sync", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "data")

	data := response["data"].(map[string]interface{})
	assert.True(t, data["success"].(bool))
	assert.Equal(t, "Sync completed successfully", data["message"])
}

func TestOfflineSyncAPI_ForceSync(t *testing.T) {
	router, db, redisClient, cleanup := setupIntegrationTest()
	defer cleanup()

	// Create test user
	user, token := createTestUser(db)

	// Test force sync
	req, _ := http.NewRequest("POST", "/api/v1/sync/force", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "data")

	data := response["data"].(map[string]interface{})
	assert.True(t, data["success"].(bool))
	assert.Equal(t, "Force sync completed successfully", data["message"])
}

func TestOfflineSyncAPI_GetPendingOperations(t *testing.T) {
	router, db, redisClient, cleanup := setupIntegrationTest()
	defer cleanup()

	// Create test user
	user, token := createTestUser(db)

	// Create test operations
	operations := []models.OfflineOperation{
		{
			UserID:        user.ID,
			OperationID:   "op1",
			OperationType: models.OperationTypeCreate,
			TableName:     "users",
			RecordID:      "123",
			Data:          map[string]interface{}{"name": "Test User"},
			Status:        models.OperationStatusPending,
		},
		{
			UserID:        user.ID,
			OperationID:   "op2",
			OperationType: models.OperationTypeUpdate,
			TableName:     "users",
			RecordID:      "456",
			Data:          map[string]interface{}{"name": "Updated User"},
			Status:        models.OperationStatusPending,
		},
	}

	for _, op := range operations {
		db.Create(&op)
	}

	// Test get pending operations
	req, _ := http.NewRequest("GET", "/api/v1/sync/operations?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "data")

	data := response["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestOfflineSyncAPI_GetConflicts(t *testing.T) {
	router, db, redisClient, cleanup := setupIntegrationTest()
	defer cleanup()

	// Create test user
	user, token := createTestUser(db)

	// Create test conflicts
	conflicts := []models.SyncConflict{
		{
			UserID:       user.ID,
			TableName:    "users",
			RecordID:     "123",
			ConflictType: models.ConflictTypeVersionMismatch,
			LocalData:    map[string]interface{}{"name": "Local User"},
			ServerData:   map[string]interface{}{"name": "Server User"},
			Status:       models.ConflictStatusPending,
		},
	}

	for _, conflict := range conflicts {
		db.Create(&conflict)
	}

	// Test get conflicts
	req, _ := http.NewRequest("GET", "/api/v1/sync/conflicts?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "data")

	data := response["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestOfflineSyncAPI_ResolveConflict(t *testing.T) {
	router, db, redisClient, cleanup := setupIntegrationTest()
	defer cleanup()

	// Create test user
	user, token := createTestUser(db)

	// Create test conflict
	conflict := &models.SyncConflict{
		UserID:       user.ID,
		TableName:    "users",
		RecordID:     "123",
		ConflictType: models.ConflictTypeVersionMismatch,
		LocalData:    map[string]interface{}{"name": "Local User"},
		ServerData:   map[string]interface{}{"name": "Server User"},
		Status:       models.ConflictStatusPending,
	}
	db.Create(conflict)

	// Test resolve conflict
	resolveData := map[string]interface{}{
		"strategy": "server_wins",
	}

	jsonData, _ := json.Marshal(resolveData)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/sync/conflicts/%d/resolve", conflict.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "data")
}

func TestOfflineSyncAPI_SetUserOnline(t *testing.T) {
	router, db, redisClient, cleanup := setupIntegrationTest()
	defer cleanup()

	// Create test user
	user, token := createTestUser(db)

	// Test set user online
	req, _ := http.NewRequest("POST", "/api/v1/sync/online", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "data")

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "online", data["status"])

	// Verify sync status was created/updated
	var syncStatus models.SyncStatus
	err = db.Where("user_id = ?", user.ID).First(&syncStatus).Error
	assert.NoError(t, err)
	assert.True(t, syncStatus.IsOnline)
}

func TestOfflineSyncAPI_SetUserOffline(t *testing.T) {
	router, db, redisClient, cleanup := setupIntegrationTest()
	defer cleanup()

	// Create test user
	user, token := createTestUser(db)

	// Create sync status
	syncStatus := &models.SyncStatus{
		UserID:       user.ID,
		IsOnline:     true,
		LastOnlineAt: time.Now(),
	}
	db.Create(syncStatus)

	// Test set user offline
	req, _ := http.NewRequest("POST", "/api/v1/sync/offline", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "data")

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "offline", data["status"])

	// Verify sync status was updated
	var updatedStatus models.SyncStatus
	err = db.Where("user_id = ?", user.ID).First(&updatedStatus).Error
	assert.NoError(t, err)
	assert.False(t, updatedStatus.IsOnline)
}

func TestOfflineSyncAPI_GetSyncHistory(t *testing.T) {
	router, db, redisClient, cleanup := setupIntegrationTest()
	defer cleanup()

	// Create test user
	user, token := createTestUser(db)

	// Create test sync history
	history := []models.SyncHistory{
		{
			UserID:              user.ID,
			SyncType:            models.SyncTypeIncremental,
			OperationsProcessed: 5,
			ConflictsResolved:   2,
			DurationMs:          1000,
			Success:             true,
		},
		{
			UserID:              user.ID,
			SyncType:            models.SyncTypeFull,
			OperationsProcessed: 10,
			ConflictsResolved:   0,
			DurationMs:          2000,
			Success:             true,
		},
	}

	for _, h := range history {
		db.Create(&h)
	}

	// Test get sync history
	req, _ := http.NewRequest("GET", "/api/v1/sync/history?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "data")

	data := response["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestOfflineSyncAPI_Unauthorized(t *testing.T) {
	router, _, _, cleanup := setupIntegrationTest()
	defer cleanup()

	// Test without authorization header
	req, _ := http.NewRequest("GET", "/api/v1/sync/status", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
	assert.Contains(t, response["message"], "User not authenticated")
}

func TestOfflineSyncAPI_InvalidToken(t *testing.T) {
	router, _, _, cleanup := setupIntegrationTest()
	defer cleanup()

	// Test with invalid token
	req, _ := http.NewRequest("GET", "/api/v1/sync/status", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "error", response["status"])
}
