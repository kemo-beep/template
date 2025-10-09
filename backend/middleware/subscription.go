package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"mobile-backend/models"
	"mobile-backend/utils"
)

// SubscriptionMiddleware handles subscription validation
type SubscriptionMiddleware struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSubscriptionMiddleware creates a new subscription middleware
func NewSubscriptionMiddleware(db *gorm.DB, logger *zap.Logger) *SubscriptionMiddleware {
	return &SubscriptionMiddleware{
		db:     db,
		logger: logger,
	}
}

// RequireProSubscription middleware that requires pro subscription
func (sm *SubscriptionMiddleware) RequireProSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
			c.Abort()
			return
		}

		userIDUint, ok := userID.(uint)
		if !ok {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
			c.Abort()
			return
		}

		// Get user with subscription info
		var user models.User
		if err := sm.db.Preload("ActiveSubscription").First(&user, userIDUint).Error; err != nil {
			sm.logger.Error("Failed to get user", zap.Error(err), zap.Uint("user_id", userIDUint))
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get user", nil)
			c.Abort()
			return
		}

		// Check if user has pro access
		if !user.IsProUser() {
			utils.SendErrorResponse(c, http.StatusForbidden, "Pro subscription required", map[string]interface{}{
				"subscription_status": user.SubscriptionStatus,
				"is_pro":              user.IsPro,
				"upgrade_required":    true,
			})
			c.Abort()
			return
		}

		// Set user info in context for use in handlers
		c.Set("user", user)
		c.Set("subscription_status", user.SubscriptionStatus)
		c.Set("is_pro", user.IsPro)

		c.Next()
	}
}

// RequireActiveSubscription middleware that requires any active subscription (pro or trial)
func (sm *SubscriptionMiddleware) RequireActiveSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
			c.Abort()
			return
		}

		userIDUint, ok := userID.(uint)
		if !ok {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
			c.Abort()
			return
		}

		// Get user with subscription info
		var user models.User
		if err := sm.db.Preload("ActiveSubscription").First(&user, userIDUint).Error; err != nil {
			sm.logger.Error("Failed to get user", zap.Error(err), zap.Uint("user_id", userIDUint))
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get user", nil)
			c.Abort()
			return
		}

		// Check if user has active subscription
		if !user.IsSubscriptionActive() {
			utils.SendErrorResponse(c, http.StatusForbidden, "Active subscription required", map[string]interface{}{
				"subscription_status":   user.SubscriptionStatus,
				"is_pro":                user.IsPro,
				"subscription_required": true,
			})
			c.Abort()
			return
		}

		// Set user info in context for use in handlers
		c.Set("user", user)
		c.Set("subscription_status", user.SubscriptionStatus)
		c.Set("is_pro", user.IsPro)

		c.Next()
	}
}

// OptionalSubscription middleware that adds subscription info but doesn't require it
func (sm *SubscriptionMiddleware) OptionalSubscription() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userIDUint, ok := userID.(uint)
		if !ok {
			c.Next()
			return
		}

		// Get user with subscription info
		var user models.User
		if err := sm.db.Preload("ActiveSubscription").First(&user, userIDUint).Error; err != nil {
			sm.logger.Error("Failed to get user", zap.Error(err), zap.Uint("user_id", userIDUint))
			c.Next()
			return
		}

		// Set user info in context for use in handlers
		c.Set("user", user)
		c.Set("subscription_status", user.SubscriptionStatus)
		c.Set("is_pro", user.IsPro)

		c.Next()
	}
}

// CheckSubscriptionLimits middleware that checks subscription limits
func (sm *SubscriptionMiddleware) CheckSubscriptionLimits(limitType string, limitValue int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
			c.Abort()
			return
		}

		userIDUint, ok := userID.(uint)
		if !ok {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
			c.Abort()
			return
		}

		// Get user with subscription info
		var user models.User
		if err := sm.db.Preload("ActiveSubscription").First(&user, userIDUint).Error; err != nil {
			sm.logger.Error("Failed to get user", zap.Error(err), zap.Uint("user_id", userIDUint))
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get user", nil)
			c.Abort()
			return
		}

		// Check limits based on subscription status
		var allowedLimit int
		switch user.SubscriptionStatus {
		case "active", "trial":
			allowedLimit = limitValue
		case "free":
			// Free users get reduced limits
			allowedLimit = limitValue / 4
		default:
			allowedLimit = 0
		}

		// Set limit info in context
		c.Set("subscription_limit", allowedLimit)
		c.Set("subscription_status", user.SubscriptionStatus)
		c.Set("is_pro", user.IsPro)

		c.Next()
	}
}

// SubscriptionStatusResponse adds subscription status to response
func (sm *SubscriptionMiddleware) SubscriptionStatusResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}

		userIDUint, ok := userID.(uint)
		if !ok {
			c.Next()
			return
		}

		// Get user with subscription info
		var user models.User
		if err := sm.db.Preload("ActiveSubscription").First(&user, userIDUint).Error; err != nil {
			sm.logger.Error("Failed to get user", zap.Error(err), zap.Uint("user_id", userIDUint))
			c.Next()
			return
		}

		// Add subscription info to response headers
		c.Header("X-Subscription-Status", user.SubscriptionStatus)
		c.Header("X-Is-Pro", strconv.FormatBool(user.IsPro))
		c.Header("X-Subscription-Ends-At", user.SubscriptionEndsAt.Format("2006-01-02T15:04:05Z"))

		c.Next()
	}
}
