package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"mobile-backend/services"
	"mobile-backend/utils"
)

// WebSocketController handles WebSocket connections and real-time features
type WebSocketController struct {
	websocketService *services.WebSocketService
	hub              *services.Hub
	logger           *zap.Logger
}

// NewWebSocketController creates a new WebSocket controller
func NewWebSocketController(websocketService *services.WebSocketService, hub *services.Hub, logger *zap.Logger) *WebSocketController {
	return &WebSocketController{
		websocketService: websocketService,
		hub:              hub,
		logger:           logger,
	}
}

// WebSocket upgrader configuration
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, implement proper origin checking
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Connect handles WebSocket connection requests
func (wc *WebSocketController) Connect(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", nil)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		wc.logger.Error("Failed to upgrade connection", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusBadRequest, "Failed to upgrade connection", nil)
		return
	}

	// Register client with hub
	wc.hub.ServeWS(conn, userID)

	wc.logger.Info("WebSocket connection established", zap.Uint("user_id", userID))
}

// GetNotifications retrieves user notifications
func (wc *WebSocketController) GetNotifications(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", nil)
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	// Get notifications
	notifications, err := wc.websocketService.GetUserNotifications(c.Request.Context(), userID, limit)
	if err != nil {
		wc.logger.Error("Failed to get notifications", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get notifications", nil)
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"notifications": notifications,
		"count":         len(notifications),
	}, "Notifications retrieved successfully")
}

// MarkNotificationAsRead marks a notification as read
func (wc *WebSocketController) MarkNotificationAsRead(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", nil)
		return
	}

	// Get notification ID from URL parameter
	notificationIDStr := c.Param("id")
	notificationID, err := strconv.ParseUint(notificationIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid notification ID", nil)
		return
	}

	// Mark as read
	if err := wc.websocketService.MarkNotificationAsRead(c.Request.Context(), userID, uint(notificationID)); err != nil {
		wc.logger.Error("Failed to mark notification as read", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to mark notification as read", nil)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "Notification marked as read"}, "Notification marked as read")
}

// GetConnectionStats returns WebSocket connection statistics
func (wc *WebSocketController) GetConnectionStats(c *gin.Context) {
	stats := wc.websocketService.GetConnectionStats()
	utils.SendSuccessResponse(c, stats, "Connection stats retrieved successfully")
}

// GetUserConnections returns active connections for a user
func (wc *WebSocketController) GetUserConnections(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", nil)
		return
	}

	connections := wc.websocketService.GetUserConnections(userID)

	// Convert connections to response format
	var connectionData []gin.H
	for _, conn := range connections {
		connectionData = append(connectionData, gin.H{
			"client_id": conn.ID,
			"user_id":   conn.UserID,
			"rooms":     conn.Rooms,
			"last_ping": conn.LastPing,
		})
	}

	utils.SendSuccessResponse(c, gin.H{
		"connections": connectionData,
		"count":       len(connectionData),
	}, "User connections retrieved successfully")
}

// SendNotification sends a notification to a user (admin only)
func (wc *WebSocketController) SendNotification(c *gin.Context) {
	// Check if user is authenticated
	_, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Check if user is admin (you might want to implement proper role checking)
	// For now, we'll allow any authenticated user to send notifications

	var req struct {
		TargetUserID uint                   `json:"target_user_id" binding:"required"`
		Type         string                 `json:"type" binding:"required"`
		Title        string                 `json:"title" binding:"required"`
		Message      string                 `json:"message" binding:"required"`
		Data         map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Send notification
	if err := wc.websocketService.SendNotification(
		c.Request.Context(),
		req.TargetUserID,
		req.Type,
		req.Title,
		req.Message,
		req.Data,
	); err != nil {
		wc.logger.Error("Failed to send notification", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to send notification", nil)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "Notification sent successfully"}, "Notification sent successfully")
}

// SendSystemNotification sends a system-wide notification (admin only)
func (wc *WebSocketController) SendSystemNotification(c *gin.Context) {
	var req struct {
		Type    string                 `json:"type" binding:"required"`
		Title   string                 `json:"title" binding:"required"`
		Message string                 `json:"message" binding:"required"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Send system notification
	if err := wc.websocketService.SendSystemNotification(
		c.Request.Context(),
		req.Type,
		req.Title,
		req.Message,
		req.Data,
	); err != nil {
		wc.logger.Error("Failed to send system notification", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to send system notification", nil)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "System notification sent successfully"}, "System notification sent successfully")
}

// SendRoomNotification sends a notification to a specific room
func (wc *WebSocketController) SendRoomNotification(c *gin.Context) {
	var req struct {
		Room    string                 `json:"room" binding:"required"`
		Type    string                 `json:"type" binding:"required"`
		Title   string                 `json:"title" binding:"required"`
		Message string                 `json:"message" binding:"required"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Send room notification
	if err := wc.websocketService.SendRoomNotification(
		c.Request.Context(),
		req.Room,
		req.Type,
		req.Title,
		req.Message,
		req.Data,
	); err != nil {
		wc.logger.Error("Failed to send room notification", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to send room notification", nil)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "Room notification sent successfully"}, "Room notification sent successfully")
}

// SendLiveUpdate sends a live update to a user
func (wc *WebSocketController) SendLiveUpdate(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", nil)
		return
	}

	var req struct {
		UpdateType string                 `json:"update_type" binding:"required"`
		Data       map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Send live update
	if err := wc.websocketService.SendLiveUpdate(
		c.Request.Context(),
		userID,
		req.UpdateType,
		req.Data,
	); err != nil {
		wc.logger.Error("Failed to send live update", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to send live update", nil)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "Live update sent successfully"}, "Live update sent successfully")
}

// SendDataUpdate sends a data update to a user
func (wc *WebSocketController) SendDataUpdate(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", nil)
		return
	}

	var req struct {
		DataType string                 `json:"data_type" binding:"required"`
		Data     map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Send data update
	if err := wc.websocketService.SendDataUpdate(
		c.Request.Context(),
		userID,
		req.DataType,
		req.Data,
	); err != nil {
		wc.logger.Error("Failed to send data update", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to send data update", nil)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "Data update sent successfully"}, "Data update sent successfully")
}

// BroadcastDataUpdate broadcasts a data update to all connected users
func (wc *WebSocketController) BroadcastDataUpdate(c *gin.Context) {
	var req struct {
		DataType string                 `json:"data_type" binding:"required"`
		Data     map[string]interface{} `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Broadcast data update
	if err := wc.websocketService.BroadcastDataUpdate(
		c.Request.Context(),
		req.DataType,
		req.Data,
	); err != nil {
		wc.logger.Error("Failed to broadcast data update", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to broadcast data update", nil)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "Data update broadcasted successfully"}, "Data update broadcasted successfully")
}

// SendTypingIndicator sends a typing indicator
func (wc *WebSocketController) SendTypingIndicator(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", nil)
		return
	}

	var req struct {
		Room     string `json:"room" binding:"required"`
		IsTyping bool   `json:"is_typing" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Send typing indicator
	if err := wc.websocketService.SendTypingIndicator(
		c.Request.Context(),
		userID,
		req.Room,
		req.IsTyping,
	); err != nil {
		wc.logger.Error("Failed to send typing indicator", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to send typing indicator", nil)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "Typing indicator sent successfully"}, "Typing indicator sent successfully")
}

// SendPresenceUpdate sends a presence update
func (wc *WebSocketController) SendPresenceUpdate(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid user ID", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=online offline away busy"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Send presence update
	if err := wc.websocketService.SendPresenceUpdate(
		c.Request.Context(),
		userID,
		req.Status,
	); err != nil {
		wc.logger.Error("Failed to send presence update", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to send presence update", nil)
		return
	}

	utils.SendSuccessResponse(c, gin.H{"message": "Presence update sent successfully"}, "Presence update sent successfully")
}
