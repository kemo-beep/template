package services

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"mobile-backend/models"
)

// OfflineSyncService handles offline-first data synchronization
type OfflineSyncService struct {
	db               *gorm.DB
	redis            *redis.Client
	cache            *CacheService
	wsService        *WebSocketService
	logger           *zap.Logger
	conflictResolver *ConflictResolver
}

// NewOfflineSyncService creates a new offline sync service
func NewOfflineSyncService(db *gorm.DB, redis *redis.Client, cache *CacheService, wsService *WebSocketService, logger *zap.Logger) *OfflineSyncService {
	return &OfflineSyncService{
		db:               db,
		redis:            redis,
		cache:            cache,
		wsService:        wsService,
		logger:           logger,
		conflictResolver: NewConflictResolver(db, logger),
	}
}

// QueueOperation queues an operation for offline sync
func (os *OfflineSyncService) QueueOperation(ctx context.Context, userID uint, operation *models.OfflineOperation) error {
	// Set user ID
	operation.UserID = userID

	// Generate operation ID if not provided
	if operation.OperationID == "" {
		operation.OperationID = generateOperationID()
	}

	// Set default status
	operation.Status = models.OperationStatusPending

	// Save to database
	if err := os.db.Create(operation).Error; err != nil {
		return fmt.Errorf("failed to queue operation: %w", err)
	}

	// Update sync status
	if err := os.updateSyncStatus(ctx, userID, 1, 0); err != nil {
		os.logger.Warn("Failed to update sync status", zap.Error(err))
	}

	// Notify user via WebSocket
	if os.wsService != nil {
		os.wsService.SendDataUpdate(ctx, userID, "offline_operation_queued", map[string]interface{}{
			"operation_id":   operation.OperationID,
			"operation_type": operation.OperationType,
			"table_name":     operation.TableName,
			"record_id":      operation.RecordID,
		})
	}

	os.logger.Info("Operation queued for offline sync",
		zap.Uint("user_id", userID),
		zap.String("operation_id", operation.OperationID),
		zap.String("operation_type", operation.OperationType),
		zap.String("table_name", operation.TableName))

	return nil
}

// SyncUserData syncs all pending operations for a user
func (os *OfflineSyncService) SyncUserData(ctx context.Context, userID uint) error {
	startTime := time.Now()

	// Get pending operations
	var operations []models.OfflineOperation
	if err := os.db.Where("user_id = ? AND status = ?", userID, models.OperationStatusPending).
		Order("created_at ASC").Find(&operations).Error; err != nil {
		return fmt.Errorf("failed to get pending operations: %w", err)
	}

	operationsProcessed := 0
	conflictsResolved := 0

	// Process each operation
	for _, operation := range operations {
		if err := os.processOperation(ctx, &operation); err != nil {
			os.logger.Error("Failed to process operation",
				zap.String("operation_id", operation.OperationID),
				zap.Error(err))
			continue
		}
		operationsProcessed++
	}

	// Resolve conflicts
	conflictsResolved, err := os.resolveConflicts(ctx, userID)
	if err != nil {
		os.logger.Error("Failed to resolve conflicts", zap.Error(err))
	}

	// Update sync status
	durationMs := int(time.Since(startTime).Milliseconds())
	syncToken := generateSyncToken()

	// Get or create sync status
	var syncStatus models.SyncStatus
	if err := os.db.Where("user_id = ?", userID).First(&syncStatus).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			syncStatus = models.SyncStatus{
				UserID:       userID,
				IsOnline:     true,
				LastOnlineAt: time.Now(),
			}
			os.db.Create(&syncStatus)
		} else {
			return fmt.Errorf("failed to get sync status: %w", err)
		}
	}

	// Update sync status
	syncStatus.UpdateSyncStatus(models.SyncTypeIncremental, operationsProcessed, conflictsResolved, durationMs, true, "", syncToken)
	os.db.Save(&syncStatus)

	// Create sync history
	history := &models.SyncHistory{
		UserID:              userID,
		SyncType:            models.SyncTypeIncremental,
		OperationsProcessed: operationsProcessed,
		ConflictsResolved:   conflictsResolved,
		DurationMs:          durationMs,
		Success:             true,
		SyncToken:           syncToken,
	}
	os.db.Create(history)

	// Notify user via WebSocket
	if os.wsService != nil {
		os.wsService.SendDataUpdate(ctx, userID, "sync_completed", map[string]interface{}{
			"operations_processed": operationsProcessed,
			"conflicts_resolved":   conflictsResolved,
			"duration_ms":          durationMs,
			"sync_token":           syncToken,
		})
	}

	os.logger.Info("User data sync completed",
		zap.Uint("user_id", userID),
		zap.Int("operations_processed", operationsProcessed),
		zap.Int("conflicts_resolved", conflictsResolved),
		zap.Int("duration_ms", durationMs))

	return nil
}

// SyncUserDataSelective performs selective sync for a user
func (os *OfflineSyncService) SyncUserDataSelective(ctx context.Context, userID uint, lastSyncTime time.Time) error {
	startTime := time.Now()

	// Get selective sync data
	selectiveData, err := os.getSelectiveSyncData(ctx, userID, lastSyncTime)
	if err != nil {
		return fmt.Errorf("failed to get selective sync data: %w", err)
	}

	// Get pending operations
	var operations []models.OfflineOperation
	if err := os.db.Where("user_id = ? AND status = ?", userID, models.OperationStatusPending).
		Order("created_at ASC").Find(&operations).Error; err != nil {
		return fmt.Errorf("failed to get pending operations: %w", err)
	}

	operationsProcessed := 0
	conflictsResolved := 0

	// Process each operation
	for _, operation := range operations {
		if err := os.processOperation(ctx, &operation); err != nil {
			os.logger.Error("Failed to process operation",
				zap.String("operation_id", operation.OperationID),
				zap.Error(err))
			continue
		}
		operationsProcessed++
	}

	// Resolve conflicts
	conflictsResolved, err = os.resolveConflicts(ctx, userID)
	if err != nil {
		os.logger.Error("Failed to resolve conflicts", zap.Error(err))
	}

	// Update sync status
	durationMs := int(time.Since(startTime).Milliseconds())
	syncToken := generateSyncToken()

	// Get or create sync status
	var syncStatus models.SyncStatus
	if err := os.db.Where("user_id = ?", userID).First(&syncStatus).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			syncStatus = models.SyncStatus{
				UserID:       userID,
				IsOnline:     true,
				LastOnlineAt: time.Now(),
			}
			os.db.Create(&syncStatus)
		} else {
			return fmt.Errorf("failed to get sync status: %w", err)
		}
	}

	// Update sync status
	syncStatus.UpdateSyncStatus(models.SyncTypeSelective, operationsProcessed, conflictsResolved, durationMs, true, "", syncToken)
	os.db.Save(&syncStatus)

	// Create sync history
	history := &models.SyncHistory{
		UserID:              userID,
		SyncType:            models.SyncTypeSelective,
		OperationsProcessed: operationsProcessed,
		ConflictsResolved:   conflictsResolved,
		DurationMs:          durationMs,
		Success:             true,
		SyncToken:           syncToken,
	}
	os.db.Create(history)

	// Notify user via WebSocket with selective data
	if os.wsService != nil {
		os.wsService.SendDataUpdate(ctx, userID, "sync_completed", map[string]interface{}{
			"operations_processed": operationsProcessed,
			"conflicts_resolved":   conflictsResolved,
			"duration_ms":          durationMs,
			"sync_token":           syncToken,
			"selective_data":       selectiveData,
		})
	}

	os.logger.Info("User selective sync completed",
		zap.Uint("user_id", userID),
		zap.Int("operations_processed", operationsProcessed),
		zap.Int("conflicts_resolved", conflictsResolved),
		zap.Int("duration_ms", durationMs))

	return nil
}

// processOperation processes a single offline operation
func (os *OfflineSyncService) processOperation(ctx context.Context, operation *models.OfflineOperation) error {
	operation.MarkAsProcessing()
	os.db.Save(operation)

	// Check for conflicts before processing
	conflict, err := os.checkForConflicts(ctx, operation)
	if err != nil {
		return fmt.Errorf("failed to check for conflicts: %w", err)
	}

	if conflict != nil {
		// Handle conflict
		if err := os.handleConflict(ctx, operation, conflict); err != nil {
			operation.MarkAsFailed(fmt.Sprintf("conflict resolution failed: %v", err))
			os.db.Save(operation)
			return err
		}
	}

	// Process the operation based on type
	switch operation.OperationType {
	case models.OperationTypeCreate:
		err = os.processCreateOperation(ctx, operation)
	case models.OperationTypeUpdate:
		err = os.processUpdateOperation(ctx, operation)
	case models.OperationTypeDelete:
		err = os.processDeleteOperation(ctx, operation)
	default:
		err = fmt.Errorf("unknown operation type: %s", operation.OperationType)
	}

	if err != nil {
		operation.MarkAsFailed(err.Error())
		os.db.Save(operation)
		return err
	}

	operation.MarkAsCompleted()
	os.db.Save(operation)

	return nil
}

// processCreateOperation processes a create operation
func (os *OfflineSyncService) processCreateOperation(ctx context.Context, operation *models.OfflineOperation) error {
	os.logger.Info("Processing create operation",
		zap.String("operation_id", operation.OperationID),
		zap.String("table_name", operation.TableName),
		zap.String("record_id", operation.RecordID))

	// Perform actual database operation based on table name
	if err := os.executeCreateOperation(ctx, operation); err != nil {
		return fmt.Errorf("failed to execute create operation: %w", err)
	}

	// Update data version
	return os.updateDataVersion(ctx, operation.UserID, operation.TableName, operation.RecordID, "server")
}

// processUpdateOperation processes an update operation
func (os *OfflineSyncService) processUpdateOperation(ctx context.Context, operation *models.OfflineOperation) error {
	os.logger.Info("Processing update operation",
		zap.String("operation_id", operation.OperationID),
		zap.String("table_name", operation.TableName),
		zap.String("record_id", operation.RecordID))

	// Perform actual database operation based on table name
	if err := os.executeUpdateOperation(ctx, operation); err != nil {
		return fmt.Errorf("failed to execute update operation: %w", err)
	}

	// Update data version
	return os.updateDataVersion(ctx, operation.UserID, operation.TableName, operation.RecordID, "server")
}

// processDeleteOperation processes a delete operation
func (os *OfflineSyncService) processDeleteOperation(ctx context.Context, operation *models.OfflineOperation) error {
	os.logger.Info("Processing delete operation",
		zap.String("operation_id", operation.OperationID),
		zap.String("table_name", operation.TableName),
		zap.String("record_id", operation.RecordID))

	// Perform actual database operation based on table name
	if err := os.executeDeleteOperation(ctx, operation); err != nil {
		return fmt.Errorf("failed to execute delete operation: %w", err)
	}

	// Update data version
	return os.updateDataVersion(ctx, operation.UserID, operation.TableName, operation.RecordID, "server")
}

// checkForConflicts checks if there are any conflicts for the operation
func (os *OfflineSyncService) checkForConflicts(ctx context.Context, operation *models.OfflineOperation) (*models.SyncConflict, error) {
	// Get current data version
	var dataVersion models.DataVersion
	err := os.db.Where("user_id = ? AND table_name = ? AND record_id = ?",
		operation.UserID, operation.TableName, operation.RecordID).First(&dataVersion).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No existing version, no conflict
			return nil, nil
		}
		return nil, err
	}

	// Check if there's a version mismatch
	if operation.Data != nil {
		// Calculate checksum of operation data
		dataJSON, _ := json.Marshal(operation.Data)
		checksum := fmt.Sprintf("%x", md5.Sum(dataJSON))

		if dataVersion.Checksum != "" && dataVersion.Checksum != checksum {
			// Version mismatch detected
			conflict := &models.SyncConflict{
				UserID:       operation.UserID,
				TableName:    operation.TableName,
				RecordID:     operation.RecordID,
				LocalData:    operation.Data,
				ServerData:   operation.Data, // This would be the actual server data
				ConflictType: models.ConflictTypeVersionMismatch,
				Status:       models.ConflictStatusPending,
			}

			return conflict, nil
		}
	}

	return nil, nil
}

// handleConflict handles a sync conflict
func (os *OfflineSyncService) handleConflict(ctx context.Context, operation *models.OfflineOperation, conflict *models.SyncConflict) error {
	// Save conflict to database
	if err := os.db.Create(conflict).Error; err != nil {
		return fmt.Errorf("failed to save conflict: %w", err)
	}

	// Use conflict resolver to resolve the conflict
	resolvedData, err := os.conflictResolver.ResolveConflict(ctx, conflict)
	if err != nil {
		return fmt.Errorf("failed to resolve conflict: %w", err)
	}

	// Update operation with resolved data
	operation.Data = resolvedData
	os.db.Save(operation)

	// Mark conflict as resolved
	conflict.Resolve(resolvedData)
	os.db.Save(conflict)

	return nil
}

// resolveConflicts resolves all pending conflicts for a user
func (os *OfflineSyncService) resolveConflicts(ctx context.Context, userID uint) (int, error) {
	var conflicts []models.SyncConflict
	if err := os.db.Where("user_id = ? AND status = ?", userID, models.ConflictStatusPending).Find(&conflicts).Error; err != nil {
		return 0, fmt.Errorf("failed to get pending conflicts: %w", err)
	}

	resolvedCount := 0
	for _, conflict := range conflicts {
		resolvedData, err := os.conflictResolver.ResolveConflict(ctx, &conflict)
		if err != nil {
			os.logger.Error("Failed to resolve conflict",
				zap.Uint("conflict_id", conflict.ID),
				zap.Error(err))
			continue
		}

		conflict.Resolve(resolvedData)
		os.db.Save(&conflict)
		resolvedCount++
	}

	return resolvedCount, nil
}

// updateDataVersion updates the data version for a record
func (os *OfflineSyncService) updateDataVersion(ctx context.Context, userID uint, tableName, recordID, modifiedBy string) error {
	var dataVersion models.DataVersion
	err := os.db.Where("user_id = ? AND table_name = ? AND record_id = ?",
		userID, tableName, recordID).First(&dataVersion).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new data version
			dataVersion = models.DataVersion{
				UserID:         userID,
				TableName:      tableName,
				RecordID:       recordID,
				Version:        1,
				LastModifiedBy: modifiedBy,
				LastModifiedAt: time.Now(),
			}
		} else {
			return err
		}
	} else {
		// Update existing data version
		dataVersion.UpdateVersion(modifiedBy)
	}

	// Calculate checksum
	dataJSON, _ := json.Marshal(map[string]interface{}{
		"table_name":  tableName,
		"record_id":   recordID,
		"version":     dataVersion.Version,
		"modified_by": modifiedBy,
	})
	dataVersion.Checksum = fmt.Sprintf("%x", md5.Sum(dataJSON))

	// Save data version
	if dataVersion.ID == 0 {
		return os.db.Create(&dataVersion).Error
	}
	return os.db.Save(&dataVersion).Error
}

// updateSyncStatus updates the sync status for a user
func (os *OfflineSyncService) updateSyncStatus(ctx context.Context, userID uint, pendingOpsDelta, conflictsDelta int) error {
	var syncStatus models.SyncStatus
	err := os.db.Where("user_id = ?", userID).First(&syncStatus).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new sync status
			syncStatus = models.SyncStatus{
				UserID:                 userID,
				PendingOperationsCount: pendingOpsDelta,
				ConflictsCount:         conflictsDelta,
				IsOnline:               true,
				LastOnlineAt:           time.Now(),
			}
			return os.db.Create(&syncStatus).Error
		}
		return err
	}

	// Update counts
	syncStatus.PendingOperationsCount += pendingOpsDelta
	syncStatus.ConflictsCount += conflictsDelta

	return os.db.Save(&syncStatus).Error
}

// GetSyncStatus returns the sync status for a user
func (os *OfflineSyncService) GetSyncStatus(ctx context.Context, userID uint) (*models.SyncStatus, error) {
	var syncStatus models.SyncStatus
	err := os.db.Where("user_id = ?", userID).First(&syncStatus).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default sync status
			return &models.SyncStatus{
				UserID:                 userID,
				PendingOperationsCount: 0,
				ConflictsCount:         0,
				IsOnline:               true,
				LastOnlineAt:           time.Now(),
			}, nil
		}
		return nil, err
	}

	return &syncStatus, nil
}

// GetPendingOperations returns pending operations for a user
func (os *OfflineSyncService) GetPendingOperations(ctx context.Context, userID uint, limit int) ([]models.OfflineOperation, error) {
	var operations []models.OfflineOperation
	query := os.db.Where("user_id = ? AND status = ?", userID, models.OperationStatusPending).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&operations).Error; err != nil {
		return nil, err
	}

	return operations, nil
}

// GetConflicts returns pending conflicts for a user
func (os *OfflineSyncService) GetConflicts(ctx context.Context, userID uint, limit int) ([]models.SyncConflict, error) {
	var conflicts []models.SyncConflict
	query := os.db.Where("user_id = ? AND status = ?", userID, models.ConflictStatusPending).
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&conflicts).Error; err != nil {
		return nil, err
	}

	return conflicts, nil
}

// GetSyncHistory returns sync history for a user
func (os *OfflineSyncService) GetSyncHistory(ctx context.Context, userID uint, limit int) ([]models.SyncHistory, error) {
	var history []models.SyncHistory
	query := os.db.Where("user_id = ?", userID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&history).Error; err != nil {
		return nil, err
	}

	return history, nil
}

// ForceSync forces a sync for a user
func (os *OfflineSyncService) ForceSync(ctx context.Context, userID uint) error {
	// Set user as online
	var syncStatus models.SyncStatus
	if err := os.db.Where("user_id = ?", userID).First(&syncStatus).Error; err == nil {
		syncStatus.SetOnline()
		os.db.Save(&syncStatus)
	}

	// Perform sync
	return os.SyncUserData(ctx, userID)
}

// SetUserOnline sets a user as online
func (os *OfflineSyncService) SetUserOnline(ctx context.Context, userID uint) error {
	var syncStatus models.SyncStatus
	err := os.db.Where("user_id = ?", userID).First(&syncStatus).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new sync status
			syncStatus = models.SyncStatus{
				UserID:       userID,
				IsOnline:     true,
				LastOnlineAt: time.Now(),
			}
			return os.db.Create(&syncStatus).Error
		}
		return err
	}

	syncStatus.SetOnline()
	return os.db.Save(&syncStatus).Error
}

// SetUserOffline sets a user as offline
func (os *OfflineSyncService) SetUserOffline(ctx context.Context, userID uint) error {
	var syncStatus models.SyncStatus
	err := os.db.Where("user_id = ?", userID).First(&syncStatus).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new sync status
			syncStatus = models.SyncStatus{
				UserID:       userID,
				IsOnline:     false,
				LastOnlineAt: time.Now(),
			}
			return os.db.Create(&syncStatus).Error
		}
		return err
	}

	syncStatus.SetOffline()
	return os.db.Save(&syncStatus).Error
}

// generateOperationID generates a unique operation ID
func generateOperationID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// generateSyncToken generates a unique sync token
func generateSyncToken() string {
	return time.Now().Format("20060102150405") + "-" + randomString(16)
}

// executeCreateOperation performs the actual database create operation
func (os *OfflineSyncService) executeCreateOperation(ctx context.Context, operation *models.OfflineOperation) error {
	// Map table names to actual model types
	switch operation.TableName {
	case "users":
		return os.createUserRecord(ctx, operation)
	case "products":
		return os.createProductRecord(ctx, operation)
	case "orders":
		return os.createOrderRecord(ctx, operation)
	default:
		// Generic create operation for unknown tables
		return os.createGenericRecord(ctx, operation)
	}
}

// executeUpdateOperation performs the actual database update operation
func (os *OfflineSyncService) executeUpdateOperation(ctx context.Context, operation *models.OfflineOperation) error {
	// Map table names to actual model types
	switch operation.TableName {
	case "users":
		return os.updateUserRecord(ctx, operation)
	case "products":
		return os.updateProductRecord(ctx, operation)
	case "orders":
		return os.updateOrderRecord(ctx, operation)
	default:
		// Generic update operation for unknown tables
		return os.updateGenericRecord(ctx, operation)
	}
}

// executeDeleteOperation performs the actual database delete operation
func (os *OfflineSyncService) executeDeleteOperation(ctx context.Context, operation *models.OfflineOperation) error {
	// Map table names to actual model types
	switch operation.TableName {
	case "users":
		return os.deleteUserRecord(ctx, operation)
	case "products":
		return os.deleteProductRecord(ctx, operation)
	case "orders":
		return os.deleteOrderRecord(ctx, operation)
	default:
		// Generic delete operation for unknown tables
		return os.deleteGenericRecord(ctx, operation)
	}
}

// createUserRecord creates a user record
func (os *OfflineSyncService) createUserRecord(ctx context.Context, operation *models.OfflineOperation) error {
	var user models.User
	dataBytes, err := json.Marshal(operation.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &user); err != nil {
		return fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	// Set the ID from record_id if provided
	if operation.RecordID != "" {
		// Parse record_id as uint for user ID
		if id, err := strconv.ParseUint(operation.RecordID, 10, 32); err == nil {
			user.ID = uint(id)
		}
	}

	return os.db.Create(&user).Error
}

// updateUserRecord updates a user record
func (os *OfflineSyncService) updateUserRecord(ctx context.Context, operation *models.OfflineOperation) error {
	var user models.User
	dataBytes, err := json.Marshal(operation.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &user); err != nil {
		return fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	// Parse record_id as uint for user ID
	id, err := strconv.ParseUint(operation.RecordID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid record ID: %w", err)
	}

	return os.db.Model(&models.User{}).Where("id = ?", uint(id)).Updates(&user).Error
}

// deleteUserRecord deletes a user record
func (os *OfflineSyncService) deleteUserRecord(ctx context.Context, operation *models.OfflineOperation) error {
	id, err := strconv.ParseUint(operation.RecordID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid record ID: %w", err)
	}

	return os.db.Delete(&models.User{}, uint(id)).Error
}

// createProductRecord creates a product record
func (os *OfflineSyncService) createProductRecord(ctx context.Context, operation *models.OfflineOperation) error {
	var product models.Product
	dataBytes, err := json.Marshal(operation.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal product data: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &product); err != nil {
		return fmt.Errorf("failed to unmarshal product data: %w", err)
	}

	if operation.RecordID != "" {
		if id, err := strconv.ParseUint(operation.RecordID, 10, 32); err == nil {
			product.ID = uint(id)
		}
	}

	return os.db.Create(&product).Error
}

// updateProductRecord updates a product record
func (os *OfflineSyncService) updateProductRecord(ctx context.Context, operation *models.OfflineOperation) error {
	var product models.Product
	dataBytes, err := json.Marshal(operation.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal product data: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &product); err != nil {
		return fmt.Errorf("failed to unmarshal product data: %w", err)
	}

	id, err := strconv.ParseUint(operation.RecordID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid record ID: %w", err)
	}

	return os.db.Model(&models.Product{}).Where("id = ?", uint(id)).Updates(&product).Error
}

// deleteProductRecord deletes a product record
func (os *OfflineSyncService) deleteProductRecord(ctx context.Context, operation *models.OfflineOperation) error {
	id, err := strconv.ParseUint(operation.RecordID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid record ID: %w", err)
	}

	return os.db.Delete(&models.Product{}, uint(id)).Error
}

// createOrderRecord creates an order record
func (os *OfflineSyncService) createOrderRecord(ctx context.Context, operation *models.OfflineOperation) error {
	var order models.Order
	dataBytes, err := json.Marshal(operation.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal order data: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order data: %w", err)
	}

	if operation.RecordID != "" {
		if id, err := strconv.ParseUint(operation.RecordID, 10, 32); err == nil {
			order.ID = uint(id)
		}
	}

	return os.db.Create(&order).Error
}

// updateOrderRecord updates an order record
func (os *OfflineSyncService) updateOrderRecord(ctx context.Context, operation *models.OfflineOperation) error {
	var order models.Order
	dataBytes, err := json.Marshal(operation.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal order data: %w", err)
	}
	if err := json.Unmarshal(dataBytes, &order); err != nil {
		return fmt.Errorf("failed to unmarshal order data: %w", err)
	}

	id, err := strconv.ParseUint(operation.RecordID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid record ID: %w", err)
	}

	return os.db.Model(&models.Order{}).Where("id = ?", uint(id)).Updates(&order).Error
}

// deleteOrderRecord deletes an order record
func (os *OfflineSyncService) deleteOrderRecord(ctx context.Context, operation *models.OfflineOperation) error {
	id, err := strconv.ParseUint(operation.RecordID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid record ID: %w", err)
	}

	return os.db.Delete(&models.Order{}, uint(id)).Error
}

// createGenericRecord creates a record in any table
func (os *OfflineSyncService) createGenericRecord(ctx context.Context, operation *models.OfflineOperation) error {
	// For generic records, we'll use raw SQL
	query := fmt.Sprintf("INSERT INTO %s (data) VALUES (?)", operation.TableName)
	return os.db.Exec(query, operation.Data).Error
}

// updateGenericRecord updates a record in any table
func (os *OfflineSyncService) updateGenericRecord(ctx context.Context, operation *models.OfflineOperation) error {
	query := fmt.Sprintf("UPDATE %s SET data = ? WHERE id = ?", operation.TableName)
	return os.db.Exec(query, operation.Data, operation.RecordID).Error
}

// deleteGenericRecord deletes a record from any table
func (os *OfflineSyncService) deleteGenericRecord(ctx context.Context, operation *models.OfflineOperation) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", operation.TableName)
	return os.db.Exec(query, operation.RecordID).Error
}

// retryFailedOperations retries failed operations with exponential backoff
func (os *OfflineSyncService) retryFailedOperations(ctx context.Context) error {
	var failedOps []models.OfflineOperation
	if err := os.db.Where("status = ? AND retry_count < max_retries", models.OperationStatusFailed).
		Find(&failedOps).Error; err != nil {
		return fmt.Errorf("failed to get failed operations: %w", err)
	}

	for _, operation := range failedOps {
		// Calculate exponential backoff delay
		delay := time.Duration(operation.RetryCount+1) * time.Minute

		// Check if enough time has passed since last retry
		if time.Since(operation.UpdatedAt) < delay {
			continue
		}

		// Reset operation status and retry
		operation.Status = models.OperationStatusPending
		operation.RetryCount++
		operation.ErrorMessage = ""
		os.db.Save(&operation)

		// Process the operation
		if err := os.processOperation(ctx, &operation); err != nil {
			os.logger.Error("Retry failed",
				zap.String("operation_id", operation.OperationID),
				zap.Int("retry_count", operation.RetryCount),
				zap.Error(err))
		}
	}

	return nil
}

// getSelectiveSyncData returns only changed data for a user since last sync
func (os *OfflineSyncService) getSelectiveSyncData(ctx context.Context, userID uint, lastSyncTime time.Time) (map[string]interface{}, error) {
	// Get all tables that have been modified since last sync
	var modifiedTables []string

	// Check each table for modifications
	tables := []string{"users", "products", "orders"}
	for _, table := range tables {
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE updated_at > ?", table)
		if err := os.db.Raw(query, lastSyncTime).Scan(&count).Error; err != nil {
			continue // Skip tables that don't exist or have errors
		}

		if count > 0 {
			modifiedTables = append(modifiedTables, table)
		}
	}

	// Build selective sync data
	syncData := make(map[string]interface{})
	for _, table := range modifiedTables {
		query := fmt.Sprintf("SELECT * FROM %s WHERE updated_at > ? ORDER BY updated_at ASC", table)
		var results []map[string]interface{}

		if err := os.db.Raw(query, lastSyncTime).Scan(&results).Error; err != nil {
			os.logger.Error("Failed to get selective sync data",
				zap.String("table", table),
				zap.Error(err))
			continue
		}

		syncData[table] = results
	}

	return syncData, nil
}

// StartRetryService starts the background retry service
func (os *OfflineSyncService) StartRetryService(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			os.logger.Info("Retry service stopped")
			return
		case <-ticker.C:
			if err := os.retryFailedOperations(ctx); err != nil {
				os.logger.Error("Failed to retry operations", zap.Error(err))
			}
		}
	}
}
