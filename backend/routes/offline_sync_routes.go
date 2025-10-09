package routes

import (
	"mobile-backend/controllers"
	"mobile-backend/middleware"

	"github.com/gin-gonic/gin"
)

// SetupOfflineSyncRoutes sets up offline sync routes
func SetupOfflineSyncRoutes(router *gin.Engine, offlineSyncController *controllers.OfflineSyncController) {
	// Create a group for offline sync routes with authentication middleware
	offlineSync := router.Group("/api/v1/sync")
	offlineSync.Use(middleware.AuthMiddleware())

	// Queue operations
	offlineSync.POST("/queue", offlineSyncController.QueueOperation)

	// Sync status and operations
	offlineSync.GET("/status", offlineSyncController.GetSyncStatus)
	offlineSync.POST("/sync", offlineSyncController.SyncUserData)
	offlineSync.POST("/force", offlineSyncController.ForceSync)

	// Pending operations and conflicts
	offlineSync.GET("/operations", offlineSyncController.GetPendingOperations)
	offlineSync.GET("/conflicts", offlineSyncController.GetConflicts)
	offlineSync.POST("/conflicts/:id/resolve", offlineSyncController.ResolveConflict)

	// User online/offline status
	offlineSync.POST("/online", offlineSyncController.SetUserOnline)
	offlineSync.POST("/offline", offlineSyncController.SetUserOffline)

	// Sync history
	offlineSync.GET("/history", offlineSyncController.GetSyncHistory)
}
