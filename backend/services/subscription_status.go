package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"mobile-backend/models"
)

// SubscriptionStatusService handles subscription status management
type SubscriptionStatusService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSubscriptionStatusService creates a new subscription status service
func NewSubscriptionStatusService(db *gorm.DB, logger *zap.Logger) *SubscriptionStatusService {
	return &SubscriptionStatusService{
		db:     db,
		logger: logger,
	}
}

// UpdateUserSubscriptionStatus updates user's subscription status based on subscription
func (s *SubscriptionStatusService) UpdateUserSubscriptionStatus(ctx context.Context, userID uint, subscription *models.Subscription) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Determine subscription status
	var status string
	var isPro bool
	var subscriptionEndsAt *time.Time
	var trialEndsAt *time.Time

	if subscription == nil {
		// No subscription - free user
		status = "free"
		isPro = false
	} else {
		// Map subscription status to user status
		switch subscription.Status {
		case "active":
			status = "active"
			isPro = true
			subscriptionEndsAt = &subscription.CurrentPeriodEnd
		case "trialing":
			status = "trial"
			isPro = true
			if subscription.TrialEnd != nil {
				trialEndsAt = subscription.TrialEnd
			}
			subscriptionEndsAt = &subscription.CurrentPeriodEnd
		case "canceled":
			status = "canceled"
			isPro = false
			if subscription.CanceledAt != nil {
				subscriptionEndsAt = subscription.CanceledAt
			}
		case "past_due":
			status = "past_due"
			isPro = false
			subscriptionEndsAt = &subscription.CurrentPeriodEnd
		default:
			status = "free"
			isPro = false
		}
	}

	// Update user subscription status
	updates := map[string]interface{}{
		"subscription_status":  status,
		"is_pro":               isPro,
		"subscription_id":      subscription.ID,
		"subscription_ends_at": subscriptionEndsAt,
		"trial_ends_at":        trialEndsAt,
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update user subscription status: %w", err)
	}

	s.logger.Info("Updated user subscription status",
		zap.Uint("user_id", userID),
		zap.String("status", status),
		zap.Bool("is_pro", isPro),
		zap.Uint("subscription_id", subscription.ID),
	)

	return nil
}

// GetUserSubscriptionStatus gets user's current subscription status
func (s *SubscriptionStatusService) GetUserSubscriptionStatus(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("ActiveSubscription").First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// CheckSubscriptionExpiry checks and updates expired subscriptions
func (s *SubscriptionStatusService) CheckSubscriptionExpiry(ctx context.Context) error {
	now := time.Now()

	// Find users with expired subscriptions
	var users []models.User
	if err := s.db.Where("subscription_status IN ? AND subscription_ends_at < ?",
		[]string{"active", "trial"}, now).Find(&users).Error; err != nil {
		return fmt.Errorf("failed to find expired subscriptions: %w", err)
	}

	for _, user := range users {
		// Update user status to expired
		updates := map[string]interface{}{
			"subscription_status": "canceled",
			"is_pro":              false,
		}

		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			s.logger.Error("Failed to update expired subscription",
				zap.Uint("user_id", user.ID),
				zap.Error(err),
			)
			continue
		}

		s.logger.Info("Updated expired subscription",
			zap.Uint("user_id", user.ID),
			zap.String("previous_status", user.SubscriptionStatus),
		)
	}

	return nil
}

// GetSubscriptionStats gets subscription statistics
func (s *SubscriptionStatusService) GetSubscriptionStats(ctx context.Context) (map[string]interface{}, error) {
	var stats struct {
		TotalUsers          int64 `json:"total_users"`
		FreeUsers           int64 `json:"free_users"`
		ProUsers            int64 `json:"pro_users"`
		TrialUsers          int64 `json:"trial_users"`
		CanceledUsers       int64 `json:"canceled_users"`
		PastDueUsers        int64 `json:"past_due_users"`
		ActiveSubscriptions int64 `json:"active_subscriptions"`
	}

	// Count users by subscription status
	if err := s.db.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count total users: %w", err)
	}

	if err := s.db.Model(&models.User{}).Where("subscription_status = ?", "free").Count(&stats.FreeUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count free users: %w", err)
	}

	if err := s.db.Model(&models.User{}).Where("subscription_status = ?", "active").Count(&stats.ProUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count pro users: %w", err)
	}

	if err := s.db.Model(&models.User{}).Where("subscription_status = ?", "trial").Count(&stats.TrialUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count trial users: %w", err)
	}

	if err := s.db.Model(&models.User{}).Where("subscription_status = ?", "canceled").Count(&stats.CanceledUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count canceled users: %w", err)
	}

	if err := s.db.Model(&models.User{}).Where("subscription_status = ?", "past_due").Count(&stats.PastDueUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count past due users: %w", err)
	}

	if err := s.db.Model(&models.Subscription{}).Where("status IN ?", []string{"active", "trialing"}).Count(&stats.ActiveSubscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to count active subscriptions: %w", err)
	}

	return map[string]interface{}{
		"total_users":          stats.TotalUsers,
		"free_users":           stats.FreeUsers,
		"pro_users":            stats.ProUsers,
		"trial_users":          stats.TrialUsers,
		"canceled_users":       stats.CanceledUsers,
		"past_due_users":       stats.PastDueUsers,
		"active_subscriptions": stats.ActiveSubscriptions,
		"pro_percentage":       float64(stats.ProUsers) / float64(stats.TotalUsers) * 100,
		"trial_percentage":     float64(stats.TrialUsers) / float64(stats.TotalUsers) * 100,
	}, nil
}

// UpgradeUserToPro upgrades a user to pro status
func (s *SubscriptionStatusService) UpgradeUserToPro(ctx context.Context, userID uint, subscriptionID uint) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	var subscription models.Subscription
	if err := s.db.First(&subscription, subscriptionID).Error; err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Update user to pro status
	updates := map[string]interface{}{
		"subscription_status":  "active",
		"is_pro":               true,
		"subscription_id":      subscriptionID,
		"subscription_ends_at": &subscription.CurrentPeriodEnd,
	}

	if subscription.TrialEnd != nil {
		updates["trial_ends_at"] = subscription.TrialEnd
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to upgrade user to pro: %w", err)
	}

	s.logger.Info("Upgraded user to pro",
		zap.Uint("user_id", userID),
		zap.Uint("subscription_id", subscriptionID),
	)

	return nil
}

// DowngradeUserFromPro downgrades a user from pro status
func (s *SubscriptionStatusService) DowngradeUserFromPro(ctx context.Context, userID uint, reason string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update user to free status
	updates := map[string]interface{}{
		"subscription_status":  "free",
		"is_pro":               false,
		"subscription_id":      nil,
		"subscription_ends_at": nil,
		"trial_ends_at":        nil,
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to downgrade user from pro: %w", err)
	}

	s.logger.Info("Downgraded user from pro",
		zap.Uint("user_id", userID),
		zap.String("reason", reason),
	)

	return nil
}

// StartSubscriptionExpiryChecker starts a background task to check for expired subscriptions
func (s *SubscriptionStatusService) StartSubscriptionExpiryChecker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Subscription expiry checker stopped")
			return
		case <-ticker.C:
			if err := s.CheckSubscriptionExpiry(ctx); err != nil {
				s.logger.Error("Failed to check subscription expiry", zap.Error(err))
			}
		}
	}
}
