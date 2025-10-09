package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"mobile-backend/models"
	"mobile-backend/services"
	"mobile-backend/utils"
)

// OfflineSyncController handles offline sync API endpoints
type OfflineSyncController struct {
	offlineSyncService *services.OfflineSyncService
	logger             *zap.Logger
}

// NewOfflineSyncController creates a new offline sync controller
func NewOfflineSyncController(offlineSyncService *services.OfflineSyncService, logger *zap.Logger) *OfflineSyncController {
	return &OfflineSyncController{
		offlineSyncService: offlineSyncService,
		logger:             logger,
	}
}

// QueueOperationRequest represents a request to queue an operation
type QueueOperationRequest struct {
	OperationType string                 `json:"operation_type" binding:"required,oneof=create update delete"`
	TableName     string                 `json:"table_name" binding:"required"`
	RecordID      string                 `json:"record_id"`
	Data          map[string]interface{} `json:"data"`
}

// QueueOperationResponse represents a response to queue operation
type QueueOperationResponse struct {
	OperationID string `json:"operation_id"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

// SyncStatusResponse represents sync status response
type SyncStatusResponse struct {
	UserID                 uint   `json:"user_id"`
	LastSyncAt             string `json:"last_sync_at,omitempty"`
	SyncToken              string `json:"sync_token,omitempty"`
	PendingOperationsCount int    `json:"pending_operations_count"`
	ConflictsCount         int    `json:"conflicts_count"`
	IsOnline               bool   `json:"is_online"`
	LastOnlineAt           string `json:"last_online_at"`
}

// SyncResponse represents sync response
type SyncResponse struct {
	Success             bool   `json:"success"`
	OperationsProcessed int    `json:"operations_processed"`
	ConflictsResolved   int    `json:"conflicts_resolved"`
	DurationMs          int    `json:"duration_ms"`
	SyncToken           string `json:"sync_token"`
	Message             string `json:"message"`
}

// ConflictResponse represents conflict response
type ConflictResponse struct {
	ID                 uint                   `json:"id"`
	TableName          string                 `json:"table_name"`
	RecordID           string                 `json:"record_id"`
	ConflictType       string                 `json:"conflict_type"`
	ResolutionStrategy string                 `json:"resolution_strategy"`
	Status             string                 `json:"status"`
	LocalData          map[string]interface{} `json:"local_data"`
	ServerData         map[string]interface{} `json:"server_data"`
	ResolvedData       map[string]interface{} `json:"resolved_data,omitempty"`
	CreatedAt          string                 `json:"created_at"`
	ResolvedAt         string                 `json:"resolved_at,omitempty"`
}

// QueueOperation queues an offline operation
func (osc *OfflineSyncController) QueueOperation(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	var req QueueOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error(), nil)
		return
	}

	// Create offline operation
	operation := &models.OfflineOperation{
		OperationType: req.OperationType,
		TableName:     req.TableName,
		RecordID:      req.RecordID,
		Data:          req.Data,
	}

	// Queue the operation
	if err := osc.offlineSyncService.QueueOperation(c.Request.Context(), userIDUint, operation); err != nil {
		osc.logger.Error("Failed to queue operation", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to queue operation", nil)
		return
	}

	response := QueueOperationResponse{
		OperationID: operation.OperationID,
		Status:      operation.Status,
		Message:     "Operation queued successfully",
	}

	utils.SendCreatedResponse(c, response, "Operation queued successfully")
}

// GetSyncStatus returns the sync status for the user
func (osc *OfflineSyncController) GetSyncStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	syncStatus, err := osc.offlineSyncService.GetSyncStatus(c.Request.Context(), userIDUint)
	if err != nil {
		osc.logger.Error("Failed to get sync status", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get sync status", nil)
		return
	}

	response := SyncStatusResponse{
		UserID:                 syncStatus.UserID,
		PendingOperationsCount: syncStatus.PendingOperationsCount,
		ConflictsCount:         syncStatus.ConflictsCount,
		IsOnline:               syncStatus.IsOnline,
		LastOnlineAt:           syncStatus.LastOnlineAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if syncStatus.LastSyncAt != nil {
		response.LastSyncAt = syncStatus.LastSyncAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if syncStatus.SyncToken != "" {
		response.SyncToken = syncStatus.SyncToken
	}

	utils.SendSuccessResponse(c, response, "Sync status retrieved successfully")
}

// SyncUserData syncs all pending operations for the user
func (osc *OfflineSyncController) SyncUserData(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	// Check if selective sync is requested
	lastSyncTimeStr := c.Query("last_sync")
	if lastSyncTimeStr != "" {
		// Parse last sync time
		lastSyncTime, err := time.Parse(time.RFC3339, lastSyncTimeStr)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid last_sync format. Use RFC3339 format", nil)
			return
		}

		// Perform selective sync
		if err := osc.offlineSyncService.SyncUserDataSelective(c.Request.Context(), userIDUint, lastSyncTime); err != nil {
			osc.logger.Error("Failed to perform selective sync", zap.Error(err))
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to perform selective sync", nil)
			return
		}
	} else {
		// Perform full sync
		if err := osc.offlineSyncService.SyncUserData(c.Request.Context(), userIDUint); err != nil {
			osc.logger.Error("Failed to sync user data", zap.Error(err))
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to sync user data", nil)
			return
		}
	}

	// Get updated sync status
	syncStatus, err := osc.offlineSyncService.GetSyncStatus(c.Request.Context(), userIDUint)
	if err != nil {
		osc.logger.Error("Failed to get updated sync status", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get updated sync status", nil)
		return
	}

	response := SyncResponse{
		Success:             true,
		OperationsProcessed: 0, // This would be set from the sync operation
		ConflictsResolved:   0, // This would be set from the sync operation
		DurationMs:          0, // This would be set from the sync operation
		Message:             "Sync completed successfully",
	}

	if syncStatus.SyncToken != "" {
		response.SyncToken = syncStatus.SyncToken
	}

	utils.SendSuccessResponse(c, response, "Sync status retrieved successfully")
}

// ForceSync forces a sync for the user
func (osc *OfflineSyncController) ForceSync(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	// Force sync
	if err := osc.offlineSyncService.ForceSync(c.Request.Context(), userIDUint); err != nil {
		osc.logger.Error("Failed to force sync", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to force sync", nil)
		return
	}

	response := SyncResponse{
		Success: true,
		Message: "Force sync completed successfully",
	}

	utils.SendSuccessResponse(c, response, "Sync status retrieved successfully")
}

// GetPendingOperations returns pending operations for the user
func (osc *OfflineSyncController) GetPendingOperations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	operations, err := osc.offlineSyncService.GetPendingOperations(c.Request.Context(), userIDUint, limit)
	if err != nil {
		osc.logger.Error("Failed to get pending operations", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get pending operations", nil)
		return
	}

	utils.SendSuccessResponse(c, operations, "Pending operations retrieved successfully")
}

// GetConflicts returns pending conflicts for the user
func (osc *OfflineSyncController) GetConflicts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	conflicts, err := osc.offlineSyncService.GetConflicts(c.Request.Context(), userIDUint, limit)
	if err != nil {
		osc.logger.Error("Failed to get conflicts", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get conflicts", nil)
		return
	}

	// Convert to response format
	var response []ConflictResponse
	for _, conflict := range conflicts {
		conflictResp := ConflictResponse{
			ID:                 conflict.ID,
			TableName:          conflict.TableName,
			RecordID:           conflict.RecordID,
			ConflictType:       conflict.ConflictType,
			ResolutionStrategy: conflict.ResolutionStrategy,
			Status:             conflict.Status,
			LocalData:          conflict.LocalData,
			ServerData:         conflict.ServerData,
			ResolvedData:       conflict.ResolvedData,
			CreatedAt:          conflict.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if conflict.ResolvedAt != nil {
			conflictResp.ResolvedAt = conflict.ResolvedAt.Format("2006-01-02T15:04:05Z07:00")
		}

		response = append(response, conflictResp)
	}

	utils.SendSuccessResponse(c, response, "Sync status retrieved successfully")
}

// ResolveConflict resolves a specific conflict
func (osc *OfflineSyncController) ResolveConflict(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	_, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	// Get conflict ID from URL parameter
	conflictIDStr := c.Param("id")
	conflictID, err := strconv.ParseUint(conflictIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid conflict ID", nil)
		return
	}

	// Get resolution strategy from request body
	var req struct {
		Strategy string `json:"strategy" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error(), nil)
		return
	}

	// Resolve conflict
	// Note: This would typically use the conflict resolver service
	// For now, we'll just return a success response
	response := map[string]interface{}{
		"conflict_id": conflictID,
		"strategy":    req.Strategy,
		"status":      "resolved",
		"message":     "Conflict resolved successfully",
	}

	utils.SendSuccessResponse(c, response, "Sync status retrieved successfully")
}

// SetUserOnline sets the user as online
func (osc *OfflineSyncController) SetUserOnline(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	if err := osc.offlineSyncService.SetUserOnline(c.Request.Context(), userIDUint); err != nil {
		osc.logger.Error("Failed to set user online", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to set user online", nil)
		return
	}

	response := map[string]interface{}{
		"status":  "online",
		"message": "User set as online",
	}

	utils.SendSuccessResponse(c, response, "Sync status retrieved successfully")
}

// SetUserOffline sets the user as offline
func (osc *OfflineSyncController) SetUserOffline(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	if err := osc.offlineSyncService.SetUserOffline(c.Request.Context(), userIDUint); err != nil {
		osc.logger.Error("Failed to set user offline", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to set user offline", nil)
		return
	}

	response := map[string]interface{}{
		"status":  "offline",
		"message": "User set as offline",
	}

	utils.SendSuccessResponse(c, response, "Sync status retrieved successfully")
}

// GetSyncHistory returns sync history for the user
func (osc *OfflineSyncController) GetSyncHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Invalid user ID", nil)
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	// Get sync history from database
	history, err := osc.offlineSyncService.GetSyncHistory(c.Request.Context(), userIDUint, limit)
	if err != nil {
		osc.logger.Error("Failed to get sync history", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get sync history", nil)
		return
	}

	utils.SendSuccessResponse(c, history, "Sync history retrieved successfully")
}
