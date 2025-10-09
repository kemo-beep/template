package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type CacheStrategy string

const (
	CacheAside   CacheStrategy = "cache_aside"
	WriteThrough CacheStrategy = "write_through"
	WriteBehind  CacheStrategy = "write_behind"
	RefreshAhead CacheStrategy = "refresh_ahead"
)

type CacheService struct {
	redis    *redis.Client
	strategy CacheStrategy
}

func NewCacheService(redis *redis.Client) *CacheService {
	return &CacheService{
		redis:    redis,
		strategy: CacheAside, // Default strategy
	}
}

func NewCacheServiceWithStrategy(redis *redis.Client, strategy CacheStrategy) *CacheService {
	return &CacheService{
		redis:    redis,
		strategy: strategy,
	}
}

// Basic cache operations
func (c *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()
	defer func() {
		c.trackPerformance(ctx, "set", time.Since(start))
	}()

	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, key, jsonData, expiration).Err()
}

func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	start := time.Now()
	defer func() {
		c.trackPerformance(ctx, "get", time.Since(start))
	}()

	val, err := c.redis.Get(ctx, key).Result()
	if err != nil {
		// Track cache miss
		c.redis.Incr(ctx, "cache:misses")
		return err
	}
	// Track cache hit
	c.redis.Incr(ctx, "cache:hits")
	return json.Unmarshal([]byte(val), dest)
}

func (c *CacheService) Delete(ctx context.Context, key string) error {
	return c.redis.Del(ctx, key).Err()
}

func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	result := c.redis.Exists(ctx, key)
	return result.Val() > 0, result.Err()
}

func (c *CacheService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.redis.Expire(ctx, key, expiration).Err()
}

// Advanced cache operations
func (c *CacheService) GetOrSet(ctx context.Context, key string, dest interface{},
	fetchFunc func() (interface{}, error), expiration time.Duration) error {

	err := c.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	if err != redis.Nil {
		return err
	}

	data, err := fetchFunc()
	if err != nil {
		return err
	}

	err = c.Set(ctx, key, data, expiration)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data.(string)), dest)
}

// Cache-aside pattern
func (c *CacheService) CacheAside(ctx context.Context, key string, dest interface{},
	fetchFunc func() (interface{}, error), expiration time.Duration) error {

	// Try to get from cache first
	err := c.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	if err != redis.Nil {
		return err
	}

	// Cache miss - fetch from source
	data, err := fetchFunc()
	if err != nil {
		return err
	}

	// Store in cache for next time
	c.Set(ctx, key, data, expiration)

	// Return the data
	return json.Unmarshal([]byte(data.(string)), dest)
}

// Write-through pattern
func (c *CacheService) WriteThrough(ctx context.Context, key string, value interface{},
	writeFunc func(interface{}) error, expiration time.Duration) error {

	// Write to source first
	err := writeFunc(value)
	if err != nil {
		return err
	}

	// Then write to cache
	return c.Set(ctx, key, value, expiration)
}

// Cache invalidation
func (c *CacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	keys, err := c.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.redis.Del(ctx, keys...).Err()
	}

	return nil
}

// Cache warming
func (c *CacheService) WarmCache(ctx context.Context, keys []string,
	fetchFunc func(string) (interface{}, error), expiration time.Duration) error {

	for _, key := range keys {
		data, err := fetchFunc(key)
		if err != nil {
			continue // Skip failed keys
		}

		c.Set(ctx, key, data, expiration)
	}

	return nil
}

// Cache statistics
func (c *CacheService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	info, err := c.redis.Info(ctx, "memory").Result()
	if err != nil {
		return nil, err
	}

	// Get cache metrics
	keyspace, err := c.redis.Info(ctx, "keyspace").Result()
	if err != nil {
		keyspace = "unavailable"
	}

	// Get hit/miss ratios
	statsInfo, err := c.redis.Info(ctx, "stats").Result()
	if err != nil {
		statsInfo = "unavailable"
	}

	// Parse Redis info (simplified)
	stats := map[string]interface{}{
		"redis_info":    info,
		"keyspace_info": keyspace,
		"stats_info":    statsInfo,
		"strategy":      c.strategy,
		"cache_metrics": c.getCacheMetrics(ctx),
		"performance":   c.getPerformanceMetrics(ctx),
	}

	return stats, nil
}

// Get detailed cache metrics
func (c *CacheService) getCacheMetrics(ctx context.Context) map[string]interface{} {
	// Get total keys
	totalKeys, _ := c.redis.DBSize(ctx).Result()

	// Get memory usage
	memoryUsage, _ := c.redis.MemoryUsage(ctx, "cache:*").Result()

	// Get cache hit/miss ratio (simplified)
	hitCount, _ := c.redis.Get(ctx, "cache:hits").Int64()
	missCount, _ := c.redis.Get(ctx, "cache:misses").Int64()

	totalRequests := hitCount + missCount
	hitRatio := 0.0
	if totalRequests > 0 {
		hitRatio = float64(hitCount) / float64(totalRequests)
	}

	return map[string]interface{}{
		"total_keys":     totalKeys,
		"memory_usage":   memoryUsage,
		"hit_count":      hitCount,
		"miss_count":     missCount,
		"hit_ratio":      hitRatio,
		"total_requests": totalRequests,
	}
}

// Get performance metrics
func (c *CacheService) getPerformanceMetrics(ctx context.Context) map[string]interface{} {
	// Get average response time (simplified)
	avgResponseTime, _ := c.redis.Get(ctx, "cache:avg_response_time").Float64()

	// Get cache efficiency by strategy
	efficiency := map[string]float64{
		"cache_aside":   0.85, // Example values
		"write_through": 0.90,
		"write_behind":  0.95,
		"refresh_ahead": 0.88,
	}

	return map[string]interface{}{
		"avg_response_time":   avgResponseTime,
		"strategy_efficiency": efficiency[string(c.strategy)],
		"cache_efficiency":    efficiency,
	}
}

// Cache with tags for easier invalidation
func (c *CacheService) SetWithTags(ctx context.Context, key string, value interface{},
	tags []string, expiration time.Duration) error {

	// Store the data
	err := c.Set(ctx, key, value, expiration)
	if err != nil {
		return err
	}

	// Store tags for this key
	for _, tag := range tags {
		tagKey := fmt.Sprintf("tag:%s", tag)
		c.redis.SAdd(ctx, tagKey, key)
		c.redis.Expire(ctx, tagKey, expiration)
	}

	return nil
}

// Invalidate by tags
func (c *CacheService) InvalidateByTags(ctx context.Context, tags []string) error {
	for _, tag := range tags {
		tagKey := fmt.Sprintf("tag:%s", tag)
		keys, err := c.redis.SMembers(ctx, tagKey).Result()
		if err != nil {
			continue
		}

		if len(keys) > 0 {
			c.redis.Del(ctx, keys...)
		}

		c.redis.Del(ctx, tagKey)
	}

	return nil
}

// Cache with compression for large values
func (c *CacheService) SetCompressed(ctx context.Context, key string, value interface{},
	expiration time.Duration) error {

	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// For now, just store as JSON
	// In production, you'd compress the data here
	return c.redis.Set(ctx, key, jsonData, expiration).Err()
}

// Distributed locking
func (c *CacheService) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	result := c.redis.SetNX(ctx, "lock:"+key, "1", expiration)
	return result.Val(), result.Err()
}

func (c *CacheService) Unlock(ctx context.Context, key string) error {
	return c.redis.Del(ctx, "lock:"+key).Err()
}

func (c *CacheService) Increment(ctx context.Context, key string) (int64, error) {
	return c.redis.Incr(ctx, key).Result()
}

// Track cache performance
func (c *CacheService) trackPerformance(ctx context.Context, operation string, duration time.Duration) {
	// Update average response time
	c.redis.ZAdd(ctx, "cache:response_times", &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: duration.Seconds(),
	})

	// Keep only last 1000 entries
	c.redis.ZRemRangeByRank(ctx, "cache:response_times", 0, -1001)

	// Calculate and store average
	responseTimes, _ := c.redis.ZRange(ctx, "cache:response_times", 0, -1).Result()
	if len(responseTimes) > 0 {
		var total float64
		for _, rt := range responseTimes {
			if val, err := strconv.ParseFloat(rt, 64); err == nil {
				total += val
			}
		}
		avg := total / float64(len(responseTimes))
		c.redis.Set(ctx, "cache:avg_response_time", avg, 0)
	}
}

func (c *CacheService) SetExpiration(ctx context.Context, key string, expiration time.Duration) error {
	return c.redis.Expire(ctx, key, expiration).Err()
}
