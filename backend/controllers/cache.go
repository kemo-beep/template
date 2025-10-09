package controllers

import (
	"net/http"
	"time"

	"mobile-backend/services"
	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
)

type CacheController struct {
	cacheService        *services.CacheService
	cacheMetricsService *services.CacheMetricsService
}

func NewCacheController(cacheService *services.CacheService, cacheMetricsService *services.CacheMetricsService) *CacheController {
	return &CacheController{
		cacheService:        cacheService,
		cacheMetricsService: cacheMetricsService,
	}
}

type CacheStatsResponse struct {
	Stats map[string]interface{} `json:"stats"`
}

type CacheKeyRequest struct {
	Key string `json:"key" binding:"required"`
}

type CacheSetRequest struct {
	Key        string      `json:"key" binding:"required"`
	Value      interface{} `json:"value" binding:"required"`
	Expiration int         `json:"expiration"` // seconds
}

type CacheInvalidateRequest struct {
	Pattern string `json:"pattern" binding:"required"`
}

type CacheWarmRequest struct {
	Keys       []string `json:"keys" binding:"required"`
	Expiration int      `json:"expiration"` // seconds
}

// GetCacheStats godoc
// @Summary Get cache statistics
// @Description Get Redis cache statistics and performance metrics
// @Tags cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=CacheStatsResponse}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/cache/stats [get]
func (cc *CacheController) GetCacheStats(c *gin.Context) {
	stats, err := cc.cacheService.GetStats(c.Request.Context())
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to get cache stats")
		return
	}

	response := CacheStatsResponse{Stats: stats}
	utils.SendSuccessResponse(c, response, "Cache statistics retrieved successfully")
}

// GetCacheKey godoc
// @Summary Get cache key
// @Description Get value from cache by key
// @Tags cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param key query string true "Cache key"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/cache/key [get]
func (cc *CacheController) GetCacheKey(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Key parameter is required", nil)
		return
	}

	var value interface{}
	err := cc.cacheService.Get(c.Request.Context(), key, &value)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Key not found", nil)
		return
	}

	utils.SendSuccessResponse(c, value, "Cache key retrieved successfully")
}

// GetCacheMetrics godoc
// @Summary Get comprehensive cache metrics
// @Description Get detailed cache performance metrics and statistics
// @Tags cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/cache/metrics [get]
func (cc *CacheController) GetCacheMetrics(c *gin.Context) {
	metrics, err := cc.cacheMetricsService.GetCacheMetrics(c.Request.Context())
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to get cache metrics")
		return
	}

	utils.SendSuccessResponse(c, metrics, "Cache metrics retrieved successfully")
}

// GetCacheHealth godoc
// @Summary Get cache health status
// @Description Get cache health status and recommendations
// @Tags cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/cache/health [get]
func (cc *CacheController) GetCacheHealth(c *gin.Context) {
	health := cc.cacheMetricsService.GetCacheHealth(c.Request.Context())
	utils.SendSuccessResponse(c, health, "Cache health status retrieved successfully")
}

// GetCacheRecommendations godoc
// @Summary Get cache optimization recommendations
// @Description Get recommendations for optimizing cache performance
// @Tags cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/cache/recommendations [get]
func (cc *CacheController) GetCacheRecommendations(c *gin.Context) {
	recommendations := cc.cacheMetricsService.GetRecommendations(c.Request.Context())
	utils.SendSuccessResponse(c, recommendations, "Cache recommendations retrieved successfully")
}

// ResetCacheMetrics godoc
// @Summary Reset cache metrics
// @Description Reset all cache performance metrics
// @Tags cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/cache/metrics/reset [post]
func (cc *CacheController) ResetCacheMetrics(c *gin.Context) {
	err := cc.cacheMetricsService.ResetMetrics(c.Request.Context())
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to reset cache metrics")
		return
	}

	utils.SendSuccessResponse(c, nil, "Cache metrics reset successfully")
}

// SetCacheKey godoc
// @Summary Set cache key
// @Description Set value in cache with optional expiration
// @Tags cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CacheSetRequest true "Cache set data"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/cache/key [post]
func (cc *CacheController) SetCacheKey(c *gin.Context) {
	var req CacheSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationErrorResponse(c, parseValidationErrors(err))
		return
	}

	expiration := time.Duration(req.Expiration) * time.Second
	if expiration == 0 {
		expiration = time.Hour // Default expiration
	}

	err := cc.cacheService.Set(c.Request.Context(), req.Key, req.Value, expiration)
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to set cache key")
		return
	}

	utils.SendSuccessResponse(c, nil, "Cache key set successfully")
}

// DeleteCacheKey godoc
// @Summary Delete cache key
// @Description Delete cache key by pattern
// @Tags cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CacheInvalidateRequest true "Cache invalidation data"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/cache/invalidate [post]
func (cc *CacheController) InvalidateCache(c *gin.Context) {
	var req CacheInvalidateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationErrorResponse(c, parseValidationErrors(err))
		return
	}

	err := cc.cacheService.InvalidatePattern(c.Request.Context(), req.Pattern)
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to invalidate cache")
		return
	}

	utils.SendSuccessResponse(c, nil, "Cache invalidated successfully")
}

// WarmCache godoc
// @Summary Warm cache
// @Description Pre-populate cache with data
// @Tags cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CacheWarmRequest true "Cache warming data"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/cache/warm [post]
func (cc *CacheController) WarmCache(c *gin.Context) {
	var req CacheWarmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationErrorResponse(c, parseValidationErrors(err))
		return
	}

	expiration := time.Duration(req.Expiration) * time.Second
	if expiration == 0 {
		expiration = time.Hour // Default expiration
	}

	// Simple fetch function that returns the key as value
	fetchFunc := func(key string) (interface{}, error) {
		return map[string]string{"key": key, "value": "cached_data"}, nil
	}

	err := cc.cacheService.WarmCache(c.Request.Context(), req.Keys, fetchFunc, expiration)
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to warm cache")
		return
	}

	utils.SendSuccessResponse(c, nil, "Cache warmed successfully")
}

// ClearCache godoc
// @Summary Clear all cache
// @Description Clear all cache data (use with caution)
// @Tags cache
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/cache/clear [post]
func (cc *CacheController) ClearCache(c *gin.Context) {
	err := cc.cacheService.InvalidatePattern(c.Request.Context(), "*")
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to clear cache")
		return
	}

	utils.SendSuccessResponse(c, nil, "Cache cleared successfully")
}
