package utils

import "time"

// UserResponse represents a user response
type UserResponse struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Subscription status fields
	SubscriptionStatus   string     `json:"subscription_status"`
	IsPro                bool       `json:"is_pro"`
	SubscriptionEndsAt   *time.Time `json:"subscription_ends_at,omitempty"`
	TrialEndsAt          *time.Time `json:"trial_ends_at,omitempty"`
	StatusDisplay        string     `json:"status_display"`
	HasTrialAccess       bool       `json:"has_trial_access"`
	IsSubscriptionActive bool       `json:"is_subscription_active"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int          `json:"expires_in"`
}

// RefreshTokenResponse represents a refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// parseValidationErrors converts validation errors to ValidationError slice
func parseValidationErrors(err error) []ValidationError {
	// This is a placeholder implementation
	// In a real implementation, you would parse the validation errors
	// and convert them to ValidationError structs
	return []ValidationError{
		{
			Field:   "general",
			Message: err.Error(),
		},
	}
}
