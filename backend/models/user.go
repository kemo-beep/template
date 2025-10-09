package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	BaseModel
	Email     string     `json:"email" gorm:"uniqueIndex;not null"`
	Password  string     `json:"-" gorm:"not null"`
	Name      string     `json:"name"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time `json:"last_login,omitempty"`

	// Subscription status fields
	SubscriptionStatus string     `json:"subscription_status" gorm:"default:'free'" validate:"oneof=free trial active canceled past_due"`
	IsPro              bool       `json:"is_pro" gorm:"default:false"`
	SubscriptionID     *uint      `json:"subscription_id,omitempty"`
	SubscriptionEndsAt *time.Time `json:"subscription_ends_at,omitempty"`
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`

	// Relationships
	ActiveSubscription *Subscription `json:"active_subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
}

// HashPassword hashes the user's password
func (u *User) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword verifies the user's password
func (u *User) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// BeforeCreate hook to hash password before creating user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Password != "" {
		return u.HashPassword(u.Password)
	}
	return nil
}

// IsSubscriptionActive checks if user has an active subscription
func (u *User) IsSubscriptionActive() bool {
	return u.SubscriptionStatus == "active" || u.SubscriptionStatus == "trial"
}

// IsProUser checks if user has pro access
func (u *User) IsProUser() bool {
	return u.IsPro && u.IsSubscriptionActive()
}

// HasTrialAccess checks if user has trial access
func (u *User) HasTrialAccess() bool {
	if u.TrialEndsAt == nil {
		return false
	}
	return time.Now().Before(*u.TrialEndsAt) && u.SubscriptionStatus == "trial"
}

// GetSubscriptionStatus returns a human-readable subscription status
func (u *User) GetSubscriptionStatus() string {
	if u.IsProUser() {
		return "Pro"
	}
	if u.HasTrialAccess() {
		return "Trial"
	}
	return "Free"
}
