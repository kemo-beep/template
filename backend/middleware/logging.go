package middleware

import (
	"time"

	"mobile-backend/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LoggerMiddleware(logger *config.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create log context
		logCtx := config.NewLogContext()

		// Add trace information if available
		if traceID := c.GetHeader("X-Trace-ID"); traceID != "" {
			logCtx.TraceID = traceID
		}
		if spanID := c.GetHeader("X-Span-ID"); spanID != "" {
			logCtx.SpanID = spanID
		}

		// Add to context
		ctx := config.WithLogContext(c.Request.Context(), logCtx)
		c.Request = c.Request.WithContext(ctx)

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get user ID if available
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uint); ok {
				logCtx.UserID = uid
			}
		}

		// Log the request
		logger.InfoWithContext(ctx, "HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("body_size", c.Writer.Size()),
			zap.String("referer", c.Request.Referer()),
		)

		// Log performance metrics
		if latency > 1*time.Second {
			logger.LogPerformance(ctx, "slow_request", latency,
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
			)
		}

		// Log errors
		if c.Writer.Status() >= 400 {
			logger.ErrorWithContext(ctx, "HTTP Error",
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
			)
		}
	}
}
