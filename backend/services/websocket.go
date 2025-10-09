package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WebSocketService handles WebSocket operations and real-time notifications
type WebSocketService struct {
	hub    *Hub
	db     *gorm.DB
	redis  *redis.Client
	logger *zap.Logger
	cache  *CacheService
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(hub *Hub, db *gorm.DB, redis *redis.Client, cache *CacheService, logger *zap.Logger) *WebSocketService {
	return &WebSocketService{
		hub:    hub,
		db:     db,
		redis:  redis,
		logger: logger,
		cache:  cache,
	}
}

// NotificationData represents notification data structure
type NotificationData struct {
	ID        uint                   `json:"id"`
	UserID    uint                   `json:"user_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Read      bool                   `json:"read"`
	CreatedAt time.Time              `json:"created_at"`
}

// SendNotification sends a real-time notification to a user
func (ws *WebSocketService) SendNotification(ctx context.Context, userID uint, notificationType, title, message string, data map[string]interface{}) error {
	// Create notification record in database
	notification := NotificationData{
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		Data:      data,
		Read:      false,
		CreatedAt: time.Now(),
	}

	// Store in Redis for quick access
	notificationKey := fmt.Sprintf("notification:%d:%d", userID, time.Now().Unix())
	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Store in Redis with 24 hour expiration
	if err := ws.redis.Set(ctx, notificationKey, notificationJSON, 24*time.Hour).Err(); err != nil {
		ws.logger.Warn("Failed to store notification in Redis", zap.Error(err))
	}

	// Send real-time notification via WebSocket
	ws.hub.SendNotification(userID, notificationType, map[string]interface{}{
		"title":   title,
		"message": message,
		"data":    data,
		"id":      notification.ID,
	})

	ws.logger.Info("Notification sent",
		zap.Uint("user_id", userID),
		zap.String("type", notificationType),
		zap.String("title", title))

	return nil
}

// SendPaymentNotification sends a payment-related notification
func (ws *WebSocketService) SendPaymentNotification(ctx context.Context, userID uint, paymentType string, paymentData map[string]interface{}) error {
	title := "Payment Update"
	message := "Your payment status has been updated"

	switch paymentType {
	case "payment_succeeded":
		title = "Payment Successful"
		message = "Your payment has been processed successfully"
	case "payment_failed":
		title = "Payment Failed"
		message = "Your payment could not be processed"
	case "subscription_created":
		title = "Subscription Created"
		message = "Your subscription has been activated"
	case "subscription_cancelled":
		title = "Subscription Cancelled"
		message = "Your subscription has been cancelled"
	case "invoice_paid":
		title = "Invoice Paid"
		message = "Your invoice has been paid"
	}

	return ws.SendNotification(ctx, userID, paymentType, title, message, paymentData)
}

// SendSystemNotification sends a system-wide notification
func (ws *WebSocketService) SendSystemNotification(ctx context.Context, notificationType, title, message string, data map[string]interface{}) error {
	// Store system notification in Redis
	systemNotification := map[string]interface{}{
		"type":      notificationType,
		"title":     title,
		"message":   message,
		"data":      data,
		"timestamp": time.Now(),
	}

	systemNotificationJSON, err := json.Marshal(systemNotification)
	if err != nil {
		return fmt.Errorf("failed to marshal system notification: %w", err)
	}

	// Store in Redis with 1 hour expiration
	notificationKey := fmt.Sprintf("system_notification:%d", time.Now().Unix())
	if err := ws.redis.Set(ctx, notificationKey, systemNotificationJSON, time.Hour).Err(); err != nil {
		ws.logger.Warn("Failed to store system notification in Redis", zap.Error(err))
	}

	// Broadcast to all connected clients
	ws.hub.SendSystemMessage(notificationType, map[string]interface{}{
		"title":   title,
		"message": message,
		"data":    data,
	})

	ws.logger.Info("System notification sent",
		zap.String("type", notificationType),
		zap.String("title", title))

	return nil
}

// SendRoomNotification sends a notification to all users in a specific room
func (ws *WebSocketService) SendRoomNotification(ctx context.Context, room, notificationType, title, message string, data map[string]interface{}) error {
	roomNotification := map[string]interface{}{
		"type":      notificationType,
		"title":     title,
		"message":   message,
		"data":      data,
		"timestamp": time.Now(),
	}

	// Store room notification in Redis
	roomNotificationJSON, err := json.Marshal(roomNotification)
	if err != nil {
		return fmt.Errorf("failed to marshal room notification: %w", err)
	}

	notificationKey := fmt.Sprintf("room_notification:%s:%d", room, time.Now().Unix())
	if err := ws.redis.Set(ctx, notificationKey, roomNotificationJSON, time.Hour).Err(); err != nil {
		ws.logger.Warn("Failed to store room notification in Redis", zap.Error(err))
	}

	// Send to room via WebSocket
	ws.hub.SendToRoom(room, WebSocketMessage{
		Type: "room_notification",
		Data: map[string]interface{}{
			"notification_type": notificationType,
			"title":             title,
			"message":           message,
			"payload":           data,
		},
		Timestamp: time.Now(),
	})

	ws.logger.Info("Room notification sent",
		zap.String("room", room),
		zap.String("type", notificationType),
		zap.String("title", title))

	return nil
}

// GetUserNotifications retrieves notifications for a user
func (ws *WebSocketService) GetUserNotifications(ctx context.Context, userID uint, limit int) ([]NotificationData, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("user_notifications:%d", userID)
	var cached string
	err := ws.cache.Get(ctx, cacheKey, &cached)
	if err == nil && cached != "" {
		var notifications []NotificationData
		if err := json.Unmarshal([]byte(cached), &notifications); err == nil {
			return notifications, nil
		}
	}

	// Get from Redis
	pattern := fmt.Sprintf("notification:%d:*", userID)
	keys, err := ws.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get notification keys: %w", err)
	}

	var notifications []NotificationData
	for _, key := range keys {
		notificationJSON, err := ws.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var notification NotificationData
		if err := json.Unmarshal([]byte(notificationJSON), &notification); err != nil {
			continue
		}

		notifications = append(notifications, notification)
	}

	// Sort by creation time (newest first)
	for i := 0; i < len(notifications)-1; i++ {
		for j := i + 1; j < len(notifications); j++ {
			if notifications[i].CreatedAt.Before(notifications[j].CreatedAt) {
				notifications[i], notifications[j] = notifications[j], notifications[i]
			}
		}
	}

	// Limit results
	if limit > 0 && len(notifications) > limit {
		notifications = notifications[:limit]
	}

	// Cache the result
	if len(notifications) > 0 {
		notificationsJSON, _ := json.Marshal(notifications)
		ws.cache.Set(ctx, cacheKey, string(notificationsJSON), 5*time.Minute)
	}

	return notifications, nil
}

// MarkNotificationAsRead marks a notification as read
func (ws *WebSocketService) MarkNotificationAsRead(ctx context.Context, userID uint, notificationID uint) error {
	// This would typically update a database record
	// For now, we'll just log it
	ws.logger.Info("Notification marked as read",
		zap.Uint("user_id", userID),
		zap.Uint("notification_id", notificationID))

	// Clear cache
	cacheKey := fmt.Sprintf("user_notifications:%d", userID)
	ws.cache.Delete(ctx, cacheKey)

	return nil
}

// GetConnectionStats returns WebSocket connection statistics
func (ws *WebSocketService) GetConnectionStats() map[string]interface{} {
	return map[string]interface{}{
		"total_connections": ws.hub.GetClientCount(),
		"timestamp":         time.Now(),
	}
}

// GetUserConnections returns active connections for a user
func (ws *WebSocketService) GetUserConnections(userID uint) []*Client {
	return ws.hub.GetUserClients(userID)
}

// SendLiveUpdate sends a live update to a user
func (ws *WebSocketService) SendLiveUpdate(ctx context.Context, userID uint, updateType string, data map[string]interface{}) error {
	ws.hub.SendToUser(userID, WebSocketMessage{
		Type: "live_update",
		Data: map[string]interface{}{
			"update_type": updateType,
			"payload":     data,
		},
		Timestamp: time.Now(),
	})

	ws.logger.Info("Live update sent",
		zap.Uint("user_id", userID),
		zap.String("update_type", updateType))

	return nil
}

// SendDataUpdate sends a data update to a user
func (ws *WebSocketService) SendDataUpdate(ctx context.Context, userID uint, dataType string, data map[string]interface{}) error {
	ws.hub.SendToUser(userID, WebSocketMessage{
		Type: "data_update",
		Data: map[string]interface{}{
			"data_type": dataType,
			"payload":   data,
		},
		Timestamp: time.Now(),
	})

	ws.logger.Info("Data update sent",
		zap.Uint("user_id", userID),
		zap.String("data_type", dataType))

	return nil
}

// BroadcastDataUpdate broadcasts a data update to all connected users
func (ws *WebSocketService) BroadcastDataUpdate(ctx context.Context, dataType string, data map[string]interface{}) error {
	ws.hub.Broadcast(WebSocketMessage{
		Type: "broadcast_data_update",
		Data: map[string]interface{}{
			"data_type": dataType,
			"payload":   data,
		},
		Timestamp: time.Now(),
	})

	ws.logger.Info("Data update broadcasted",
		zap.String("data_type", dataType))

	return nil
}

// SendTypingIndicator sends a typing indicator to other users in a room
func (ws *WebSocketService) SendTypingIndicator(ctx context.Context, userID uint, room string, isTyping bool) error {
	ws.hub.SendToRoom(room, WebSocketMessage{
		Type: "typing_indicator",
		Data: map[string]interface{}{
			"user_id":   userID,
			"is_typing": isTyping,
		},
		Timestamp: time.Now(),
	})

	return nil
}

// SendPresenceUpdate sends a presence update (online/offline status)
func (ws *WebSocketService) SendPresenceUpdate(ctx context.Context, userID uint, status string) error {
	ws.hub.Broadcast(WebSocketMessage{
		Type: "presence_update",
		Data: map[string]interface{}{
			"user_id": userID,
			"status":  status,
		},
		Timestamp: time.Now(),
	})

	return nil
}
