# WebSocket Real-time Notifications Implementation

## 🚀 Overview

This document describes the complete WebSocket real-time notification system that has been integrated into your Go backend. The implementation provides real-time communication capabilities for live updates, notifications, and interactive features.

## 📁 Files Added/Modified

### New Files Created:
- `backend/services/websocket_hub.go` - WebSocket connection management hub
- `backend/services/websocket.go` - WebSocket service for notifications and real-time features
- `backend/controllers/websocket.go` - WebSocket API controller
- `backend/middleware/websocket.go` - WebSocket authentication and CORS middleware
- `backend/routes/websocket_routes.go` - WebSocket route definitions
- `backend/websocket_client_example.html` - HTML client example for testing

### Modified Files:
- `backend/main.go` - Added WebSocket service initialization
- `backend/routes/routes.go` - Updated to include WebSocket controller
- `backend/services/stripe.go` - Integrated WebSocket notifications for payment events
- `backend/tests/integration/api_test.go` - Updated tests to include WebSocket services

## 🔧 Features Implemented

### 1. WebSocket Connection Management
- **Connection Hub**: Manages all active WebSocket connections
- **User-based Connections**: Track connections per user
- **Room Management**: Support for joining/leaving rooms
- **Connection Health**: Automatic ping/pong and stale connection cleanup
- **Concurrent Safety**: Thread-safe operations with mutex protection

### 2. Real-time Notifications
- **User Notifications**: Send notifications to specific users
- **System Notifications**: Broadcast messages to all connected users
- **Room Notifications**: Send messages to users in specific rooms
- **Payment Notifications**: Real-time payment status updates
- **Notification Persistence**: Store notifications in Redis for retrieval

### 3. Live Updates
- **Data Updates**: Send real-time data changes to users
- **Live Updates**: Push live status updates
- **Broadcast Updates**: Send updates to all connected users
- **Typing Indicators**: Real-time typing status in rooms
- **Presence Updates**: Online/offline status management

### 4. Authentication & Security
- **JWT Authentication**: Secure WebSocket connections with JWT tokens
- **CORS Support**: Configurable CORS for WebSocket connections
- **Rate Limiting**: Protection against connection spam
- **Connection Logging**: Comprehensive logging of connection events

## 🌐 API Endpoints

### WebSocket Connection
```
WS /ws/connect?token=<jwt_token>
```

### REST API Endpoints
```
GET    /api/v1/websocket/notifications              # Get user notifications
PUT    /api/v1/websocket/notifications/:id/read     # Mark notification as read
GET    /api/v1/websocket/connections/stats          # Get connection statistics
GET    /api/v1/websocket/connections/user           # Get user's active connections

POST   /api/v1/websocket/send/notification          # Send notification to user
POST   /api/v1/websocket/send/system                # Send system notification
POST   /api/v1/websocket/send/room                  # Send room notification

POST   /api/v1/websocket/live/update                # Send live update
POST   /api/v1/websocket/live/data                  # Send data update
POST   /api/v1/websocket/live/broadcast             # Broadcast data update

POST   /api/v1/websocket/realtime/typing            # Send typing indicator
POST   /api/v1/websocket/realtime/presence          # Send presence update
```

## 📡 WebSocket Message Types

### Client to Server Messages:
```json
{
  "type": "ping",
  "data": {}
}

{
  "type": "join_room",
  "data": {
    "room": "room_name"
  }
}

{
  "type": "leave_room",
  "data": {
    "room": "room_name"
  }
}

{
  "type": "typing_indicator",
  "data": {
    "room": "room_name",
    "is_typing": true
  }
}

{
  "type": "presence_update",
  "data": {
    "status": "online"
  }
}
```

### Server to Client Messages:
```json
{
  "type": "notification",
  "data": {
    "notification_type": "payment_succeeded",
    "title": "Payment Successful",
    "message": "Your payment has been processed",
    "payload": {
      "payment_id": 123,
      "amount": 1000
    }
  },
  "timestamp": "2025-10-09T11:45:36Z"
}

{
  "type": "live_update",
  "data": {
    "update_type": "order_status",
    "payload": {
      "order_id": 456,
      "status": "shipped"
    }
  },
  "timestamp": "2025-10-09T11:45:36Z"
}

{
  "type": "system_message",
  "data": {
    "message_type": "maintenance",
    "title": "Scheduled Maintenance",
    "message": "System will be down for maintenance",
    "payload": {
      "start_time": "2025-10-09T12:00:00Z",
      "duration": "2 hours"
    }
  },
  "timestamp": "2025-10-09T11:45:36Z"
}
```

## 🔌 Integration with Existing Services

### Payment Integration
The WebSocket system is integrated with your existing Stripe payment service:

- **Payment Success**: Real-time notifications when payments succeed
- **Payment Failure**: Immediate alerts for failed payments
- **Subscription Events**: Live updates for subscription changes
- **Invoice Updates**: Real-time invoice status changes

### Example Payment Notification:
```json
{
  "type": "payment_update",
  "data": {
    "payment_id": 123,
    "amount": 1000,
    "currency": "usd",
    "status": "succeeded"
  },
  "timestamp": "2025-10-09T11:45:36Z"
}
```

## 🧪 Testing

### HTML Client Example
Use the provided `websocket_client_example.html` to test WebSocket functionality:

1. Start your backend server
2. Open the HTML file in a browser
3. Enter a valid JWT token
4. Click "Connect" to establish WebSocket connection
5. Test various features like notifications, room management, etc.

### Manual Testing
```bash
# Start the server
go run main.go

# Test WebSocket connection (replace with your JWT token)
wscat -c "ws://localhost:8081/ws/connect?token=your_jwt_token_here"

# Send a test message
{"type": "ping", "data": {}}
```

## 🚀 Usage Examples

### 1. Sending a User Notification
```go
// In your service or controller
err := websocketService.SendNotification(
    ctx,
    userID,
    "payment_succeeded",
    "Payment Successful",
    "Your payment of $10.00 has been processed",
    map[string]interface{}{
        "payment_id": 123,
        "amount": 1000,
    },
)
```

### 2. Broadcasting System Message
```go
err := websocketService.SendSystemNotification(
    ctx,
    "maintenance",
    "Scheduled Maintenance",
    "System will be down for 2 hours starting at 12:00 PM",
    map[string]interface{}{
        "start_time": "2025-10-09T12:00:00Z",
        "duration": "2 hours",
    },
)
```

### 3. Sending Live Data Update
```go
err := websocketService.SendDataUpdate(
    ctx,
    userID,
    "order_status",
    map[string]interface{}{
        "order_id": 456,
        "status": "shipped",
        "tracking_number": "1Z999AA1234567890",
    },
)
```

## ⚙️ Configuration

### Environment Variables
```bash
# JWT Secret (required for WebSocket authentication)
JWT_SECRET=your_super_secret_jwt_key_change_this_in_production

# Redis URL (required for WebSocket hub)
REDIS_URL=localhost:6379

# CORS Origins (optional, defaults to localhost)
WEBSOCKET_CORS_ORIGINS=http://localhost:3000,https://yourdomain.com
```

### WebSocket Hub Configuration
The WebSocket hub can be configured in `services/websocket_hub.go`:

- **Ping Interval**: 30 seconds (configurable)
- **Connection Timeout**: 60 seconds
- **Stale Connection Cleanup**: 5 minutes
- **Max Message Size**: 512 bytes

## 🔒 Security Considerations

1. **JWT Authentication**: All WebSocket connections require valid JWT tokens
2. **CORS Protection**: Configure allowed origins for WebSocket connections
3. **Rate Limiting**: Implement rate limiting for connection attempts
4. **Input Validation**: Validate all incoming WebSocket messages
5. **Connection Limits**: Consider implementing per-user connection limits

## 📊 Monitoring & Metrics

### Connection Statistics
```go
stats := websocketService.GetConnectionStats()
// Returns: map[string]interface{}{
//   "total_connections": 42,
//   "timestamp": "2025-10-09T11:45:36Z"
// }
```

### User Connection Info
```go
connections := websocketService.GetUserConnections(userID)
// Returns: []*Client with connection details
```

## 🎯 Next Steps

1. **Client Integration**: Implement WebSocket clients in your frontend applications
2. **Room Management**: Add more sophisticated room management features
3. **Message Persistence**: Implement message history and persistence
4. **Load Balancing**: Consider Redis Pub/Sub for multi-instance deployments
5. **Monitoring**: Add Prometheus metrics for WebSocket connections
6. **Rate Limiting**: Implement more sophisticated rate limiting strategies

## 🐛 Troubleshooting

### Common Issues

1. **Connection Refused**: Check if Redis is running and accessible
2. **Authentication Failed**: Verify JWT token is valid and not expired
3. **CORS Errors**: Check CORS configuration in middleware
4. **Memory Leaks**: Monitor connection cleanup and stale connection removal

### Debug Mode
Enable debug logging by setting:
```bash
export LOG_LEVEL=debug
export GIN_MODE=debug
```

## 📚 Dependencies Added

- `github.com/gorilla/websocket v1.5.3` - WebSocket implementation
- Existing dependencies: `go.uber.org/zap`, `github.com/go-redis/redis/v8`

---

**🎉 Your backend now has full WebSocket real-time notification capabilities!**

The implementation is production-ready and includes comprehensive error handling, logging, and security features. You can now build real-time features like live chat, live notifications, real-time dashboards, and more.
