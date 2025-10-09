package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"mobile-backend/models"
	"mobile-backend/services"
	"mobile-backend/utils"
)

// PushNotificationController handles push notification API endpoints
type PushNotificationController struct {
	pushService *services.PushNotificationService
	logger      *zap.Logger
}

// NewPushNotificationController creates a new push notification controller
func NewPushNotificationController(pushService *services.PushNotificationService, logger *zap.Logger) *PushNotificationController {
	return &PushNotificationController{
		pushService: pushService,
		logger:      logger,
	}
}

// SendNotificationRequest represents a request to send a notification
type SendNotificationRequest struct {
	Title       string                    `json:"title" binding:"required"`
	Body        string                    `json:"body" binding:"required"`
	Data        map[string]interface{}    `json:"data"`
	Target      models.NotificationTarget `json:"target" binding:"required"`
	ScheduledAt *time.Time                `json:"scheduled_at"`
	ExpiresAt   *time.Time                `json:"expires_at"`
	Priority    string                    `json:"priority"`
	Platform    string                    `json:"platform"`
	TemplateID  *uint                     `json:"template_id"`
	UserID      *uint                     `json:"user_id"`
	SegmentID   *uint                     `json:"segment_id"`
}

// SendNotificationResponse represents a response to send notification
type SendNotificationResponse struct {
	NotificationID uint   `json:"notification_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

// CreateTemplateRequest represents a request to create a template
type CreateTemplateRequest struct {
	Name      string                 `json:"name" binding:"required"`
	Title     string                 `json:"title" binding:"required"`
	Body      string                 `json:"body" binding:"required"`
	Data      map[string]interface{} `json:"data"`
	Platform  string                 `json:"platform"`
	Category  string                 `json:"category"`
	Variables []string               `json:"variables"`
}

// CreateSegmentRequest represents a request to create a segment
type CreateSegmentRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Query       string `json:"query" binding:"required"`
}

// RegisterDeviceRequest represents a request to register a device
type RegisterDeviceRequest struct {
	Token      string     `json:"token" binding:"required"`
	Platform   string     `json:"platform" binding:"required"`
	DeviceID   string     `json:"device_id"`
	AppVersion string     `json:"app_version"`
	OSVersion  string     `json:"os_version"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

// SendNotification sends a push notification
func (pnc *PushNotificationController) SendNotification(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	_, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	var req SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error(), nil)
		return
	}

	// Create notification
	notification := &models.PushNotification{
		Title:       req.Title,
		Body:        req.Body,
		Data:        req.Data,
		Target:      req.Target,
		ScheduledAt: req.ScheduledAt,
		ExpiresAt:   req.ExpiresAt,
		Priority:    req.Priority,
		Platform:    req.Platform,
		TemplateID:  req.TemplateID,
		UserID:      req.UserID,
		SegmentID:   req.SegmentID,
	}

	// Set defaults
	if notification.Priority == "" {
		notification.Priority = models.NotificationPriorityNormal
	}
	if notification.Platform == "" {
		notification.Platform = models.PlatformAll
	}

	// Send or schedule notification
	var err error
	if notification.ScheduledAt != nil {
		err = pnc.pushService.ScheduleNotification(c.Request.Context(), notification)
	} else {
		err = pnc.pushService.SendNotification(c.Request.Context(), notification)
	}

	if err != nil {
		pnc.logger.Error("Failed to send notification", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to send notification", nil)
		return
	}

	response := SendNotificationResponse{
		NotificationID: notification.ID,
		Status:         notification.Status,
		Message:        "Notification sent successfully",
	}

	utils.SendCreatedResponse(c, response, "Notification sent successfully")
}

// GetNotifications returns notifications for the user
func (pnc *PushNotificationController) GetNotifications(c *gin.Context) {
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

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	// Get notifications for user
	var notifications []models.PushNotification
	query := pnc.pushService.GetDB().Where("user_id = ? OR target_type = ?", userIDUint, models.TargetTypeAll).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&notifications).Error; err != nil {
		pnc.logger.Error("Failed to get notifications", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get notifications", nil)
		return
	}

	utils.SendSuccessResponse(c, notifications, "Notifications retrieved successfully")
}

// CreateTemplate creates a notification template
func (pnc *PushNotificationController) CreateTemplate(c *gin.Context) {
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

	var req CreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error(), nil)
		return
	}

	// Create template
	template := &models.NotificationTemplate{
		Name:      req.Name,
		Title:     req.Title,
		Body:      req.Body,
		Data:      req.Data,
		Platform:  req.Platform,
		Category:  req.Category,
		Variables: req.Variables,
		CreatedBy: userIDUint,
	}

	// Set defaults
	if template.Platform == "" {
		template.Platform = models.PlatformAll
	}

	if err := pnc.pushService.CreateTemplate(c.Request.Context(), template); err != nil {
		pnc.logger.Error("Failed to create template", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create template", nil)
		return
	}

	utils.SendCreatedResponse(c, template, "Template created successfully")
}

// GetTemplates returns notification templates
func (pnc *PushNotificationController) GetTemplates(c *gin.Context) {
	var templates []models.NotificationTemplate
	if err := pnc.pushService.GetDB().Where("is_active = ?", true).Find(&templates).Error; err != nil {
		pnc.logger.Error("Failed to get templates", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get templates", nil)
		return
	}

	utils.SendSuccessResponse(c, templates, "Templates retrieved successfully")
}

// CreateSegment creates a notification segment
func (pnc *PushNotificationController) CreateSegment(c *gin.Context) {
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

	var req CreateSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error(), nil)
		return
	}

	// Create segment
	segment := &models.NotificationSegment{
		Name:        req.Name,
		Description: req.Description,
		Query:       req.Query,
		CreatedBy:   userIDUint,
	}

	if err := pnc.pushService.CreateSegment(c.Request.Context(), segment); err != nil {
		pnc.logger.Error("Failed to create segment", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create segment", nil)
		return
	}

	utils.SendCreatedResponse(c, segment, "Segment created successfully")
}

// GetSegments returns notification segments
func (pnc *PushNotificationController) GetSegments(c *gin.Context) {
	var segments []models.NotificationSegment
	if err := pnc.pushService.GetDB().Where("is_active = ?", true).Find(&segments).Error; err != nil {
		pnc.logger.Error("Failed to get segments", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get segments", nil)
		return
	}

	utils.SendSuccessResponse(c, segments, "Segments retrieved successfully")
}

// RegisterDevice registers a device token
func (pnc *PushNotificationController) RegisterDevice(c *gin.Context) {
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

	var req RegisterDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error(), nil)
		return
	}

	// Create device token
	deviceToken := &models.DeviceToken{
		UserID:     userIDUint,
		Token:      req.Token,
		Platform:   req.Platform,
		DeviceID:   req.DeviceID,
		AppVersion: req.AppVersion,
		OSVersion:  req.OSVersion,
		ExpiresAt:  req.ExpiresAt,
	}

	if err := pnc.pushService.RegisterDeviceToken(c.Request.Context(), deviceToken); err != nil {
		pnc.logger.Error("Failed to register device token", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to register device token", nil)
		return
	}

	utils.SendCreatedResponse(c, deviceToken, "Device registered successfully")
}

// UnregisterDevice unregisters a device token
func (pnc *PushNotificationController) UnregisterDevice(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	_, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error(), nil)
		return
	}

	if err := pnc.pushService.UnregisterDeviceToken(c.Request.Context(), req.Token); err != nil {
		pnc.logger.Error("Failed to unregister device token", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to unregister device token", nil)
		return
	}

	utils.SendSuccessResponse(c, map[string]string{"message": "Device unregistered successfully"}, "Device unregistered successfully")
}

// GetNotificationAnalytics returns analytics for a notification
func (pnc *PushNotificationController) GetNotificationAnalytics(c *gin.Context) {
	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", nil)
		return
	}

	analytics, err := pnc.pushService.GetNotificationAnalytics(c.Request.Context(), uint(notificationID))
	if err != nil {
		pnc.logger.Error("Failed to get notification analytics", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get notification analytics", nil)
		return
	}

	utils.SendSuccessResponse(c, analytics, "Analytics retrieved successfully")
}

// MarkAsOpened marks a notification as opened
func (pnc *PushNotificationController) MarkAsOpened(c *gin.Context) {
	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", nil)
		return
	}

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

	// Update notification status
	var notification models.PushNotification
	if err := pnc.pushService.GetDB().First(&notification, notificationID).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Notification not found", nil)
		return
	}

	notification.MarkAsOpened()
	pnc.pushService.GetDB().Save(&notification)

	// Record analytics event
	analytic := &models.NotificationAnalytics{
		NotificationID: uint(notificationID),
		UserID:         userIDUint,
		Event:          models.AnalyticsEventOpened,
		Timestamp:      time.Now(),
	}
	pnc.pushService.GetDB().Create(analytic)

	utils.SendSuccessResponse(c, map[string]string{"message": "Notification marked as opened"}, "Notification marked as opened")
}

// GetDeviceTokens returns device tokens for the user
func (pnc *PushNotificationController) GetDeviceTokens(c *gin.Context) {
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

	var tokens []models.DeviceToken
	if err := pnc.pushService.GetDB().Where("user_id = ? AND is_active = ?", userIDUint, true).Find(&tokens).Error; err != nil {
		pnc.logger.Error("Failed to get device tokens", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get device tokens", nil)
		return
	}

	utils.SendSuccessResponse(c, tokens, "Device tokens retrieved successfully")
}
