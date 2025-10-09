package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type RateLimitStrategy string

const (
	FixedWindow   RateLimitStrategy = "fixed_window"
	SlidingWindow RateLimitStrategy = "sliding_window"
	TokenBucket   RateLimitStrategy = "token_bucket"
	LeakyBucket   RateLimitStrategy = "leaky_bucket"
)

type RateLimiter struct {
	redis    *redis.Client
	strategy RateLimitStrategy
}

type RateLimitConfig struct {
	Limit    int
	Window   time.Duration
	Burst    int
	Strategy RateLimitStrategy
	KeyFunc  func(*gin.Context) string
	SkipFunc func(*gin.Context) bool
}

func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{
		redis:    redis,
		strategy: FixedWindow,
	}
}

func NewRateLimiterWithStrategy(redis *redis.Client, strategy RateLimitStrategy) *RateLimiter {
	return &RateLimiter{
		redis:    redis,
		strategy: strategy,
	}
}

// Default key function - uses IP address
func DefaultKeyFunc(c *gin.Context) string {
	return fmt.Sprintf("rate_limit:%s", c.ClientIP())
}

// User-based key function
func UserKeyFunc(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("rate_limit:user:%v", userID)
	}
	return DefaultKeyFunc(c)
}

// Endpoint-based key function
func EndpointKeyFunc(c *gin.Context) string {
	return fmt.Sprintf("rate_limit:%s:%s", c.Request.Method, c.FullPath())
}

// Combined key function (IP + endpoint)
func CombinedKeyFunc(c *gin.Context) string {
	return fmt.Sprintf("rate_limit:%s:%s:%s", c.ClientIP(), c.Request.Method, c.FullPath())
}

// RateLimit creates a rate limiting middleware
func (rl *RateLimiter) RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	config := RateLimitConfig{
		Limit:    limit,
		Window:   window,
		Strategy: rl.strategy,
		KeyFunc:  DefaultKeyFunc,
	}
	return rl.RateLimitWithConfig(config)
}

// RateLimitWithConfig creates a rate limiting middleware with custom configuration
func (rl *RateLimiter) RateLimitWithConfig(config RateLimitConfig) gin.HandlerFunc {
	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}

	return func(c *gin.Context) {
		// Skip rate limiting if skip function returns true
		if config.SkipFunc != nil && config.SkipFunc(c) {
			c.Next()
			return
		}

		key := config.KeyFunc(c)

		// Apply rate limiting based on strategy
		allowed, remaining, resetTime, err := rl.checkRateLimit(c.Request.Context(), key, config)
		if err != nil {
			utils.HandleError(c, utils.ErrInternalError.WithDetails(map[string]interface{}{
				"rate_limit_error": err.Error(),
			}))
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.Limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))

		if !allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(config.Window.Seconds()), 10))
			utils.HandleError(c, utils.ErrTooManyRequests.WithDetails(map[string]interface{}{
				"retry_after": config.Window.Seconds(),
				"limit":       config.Limit,
				"window":      config.Window.String(),
			}))
			return
		}

		c.Next()
	}
}

// checkRateLimit implements different rate limiting strategies
func (rl *RateLimiter) checkRateLimit(ctx context.Context, key string, config RateLimitConfig) (bool, int, int64, error) {
	switch config.Strategy {
	case FixedWindow:
		return rl.fixedWindow(ctx, key, config)
	case SlidingWindow:
		return rl.slidingWindow(ctx, key, config)
	case TokenBucket:
		return rl.tokenBucket(ctx, key, config)
	case LeakyBucket:
		return rl.leakyBucket(ctx, key, config)
	default:
		return rl.fixedWindow(ctx, key, config)
	}
}

// Fixed window rate limiting
func (rl *RateLimiter) fixedWindow(ctx context.Context, key string, config RateLimitConfig) (bool, int, int64, error) {
	now := time.Now()
	windowStart := now.Truncate(config.Window)
	windowKey := fmt.Sprintf("%s:%d", key, windowStart.Unix())

	pipe := rl.redis.Pipeline()
	incr := pipe.Incr(ctx, windowKey)
	pipe.Expire(ctx, windowKey, config.Window)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return false, 0, 0, err
	}

	current := incr.Val()
	remaining := config.Limit - int(current)
	resetTime := windowStart.Add(config.Window).Unix()

	return current <= int64(config.Limit), remaining, resetTime, nil
}

// Sliding window rate limiting
func (rl *RateLimiter) slidingWindow(ctx context.Context, key string, config RateLimitConfig) (bool, int, int64, error) {
	now := time.Now()
	windowStart := now.Add(-config.Window)

	// Use Redis sorted set for sliding window
	pipe := rl.redis.Pipeline()

	// Remove old entries
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.Unix()))

	// Add current request
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now.Unix()),
		Member: now.UnixNano(),
	})

	// Count requests in window
	pipe.ZCard(ctx, key)

	// Set expiration
	pipe.Expire(ctx, key, config.Window)

	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, 0, err
	}

	count := results[2].(*redis.IntCmd).Val()
	remaining := config.Limit - int(count)
	resetTime := now.Add(config.Window).Unix()

	return count <= int64(config.Limit), remaining, resetTime, nil
}

// Token bucket rate limiting
func (rl *RateLimiter) tokenBucket(ctx context.Context, key string, config RateLimitConfig) (bool, int, int64, error) {
	now := time.Now()

	// Get current bucket state
	pipe := rl.redis.Pipeline()
	tokens := pipe.HGet(ctx, key, "tokens")
	lastRefill := pipe.HGet(ctx, key, "last_refill")
	pipe.Exec(ctx)

	currentTokens := 0
	lastRefillTime := now

	if tokens.Val() != "" {
		if t, err := strconv.Atoi(tokens.Val()); err == nil {
			currentTokens = t
		}
	}

	if lastRefill.Val() != "" {
		if t, err := strconv.ParseInt(lastRefill.Val(), 10, 64); err == nil {
			lastRefillTime = time.Unix(t, 0)
		}
	}

	// Calculate tokens to add based on time passed
	timePassed := now.Sub(lastRefillTime)
	tokensToAdd := int(timePassed.Seconds()) * (config.Limit / int(config.Window.Seconds()))

	if tokensToAdd > 0 {
		currentTokens = min(config.Limit, currentTokens+tokensToAdd)
		lastRefillTime = now
	}

	// Check if we have tokens
	if currentTokens <= 0 {
		// Update bucket state
		rl.redis.HMSet(ctx, key, map[string]interface{}{
			"tokens":      currentTokens,
			"last_refill": lastRefillTime.Unix(),
		})
		rl.redis.Expire(ctx, key, config.Window)

		return false, 0, now.Add(config.Window).Unix(), nil
	}

	// Consume token
	currentTokens--

	// Update bucket state
	rl.redis.HMSet(ctx, key, map[string]interface{}{
		"tokens":      currentTokens,
		"last_refill": lastRefillTime.Unix(),
	})
	rl.redis.Expire(ctx, key, config.Window)

	return true, currentTokens, now.Add(config.Window).Unix(), nil
}

// Leaky bucket rate limiting
func (rl *RateLimiter) leakyBucket(ctx context.Context, key string, config RateLimitConfig) (bool, int, int64, error) {
	now := time.Now()

	// Get current bucket state
	pipe := rl.redis.Pipeline()
	level := pipe.HGet(ctx, key, "level")
	lastLeak := pipe.HGet(ctx, key, "last_leak")
	pipe.Exec(ctx)

	currentLevel := 0
	lastLeakTime := now

	if level.Val() != "" {
		if l, err := strconv.Atoi(level.Val()); err == nil {
			currentLevel = l
		}
	}

	if lastLeak.Val() != "" {
		if t, err := strconv.ParseInt(lastLeak.Val(), 10, 64); err == nil {
			lastLeakTime = time.Unix(t, 0)
		}
	}

	// Calculate leaked tokens
	timePassed := now.Sub(lastLeakTime)
	leakedTokens := int(timePassed.Seconds()) * (config.Limit / int(config.Window.Seconds()))

	if leakedTokens > 0 {
		currentLevel = max(0, currentLevel-leakedTokens)
		lastLeakTime = now
	}

	// Check if bucket has space
	if currentLevel >= config.Limit {
		// Update bucket state
		rl.redis.HMSet(ctx, key, map[string]interface{}{
			"level":     currentLevel,
			"last_leak": lastLeakTime.Unix(),
		})
		rl.redis.Expire(ctx, key, config.Window)

		return false, 0, now.Add(config.Window).Unix(), nil
	}

	// Add request to bucket
	currentLevel++

	// Update bucket state
	rl.redis.HMSet(ctx, key, map[string]interface{}{
		"level":     currentLevel,
		"last_leak": lastLeakTime.Unix(),
	})
	rl.redis.Expire(ctx, key, config.Window)

	remaining := config.Limit - currentLevel
	return true, remaining, now.Add(config.Window).Unix(), nil
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Predefined rate limit configurations
func (rl *RateLimiter) AuthRateLimit() gin.HandlerFunc {
	return rl.RateLimitWithConfig(RateLimitConfig{
		Limit:    5,
		Window:   time.Minute,
		Strategy: FixedWindow,
		KeyFunc:  DefaultKeyFunc,
	})
}

func (rl *RateLimiter) APIRateLimit() gin.HandlerFunc {
	return rl.RateLimitWithConfig(RateLimitConfig{
		Limit:    100,
		Window:   time.Minute,
		Strategy: SlidingWindow,
		KeyFunc:  DefaultKeyFunc,
	})
}

func (rl *RateLimiter) UserRateLimit() gin.HandlerFunc {
	return rl.RateLimitWithConfig(RateLimitConfig{
		Limit:    1000,
		Window:   time.Hour,
		Strategy: TokenBucket,
		KeyFunc:  UserKeyFunc,
	})
}

func (rl *RateLimiter) UploadRateLimit() gin.HandlerFunc {
	return rl.RateLimitWithConfig(RateLimitConfig{
		Limit:    10,
		Window:   time.Minute,
		Strategy: LeakyBucket,
		KeyFunc:  UserKeyFunc,
	})
}
