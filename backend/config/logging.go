package config

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

type LogContext struct {
	CorrelationID string
	UserID        uint
	RequestID     string
	TraceID       string
	SpanID        string
}

func SetupLogger() *Logger {
	config := zap.NewProductionConfig()

	if os.Getenv("GIN_MODE") == "debug" {
		config = zap.NewDevelopmentConfig()
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	// Add custom fields
	config.InitialFields = map[string]interface{}{
		"service": "mobile-backend",
		"version": "1.0.0",
	}

	logger, _ := config.Build()
	return &Logger{Logger: logger}
}

// Context-aware logging methods
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	logCtx := GetLogContext(ctx)

	fields := []zap.Field{
		zap.String("correlation_id", logCtx.CorrelationID),
		zap.String("request_id", logCtx.RequestID),
	}

	if logCtx.UserID > 0 {
		fields = append(fields, zap.Uint("user_id", logCtx.UserID))
	}

	if logCtx.TraceID != "" {
		fields = append(fields, zap.String("trace_id", logCtx.TraceID))
	}

	if logCtx.SpanID != "" {
		fields = append(fields, zap.String("span_id", logCtx.SpanID))
	}

	return l.Logger.With(fields...)
}

// Create log context
func NewLogContext() *LogContext {
	return &LogContext{
		CorrelationID: uuid.New().String(),
		RequestID:     uuid.New().String(),
	}
}

// Get log context from context
func GetLogContext(ctx context.Context) *LogContext {
	if logCtx, ok := ctx.Value("log_context").(*LogContext); ok {
		return logCtx
	}
	return NewLogContext()
}

// Set log context in context
func WithLogContext(ctx context.Context, logCtx *LogContext) context.Context {
	return context.WithValue(ctx, "log_context", logCtx)
}

// Enhanced logging methods
func (l *Logger) InfoWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Info(msg, fields...)
}

func (l *Logger) ErrorWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Error(msg, fields...)
}

func (l *Logger) WarnWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Warn(msg, fields...)
}

func (l *Logger) DebugWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	l.WithContext(ctx).Debug(msg, fields...)
}

// Performance logging
func (l *Logger) LogPerformance(ctx context.Context, operation string, duration time.Duration, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("operation", operation),
		zap.Duration("duration", duration),
		zap.String("performance_category", "timing"),
	)
	l.InfoWithContext(ctx, "Performance metric", allFields...)
}

// Security logging
func (l *Logger) LogSecurityEvent(ctx context.Context, event string, severity string, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("security_event", event),
		zap.String("severity", severity),
		zap.String("log_category", "security"),
	)
	l.WarnWithContext(ctx, "Security event", allFields...)
}

// Business logic logging
func (l *Logger) LogBusinessEvent(ctx context.Context, event string, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("business_event", event),
		zap.String("log_category", "business"),
	)
	l.InfoWithContext(ctx, "Business event", allFields...)
}

// Error logging with stack trace
func (l *Logger) LogError(ctx context.Context, err error, msg string, fields ...zap.Field) {
	allFields := append(fields,
		zap.Error(err),
		zap.String("log_category", "error"),
	)
	l.ErrorWithContext(ctx, msg, allFields...)
}

// Audit logging
func (l *Logger) LogAudit(ctx context.Context, action string, resource string, fields ...zap.Field) {
	allFields := append(fields,
		zap.String("audit_action", action),
		zap.String("audit_resource", resource),
		zap.String("log_category", "audit"),
	)
	l.InfoWithContext(ctx, "Audit log", allFields...)
}
