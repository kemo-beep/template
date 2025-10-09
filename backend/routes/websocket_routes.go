package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"mobile-backend/controllers"
	"mobile-backend/middleware"
)

// SetupWebSocketRoutes sets up WebSocket-related routes
func SetupWebSocketRoutes(r *gin.Engine, websocketController *controllers.WebSocketController, logger *zap.Logger) {
	// WebSocket connection endpoint
	ws := r.Group("/ws")
	{
		// Apply WebSocket middleware
		ws.Use(middleware.WebSocketCORSMiddleware())
		ws.Use(middleware.WebSocketRateLimitMiddleware())
		ws.Use(middleware.WebSocketLoggingMiddleware(logger))
		ws.Use(middleware.WebSocketAuthMiddleware(logger))

		// WebSocket connection
		ws.GET("/connect", websocketController.Connect)
	}

	// WebSocket API endpoints (REST API for WebSocket management)
	api := r.Group("/api/v1/websocket")
	{
		// Apply authentication middleware
		api.Use(middleware.AuthMiddleware())

		// Notification endpoints
		notifications := api.Group("/notifications")
		{
			notifications.GET("", websocketController.GetNotifications)
			notifications.PUT("/:id/read", websocketController.MarkNotificationAsRead)
		}

		// Connection management endpoints
		connections := api.Group("/connections")
		{
			connections.GET("/stats", websocketController.GetConnectionStats)
			connections.GET("/user", websocketController.GetUserConnections)
		}

		// Notification sending endpoints
		send := api.Group("/send")
		{
			send.POST("/notification", websocketController.SendNotification)
			send.POST("/system", websocketController.SendSystemNotification)
			send.POST("/room", websocketController.SendRoomNotification)
		}

		// Live update endpoints
		live := api.Group("/live")
		{
			live.POST("/update", websocketController.SendLiveUpdate)
			live.POST("/data", websocketController.SendDataUpdate)
			live.POST("/broadcast", websocketController.BroadcastDataUpdate)
		}

		// Real-time features
		realtime := api.Group("/realtime")
		{
			realtime.POST("/typing", websocketController.SendTypingIndicator)
			realtime.POST("/presence", websocketController.SendPresenceUpdate)
		}
	}
}
