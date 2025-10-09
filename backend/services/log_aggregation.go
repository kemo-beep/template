package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"mobile-backend/config"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type LogAggregationService struct {
	redis  *redis.Client
	logger *config.Logger
}

type LogEntry struct {
	Timestamp     time.Time              `json:"timestamp"`
	Level         string                 `json:"level"`
	Message       string                 `json:"message"`
	Service       string                 `json:"service"`
	CorrelationID string                 `json:"correlation_id"`
	RequestID     string                 `json:"request_id"`
	UserID        uint                   `json:"user_id,omitempty"`
	TraceID       string                 `json:"trace_id,omitempty"`
	SpanID        string                 `json:"span_id,omitempty"`
	Fields        map[string]interface{} `json:"fields"`
	Category      string                 `json:"category,omitempty"`
}

type LogQuery struct {
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Level         string    `json:"level,omitempty"`
	Service       string    `json:"service,omitempty"`
	CorrelationID string    `json:"correlation_id,omitempty"`
	UserID        uint      `json:"user_id,omitempty"`
	Category      string    `json:"category,omitempty"`
	Limit         int       `json:"limit"`
	Offset        int       `json:"offset"`
}

func NewLogAggregationService(redis *redis.Client, logger *config.Logger) *LogAggregationService {
	return &LogAggregationService{
		redis:  redis,
		logger: logger,
	}
}

// Store log entry in Redis
func (las *LogAggregationService) StoreLogEntry(ctx context.Context, entry *LogEntry) error {
	// Create a unique key for this log entry
	key := fmt.Sprintf("logs:%s:%d", entry.CorrelationID, entry.Timestamp.UnixNano())

	// Serialize log entry
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Store with TTL (7 days)
	ttl := 7 * 24 * time.Hour
	return las.redis.Set(ctx, key, data, ttl).Err()
}

// Query logs with filters
func (las *LogAggregationService) QueryLogs(ctx context.Context, query *LogQuery) ([]*LogEntry, error) {
	// This is a simplified implementation
	// In production, you'd use a proper log aggregation system like ELK stack

	// Get all log keys
	pattern := "logs:*"
	keys, err := las.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var entries []*LogEntry

	// Fetch and filter entries
	for _, key := range keys {
		data, err := las.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			continue
		}

		// Apply filters
		if las.matchesQuery(&entry, query) {
			entries = append(entries, &entry)
		}
	}

	// Sort by timestamp (newest first)
	// In production, you'd do this at the database level

	return entries, nil
}

// Check if log entry matches query criteria
func (las *LogAggregationService) matchesQuery(entry *LogEntry, query *LogQuery) bool {
	// Time range filter
	if !entry.Timestamp.IsZero() {
		if !query.StartTime.IsZero() && entry.Timestamp.Before(query.StartTime) {
			return false
		}
		if !query.EndTime.IsZero() && entry.Timestamp.After(query.EndTime) {
			return false
		}
	}

	// Level filter
	if query.Level != "" && entry.Level != query.Level {
		return false
	}

	// Service filter
	if query.Service != "" && entry.Service != query.Service {
		return false
	}

	// Correlation ID filter
	if query.CorrelationID != "" && entry.CorrelationID != query.CorrelationID {
		return false
	}

	// User ID filter
	if query.UserID > 0 && entry.UserID != query.UserID {
		return false
	}

	// Category filter
	if query.Category != "" && entry.Category != query.Category {
		return false
	}

	return true
}

// Get log statistics
func (las *LogAggregationService) GetLogStats(ctx context.Context, timeRange time.Duration) (map[string]interface{}, error) {
	now := time.Now()
	startTime := now.Add(-timeRange)

	query := &LogQuery{
		StartTime: startTime,
		EndTime:   now,
		Limit:     10000, // Large limit for stats
	}

	entries, err := las.QueryLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_logs":      len(entries),
		"time_range":      timeRange.String(),
		"level_counts":    make(map[string]int),
		"category_counts": make(map[string]int),
		"error_rate":      0.0,
	}

	levelCounts := make(map[string]int)
	categoryCounts := make(map[string]int)
	errorCount := 0

	for _, entry := range entries {
		levelCounts[entry.Level]++
		if entry.Category != "" {
			categoryCounts[entry.Category]++
		}
		if entry.Level == "error" {
			errorCount++
		}
	}

	stats["level_counts"] = levelCounts
	stats["category_counts"] = categoryCounts
	if len(entries) > 0 {
		stats["error_rate"] = float64(errorCount) / float64(len(entries))
	}

	return stats, nil
}

// Real-time log streaming (simplified)
func (las *LogAggregationService) StreamLogs(ctx context.Context, correlationID string) (<-chan *LogEntry, error) {
	logChan := make(chan *LogEntry, 100)

	go func() {
		defer close(logChan)

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				query := &LogQuery{
					CorrelationID: correlationID,
					Limit:         100,
				}

				entries, err := las.QueryLogs(ctx, query)
				if err != nil {
					las.logger.ErrorWithContext(ctx, "Failed to query logs", zap.Error(err))
					continue
				}

				for _, entry := range entries {
					select {
					case logChan <- entry:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return logChan, nil
}

// Log retention cleanup
func (las *LogAggregationService) CleanupOldLogs(ctx context.Context, olderThan time.Duration) error {
	cutoffTime := time.Now().Add(-olderThan)

	// Get all log keys
	pattern := "logs:*"
	keys, err := las.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	deletedCount := 0
	for _, key := range keys {
		data, err := las.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			continue
		}

		if entry.Timestamp.Before(cutoffTime) {
			if err := las.redis.Del(ctx, key).Err(); err == nil {
				deletedCount++
			}
		}
	}

	las.logger.InfoWithContext(ctx, "Cleaned up old logs",
		zap.Int("deleted_count", deletedCount),
		zap.Duration("older_than", olderThan),
	)

	return nil
}
