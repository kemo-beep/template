package services

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type NotificationService struct {
	redis *redis.Client
}

func NewNotificationService(redis *redis.Client) *NotificationService {
	return &NotificationService{
		redis: redis,
	}
}

// SendPushNotification sends a push notification to a specific user
func (n *NotificationService) SendPushNotification(ctx context.Context, userID uint, title, body string) error {
	// In a real application, you would retrieve the user's FCM token from the database or cache
	// For demonstration, we assume it's stored in Redis.
	token, err := n.redis.Get(ctx, fmt.Sprintf("fcm_token:%d", userID)).Result()
	if err != nil {
		return fmt.Errorf("failed to get FCM token for user %d: %w", userID, err)
	}

	// Example using Firebase Cloud Messaging (FCM)
	// This is a placeholder implementation
	// In production, you would integrate with FCM or APNS
	fmt.Printf("Simulating push notification to user %d (token: %s): %s - %s\n", userID, token, title, body)
	return nil
}

// RegisterToken stores a user's FCM token in Redis
func (n *NotificationService) RegisterToken(ctx context.Context, userID uint, token string) error {
	// Store token indefinitely or with a suitable expiration
	return n.redis.Set(ctx, fmt.Sprintf("fcm_token:%d", userID), token, 0).Err()
}

// SendBulkNotification sends notifications to multiple users
func (n *NotificationService) SendBulkNotification(ctx context.Context, userIDs []uint, title, body string) error {
	for _, userID := range userIDs {
		if err := n.SendPushNotification(ctx, userID, title, body); err != nil {
			fmt.Printf("Failed to send notification to user %d: %v\n", userID, err)
			// Continue with other users even if one fails
		}
	}
	return nil
}

// SendDataNotification sends a data-only notification
func (n *NotificationService) SendDataNotification(ctx context.Context, userID uint, data map[string]string) error {
	// This is a placeholder for data-only notifications
	fmt.Printf("Simulating data notification to user %d: %+v\n", userID, data)
	return nil
}