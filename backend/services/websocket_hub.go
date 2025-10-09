package services

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketMessage represents a message sent through WebSocket
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    uint                   `json:"user_id,omitempty"`
	Room      string                 `json:"room,omitempty"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	UserID   uint
	Conn     *websocket.Conn
	Send     chan WebSocketMessage
	Hub      *Hub
	Rooms    map[string]bool
	LastPing time.Time
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to all clients
	broadcast chan WebSocketMessage

	// Send message to specific user
	sendToUser chan WebSocketMessage

	// Send message to specific room
	sendToRoom chan WebSocketMessage

	// Mutex for thread-safe operations
	mutex sync.RWMutex

	// Logger
	logger *zap.Logger
}

// NewHub creates a new WebSocket hub
func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan WebSocketMessage),
		sendToUser: make(chan WebSocketMessage),
		sendToRoom: make(chan WebSocketMessage),
		logger:     logger,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	// Start ping ticker for connection health checks
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Start cleanup ticker for stale connections
	cleanupTicker := time.NewTicker(1 * time.Minute)
	defer cleanupTicker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			h.logger.Info("Client registered", zap.String("client_id", client.ID), zap.Uint("user_id", client.UserID))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				h.logger.Info("Client unregistered", zap.String("client_id", client.ID))
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()

		case message := <-h.sendToUser:
			h.mutex.RLock()
			for client := range h.clients {
				if client.UserID == message.UserID {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}
			h.mutex.RUnlock()

		case message := <-h.sendToRoom:
			h.mutex.RLock()
			for client := range h.clients {
				if client.Rooms[message.Room] {
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(h.clients, client)
					}
				}
			}
			h.mutex.RUnlock()

		case <-pingTicker.C:
			h.sendPingToAll()

		case <-cleanupTicker.C:
			h.cleanupStaleConnections()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message WebSocketMessage) {
	select {
	case h.broadcast <- message:
	default:
		h.logger.Warn("Broadcast channel is full, dropping message")
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID uint, message WebSocketMessage) {
	message.UserID = userID
	select {
	case h.sendToUser <- message:
	default:
		h.logger.Warn("SendToUser channel is full, dropping message", zap.Uint("user_id", userID))
	}
}

// SendToRoom sends a message to all clients in a specific room
func (h *Hub) SendToRoom(room string, message WebSocketMessage) {
	message.Room = room
	select {
	case h.sendToRoom <- message:
	default:
		h.logger.Warn("SendToRoom channel is full, dropping message", zap.String("room", room))
	}
}

// JoinRoom adds a client to a room
func (h *Hub) JoinRoom(client *Client, room string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if client.Rooms == nil {
		client.Rooms = make(map[string]bool)
	}
	client.Rooms[room] = true

	h.logger.Info("Client joined room", zap.String("client_id", client.ID), zap.String("room", room))
}

// LeaveRoom removes a client from a room
func (h *Hub) LeaveRoom(client *Client, room string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if client.Rooms != nil {
		delete(client.Rooms, room)
	}

	h.logger.Info("Client left room", zap.String("client_id", client.ID), zap.String("room", room))
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// GetUserClients returns all clients for a specific user
func (h *Hub) GetUserClients(userID uint) []*Client {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	var userClients []*Client
	for client := range h.clients {
		if client.UserID == userID {
			userClients = append(userClients, client)
		}
	}
	return userClients
}

// sendPingToAll sends ping messages to all clients
func (h *Hub) sendPingToAll() {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	now := time.Now()
	for client := range h.clients {
		// Check if client hasn't pinged in the last 2 minutes
		if now.Sub(client.LastPing) > 2*time.Minute {
			// Send ping
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.logger.Warn("Failed to send ping", zap.String("client_id", client.ID), zap.Error(err))
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

// cleanupStaleConnections removes stale connections
func (h *Hub) cleanupStaleConnections() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	now := time.Now()
	for client := range h.clients {
		// Remove clients that haven't pinged in the last 5 minutes
		if now.Sub(client.LastPing) > 5*time.Minute {
			h.logger.Info("Removing stale connection", zap.String("client_id", client.ID))
			close(client.Send)
			delete(h.clients, client)
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	// Set read deadline
	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.LastPing = time.Now()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message WebSocketMessage
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error("WebSocket error", zap.String("client_id", c.ID), zap.Error(err))
			}
			break
		}

		// Update last ping time
		c.LastPing = time.Now()

		// Handle different message types
		switch message.Type {
		case "ping":
			// Respond with pong
			c.Hub.SendToUser(c.UserID, WebSocketMessage{
				Type:      "pong",
				Timestamp: time.Now(),
			})
		case "join_room":
			if room, ok := message.Data["room"].(string); ok {
				c.Hub.JoinRoom(c, room)
			}
		case "leave_room":
			if room, ok := message.Data["room"].(string); ok {
				c.Hub.LeaveRoom(c, room)
			}
		case "subscribe":
			// Handle subscription requests
			c.Hub.logger.Info("Client subscription request", zap.String("client_id", c.ID), zap.Any("data", message.Data))
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				c.Hub.logger.Error("Failed to write message", zap.String("client_id", c.ID), zap.Error(err))
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWS handles websocket requests from clients
func (h *Hub) ServeWS(conn *websocket.Conn, userID uint) {
	client := &Client{
		ID:       generateClientID(),
		UserID:   userID,
		Conn:     conn,
		Send:     make(chan WebSocketMessage, 256),
		Hub:      h,
		Rooms:    make(map[string]bool),
		LastPing: time.Now(),
	}

	client.Hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// generateClientID generates a unique client ID
func generateClientID() string {
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

// SendNotification sends a notification to a user
func (h *Hub) SendNotification(userID uint, notificationType string, data map[string]interface{}) {
	message := WebSocketMessage{
		Type: "notification",
		Data: map[string]interface{}{
			"notification_type": notificationType,
			"payload":           data,
		},
		Timestamp: time.Now(),
	}
	h.SendToUser(userID, message)
}

// SendPaymentUpdate sends a payment update to a user
func (h *Hub) SendPaymentUpdate(userID uint, paymentData map[string]interface{}) {
	message := WebSocketMessage{
		Type:      "payment_update",
		Data:      paymentData,
		Timestamp: time.Now(),
	}
	h.SendToUser(userID, message)
}

// SendSystemMessage sends a system message to all users
func (h *Hub) SendSystemMessage(messageType string, data map[string]interface{}) {
	message := WebSocketMessage{
		Type: "system_message",
		Data: map[string]interface{}{
			"message_type": messageType,
			"payload":      data,
		},
		Timestamp: time.Now(),
	}
	h.Broadcast(message)
}
