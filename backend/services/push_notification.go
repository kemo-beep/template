package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"mobile-backend/models"
)

// PushNotificationService handles push notification operations
type PushNotificationService struct {
	db             *gorm.DB
	redis          *redis.Client
	cache          *CacheService
	wsService      *WebSocketService
	logger         *zap.Logger
	fcmService     *FCMService
	apnsService    *APNSService
	webPushService *WebPushService
}

// FCMService handles Firebase Cloud Messaging
type FCMService struct {
	serverKey string
	projectID string
}

// APNSService handles Apple Push Notification Service
type APNSService struct {
	certPath   string
	keyPath    string
	teamID     string
	keyID      string
	bundleID   string
	production bool
}

// WebPushService handles Web Push notifications
type WebPushService struct {
	vapidPublicKey  string
	vapidPrivateKey string
	vapidEmail      string
}

// NewPushNotificationService creates a new push notification service
func NewPushNotificationService(db *gorm.DB, redis *redis.Client, cache *CacheService, wsService *WebSocketService, logger *zap.Logger) *PushNotificationService {
	return &PushNotificationService{
		db:             db,
		redis:          redis,
		cache:          cache,
		wsService:      wsService,
		logger:         logger,
		fcmService:     NewFCMService(),
		apnsService:    NewAPNSService(),
		webPushService: NewWebPushService(),
	}
}

// NewFCMService creates a new FCM service
func NewFCMService() *FCMService {
	return &FCMService{
		serverKey: "your-fcm-server-key",
		projectID: "your-project-id",
	}
}

// NewAPNSService creates a new APNS service
func NewAPNSService() *APNSService {
	return &APNSService{
		certPath:   "path/to/cert.pem",
		keyPath:    "path/to/key.pem",
		teamID:     "your-team-id",
		keyID:      "your-key-id",
		bundleID:   "your-bundle-id",
		production: false,
	}
}

// NewWebPushService creates a new Web Push service
func NewWebPushService() *WebPushService {
	return &WebPushService{
		vapidPublicKey:  "your-vapid-public-key",
		vapidPrivateKey: "your-vapid-private-key",
		vapidEmail:      "your-email@example.com",
	}
}

// SendNotification sends a push notification
func (pns *PushNotificationService) SendNotification(ctx context.Context, notification *models.PushNotification) error {
	// Validate notification
	if err := pns.validateNotification(notification); err != nil {
		return fmt.Errorf("invalid notification: %w", err)
	}

	// Get target devices
	devices, err := pns.getTargetDevices(ctx, &notification.Target)
	if err != nil {
		return fmt.Errorf("failed to get target devices: %w", err)
	}

	if len(devices) == 0 {
		pns.logger.Warn("No target devices found for notification", zap.Uint("notification_id", notification.ID))
		return nil
	}

	// Send to each device
	successCount := 0
	for _, device := range devices {
		if err := pns.sendToDevice(ctx, notification, device); err != nil {
			pns.logger.Error("Failed to send notification to device",
				zap.Uint("notification_id", notification.ID),
				zap.String("device_token", device.Token),
				zap.Error(err))
			continue
		}
		successCount++
	}

	// Update notification status
	if successCount > 0 {
		notification.MarkAsSent()
		pns.db.Save(notification)
	}

	pns.logger.Info("Notification sent",
		zap.Uint("notification_id", notification.ID),
		zap.Int("total_devices", len(devices)),
		zap.Int("successful_sends", successCount))

	return nil
}

// ScheduleNotification schedules a notification for future delivery
func (pns *PushNotificationService) ScheduleNotification(ctx context.Context, notification *models.PushNotification) error {
	if notification.ScheduledAt == nil {
		return fmt.Errorf("scheduled_at is required for scheduled notifications")
	}

	if notification.ScheduledAt.Before(time.Now()) {
		return fmt.Errorf("scheduled_at must be in the future")
	}

	notification.Status = models.NotificationStatusScheduled
	if err := pns.db.Create(notification).Error; err != nil {
		return fmt.Errorf("failed to create scheduled notification: %w", err)
	}

	// Schedule the notification in Redis
	scheduleKey := fmt.Sprintf("scheduled_notifications:%d", notification.ID)
	scheduleData, _ := json.Marshal(notification)

	delay := time.Until(*notification.ScheduledAt)
	pns.redis.SetEX(ctx, scheduleKey, string(scheduleData), delay)

	pns.logger.Info("Notification scheduled",
		zap.Uint("notification_id", notification.ID),
		zap.Time("scheduled_at", *notification.ScheduledAt))

	return nil
}

// ProcessScheduledNotifications processes scheduled notifications
func (pns *PushNotificationService) ProcessScheduledNotifications(ctx context.Context) error {
	// Get notifications that are ready to be sent
	var notifications []models.PushNotification
	err := pns.db.Where("status = ? AND scheduled_at <= ?",
		models.NotificationStatusScheduled, time.Now()).Find(&notifications).Error
	if err != nil {
		return fmt.Errorf("failed to get scheduled notifications: %w", err)
	}

	for _, notification := range notifications {
		if err := pns.SendNotification(ctx, &notification); err != nil {
			pns.logger.Error("Failed to process scheduled notification",
				zap.Uint("notification_id", notification.ID),
				zap.Error(err))
			continue
		}
	}

	return nil
}

// CreateTemplate creates a notification template
func (pns *PushNotificationService) CreateTemplate(ctx context.Context, template *models.NotificationTemplate) error {
	if err := pns.db.Create(template).Error; err != nil {
		return fmt.Errorf("failed to create template: %w", err)
	}

	pns.logger.Info("Notification template created",
		zap.Uint("template_id", template.ID),
		zap.String("template_name", template.Name))

	return nil
}

// CreateSegment creates a notification segment
func (pns *PushNotificationService) CreateSegment(ctx context.Context, segment *models.NotificationSegment) error {
	// Update user count
	if err := segment.UpdateUserCount(pns.db); err != nil {
		pns.logger.Warn("Failed to update segment user count", zap.Error(err))
	}

	if err := pns.db.Create(segment).Error; err != nil {
		return fmt.Errorf("failed to create segment: %w", err)
	}

	pns.logger.Info("Notification segment created",
		zap.Uint("segment_id", segment.ID),
		zap.String("segment_name", segment.Name))

	return nil
}

// RegisterDeviceToken registers a device token
func (pns *PushNotificationService) RegisterDeviceToken(ctx context.Context, token *models.DeviceToken) error {
	// Check if token already exists
	var existingToken models.DeviceToken
	err := pns.db.Where("token = ?", token.Token).First(&existingToken).Error
	if err == nil {
		// Update existing token
		existingToken.UserID = token.UserID
		existingToken.Platform = token.Platform
		existingToken.DeviceID = token.DeviceID
		existingToken.AppVersion = token.AppVersion
		existingToken.OSVersion = token.OSVersion
		existingToken.IsActive = true
		existingToken.UpdateLastUsed()
		existingToken.ExpiresAt = token.ExpiresAt

		if err := pns.db.Save(&existingToken).Error; err != nil {
			return fmt.Errorf("failed to update device token: %w", err)
		}
	} else if err == gorm.ErrRecordNotFound {
		// Create new token
		token.IsActive = true
		token.UpdateLastUsed()
		if err := pns.db.Create(token).Error; err != nil {
			return fmt.Errorf("failed to create device token: %w", err)
		}
	} else {
		return fmt.Errorf("failed to check existing token: %w", err)
	}

	pns.logger.Info("Device token registered",
		zap.String("token", token.Token),
		zap.String("platform", token.Platform),
		zap.Uint("user_id", token.UserID))

	return nil
}

// UnregisterDeviceToken unregisters a device token
func (pns *PushNotificationService) UnregisterDeviceToken(ctx context.Context, token string) error {
	if err := pns.db.Model(&models.DeviceToken{}).Where("token = ?", token).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to unregister device token: %w", err)
	}

	pns.logger.Info("Device token unregistered", zap.String("token", token))
	return nil
}

// GetNotificationAnalytics gets analytics for a notification
func (pns *PushNotificationService) GetNotificationAnalytics(ctx context.Context, notificationID uint) (*NotificationAnalyticsSummary, error) {
	var analytics []models.NotificationAnalytics
	err := pns.db.Where("notification_id = ?", notificationID).Find(&analytics).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get notification analytics: %w", err)
	}

	summary := &NotificationAnalyticsSummary{
		NotificationID: notificationID,
		TotalSent:      len(analytics),
		Sent:           0,
		Delivered:      0,
		Opened:         0,
		Failed:         0,
	}

	for _, analytic := range analytics {
		switch analytic.Event {
		case models.AnalyticsEventSent:
			summary.Sent++
		case models.AnalyticsEventDelivered:
			summary.Delivered++
		case models.AnalyticsEventOpened:
			summary.Opened++
		case models.AnalyticsEventFailed:
			summary.Failed++
		}
	}

	return summary, nil
}

// GetDB returns the database instance
func (pns *PushNotificationService) GetDB() *gorm.DB {
	return pns.db
}

// validateNotification validates a notification
func (pns *PushNotificationService) validateNotification(notification *models.PushNotification) error {
	if notification.Title == "" {
		return fmt.Errorf("title is required")
	}
	if notification.Body == "" {
		return fmt.Errorf("body is required")
	}
	if notification.ExpiresAt != nil && notification.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("expires_at must be in the future")
	}
	return nil
}

// getTargetDevices gets devices based on notification target
func (pns *PushNotificationService) getTargetDevices(ctx context.Context, target *models.NotificationTarget) ([]models.DeviceToken, error) {
	var devices []models.DeviceToken
	query := pns.db.Where("is_active = ?", true)

	switch target.Type {
	case models.TargetTypeAll:
		// Get all active devices
		err := query.Find(&devices).Error
		return devices, err

	case models.TargetTypeUser:
		// Get devices for specific users
		if len(target.UserIDs) > 0 {
			err := query.Where("user_id IN ?", target.UserIDs).Find(&devices).Error
			return devices, err
		}

	case models.TargetTypeDevice:
		// Get specific devices
		if len(target.DeviceIDs) > 0 {
			err := query.Where("device_id IN ?", target.DeviceIDs).Find(&devices).Error
			return devices, err
		}

	case models.TargetTypeSegment:
		// Get devices for users in segments
		if len(target.SegmentIDs) > 0 {
			// This would typically involve a more complex query based on segment criteria
			// For now, we'll just return empty
			return devices, nil
		}

	case models.TargetTypeCustom:
		// Execute custom query
		if target.CustomQuery != "" {
			err := pns.db.Raw(target.CustomQuery).Scan(&devices).Error
			return devices, err
		}
	}

	return devices, nil
}

// sendToDevice sends notification to a specific device
func (pns *PushNotificationService) sendToDevice(ctx context.Context, notification *models.PushNotification, device models.DeviceToken) error {
	// Record analytics event
	analytic := &models.NotificationAnalytics{
		NotificationID: notification.ID,
		UserID:         device.UserID,
		Event:          models.AnalyticsEventSent,
		Timestamp:      time.Now(),
		DeviceInfo: map[string]interface{}{
			"platform":    device.Platform,
			"app_version": device.AppVersion,
			"os_version":  device.OSVersion,
		},
	}
	pns.db.Create(analytic)

	// Send based on platform
	switch device.Platform {
	case models.PlatformIOS:
		return pns.apnsService.Send(ctx, device.Token, notification)
	case models.PlatformAndroid:
		return pns.fcmService.Send(ctx, device.Token, notification)
	case models.PlatformWeb:
		return pns.webPushService.Send(ctx, device.Token, notification)
	default:
		return fmt.Errorf("unsupported platform: %s", device.Platform)
	}
}

// FCM Service Methods
func (fcm *FCMService) Send(ctx context.Context, token string, notification *models.PushNotification) error {
	// This would implement actual FCM sending
	// For now, we'll just log it
	fmt.Printf("Sending FCM notification to %s: %s\n", token, notification.Title)
	return nil
}

// APNS Service Methods
func (apns *APNSService) Send(ctx context.Context, token string, notification *models.PushNotification) error {
	// This would implement actual APNS sending
	// For now, we'll just log it
	fmt.Printf("Sending APNS notification to %s: %s\n", token, notification.Title)
	return nil
}

// Web Push Service Methods
func (wps *WebPushService) Send(ctx context.Context, token string, notification *models.PushNotification) error {
	// This would implement actual Web Push sending
	// For now, we'll just log it
	fmt.Printf("Sending Web Push notification to %s: %s\n", token, notification.Title)
	return nil
}

// NotificationAnalyticsSummary represents analytics summary
type NotificationAnalyticsSummary struct {
	NotificationID uint `json:"notification_id"`
	TotalSent      int  `json:"total_sent"`
	Sent           int  `json:"sent"`
	Delivered      int  `json:"delivered"`
	Opened         int  `json:"opened"`
	Failed         int  `json:"failed"`
}
