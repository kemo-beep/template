package models

import (
	"time"

	"gorm.io/gorm"
)

// Product represents a sellable product in the system
type Product struct {
	BaseModel
	Name          string  `json:"name" gorm:"not null" validate:"required,min=1,max=255"`
	Description   string  `json:"description" gorm:"type:text"`
	Price         int64   `json:"price" gorm:"not null" validate:"min=0"` // Price in cents
	Currency      string  `json:"currency" gorm:"default:'usd'" validate:"required,len=3"`
	IsActive      bool    `json:"is_active" gorm:"default:true"`
	IsRecurring   bool    `json:"is_recurring" gorm:"default:false"` // For subscriptions
	Interval      string  `json:"interval,omitempty"`                // monthly, yearly, etc.
	IntervalCount int     `json:"interval_count,omitempty"`          // 1, 3, 6, 12, etc.
	TrialDays     int     `json:"trial_days,omitempty"`
	Metadata      JSONMap `json:"metadata,omitempty" gorm:"type:jsonb"`

	// External IDs for payment providers
	StripeProductID string `json:"stripe_product_id,omitempty"`
	PolarProductID  string `json:"polar_product_id,omitempty"`

	// Relationships
	Subscriptions []Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:ProductID"`
	Payments      []Payment      `json:"payments,omitempty" gorm:"foreignKey:ProductID"`
}

// Plan represents a subscription plan
type Plan struct {
	BaseModel
	Name          string  `json:"name" gorm:"not null" validate:"required,min=1,max=255"`
	Description   string  `json:"description" gorm:"type:text"`
	ProductID     uint    `json:"product_id" gorm:"not null"`
	Price         int64   `json:"price" gorm:"not null" validate:"min=0"` // Price in cents
	Currency      string  `json:"currency" gorm:"default:'usd'" validate:"required,len=3"`
	Interval      string  `json:"interval" gorm:"not null" validate:"required,oneof=day week month year"`
	IntervalCount int     `json:"interval_count" gorm:"default:1" validate:"min=1"`
	IsActive      bool    `json:"is_active" gorm:"default:true"`
	TrialDays     int     `json:"trial_days,omitempty"`
	Metadata      JSONMap `json:"metadata,omitempty" gorm:"type:jsonb"`

	// External IDs for payment providers
	StripePriceID string `json:"stripe_price_id,omitempty"`
	PolarPlanID   string `json:"polar_plan_id,omitempty"`

	// Relationships
	Product       Product        `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Subscriptions []Subscription `json:"subscriptions,omitempty" gorm:"foreignKey:PlanID"`
}

// Subscription represents a user's subscription to a product/plan
type Subscription struct {
	BaseModel
	UserID             uint       `json:"user_id" gorm:"not null"`
	ProductID          uint       `json:"product_id" gorm:"not null"`
	PlanID             *uint      `json:"plan_id,omitempty"`
	Status             string     `json:"status" gorm:"not null;default:'active'" validate:"oneof=active canceled past_due incomplete incomplete_expired trialing paused"`
	CurrentPeriodStart time.Time  `json:"current_period_start" gorm:"not null"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end" gorm:"not null"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end" gorm:"default:false"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty"`
	TrialStart         *time.Time `json:"trial_start,omitempty"`
	TrialEnd           *time.Time `json:"trial_end,omitempty"`
	Quantity           int        `json:"quantity" gorm:"default:1" validate:"min=1"`
	Metadata           JSONMap    `json:"metadata,omitempty" gorm:"type:jsonb"`

	// External IDs for payment providers
	StripeSubscriptionID string `json:"stripe_subscription_id,omitempty"`
	PolarSubscriptionID  string `json:"polar_subscription_id,omitempty"`
	PaymentMethod        string `json:"payment_method,omitempty"`

	// Relationships
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Product  Product   `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Plan     *Plan     `json:"plan,omitempty" gorm:"foreignKey:PlanID"`
	Payments []Payment `json:"payments,omitempty" gorm:"foreignKey:SubscriptionID"`
}

// Payment represents a payment transaction
type Payment struct {
	BaseModel
	UserID         uint    `json:"user_id" gorm:"not null"`
	ProductID      uint    `json:"product_id" gorm:"not null"`
	SubscriptionID *uint   `json:"subscription_id,omitempty"`
	Amount         int64   `json:"amount" gorm:"not null" validate:"min=0"` // Amount in cents
	Currency       string  `json:"currency" gorm:"not null" validate:"required,len=3"`
	Status         string  `json:"status" gorm:"not null;default:'pending'" validate:"oneof=pending succeeded failed canceled refunded"`
	PaymentMethod  string  `json:"payment_method" gorm:"not null" validate:"required,oneof=stripe polar"`
	Description    string  `json:"description,omitempty"`
	Metadata       JSONMap `json:"metadata,omitempty" gorm:"type:jsonb"`

	// External IDs for payment providers
	StripePaymentIntentID string `json:"stripe_payment_intent_id,omitempty"`
	StripeChargeID        string `json:"stripe_charge_id,omitempty"`
	PolarPaymentID        string `json:"polar_payment_id,omitempty"`

	// Relationships
	User         User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Product      Product       `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Subscription *Subscription `json:"subscription,omitempty" gorm:"foreignKey:SubscriptionID"`
}

// PaymentMethod represents a user's saved payment method
type PaymentMethod struct {
	BaseModel
	UserID    uint    `json:"user_id" gorm:"not null"`
	Type      string  `json:"type" gorm:"not null" validate:"required,oneof=card bank_account"`
	IsDefault bool    `json:"is_default" gorm:"default:false"`
	Last4     string  `json:"last4,omitempty"`
	Brand     string  `json:"brand,omitempty"`
	ExpMonth  int     `json:"exp_month,omitempty"`
	ExpYear   int     `json:"exp_year,omitempty"`
	Metadata  JSONMap `json:"metadata,omitempty" gorm:"type:jsonb"`

	// External IDs for payment providers
	StripePaymentMethodID string `json:"stripe_payment_method_id,omitempty"`
	PolarPaymentMethodID  string `json:"polar_payment_method_id,omitempty"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// WebhookEvent represents a webhook event from payment providers
type WebhookEvent struct {
	BaseModel
	Provider    string     `json:"provider" gorm:"not null" validate:"required,oneof=stripe polar"`
	EventType   string     `json:"event_type" gorm:"not null"`
	EventID     string     `json:"event_id" gorm:"not null;uniqueIndex"`
	Processed   bool       `json:"processed" gorm:"default:false"`
	Data        JSONMap    `json:"data" gorm:"type:jsonb"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	Error       string     `json:"error,omitempty"`
}

// BeforeCreate hook for Product
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.Currency == "" {
		p.Currency = "usd"
	}
	return nil
}

// BeforeCreate hook for Plan
func (p *Plan) BeforeCreate(tx *gorm.DB) error {
	if p.Currency == "" {
		p.Currency = "usd"
	}
	if p.IntervalCount == 0 {
		p.IntervalCount = 1
	}
	return nil
}

// BeforeCreate hook for Payment
func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.Currency == "" {
		p.Currency = "usd"
	}
	return nil
}

// IsActive checks if subscription is currently active
func (s *Subscription) IsActive() bool {
	return s.Status == "active" || s.Status == "trialing"
}

// IsCanceled checks if subscription is canceled
func (s *Subscription) IsCanceled() bool {
	return s.Status == "canceled" || s.Status == "incomplete_expired"
}

// IsPastDue checks if subscription is past due
func (s *Subscription) IsPastDue() bool {
	return s.Status == "past_due"
}

// IsTrial checks if subscription is in trial period
func (s *Subscription) IsTrial() bool {
	if s.TrialEnd == nil {
		return false
	}
	return time.Now().Before(*s.TrialEnd)
}

// GetAmountInDollars returns the amount in dollars
func (p *Payment) GetAmountInDollars() float64 {
	return float64(p.Amount) / 100.0
}

// GetPriceInDollars returns the price in dollars
func (pr *Product) GetPriceInDollars() float64 {
	return float64(pr.Price) / 100.0
}

// GetPriceInDollars returns the price in dollars
func (pl *Plan) GetPriceInDollars() float64 {
	return float64(pl.Price) / 100.0
}
