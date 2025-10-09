package models

import (
	"time"

	"gorm.io/gorm"
)

// PushNotification represents a push notification
type PushNotification struct {
	BaseModel
	Title       string             `json:"title" gorm:"not null"`
	Body        string             `json:"body" gorm:"not null"`
	Data        JSONMap            `json:"data"`
	Target      NotificationTarget `json:"target" gorm:"embedded;embeddedPrefix:target_"`
	ScheduledAt *time.Time         `json:"scheduled_at"`
	ExpiresAt   *time.Time         `json:"expires_at"`
	Status      string             `json:"status" gorm:"default:'pending';index:idx_push_notifications_status"`
	Priority    string             `json:"priority" gorm:"default:'normal'"` // low, normal, high
	Platform    string             `json:"platform"`                         // all, ios, android, web
	TemplateID  *uint              `json:"template_id"`
	UserID      *uint              `json:"user_id"`
	SegmentID   *uint              `json:"segment_id"`
	SentAt      *time.Time         `json:"sent_at"`
	DeliveredAt *time.Time         `json:"delivered_at"`
	OpenedAt    *time.Time         `json:"opened_at"`
	FailedAt    *time.Time         `json:"failed_at"`
	ErrorMsg    string             `json:"error_msg"`
	RetryCount  int                `json:"retry_count" gorm:"default:0"`
	MaxRetries  int                `json:"max_retries" gorm:"default:3"`

	// Relationships
	Template *NotificationTemplate `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	User     *User                 `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Segment  *NotificationSegment  `json:"segment,omitempty" gorm:"foreignKey:SegmentID"`
}

// NotificationTarget represents the target for a notification
type NotificationTarget struct {
	Type        string   `json:"type"`         // all, user, segment, device, custom
	UserIDs     []uint   `json:"user_ids"`     // Specific user IDs
	DeviceIDs   []string `json:"device_ids"`   // Specific device IDs
	SegmentIDs  []uint   `json:"segment_ids"`  // Segment IDs
	CustomQuery string   `json:"custom_query"` // Custom query for targeting
}

// NotificationTemplate represents a reusable notification template
type NotificationTemplate struct {
	BaseModel
	Name      string   `json:"name" gorm:"not null;uniqueIndex"`
	Title     string   `json:"title" gorm:"not null"`
	Body      string   `json:"body" gorm:"not null"`
	Data      JSONMap  `json:"data"`
	Platform  string   `json:"platform" gorm:"default:'all'"` // all, ios, android, web
	Category  string   `json:"category"`
	IsActive  bool     `json:"is_active" gorm:"default:true"`
	Variables []string `json:"variables"` // Template variables like {{name}}, {{email}}
	CreatedBy uint     `json:"created_by"`

	// Relationships
	Creator *User `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// NotificationSegment represents a user segment for targeting
type NotificationSegment struct {
	BaseModel
	Name        string `json:"name" gorm:"not null;uniqueIndex"`
	Description string `json:"description"`
	Query       string `json:"query"`      // SQL query or filter criteria
	UserCount   int    `json:"user_count"` // Cached user count
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	CreatedBy   uint   `json:"created_by"`

	// Relationships
	Creator *User `json:"creator,omitempty" gorm:"foreignKey:CreatedBy"`
}

// DeviceToken represents a device token for push notifications
type DeviceToken struct {
	BaseModel
	UserID     uint       `json:"user_id" gorm:"not null;index:idx_device_tokens_user"`
	Token      string     `json:"token" gorm:"not null;uniqueIndex"`
	Platform   string     `json:"platform" gorm:"not null"` // ios, android, web
	DeviceID   string     `json:"device_id"`
	AppVersion string     `json:"app_version"`
	OSVersion  string     `json:"os_version"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	LastUsedAt time.Time  `json:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// NotificationAnalytics represents analytics for notifications
type NotificationAnalytics struct {
	BaseModel
	NotificationID uint      `json:"notification_id" gorm:"not null;index:idx_notification_analytics_notification"`
	UserID         uint      `json:"user_id" gorm:"not null;index:idx_notification_analytics_user"`
	Event          string    `json:"event" gorm:"not null"` // sent, delivered, opened, failed
	Timestamp      time.Time `json:"timestamp" gorm:"not null;index:idx_notification_analytics_timestamp"`
	DeviceInfo     JSONMap   `json:"device_info"`
	Location       JSONMap   `json:"location"`

	// Relationships
	Notification *PushNotification `json:"notification,omitempty" gorm:"foreignKey:NotificationID"`
	User         *User             `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// NotificationStatus constants
const (
	NotificationStatusPending   = "pending"
	NotificationStatusScheduled = "scheduled"
	NotificationStatusSending   = "sending"
	NotificationStatusSent      = "sent"
	NotificationStatusDelivered = "delivered"
	NotificationStatusOpened    = "opened"
	NotificationStatusFailed    = "failed"
	NotificationStatusCancelled = "cancelled"
)

// NotificationPriority constants
const (
	NotificationPriorityLow    = "low"
	NotificationPriorityNormal = "normal"
	NotificationPriorityHigh   = "high"
)

// Platform constants
const (
	PlatformAll     = "all"
	PlatformIOS     = "ios"
	PlatformAndroid = "android"
	PlatformWeb     = "web"
)

// TargetType constants
const (
	TargetTypeAll     = "all"
	TargetTypeUser    = "user"
	TargetTypeSegment = "segment"
	TargetTypeDevice  = "device"
	TargetTypeCustom  = "custom"
)

// AnalyticsEvent constants
const (
	AnalyticsEventSent      = "sent"
	AnalyticsEventDelivered = "delivered"
	AnalyticsEventOpened    = "opened"
	AnalyticsEventFailed    = "failed"
)

// BeforeCreate hook for PushNotification
func (pn *PushNotification) BeforeCreate(tx *gorm.DB) error {
	if pn.Status == "" {
		pn.Status = NotificationStatusPending
	}
	if pn.Priority == "" {
		pn.Priority = NotificationPriorityNormal
	}
	if pn.Platform == "" {
		pn.Platform = PlatformAll
	}
	return nil
}

// CanRetry checks if the notification can be retried
func (pn *PushNotification) CanRetry() bool {
	return pn.Status == NotificationStatusFailed && pn.RetryCount < pn.MaxRetries
}

// MarkAsSending marks the notification as being sent
func (pn *PushNotification) MarkAsSending() {
	pn.Status = NotificationStatusSending
	now := time.Now()
	pn.SentAt = &now
}

// MarkAsSent marks the notification as sent
func (pn *PushNotification) MarkAsSent() {
	pn.Status = NotificationStatusSent
	now := time.Now()
	pn.SentAt = &now
}

// MarkAsDelivered marks the notification as delivered
func (pn *PushNotification) MarkAsDelivered() {
	pn.Status = NotificationStatusDelivered
	now := time.Now()
	pn.DeliveredAt = &now
}

// MarkAsOpened marks the notification as opened
func (pn *PushNotification) MarkAsOpened() {
	pn.Status = NotificationStatusOpened
	now := time.Now()
	pn.OpenedAt = &now
}

// MarkAsFailed marks the notification as failed and increments retry count
func (pn *PushNotification) MarkAsFailed(errorMsg string) {
	pn.Status = NotificationStatusFailed
	pn.ErrorMsg = errorMsg
	pn.RetryCount++
	now := time.Now()
	pn.FailedAt = &now
}

// IsExpired checks if the notification is expired
func (pn *PushNotification) IsExpired() bool {
	if pn.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*pn.ExpiresAt)
}

// IsScheduled checks if the notification is scheduled for future delivery
func (pn *PushNotification) IsScheduled() bool {
	return pn.ScheduledAt != nil && time.Now().Before(*pn.ScheduledAt)
}

// ShouldSend checks if the notification should be sent now
func (pn *PushNotification) ShouldSend() bool {
	if pn.Status != NotificationStatusPending && pn.Status != NotificationStatusScheduled {
		return false
	}
	if pn.IsExpired() {
		return false
	}
	if pn.IsScheduled() {
		return false
	}
	return true
}

// UpdateUserCount updates the user count for a segment
func (ns *NotificationSegment) UpdateUserCount(db *gorm.DB) error {
	// This would typically execute the query and count users
	// For now, we'll just set a default value
	ns.UserCount = 0
	return nil
}

// IsValid checks if the device token is valid
func (dt *DeviceToken) IsValid() bool {
	if !dt.IsActive {
		return false
	}
	if dt.ExpiresAt != nil && time.Now().After(*dt.ExpiresAt) {
		return false
	}
	return true
}

// UpdateLastUsed updates the last used timestamp
func (dt *DeviceToken) UpdateLastUsed() {
	dt.LastUsedAt = time.Now()
}
