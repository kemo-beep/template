package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"mobile-backend/models"
	"mobile-backend/services"
	"mobile-backend/utils"
)

// SubscriptionManagementController handles subscription management operations
type SubscriptionManagementController struct {
	db                        *gorm.DB
	subscriptionStatusService *services.SubscriptionStatusService
	stripeService             *services.StripeService
	polarService              *services.PolarService
	logger                    *zap.Logger
}

// NewSubscriptionManagementController creates a new subscription management controller
func NewSubscriptionManagementController(
	db *gorm.DB,
	subscriptionStatusService *services.SubscriptionStatusService,
	stripeService *services.StripeService,
	polarService *services.PolarService,
	logger *zap.Logger,
) *SubscriptionManagementController {
	return &SubscriptionManagementController{
		db:                        db,
		subscriptionStatusService: subscriptionStatusService,
		stripeService:             stripeService,
		polarService:              polarService,
		logger:                    logger,
	}
}

// GetUserSubscriptionStatus godoc
// @Summary Get user subscription status
// @Description Get current user's subscription status and details
// @Tags subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/subscription/status [get]
func (smc *SubscriptionManagementController) GetUserSubscriptionStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	user, err := smc.subscriptionStatusService.GetUserSubscriptionStatus(c.Request.Context(), userIDUint)
	if err != nil {
		smc.logger.Error("Failed to get user subscription status", zap.Error(err), zap.Uint("user_id", userIDUint))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get subscription status", nil)
		return
	}

	// Prepare response data
	responseData := map[string]interface{}{
		"subscription_status":    user.SubscriptionStatus,
		"is_pro":                 user.IsPro,
		"subscription_ends_at":   user.SubscriptionEndsAt,
		"trial_ends_at":          user.TrialEndsAt,
		"status_display":         user.GetSubscriptionStatus(),
		"has_trial_access":       user.HasTrialAccess(),
		"is_subscription_active": user.IsSubscriptionActive(),
	}

	// Add subscription details if available
	if user.ActiveSubscription != nil {
		responseData["subscription"] = map[string]interface{}{
			"id":                   user.ActiveSubscription.ID,
			"status":               user.ActiveSubscription.Status,
			"current_period_start": user.ActiveSubscription.CurrentPeriodStart,
			"current_period_end":   user.ActiveSubscription.CurrentPeriodEnd,
			"cancel_at_period_end": user.ActiveSubscription.CancelAtPeriodEnd,
			"canceled_at":          user.ActiveSubscription.CanceledAt,
			"trial_start":          user.ActiveSubscription.TrialStart,
			"trial_end":            user.ActiveSubscription.TrialEnd,
			"quantity":             user.ActiveSubscription.Quantity,
			"payment_method":       user.ActiveSubscription.PaymentMethod,
		}
	}

	utils.SendSuccessResponse(c, responseData, "Subscription status retrieved successfully")
}

// GetSubscriptionPlans godoc
// @Summary Get available subscription plans
// @Description Get list of available subscription plans
// @Tags subscription
// @Accept json
// @Produce json
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/subscription/plans [get]
func (smc *SubscriptionManagementController) GetSubscriptionPlans(c *gin.Context) {
	var plans []models.Plan
	if err := smc.db.Preload("Product").Where("is_active = ?", true).Find(&plans).Error; err != nil {
		smc.logger.Error("Failed to get subscription plans", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get subscription plans", nil)
		return
	}

	// Format plans for response
	var formattedPlans []map[string]interface{}
	for _, plan := range plans {
		formattedPlan := map[string]interface{}{
			"id":             plan.ID,
			"name":           plan.Name,
			"description":    plan.Description,
			"price":          plan.Price,
			"price_dollars":  plan.GetPriceInDollars(),
			"currency":       plan.Currency,
			"interval":       plan.Interval,
			"interval_count": plan.IntervalCount,
			"trial_days":     plan.TrialDays,
			"is_active":      plan.IsActive,
			"product": map[string]interface{}{
				"id":          plan.Product.ID,
				"name":        plan.Product.Name,
				"description": plan.Product.Description,
			},
		}
		formattedPlans = append(formattedPlans, formattedPlan)
	}

	utils.SendSuccessResponse(c, formattedPlans, "Subscription plans retrieved successfully")
}

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Create a new subscription for the current user
// @Tags subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateSubscriptionRequest true "Subscription creation data"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/subscription [post]
func (smc *SubscriptionManagementController) CreateSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Check if user already has an active subscription
	var existingSubscription models.Subscription
	if err := smc.db.Where("user_id = ? AND status IN ?", userIDUint, []string{"active", "trialing"}).First(&existingSubscription).Error; err == nil {
		utils.SendErrorResponse(c, http.StatusConflict, "User already has an active subscription", map[string]interface{}{
			"existing_subscription_id": existingSubscription.ID,
		})
		return
	}

	// Create subscription based on payment method
	var subscription *models.Subscription
	var err error

	switch req.PaymentMethod {
	case "stripe":
		subscription, err = smc.stripeService.CreateSubscription(c.Request.Context(), userIDUint, req.PlanID, req.PaymentMethod)
	case "polar":
		subscription, err = smc.polarService.CreateSubscription(c.Request.Context(), userIDUint, req.PlanID)
	default:
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payment method", nil)
		return
	}

	if err != nil {
		smc.logger.Error("Failed to create subscription", zap.Error(err), zap.Uint("user_id", userIDUint), zap.Uint("plan_id", req.PlanID))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create subscription", map[string]interface{}{"error": err.Error()})
		return
	}

	// Update user subscription status
	if err := smc.subscriptionStatusService.UpdateUserSubscriptionStatus(c.Request.Context(), userIDUint, subscription); err != nil {
		smc.logger.Error("Failed to update user subscription status", zap.Error(err), zap.Uint("user_id", userIDUint))
		// Don't fail the request, just log the error
	}

	utils.SendSuccessResponse(c, subscription, "Subscription created successfully")
}

// CancelSubscription godoc
// @Summary Cancel user subscription
// @Description Cancel the current user's subscription
// @Tags subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CancelSubscriptionRequest true "Cancellation data"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/subscription/cancel [post]
func (smc *SubscriptionManagementController) CancelSubscription(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	var req CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user's active subscription
	var subscription models.Subscription
	if err := smc.db.Where("user_id = ? AND status IN ?", userIDUint, []string{"active", "trialing"}).First(&subscription).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "No active subscription found", nil)
		return
	}

	// Cancel subscription based on payment method
	var err error
	switch subscription.PaymentMethod {
	case "stripe":
		err = smc.stripeService.CancelSubscription(c.Request.Context(), subscription.ID, req.Immediately)
	case "polar":
		err = smc.polarService.CancelSubscription(c.Request.Context(), subscription.ID, req.Immediately)
	default:
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payment method", nil)
		return
	}

	if err != nil {
		smc.logger.Error("Failed to cancel subscription", zap.Error(err), zap.Uint("subscription_id", subscription.ID))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to cancel subscription", map[string]interface{}{"error": err.Error()})
		return
	}

	// Update user subscription status
	if err := smc.subscriptionStatusService.UpdateUserSubscriptionStatus(c.Request.Context(), userIDUint, &subscription); err != nil {
		smc.logger.Error("Failed to update user subscription status", zap.Error(err), zap.Uint("user_id", userIDUint))
		// Don't fail the request, just log the error
	}

	utils.SendSuccessResponse(c, nil, "Subscription canceled successfully")
}

// GetSubscriptionHistory godoc
// @Summary Get user subscription history
// @Description Get the current user's subscription history
// @Tags subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/subscription/history [get]
func (smc *SubscriptionManagementController) GetSubscriptionHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	var subscriptions []models.Subscription
	if err := smc.db.Preload("Product").Preload("Plan").
		Where("user_id = ?", userIDUint).
		Order("created_at DESC").
		Find(&subscriptions).Error; err != nil {
		smc.logger.Error("Failed to get subscription history", zap.Error(err), zap.Uint("user_id", userIDUint))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get subscription history", nil)
		return
	}

	utils.SendSuccessResponse(c, subscriptions, "Subscription history retrieved successfully")
}

// GetSubscriptionStats godoc
// @Summary Get subscription statistics
// @Description Get subscription statistics (admin only)
// @Tags subscription
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/subscription/stats [get]
func (smc *SubscriptionManagementController) GetSubscriptionStats(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		utils.SendErrorResponse(c, http.StatusForbidden, "Admin access required", nil)
		return
	}

	stats, err := smc.subscriptionStatusService.GetSubscriptionStats(c.Request.Context())
	if err != nil {
		smc.logger.Error("Failed to get subscription stats", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get subscription stats", nil)
		return
	}

	utils.SendSuccessResponse(c, stats, "Subscription statistics retrieved successfully")
}

// Request Types are defined in payment.go to avoid duplication
