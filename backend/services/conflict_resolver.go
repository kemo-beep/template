package services

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"mobile-backend/models"
)

// ConflictResolver handles conflict resolution strategies
type ConflictResolver struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewConflictResolver creates a new conflict resolver
func NewConflictResolver(db *gorm.DB, logger *zap.Logger) *ConflictResolver {
	return &ConflictResolver{
		db:     db,
		logger: logger,
	}
}

// ResolveConflict resolves a sync conflict using the appropriate strategy
func (cr *ConflictResolver) ResolveConflict(ctx context.Context, conflict *models.SyncConflict) (models.JSONMap, error) {
	// Determine resolution strategy if not set
	if conflict.ResolutionStrategy == "" {
		conflict.ResolutionStrategy = cr.determineResolutionStrategy(conflict)
	}

	cr.logger.Info("Resolving conflict",
		zap.Uint("conflict_id", conflict.ID),
		zap.String("conflict_type", conflict.ConflictType),
		zap.String("resolution_strategy", conflict.ResolutionStrategy))

	switch conflict.ResolutionStrategy {
	case models.ResolutionStrategyServerWins:
		return cr.resolveServerWins(conflict)
	case models.ResolutionStrategyClientWins:
		return cr.resolveClientWins(conflict)
	case models.ResolutionStrategyMerge:
		return cr.resolveMerge(conflict)
	case models.ResolutionStrategyManual:
		return cr.resolveManual(conflict)
	default:
		return nil, fmt.Errorf("unknown resolution strategy: %s", conflict.ResolutionStrategy)
	}
}

// determineResolutionStrategy determines the best resolution strategy for a conflict
func (cr *ConflictResolver) determineResolutionStrategy(conflict *models.SyncConflict) string {
	switch conflict.ConflictType {
	case models.ConflictTypeVersionMismatch:
		// For version mismatches, prefer server wins by default
		return models.ResolutionStrategyServerWins
	case models.ConflictTypeConcurrentEdit:
		// For concurrent edits, try to merge
		return models.ResolutionStrategyMerge
	case models.ConflictTypeDeletedModified:
		// For deleted-modified conflicts, prefer server wins
		return models.ResolutionStrategyServerWins
	default:
		// Default to server wins
		return models.ResolutionStrategyServerWins
	}
}

// resolveServerWins resolves conflict by using server data
func (cr *ConflictResolver) resolveServerWins(conflict *models.SyncConflict) (models.JSONMap, error) {
	cr.logger.Info("Resolving conflict with server wins strategy",
		zap.Uint("conflict_id", conflict.ID))

	// Server data takes precedence
	return conflict.ServerData, nil
}

// resolveClientWins resolves conflict by using client data
func (cr *ConflictResolver) resolveClientWins(conflict *models.SyncConflict) (models.JSONMap, error) {
	cr.logger.Info("Resolving conflict with client wins strategy",
		zap.Uint("conflict_id", conflict.ID))

	// Client data takes precedence
	return conflict.LocalData, nil
}

// resolveMerge resolves conflict by merging data
func (cr *ConflictResolver) resolveMerge(conflict *models.SyncConflict) (models.JSONMap, error) {
	cr.logger.Info("Resolving conflict with merge strategy",
		zap.Uint("conflict_id", conflict.ID))

	// Start with server data as base
	merged := make(models.JSONMap)
	for k, v := range conflict.ServerData {
		merged[k] = v
	}

	// Merge in client data, giving priority to non-nil client values
	for k, clientValue := range conflict.LocalData {
		if clientValue != nil {
			// Check if server has this field
			if serverValue, exists := conflict.ServerData[k]; exists {
				// Both have the field, use more recent or merge based on type
				merged[k] = cr.mergeFieldValues(k, serverValue, clientValue)
			} else {
				// Client has field that server doesn't, add it
				merged[k] = clientValue
			}
		}
	}

	return merged, nil
}

// resolveManual resolves conflict by requiring manual intervention
func (cr *ConflictResolver) resolveManual(conflict *models.SyncConflict) (models.JSONMap, error) {
	cr.logger.Info("Resolving conflict with manual strategy",
		zap.Uint("conflict_id", conflict.ID))

	// For manual resolution, we'll use server data as default
	// but mark it for manual review
	conflict.Status = models.ConflictStatusPending // Keep as pending for manual review

	return conflict.ServerData, nil
}

// mergeFieldValues merges two field values based on their types
func (cr *ConflictResolver) mergeFieldValues(fieldName string, serverValue, clientValue interface{}) interface{} {
	// Convert to JSON for comparison
	serverJSON, _ := json.Marshal(serverValue)
	clientJSON, _ := json.Marshal(clientValue)

	// If they're the same, return either one
	if string(serverJSON) == string(clientJSON) {
		return serverValue
	}

	// For different types, prefer the more complex one
	switch serverValue.(type) {
	case map[string]interface{}:
		// Server has object, prefer it
		return serverValue
	case []interface{}:
		// Server has array, prefer it
		return serverValue
	case string:
		// Both are strings, prefer the longer one
		if clientStr, ok := clientValue.(string); ok {
			if len(clientStr) > len(serverValue.(string)) {
				return clientValue
			}
		}
		return serverValue
	case float64:
		// Both are numbers, prefer the larger one
		if clientNum, ok := clientValue.(float64); ok {
			if clientNum > serverValue.(float64) {
				return clientValue
			}
		}
		return serverValue
	case bool:
		// Both are booleans, prefer true
		if clientBool, ok := clientValue.(bool); ok {
			if clientBool {
				return clientValue
			}
		}
		return serverValue
	default:
		// Default to server value
		return serverValue
	}
}

// GetConflictResolutionStrategies returns available resolution strategies
func (cr *ConflictResolver) GetConflictResolutionStrategies() []string {
	return []string{
		models.ResolutionStrategyServerWins,
		models.ResolutionStrategyClientWins,
		models.ResolutionStrategyMerge,
		models.ResolutionStrategyManual,
	}
}

// GetConflictTypes returns available conflict types
func (cr *ConflictResolver) GetConflictTypes() []string {
	return []string{
		models.ConflictTypeVersionMismatch,
		models.ConflictTypeConcurrentEdit,
		models.ConflictTypeDeletedModified,
	}
}

// AnalyzeConflict analyzes a conflict and provides recommendations
func (cr *ConflictResolver) AnalyzeConflict(conflict *models.SyncConflict) ConflictAnalysis {
	analysis := ConflictAnalysis{
		ConflictID:          conflict.ID,
		ConflictType:        conflict.ConflictType,
		RecommendedStrategy: cr.determineResolutionStrategy(conflict),
		Severity:            cr.calculateSeverity(conflict),
		Description:         cr.generateDescription(conflict),
		FieldsInConflict:    cr.identifyConflictingFields(conflict),
	}

	return analysis
}

// ConflictAnalysis represents the analysis of a conflict
type ConflictAnalysis struct {
	ConflictID          uint     `json:"conflict_id"`
	ConflictType        string   `json:"conflict_type"`
	RecommendedStrategy string   `json:"recommended_strategy"`
	Severity            string   `json:"severity"` // low, medium, high, critical
	Description         string   `json:"description"`
	FieldsInConflict    []string `json:"fields_in_conflict"`
}

// calculateSeverity calculates the severity of a conflict
func (cr *ConflictResolver) calculateSeverity(conflict *models.SyncConflict) string {
	// Count conflicting fields
	conflictingFields := cr.identifyConflictingFields(conflict)
	fieldCount := len(conflictingFields)

	// Determine severity based on field count and conflict type
	switch {
	case fieldCount == 0:
		return "low"
	case fieldCount <= 2:
		return "medium"
	case fieldCount <= 5:
		return "high"
	default:
		return "critical"
	}
}

// generateDescription generates a human-readable description of the conflict
func (cr *ConflictResolver) generateDescription(conflict *models.SyncConflict) string {
	conflictingFields := cr.identifyConflictingFields(conflict)

	switch conflict.ConflictType {
	case models.ConflictTypeVersionMismatch:
		return fmt.Sprintf("Version mismatch detected for record %s in table %s. %d fields have conflicting values.",
			conflict.RecordID, conflict.TableName, len(conflictingFields))
	case models.ConflictTypeConcurrentEdit:
		return fmt.Sprintf("Concurrent edit detected for record %s in table %s. %d fields were modified simultaneously.",
			conflict.RecordID, conflict.TableName, len(conflictingFields))
	case models.ConflictTypeDeletedModified:
		return fmt.Sprintf("Record %s in table %s was deleted on server but modified on client.",
			conflict.RecordID, conflict.TableName)
	default:
		return fmt.Sprintf("Unknown conflict type for record %s in table %s.",
			conflict.RecordID, conflict.TableName)
	}
}

// identifyConflictingFields identifies which fields are in conflict
func (cr *ConflictResolver) identifyConflictingFields(conflict *models.SyncConflict) []string {
	var conflictingFields []string

	// Compare each field in local and server data
	for fieldName, localValue := range conflict.LocalData {
		if serverValue, exists := conflict.ServerData[fieldName]; exists {
			// Convert to JSON for comparison
			localJSON, _ := json.Marshal(localValue)
			serverJSON, _ := json.Marshal(serverValue)

			if string(localJSON) != string(serverJSON) {
				conflictingFields = append(conflictingFields, fieldName)
			}
		}
	}

	// Check for fields that exist in server but not in local
	for fieldName := range conflict.ServerData {
		if _, exists := conflict.LocalData[fieldName]; !exists {
			conflictingFields = append(conflictingFields, fieldName)
		}
	}

	return conflictingFields
}

// ResolveConflictWithStrategy resolves a conflict using a specific strategy
func (cr *ConflictResolver) ResolveConflictWithStrategy(ctx context.Context, conflictID uint, strategy string) (models.JSONMap, error) {
	var conflict models.SyncConflict
	if err := cr.db.First(&conflict, conflictID).Error; err != nil {
		return nil, fmt.Errorf("conflict not found: %w", err)
	}

	// Update the conflict with the specified strategy
	conflict.ResolutionStrategy = strategy
	cr.db.Save(&conflict)

	// Resolve the conflict
	return cr.ResolveConflict(ctx, &conflict)
}

// GetConflictStatistics returns statistics about conflicts
func (cr *ConflictResolver) GetConflictStatistics(ctx context.Context, userID uint) (ConflictStatistics, error) {
	var stats ConflictStatistics

	// Count total conflicts
	cr.db.Model(&models.SyncConflict{}).Where("user_id = ?", userID).Count(&stats.TotalConflicts)

	// Count by status
	cr.db.Model(&models.SyncConflict{}).Where("user_id = ? AND status = ?", userID, models.ConflictStatusPending).Count(&stats.PendingConflicts)
	cr.db.Model(&models.SyncConflict{}).Where("user_id = ? AND status = ?", userID, models.ConflictStatusResolved).Count(&stats.ResolvedConflicts)
	cr.db.Model(&models.SyncConflict{}).Where("user_id = ? AND status = ?", userID, models.ConflictStatusIgnored).Count(&stats.IgnoredConflicts)

	// Count by type
	cr.db.Model(&models.SyncConflict{}).Where("user_id = ? AND conflict_type = ?", userID, models.ConflictTypeVersionMismatch).Count(&stats.VersionMismatchConflicts)
	cr.db.Model(&models.SyncConflict{}).Where("user_id = ? AND conflict_type = ?", userID, models.ConflictTypeConcurrentEdit).Count(&stats.ConcurrentEditConflicts)
	cr.db.Model(&models.SyncConflict{}).Where("user_id = ? AND conflict_type = ?", userID, models.ConflictTypeDeletedModified).Count(&stats.DeletedModifiedConflicts)

	// Count by resolution strategy
	cr.db.Model(&models.SyncConflict{}).Where("user_id = ? AND resolution_strategy = ?", userID, models.ResolutionStrategyServerWins).Count(&stats.ServerWinsResolutions)
	cr.db.Model(&models.SyncConflict{}).Where("user_id = ? AND resolution_strategy = ?", userID, models.ResolutionStrategyClientWins).Count(&stats.ClientWinsResolutions)
	cr.db.Model(&models.SyncConflict{}).Where("user_id = ? AND resolution_strategy = ?", userID, models.ResolutionStrategyMerge).Count(&stats.MergeResolutions)
	cr.db.Model(&models.SyncConflict{}).Where("user_id = ? AND resolution_strategy = ?", userID, models.ResolutionStrategyManual).Count(&stats.ManualResolutions)

	return stats, nil
}

// ConflictStatistics represents conflict statistics
type ConflictStatistics struct {
	TotalConflicts           int64 `json:"total_conflicts"`
	PendingConflicts         int64 `json:"pending_conflicts"`
	ResolvedConflicts        int64 `json:"resolved_conflicts"`
	IgnoredConflicts         int64 `json:"ignored_conflicts"`
	VersionMismatchConflicts int64 `json:"version_mismatch_conflicts"`
	ConcurrentEditConflicts  int64 `json:"concurrent_edit_conflicts"`
	DeletedModifiedConflicts int64 `json:"deleted_modified_conflicts"`
	ServerWinsResolutions    int64 `json:"server_wins_resolutions"`
	ClientWinsResolutions    int64 `json:"client_wins_resolutions"`
	MergeResolutions         int64 `json:"merge_resolutions"`
	ManualResolutions        int64 `json:"manual_resolutions"`
}
