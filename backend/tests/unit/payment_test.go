package unit

import (
	"context"
	"testing"
	"time"

	"mobile-backend/models"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockRedisClient is a mock implementation of Redis client
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Incr(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	args := m.Called(ctx, pattern)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (m *MockRedisClient) DBSize(ctx context.Context) *redis.IntCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) MemoryUsage(ctx context.Context, key string) *redis.IntCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) SMembers(ctx context.Context, key string) *redis.StringSliceCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringSliceCmd)
}

func (m *MockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.BoolCmd)
}

func (m *MockRedisClient) ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	args := m.Called(ctx, key, members)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) *redis.IntCmd {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).(*redis.StringSliceCmd)
}

// MockCacheService is a mock implementation of CacheService
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

func (m *MockCacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockCacheService) GetOrSet(ctx context.Context, key string, dest interface{}, fetchFunc func() (interface{}, error), expiration time.Duration) error {
	args := m.Called(ctx, key, dest, fetchFunc, expiration)
	return args.Error(0)
}

func (m *MockCacheService) CacheAside(ctx context.Context, key string, dest interface{}, fetchFunc func() (interface{}, error), expiration time.Duration) error {
	args := m.Called(ctx, key, dest, fetchFunc, expiration)
	return args.Error(0)
}

func (m *MockCacheService) WriteThrough(ctx context.Context, key string, value interface{}, writeFunc func(interface{}) error, expiration time.Duration) error {
	args := m.Called(ctx, key, value, writeFunc, expiration)
	return args.Error(0)
}

func (m *MockCacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCacheService) WarmCache(ctx context.Context, keys []string, fetchFunc func(string) (interface{}, error), expiration time.Duration) error {
	args := m.Called(ctx, keys, fetchFunc, expiration)
	return args.Error(0)
}

func (m *MockCacheService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockCacheService) SetWithTags(ctx context.Context, key string, value interface{}, tags []string, expiration time.Duration) error {
	args := m.Called(ctx, key, value, tags, expiration)
	return args.Error(0)
}

func (m *MockCacheService) InvalidateByTags(ctx context.Context, tags []string) error {
	args := m.Called(ctx, tags)
	return args.Error(0)
}

func (m *MockCacheService) SetCompressed(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	args := m.Called(ctx, key, expiration)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) Unlock(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheService) Increment(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheService) SetExpiration(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database")
	}

	// Auto migrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Product{},
		&models.Plan{},
		&models.Subscription{},
		&models.Payment{},
		&models.PaymentMethod{},
		&models.WebhookEvent{},
	); err != nil {
		panic("Failed to migrate test database: " + err.Error())
	}

	return db
}

func TestProductModel(t *testing.T) {
	db := setupTestDB()

	// Test product creation
	product := &models.Product{
		Name:        "Test Product",
		Description: "A test product",
		Price:       1000, // $10.00 in cents
		Currency:    "usd",
		IsActive:    true,
		IsRecurring: false,
	}

	err := db.Create(product).Error
	assert.NoError(t, err)
	assert.NotZero(t, product.ID)

	// Test product retrieval
	var retrievedProduct models.Product
	err = db.First(&retrievedProduct, product.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, product.Name, retrievedProduct.Name)
	assert.Equal(t, product.Price, retrievedProduct.Price)
	assert.Equal(t, 10.0, retrievedProduct.GetPriceInDollars())
}

func TestPlanModel(t *testing.T) {
	db := setupTestDB()

	// Create a product first
	product := &models.Product{
		Name:        "Test Product",
		Description: "A test product",
		Price:       1000,
		Currency:    "usd",
		IsActive:    true,
		IsRecurring: true,
	}
	err := db.Create(product).Error
	assert.NoError(t, err)

	// Test plan creation
	plan := &models.Plan{
		Name:          "Test Plan",
		Description:   "A test plan",
		ProductID:     product.ID,
		Price:         2000, // $20.00 in cents
		Currency:      "usd",
		Interval:      "month",
		IntervalCount: 1,
		IsActive:      true,
	}

	err = db.Create(plan).Error
	assert.NoError(t, err)
	assert.NotZero(t, plan.ID)

	// Test plan retrieval
	var retrievedPlan models.Plan
	err = db.Preload("Product").First(&retrievedPlan, plan.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, plan.Name, retrievedPlan.Name)
	assert.Equal(t, plan.Price, retrievedPlan.Price)
	assert.Equal(t, 20.0, retrievedPlan.GetPriceInDollars())
	assert.Equal(t, product.Name, retrievedPlan.Product.Name)
}

func TestSubscriptionModel(t *testing.T) {
	db := setupTestDB()

	// Create a user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		IsActive: true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	// Create a product
	product := &models.Product{
		Name:        "Test Product",
		Description: "A test product",
		Price:       1000,
		Currency:    "usd",
		IsActive:    true,
		IsRecurring: true,
	}
	err = db.Create(product).Error
	assert.NoError(t, err)

	// Create a plan
	plan := &models.Plan{
		Name:          "Test Plan",
		Description:   "A test plan",
		ProductID:     product.ID,
		Price:         2000,
		Currency:      "usd",
		Interval:      "month",
		IntervalCount: 1,
		IsActive:      true,
	}
	err = db.Create(plan).Error
	assert.NoError(t, err)

	// Test subscription creation
	now := time.Now()
	subscription := &models.Subscription{
		UserID:             user.ID,
		ProductID:          product.ID,
		PlanID:             &plan.ID,
		Status:             "active",
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   now.AddDate(0, 1, 0), // 1 month from now
		Quantity:           1,
	}

	err = db.Create(subscription).Error
	assert.NoError(t, err)
	assert.NotZero(t, subscription.ID)

	// Test subscription methods
	assert.True(t, subscription.IsActive())
	assert.False(t, subscription.IsCanceled())
	assert.False(t, subscription.IsPastDue())
	assert.False(t, subscription.IsTrial())

	// Test subscription retrieval
	var retrievedSubscription models.Subscription
	err = db.Preload("User").Preload("Product").Preload("Plan").First(&retrievedSubscription, subscription.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, subscription.Status, retrievedSubscription.Status)
	assert.Equal(t, user.Email, retrievedSubscription.User.Email)
	assert.Equal(t, product.Name, retrievedSubscription.Product.Name)
	assert.Equal(t, plan.Name, retrievedSubscription.Plan.Name)
}

func TestPaymentModel(t *testing.T) {
	db := setupTestDB()

	// Create a user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		IsActive: true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	// Create a product
	product := &models.Product{
		Name:        "Test Product",
		Description: "A test product",
		Price:       1000,
		Currency:    "usd",
		IsActive:    true,
		IsRecurring: false,
	}
	err = db.Create(product).Error
	assert.NoError(t, err)

	// Test payment creation
	payment := &models.Payment{
		UserID:        user.ID,
		ProductID:     product.ID,
		Amount:        1500, // $15.00 in cents
		Currency:      "usd",
		Status:        "succeeded",
		PaymentMethod: "stripe",
		Description:   "Test payment",
	}

	err = db.Create(payment).Error
	assert.NoError(t, err)
	assert.NotZero(t, payment.ID)

	// Test payment methods
	assert.Equal(t, 15.0, payment.GetAmountInDollars())

	// Test payment retrieval
	var retrievedPayment models.Payment
	err = db.Preload("User").Preload("Product").First(&retrievedPayment, payment.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, payment.Status, retrievedPayment.Status)
	assert.Equal(t, user.Email, retrievedPayment.User.Email)
	assert.Equal(t, product.Name, retrievedPayment.Product.Name)
}

func TestWebhookEventModel(t *testing.T) {
	db := setupTestDB()

	// Test webhook event creation
	event := &models.WebhookEvent{
		Provider:  "stripe",
		EventType: "payment_intent.succeeded",
		EventID:   "evt_test123",
		Processed: false,
		Data: models.JSONMap{
			"id":     "pi_test123",
			"amount": 1000,
		},
	}

	err := db.Create(event).Error
	assert.NoError(t, err)
	assert.NotZero(t, event.ID)

	// Test webhook event retrieval
	var retrievedEvent models.WebhookEvent
	err = db.First(&retrievedEvent, event.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, event.Provider, retrievedEvent.Provider)
	assert.Equal(t, event.EventType, retrievedEvent.EventType)
	assert.Equal(t, event.EventID, retrievedEvent.EventID)
	assert.Equal(t, event.Processed, retrievedEvent.Processed)
}

func TestPaymentMethodModel(t *testing.T) {
	db := setupTestDB()

	// Create a user
	user := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Name:     "Test User",
		IsActive: true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)

	// Test payment method creation
	paymentMethod := &models.PaymentMethod{
		UserID:                user.ID,
		Type:                  "card",
		IsDefault:             true,
		Last4:                 "4242",
		Brand:                 "visa",
		ExpMonth:              12,
		ExpYear:               2025,
		StripePaymentMethodID: "pm_test123",
	}

	err = db.Create(paymentMethod).Error
	assert.NoError(t, err)
	assert.NotZero(t, paymentMethod.ID)

	// Test payment method retrieval
	var retrievedPaymentMethod models.PaymentMethod
	err = db.Preload("User").First(&retrievedPaymentMethod, paymentMethod.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, paymentMethod.Type, retrievedPaymentMethod.Type)
	assert.Equal(t, paymentMethod.Last4, retrievedPaymentMethod.Last4)
	assert.Equal(t, user.Email, retrievedPaymentMethod.User.Email)
}

// Integration tests would go here, testing the actual service implementations
// with mocked external APIs (Stripe, Polar)
