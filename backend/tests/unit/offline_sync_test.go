package unit

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"mobile-backend-template/models"
	"mobile-backend-template/services"
)

// MockWebSocketService is a mock for WebSocketService
type MockWebSocketService struct {
	mock.Mock
}

func (m *MockWebSocketService) SendDataUpdate(ctx context.Context, userID uint, dataType string, data map[string]interface{}) error {
	args := m.Called(ctx, userID, dataType, data)
	return args.Error(0)
}

func (m *MockWebSocketService) SendNotification(ctx context.Context, userID uint, notificationType, title, message string, data map[string]interface{}) error {
	args := m.Called(ctx, userID, notificationType, title, message, data)
	return args.Error(0)
}

func (m *MockWebSocketService) SendLiveUpdate(ctx context.Context, userID uint, updateType string, data map[string]interface{}) error {
	args := m.Called(ctx, userID, updateType, data)
	return args.Error(0)
}

func (m *MockWebSocketService) BroadcastDataUpdate(ctx context.Context, dataType string, data map[string]interface{}) error {
	args := m.Called(ctx, dataType, data)
	return args.Error(0)
}

func (m *MockWebSocketService) SendPaymentNotification(ctx context.Context, userID uint, paymentType string, paymentData map[string]interface{}) error {
	args := m.Called(ctx, userID, paymentType, paymentData)
	return args.Error(0)
}

func (m *MockWebSocketService) SendSystemNotification(ctx context.Context, notificationType, title, message string, data map[string]interface{}) error {
	args := m.Called(ctx, notificationType, title, message, data)
	return args.Error(0)
}

func (m *MockWebSocketService) SendRoomNotification(ctx context.Context, room, notificationType, title, message string, data map[string]interface{}) error {
	args := m.Called(ctx, room, notificationType, title, message, data)
	return args.Error(0)
}

func (m *MockWebSocketService) GetUserNotifications(ctx context.Context, userID uint, limit int) ([]services.NotificationData, error) {
	args := m.Called(ctx, userID, limit)
	return args.Get(0).([]services.NotificationData), args.Error(1)
}

func (m *MockWebSocketService) MarkNotificationAsRead(ctx context.Context, userID uint, notificationID uint) error {
	args := m.Called(ctx, userID, notificationID)
	return args.Error(0)
}

func (m *MockWebSocketService) GetConnectionStats() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockWebSocketService) GetUserConnections(userID uint) []*services.Client {
	args := m.Called(userID)
	return args.Get(0).([]*services.Client)
}

func (m *MockWebSocketService) SendTypingIndicator(ctx context.Context, userID uint, room string, isTyping bool) error {
	args := m.Called(ctx, userID, room, isTyping)
	return args.Error(0)
}

func (m *MockWebSocketService) SendPresenceUpdate(ctx context.Context, userID uint, status string) error {
	args := m.Called(ctx, userID, status)
	return args.Error(0)
}

// MockCacheService is a mock for CacheService
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) Increment(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) Decrement(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) GetStats() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockCacheService) Clear() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheService) GetKeys(pattern string) ([]string, error) {
	args := m.Called(pattern)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCacheService) SetWithTags(ctx context.Context, key string, value interface{}, expiration time.Duration, tags []string) error {
	args := m.Called(ctx, key, value, expiration, tags)
	return args.Error(0)
}

func (m *MockCacheService) InvalidateByTags(ctx context.Context, tags []string) error {
	args := m.Called(ctx, tags)
	return args.Error(0)
}

func (m *MockCacheService) GetMetrics() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockCacheService) ResetMetrics() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCacheService) GetRecommendations() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockCacheService) GetHealth() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
}

func (m *MockCacheService) WarmCache(ctx context.Context, keys []string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockCacheService) GetCacheKey(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheService) SetCacheKey(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) InvalidateCache(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCacheService) ClearCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func setupTestDB() *gorm.DB {
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

	return db
}

func setupTestRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use different DB for tests
	})
}

func TestOfflineSyncService_QueueOperation(t *testing.T) {
	db := setupTestDB()
	redisClient := setupTestRedis()
	mockWS := &MockWebSocketService{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	offlineSyncService := services.NewOfflineSyncService(db, redisClient, mockCache, mockWS, logger)

	// Create a test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	db.Create(user)

	// Mock expectations
	mockWS.On("SendDataUpdate", mock.Anything, user.ID, "offline_operation_queued", mock.Anything).Return(nil)
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Test queue operation
	operation := &models.OfflineOperation{
		OperationType: models.OperationTypeCreate,
		TableName:     "users",
		RecordID:      "123",
		Data: map[string]interface{}{
			"name":  "Test User",
			"email": "test@example.com",
		},
	}

	err := offlineSyncService.QueueOperation(context.Background(), user.ID, operation)

	assert.NoError(t, err)
	assert.NotEmpty(t, operation.OperationID)
	assert.Equal(t, models.OperationStatusPending, operation.Status)

	// Verify operation was saved to database
	var savedOp models.OfflineOperation
	err = db.Where("operation_id = ?", operation.OperationID).First(&savedOp).Error
	assert.NoError(t, err)
	assert.Equal(t, user.ID, savedOp.UserID)
	assert.Equal(t, models.OperationTypeCreate, savedOp.OperationType)

	mockWS.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestOfflineSyncService_SyncUserData(t *testing.T) {
	db := setupTestDB()
	redisClient := setupTestRedis()
	mockWS := &MockWebSocketService{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	offlineSyncService := services.NewOfflineSyncService(db, redisClient, mockCache, mockWS, logger)

	// Create a test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	db.Create(user)

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

	// Mock expectations
	mockWS.On("SendDataUpdate", mock.Anything, user.ID, "sync_completed", mock.Anything).Return(nil)

	// Test sync
	err := offlineSyncService.SyncUserData(context.Background(), user.ID)

	assert.NoError(t, err)

	// Verify operations were processed
	var processedOps []models.OfflineOperation
	db.Where("user_id = ?", user.ID).Find(&processedOps)
	assert.Len(t, processedOps, 2)

	// All operations should be completed
	for _, op := range processedOps {
		assert.Equal(t, models.OperationStatusCompleted, op.Status)
	}

	mockWS.AssertExpectations(t)
}

func TestOfflineSyncService_GetSyncStatus(t *testing.T) {
	db := setupTestDB()
	redisClient := setupTestRedis()
	mockWS := &MockWebSocketService{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	offlineSyncService := services.NewOfflineSyncService(db, redisClient, mockCache, mockWS, logger)

	// Create a test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	db.Create(user)

	// Create sync status
	syncStatus := &models.SyncStatus{
		UserID:                 user.ID,
		PendingOperationsCount: 5,
		ConflictsCount:         2,
		IsOnline:               true,
		LastOnlineAt:           time.Now(),
	}
	db.Create(syncStatus)

	// Test get sync status
	result, err := offlineSyncService.GetSyncStatus(context.Background(), user.ID)

	assert.NoError(t, err)
	assert.Equal(t, user.ID, result.UserID)
	assert.Equal(t, 5, result.PendingOperationsCount)
	assert.Equal(t, 2, result.ConflictsCount)
	assert.True(t, result.IsOnline)
}

func TestOfflineSyncService_SetUserOnline(t *testing.T) {
	db := setupTestDB()
	redisClient := setupTestRedis()
	mockWS := &MockWebSocketService{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	offlineSyncService := services.NewOfflineSyncService(db, redisClient, mockCache, mockWS, logger)

	// Create a test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	db.Create(user)

	// Test set user online
	err := offlineSyncService.SetUserOnline(context.Background(), user.ID)

	assert.NoError(t, err)

	// Verify sync status was created/updated
	var syncStatus models.SyncStatus
	err = db.Where("user_id = ?", user.ID).First(&syncStatus).Error
	assert.NoError(t, err)
	assert.True(t, syncStatus.IsOnline)
}

func TestOfflineSyncService_SetUserOffline(t *testing.T) {
	db := setupTestDB()
	redisClient := setupTestRedis()
	mockWS := &MockWebSocketService{}
	mockCache := &MockCacheService{}
	logger := zap.NewNop()

	offlineSyncService := services.NewOfflineSyncService(db, redisClient, mockCache, mockWS, logger)

	// Create a test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	db.Create(user)

	// Create sync status
	syncStatus := &models.SyncStatus{
		UserID:       user.ID,
		IsOnline:     true,
		LastOnlineAt: time.Now(),
	}
	db.Create(syncStatus)

	// Test set user offline
	err := offlineSyncService.SetUserOffline(context.Background(), user.ID)

	assert.NoError(t, err)

	// Verify sync status was updated
	var updatedStatus models.SyncStatus
	err = db.Where("user_id = ?", user.ID).First(&updatedStatus).Error
	assert.NoError(t, err)
	assert.False(t, updatedStatus.IsOnline)
}

func TestConflictResolver_ResolveConflict(t *testing.T) {
	db := setupTestDB()
	logger := zap.NewNop()

	conflictResolver := services.NewConflictResolver(db, logger)

	// Create test conflict
	conflict := &models.SyncConflict{
		UserID:       1,
		TableName:    "users",
		RecordID:     "123",
		ConflictType: models.ConflictTypeVersionMismatch,
		LocalData: map[string]interface{}{
			"name":  "Local User",
			"email": "local@example.com",
		},
		ServerData: map[string]interface{}{
			"name":  "Server User",
			"email": "server@example.com",
		},
		Status: models.ConflictStatusPending,
	}

	// Test resolve conflict
	resolvedData, err := conflictResolver.ResolveConflict(context.Background(), conflict)

	assert.NoError(t, err)
	assert.NotNil(t, resolvedData)

	// Should use server data by default (server wins strategy)
	assert.Equal(t, "Server User", resolvedData["name"])
	assert.Equal(t, "server@example.com", resolvedData["email"])
}

func TestConflictResolver_GetConflictResolutionStrategies(t *testing.T) {
	db := setupTestDB()
	logger := zap.NewNop()

	conflictResolver := services.NewConflictResolver(db, logger)

	strategies := conflictResolver.GetConflictResolutionStrategies()

	expectedStrategies := []string{
		models.ResolutionStrategyServerWins,
		models.ResolutionStrategyClientWins,
		models.ResolutionStrategyMerge,
		models.ResolutionStrategyManual,
	}

	assert.ElementsMatch(t, expectedStrategies, strategies)
}

func TestConflictResolver_GetConflictTypes(t *testing.T) {
	db := setupTestDB()
	logger := zap.NewNop()

	conflictResolver := services.NewConflictResolver(db, logger)

	conflictTypes := conflictResolver.GetConflictTypes()

	expectedTypes := []string{
		models.ConflictTypeVersionMismatch,
		models.ConflictTypeConcurrentEdit,
		models.ConflictTypeDeletedModified,
	}

	assert.ElementsMatch(t, expectedTypes, conflictTypes)
}

func TestConflictResolver_AnalyzeConflict(t *testing.T) {
	db := setupTestDB()
	logger := zap.NewNop()

	conflictResolver := services.NewConflictResolver(db, logger)

	// Create test conflict
	conflict := &models.SyncConflict{
		ID:           1,
		UserID:       1,
		TableName:    "users",
		RecordID:     "123",
		ConflictType: models.ConflictTypeVersionMismatch,
		LocalData: map[string]interface{}{
			"name":  "Local User",
			"email": "local@example.com",
		},
		ServerData: map[string]interface{}{
			"name":  "Server User",
			"email": "server@example.com",
		},
		Status: models.ConflictStatusPending,
	}

	// Test analyze conflict
	analysis := conflictResolver.AnalyzeConflict(conflict)

	assert.Equal(t, uint(1), analysis.ConflictID)
	assert.Equal(t, models.ConflictTypeVersionMismatch, analysis.ConflictType)
	assert.Equal(t, models.ResolutionStrategyServerWins, analysis.RecommendedStrategy)
	assert.NotEmpty(t, analysis.Description)
	assert.Len(t, analysis.FieldsInConflict, 2) // name and email
}

func TestConflictResolver_GetConflictStatistics(t *testing.T) {
	db := setupTestDB()
	logger := zap.NewNop()

	conflictResolver := services.NewConflictResolver(db, logger)

	// Create test user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	db.Create(user)

	// Create test conflicts
	conflicts := []models.SyncConflict{
		{
			UserID:       user.ID,
			TableName:    "users",
			RecordID:     "1",
			ConflictType: models.ConflictTypeVersionMismatch,
			Status:       models.ConflictStatusPending,
		},
		{
			UserID:       user.ID,
			TableName:    "users",
			RecordID:     "2",
			ConflictType: models.ConflictTypeConcurrentEdit,
			Status:       models.ConflictStatusResolved,
		},
	}

	for _, conflict := range conflicts {
		db.Create(&conflict)
	}

	// Test get conflict statistics
	stats, err := conflictResolver.GetConflictStatistics(context.Background(), user.ID)

	assert.NoError(t, err)
	assert.Equal(t, 2, stats.TotalConflicts)
	assert.Equal(t, 1, stats.PendingConflicts)
	assert.Equal(t, 1, stats.ResolvedConflicts)
	assert.Equal(t, 1, stats.VersionMismatchConflicts)
	assert.Equal(t, 1, stats.ConcurrentEditConflicts)
}
