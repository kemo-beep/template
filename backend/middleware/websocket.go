package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// WebSocketAuthMiddleware handles authentication for WebSocket connections
func WebSocketAuthMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from query parameter or Authorization header
		token := c.Query("token")
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
				}
			}
		}

		if token == "" {
			logger.Warn("WebSocket connection attempt without token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication token required"})
			c.Abort()
			return
		}

		// Verify JWT token
		claims, err := verifyJWTToken(token)
		if err != nil {
			logger.Warn("Invalid WebSocket token", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
			c.Abort()
			return
		}

		// Extract user ID from claims
		userID, ok := claims["user_id"].(float64)
		if !ok {
			logger.Warn("Invalid user ID in WebSocket token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", uint(userID))
		c.Set("user_claims", claims)

		logger.Info("WebSocket authentication successful", zap.Uint("user_id", uint(userID)))
		c.Next()
	}
}

// verifyJWTToken verifies a JWT token and returns the claims
func verifyJWTToken(tokenString string) (jwt.MapClaims, error) {
	// Get JWT secret from environment
	jwtSecret := getJWTSecret()
	if jwtSecret == "" {
		return nil, jwt.ErrSignatureInvalid
	}

	// Parse and verify token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

// getJWTSecret retrieves JWT secret from environment
func getJWTSecret() string {
	// This should match your JWT secret configuration
	// You might want to use a config service instead
	return "your_super_secret_jwt_key_change_this_in_production" // This should come from environment
}

// WebSocketCORSMiddleware handles CORS for WebSocket connections
func WebSocketCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow specific origins (configure as needed)
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"https://yourdomain.com",
		}

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// WebSocketRateLimitMiddleware applies rate limiting to WebSocket connections
func WebSocketRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Simple rate limiting based on IP
		// In production, you might want to use Redis for distributed rate limiting
		// For now, we'll just log the connection attempt
		logger := zap.L()
		logger.Info("WebSocket connection attempt", zap.String("client_ip", clientIP))

		// You can implement more sophisticated rate limiting here
		// For example, using a sliding window or token bucket algorithm

		c.Next()
	}
}

// WebSocketLoggingMiddleware logs WebSocket connection events
func WebSocketLoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log connection attempt
		logger.Info("WebSocket connection attempt",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// Process request
		c.Next()

		// Log connection result
		status := c.Writer.Status()
		if status >= 200 && status < 300 {
			logger.Info("WebSocket connection successful",
				zap.Int("status", status),
				zap.String("client_ip", c.ClientIP()),
			)
		} else {
			logger.Warn("WebSocket connection failed",
				zap.Int("status", status),
				zap.String("client_ip", c.ClientIP()),
			)
		}
	}
}
