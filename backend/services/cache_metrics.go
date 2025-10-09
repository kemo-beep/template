package services

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type CacheMetricsService struct {
	redis *redis.Client
}

type CacheMetrics struct {
	HitRatio        float64            `json:"hit_ratio"`
	MissRatio       float64            `json:"miss_ratio"`
	TotalRequests   int64              `json:"total_requests"`
	TotalKeys       int64              `json:"total_keys"`
	MemoryUsage     int64              `json:"memory_usage"`
	AvgResponseTime float64            `json:"avg_response_time"`
	StrategyStats   map[string]float64 `json:"strategy_stats"`
	TopKeys         []KeyStats         `json:"top_keys"`
}

type KeyStats struct {
	Key         string `json:"key"`
	AccessCount int64  `json:"access_count"`
	LastAccess  string `json:"last_access"`
}

func NewCacheMetricsService(redis *redis.Client) *CacheMetricsService {
	return &CacheMetricsService{redis: redis}
}

// Get comprehensive cache metrics
func (cms *CacheMetricsService) GetCacheMetrics(ctx context.Context) (*CacheMetrics, error) {
	// Get basic stats
	hitCount, _ := cms.redis.Get(ctx, "cache:hits").Int64()
	missCount, _ := cms.redis.Get(ctx, "cache:misses").Int64()
	totalKeys, _ := cms.redis.DBSize(ctx).Result()
	avgResponseTime, _ := cms.redis.Get(ctx, "cache:avg_response_time").Float64()

	totalRequests := hitCount + missCount
	hitRatio := 0.0
	missRatio := 0.0
	if totalRequests > 0 {
		hitRatio = float64(hitCount) / float64(totalRequests)
		missRatio = float64(missCount) / float64(totalRequests)
	}

	// Get memory usage
	memoryUsage, _ := cms.redis.MemoryUsage(ctx, "cache:*").Result()

	// Get strategy effectiveness
	strategyStats := cms.getStrategyStats(ctx)

	// Get top accessed keys
	topKeys := cms.getTopKeys(ctx)

	return &CacheMetrics{
		HitRatio:        hitRatio,
		MissRatio:       missRatio,
		TotalRequests:   totalRequests,
		TotalKeys:       totalKeys,
		MemoryUsage:     memoryUsage,
		AvgResponseTime: avgResponseTime,
		StrategyStats:   strategyStats,
		TopKeys:         topKeys,
	}, nil
}

// Get strategy effectiveness stats
func (cms *CacheMetricsService) getStrategyStats(ctx context.Context) map[string]float64 {
	// This would typically come from actual performance data
	// For now, return example values
	return map[string]float64{
		"cache_aside":   0.85,
		"write_through": 0.90,
		"write_behind":  0.95,
		"refresh_ahead": 0.88,
	}
}

// Get top accessed keys
func (cms *CacheMetricsService) getTopKeys(ctx context.Context) []KeyStats {
	// Get keys with access counts (simplified implementation)
	keys, _ := cms.redis.Keys(ctx, "cache:access:*").Result()

	var keyStats []KeyStats
	for _, key := range keys {
		accessCount, _ := cms.redis.Get(ctx, key).Int64()
		lastAccess, _ := cms.redis.Get(ctx, key+":last_access").Result()

		keyStats = append(keyStats, KeyStats{
			Key:         key,
			AccessCount: accessCount,
			LastAccess:  lastAccess,
		})
	}

	// Sort by access count (simplified)
	// In production, you'd implement proper sorting
	return keyStats[:min(len(keyStats), 10)] // Return top 10
}

// Track key access
func (cms *CacheMetricsService) TrackKeyAccess(ctx context.Context, key string) {
	accessKey := "cache:access:" + key
	cms.redis.Incr(ctx, accessKey)
	cms.redis.Set(ctx, accessKey+":last_access", time.Now().Format(time.RFC3339), 24*time.Hour)
}

// Get cache health status
func (cms *CacheMetricsService) GetCacheHealth(ctx context.Context) map[string]interface{} {
	metrics, _ := cms.GetCacheMetrics(ctx)

	health := map[string]interface{}{
		"status": "healthy",
		"score":  100,
	}

	// Check hit ratio
	if metrics.HitRatio < 0.5 {
		health["status"] = "warning"
		health["score"] = 70
	}
	if metrics.HitRatio < 0.3 {
		health["status"] = "critical"
		health["score"] = 30
	}

	// Check response time
	if metrics.AvgResponseTime > 0.1 { // 100ms
		health["status"] = "warning"
		health["score"] = min(health["score"].(int), 80)
	}
	if metrics.AvgResponseTime > 0.5 { // 500ms
		health["status"] = "critical"
		health["score"] = min(health["score"].(int), 40)
	}

	// Check memory usage
	if metrics.MemoryUsage > 100*1024*1024 { // 100MB
		health["status"] = "warning"
		health["score"] = min(health["score"].(int), 85)
	}

	health["metrics"] = metrics
	return health
}

// Reset cache metrics
func (cms *CacheMetricsService) ResetMetrics(ctx context.Context) error {
	cms.redis.Del(ctx, "cache:hits", "cache:misses", "cache:avg_response_time")
	cms.redis.Del(ctx, "cache:response_times")
	return nil
}

// Get cache recommendations
func (cms *CacheMetricsService) GetRecommendations(ctx context.Context) []string {
	metrics, _ := cms.GetCacheMetrics(ctx)
	var recommendations []string

	if metrics.HitRatio < 0.5 {
		recommendations = append(recommendations, "Consider increasing cache TTL or improving cache key strategy")
	}

	if metrics.AvgResponseTime > 0.1 {
		recommendations = append(recommendations, "Cache response time is high, consider optimizing Redis configuration")
	}

	if metrics.MemoryUsage > 50*1024*1024 {
		recommendations = append(recommendations, "High memory usage detected, consider implementing cache eviction policies")
	}

	if len(metrics.TopKeys) < 5 {
		recommendations = append(recommendations, "Low key diversity, consider implementing cache warming strategies")
	}

	return recommendations
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
