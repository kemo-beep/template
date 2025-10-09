package models

import (
	"time"

	"gorm.io/gorm"
)

// OfflineOperation represents a queued operation for offline sync
type OfflineOperation struct {
	BaseModel
	UserID        uint       `json:"user_id" gorm:"not null;index:idx_offline_operations_user_status"`
	OperationID   string     `json:"operation_id" gorm:"uniqueIndex:idx_user_operation;not null"`
	OperationType string     `json:"operation_type" gorm:"not null;index:idx_offline_operations_operation_type"` // create, update, delete
	TableName     string     `json:"table_name" gorm:"not null"`
	RecordID      string     `json:"record_id"`
	Data          JSONMap    `json:"data" gorm:"type:text"`
	Status        string     `json:"status" gorm:"default:'pending';index:idx_offline_operations_user_status"` // pending, processing, completed, failed
	RetryCount    int        `json:"retry_count" gorm:"default:0"`
	MaxRetries    int        `json:"max_retries" gorm:"default:3"`
	ErrorMessage  string     `json:"error_message"`
	ProcessedAt   *time.Time `json:"processed_at"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// SyncConflict represents a data synchronization conflict
type SyncConflict struct {
	BaseModel
	UserID             uint       `json:"user_id" gorm:"not null;index:idx_sync_conflicts_user_status"`
	TableName          string     `json:"table_name" gorm:"not null"`
	RecordID           string     `json:"record_id" gorm:"not null"`
	LocalData          JSONMap    `json:"local_data" gorm:"type:text"`
	ServerData         JSONMap    `json:"server_data" gorm:"type:text"`
	ConflictType       string     `json:"conflict_type" gorm:"not null"` // version_mismatch, concurrent_edit, deleted_modified
	ResolutionStrategy string     `json:"resolution_strategy"`           // server_wins, client_wins, merge, manual
	ResolvedData       JSONMap    `json:"resolved_data" gorm:"type:text"`
	Status             string     `json:"status" gorm:"default:'pending';index:idx_sync_conflicts_user_status"` // pending, resolved, ignored
	ResolvedAt         *time.Time `json:"resolved_at"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// DataVersion represents version information for conflict detection
type DataVersion struct {
	BaseModel
	UserID         uint      `json:"user_id" gorm:"not null;index:idx_data_versions_user_table"`
	TableName      string    `json:"table_name" gorm:"not null"`
	RecordID       string    `json:"record_id" gorm:"not null"`
	Version        int       `json:"version" gorm:"not null;default:1"`
	LastModifiedBy string    `json:"last_modified_by" gorm:"not null"` // client or server
	LastModifiedAt time.Time `json:"last_modified_at" gorm:"not null;index:idx_data_versions_last_modified"`
	Checksum       string    `json:"checksum"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// SyncStatus represents the sync status for a user
type SyncStatus struct {
	BaseModel
	UserID                 uint       `json:"user_id" gorm:"uniqueIndex;not null"`
	LastSyncAt             *time.Time `json:"last_sync_at"`
	SyncToken              string     `json:"sync_token"`
	PendingOperationsCount int        `json:"pending_operations_count" gorm:"default:0"`
	ConflictsCount         int        `json:"conflicts_count" gorm:"default:0"`
	IsOnline               bool       `json:"is_online" gorm:"default:true"`
	LastOnlineAt           time.Time  `json:"last_online_at"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// SyncHistory represents sync operation history
type SyncHistory struct {
	BaseModel
	UserID              uint   `json:"user_id" gorm:"not null;index:idx_sync_history_user"`
	SyncType            string `json:"sync_type" gorm:"not null;index:idx_sync_history_sync_type"` // full, incremental, conflict_resolution
	OperationsProcessed int    `json:"operations_processed" gorm:"default:0"`
	ConflictsResolved   int    `json:"conflicts_resolved" gorm:"default:0"`
	DurationMs          int    `json:"duration_ms"`
	Success             bool   `json:"success" gorm:"default:true"`
	ErrorMessage        string `json:"error_message"`
	SyncToken           string `json:"sync_token"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// OperationType constants
const (
	OperationTypeCreate = "create"
	OperationTypeUpdate = "update"
	OperationTypeDelete = "delete"
)

// OperationStatus constants
const (
	OperationStatusPending    = "pending"
	OperationStatusProcessing = "processing"
	OperationStatusCompleted  = "completed"
	OperationStatusFailed     = "failed"
)

// ConflictType constants
const (
	ConflictTypeVersionMismatch = "version_mismatch"
	ConflictTypeConcurrentEdit  = "concurrent_edit"
	ConflictTypeDeletedModified = "deleted_modified"
)

// ResolutionStrategy constants
const (
	ResolutionStrategyServerWins = "server_wins"
	ResolutionStrategyClientWins = "client_wins"
	ResolutionStrategyMerge      = "merge"
	ResolutionStrategyManual     = "manual"
)

// ConflictStatus constants
const (
	ConflictStatusPending  = "pending"
	ConflictStatusResolved = "resolved"
	ConflictStatusIgnored  = "ignored"
)

// SyncType constants
const (
	SyncTypeFull               = "full"
	SyncTypeIncremental        = "incremental"
	SyncTypeSelective          = "selective"
	SyncTypeConflictResolution = "conflict_resolution"
)

// BeforeCreate hook for OfflineOperation
func (o *OfflineOperation) BeforeCreate(tx *gorm.DB) error {
	if o.OperationID == "" {
		o.OperationID = generateOperationID()
	}
	return nil
}

// CanRetry checks if the operation can be retried
func (o *OfflineOperation) CanRetry() bool {
	return o.Status == OperationStatusFailed && o.RetryCount < o.MaxRetries
}

// MarkAsProcessing marks the operation as processing
func (o *OfflineOperation) MarkAsProcessing() {
	o.Status = OperationStatusProcessing
	now := time.Now()
	o.ProcessedAt = &now
}

// MarkAsCompleted marks the operation as completed
func (o *OfflineOperation) MarkAsCompleted() {
	o.Status = OperationStatusCompleted
}

// MarkAsFailed marks the operation as failed and increments retry count
func (o *OfflineOperation) MarkAsFailed(errorMessage string) {
	o.Status = OperationStatusFailed
	o.ErrorMessage = errorMessage
	o.RetryCount++
}

// IsResolved checks if the conflict is resolved
func (c *SyncConflict) IsResolved() bool {
	return c.Status == ConflictStatusResolved
}

// Resolve marks the conflict as resolved
func (c *SyncConflict) Resolve(resolvedData JSONMap) {
	c.Status = ConflictStatusResolved
	c.ResolvedData = resolvedData
	now := time.Now()
	c.ResolvedAt = &now
}

// Ignore marks the conflict as ignored
func (c *SyncConflict) Ignore() {
	c.Status = ConflictStatusIgnored
	now := time.Now()
	c.ResolvedAt = &now
}

// UpdateVersion increments the version number
func (dv *DataVersion) UpdateVersion(modifiedBy string) {
	dv.Version++
	dv.LastModifiedBy = modifiedBy
	dv.LastModifiedAt = time.Now()
}

// IsStale checks if the data version is stale compared to another version
func (dv *DataVersion) IsStale(other *DataVersion) bool {
	return dv.Version < other.Version
}

// UpdateSyncStatus updates the sync status for a user
func (ss *SyncStatus) UpdateSyncStatus(syncType string, operationsProcessed, conflictsResolved int, durationMs int, success bool, errorMessage, syncToken string) {
	now := time.Now()
	ss.LastSyncAt = &now
	ss.SyncToken = syncToken
	ss.PendingOperationsCount -= operationsProcessed
	ss.ConflictsCount -= conflictsResolved

	// Create sync history record
	history := &SyncHistory{
		UserID:              ss.UserID,
		SyncType:            syncType,
		OperationsProcessed: operationsProcessed,
		ConflictsResolved:   conflictsResolved,
		DurationMs:          durationMs,
		Success:             success,
		ErrorMessage:        errorMessage,
		SyncToken:           syncToken,
	}

	// This would typically be saved to the database
	_ = history
}

// SetOnline sets the user as online
func (ss *SyncStatus) SetOnline() {
	ss.IsOnline = true
	ss.LastOnlineAt = time.Now()
}

// SetOffline sets the user as offline
func (ss *SyncStatus) SetOffline() {
	ss.IsOnline = false
}

// generateOperationID generates a unique operation ID
func generateOperationID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
