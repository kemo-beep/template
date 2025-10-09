package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// JobQueueService handles background job processing
type JobQueueService struct {
	client    *asynq.Client
	server    *asynq.Server
	mux       *asynq.ServeMux
	db        *gorm.DB
	logger    *zap.Logger
	redisAddr string
}

// Job types
const (
	TypeEmailNotification     = "email:notification"
	TypeEmailBulk             = "email:bulk"
	TypeDataCleanup           = "data:cleanup"
	TypeReportGeneration      = "report:generation"
	TypePaymentReconciliation = "payment:reconciliation"
	TypeWebhookRetry          = "webhook:retry"
	TypeCacheWarmup           = "cache:warmup"
	TypeUserActivity          = "user:activity"
	TypeSubscriptionReminder  = "subscription:reminder"
	TypeBackupTask            = "backup:task"
)

// Job payloads
type EmailNotificationPayload struct {
	UserID   uint                   `json:"user_id"`
	Email    string                 `json:"email"`
	Subject  string                 `json:"subject"`
	Body     string                 `json:"body"`
	Template string                 `json:"template,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
	Priority int                    `json:"priority"`
}

type DataCleanupPayload struct {
	TableName  string    `json:"table_name"`
	OlderThan  time.Time `json:"older_than"`
	BatchSize  int       `json:"batch_size"`
	SoftDelete bool      `json:"soft_delete"`
}

type ReportGenerationPayload struct {
	ReportType string                 `json:"report_type"`
	UserID     uint                   `json:"user_id"`
	Parameters map[string]interface{} `json:"parameters"`
	Format     string                 `json:"format"` // pdf, csv, xlsx
}

type PaymentReconciliationPayload struct {
	Provider  string    `json:"provider"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	BatchSize int       `json:"batch_size"`
}

type WebhookRetryPayload struct {
	WebhookID  uint      `json:"webhook_id"`
	MaxRetries int       `json:"max_retries"`
	RetryCount int       `json:"retry_count"`
	NextRetry  time.Time `json:"next_retry"`
}

type CacheWarmupPayload struct {
	CacheKeys  []string `json:"cache_keys"`
	Pattern    string   `json:"pattern,omitempty"`
	Expiration int      `json:"expiration"`
}

type UserActivityPayload struct {
	UserID    uint                   `json:"user_id"`
	Activity  string                 `json:"activity"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

type SubscriptionReminderPayload struct {
	SubscriptionID uint   `json:"subscription_id"`
	ReminderType   string `json:"reminder_type"` // trial_ending, payment_due, expired
	DaysBefore     int    `json:"days_before"`
}

type BackupTaskPayload struct {
	BackupType  string                 `json:"backup_type"` // database, files, full
	Retention   int                    `json:"retention_days"`
	Compression bool                   `json:"compression"`
	Encryption  bool                   `json:"encryption"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewJobQueueService creates a new job queue service
func NewJobQueueService(redisAddr string, db *gorm.DB, logger *zap.Logger) *JobQueueService {
	// Redis client for enqueueing jobs
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})

	// Redis server for processing jobs
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			RetryDelayFunc: asynq.DefaultRetryDelayFunc,
			Logger:         nil,
		},
	)

	// Mux for routing jobs to handlers
	mux := asynq.NewServeMux()

	service := &JobQueueService{
		client:    client,
		server:    server,
		mux:       mux,
		db:        db,
		logger:    logger,
		redisAddr: redisAddr,
	}

	// Register job handlers
	service.registerHandlers()

	return service
}

// registerHandlers registers all job handlers
func (j *JobQueueService) registerHandlers() {
	// Email jobs
	j.mux.HandleFunc(TypeEmailNotification, j.handleEmailNotification)
	j.mux.HandleFunc(TypeEmailBulk, j.handleEmailBulk)

	// Data management jobs
	j.mux.HandleFunc(TypeDataCleanup, j.handleDataCleanup)
	j.mux.HandleFunc(TypeReportGeneration, j.handleReportGeneration)

	// Payment jobs
	j.mux.HandleFunc(TypePaymentReconciliation, j.handlePaymentReconciliation)
	j.mux.HandleFunc(TypeWebhookRetry, j.handleWebhookRetry)

	// System jobs
	j.mux.HandleFunc(TypeCacheWarmup, j.handleCacheWarmup)
	j.mux.HandleFunc(TypeUserActivity, j.handleUserActivity)
	j.mux.HandleFunc(TypeSubscriptionReminder, j.handleSubscriptionReminder)
	j.mux.HandleFunc(TypeBackupTask, j.handleBackupTask)
}

// Start starts the job queue server
func (j *JobQueueService) Start() error {
	j.logger.Info("Starting job queue server...")
	return j.server.Run(j.mux)
}

// Stop stops the job queue server
func (j *JobQueueService) Stop() {
	j.logger.Info("Stopping job queue server...")
	j.server.Shutdown()
	j.client.Close()
}

// EnqueueEmailNotification enqueues an email notification job
func (j *JobQueueService) EnqueueEmailNotification(payload EmailNotificationPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TypeEmailNotification, payloadBytes)
	return j.client.Enqueue(task, opts...)
}

// EnqueueEmailBulk enqueues a bulk email job
func (j *JobQueueService) EnqueueEmailBulk(userIDs []uint, subject, body string, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payload := map[string]interface{}{
		"user_ids": userIDs,
		"subject":  subject,
		"body":     body,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TypeEmailBulk, payloadBytes)
	return j.client.Enqueue(task, opts...)
}

// EnqueueDataCleanup enqueues a data cleanup job
func (j *JobQueueService) EnqueueDataCleanup(payload DataCleanupPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TypeDataCleanup, payloadBytes)
	return j.client.Enqueue(task, opts...)
}

// EnqueueReportGeneration enqueues a report generation job
func (j *JobQueueService) EnqueueReportGeneration(payload ReportGenerationPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TypeReportGeneration, payloadBytes)
	return j.client.Enqueue(task, opts...)
}

// EnqueuePaymentReconciliation enqueues a payment reconciliation job
func (j *JobQueueService) EnqueuePaymentReconciliation(payload PaymentReconciliationPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TypePaymentReconciliation, payloadBytes)
	return j.client.Enqueue(task, opts...)
}

// EnqueueWebhookRetry enqueues a webhook retry job
func (j *JobQueueService) EnqueueWebhookRetry(payload WebhookRetryPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TypeWebhookRetry, payloadBytes)
	return j.client.Enqueue(task, opts...)
}

// EnqueueCacheWarmup enqueues a cache warmup job
func (j *JobQueueService) EnqueueCacheWarmup(payload CacheWarmupPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TypeCacheWarmup, payloadBytes)
	return j.client.Enqueue(task, opts...)
}

// EnqueueUserActivity enqueues a user activity tracking job
func (j *JobQueueService) EnqueueUserActivity(payload UserActivityPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TypeUserActivity, payloadBytes)
	return j.client.Enqueue(task, opts...)
}

// EnqueueSubscriptionReminder enqueues a subscription reminder job
func (j *JobQueueService) EnqueueSubscriptionReminder(payload SubscriptionReminderPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TypeSubscriptionReminder, payloadBytes)
	return j.client.Enqueue(task, opts...)
}

// EnqueueBackupTask enqueues a backup task job
func (j *JobQueueService) EnqueueBackupTask(payload BackupTaskPayload, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	task := asynq.NewTask(TypeBackupTask, payloadBytes)
	return j.client.Enqueue(task, opts...)
}

// Job Handlers

func (j *JobQueueService) handleEmailNotification(ctx context.Context, t *asynq.Task) error {
	var payload EmailNotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal email notification payload: %w", err)
	}

	j.logger.Info("Processing email notification",
		zap.Uint("user_id", payload.UserID),
		zap.String("email", payload.Email),
		zap.String("subject", payload.Subject))

	// TODO: Implement actual email sending logic
	// This would integrate with your email service
	fmt.Printf("Sending email to %s: %s\n", payload.Email, payload.Subject)

	return nil
}

func (j *JobQueueService) handleEmailBulk(ctx context.Context, t *asynq.Task) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal bulk email payload: %w", err)
	}

	userIDs, _ := payload["user_ids"].([]uint)
	subject, _ := payload["subject"].(string)
	body, _ := payload["body"].(string)

	j.logger.Info("Processing bulk email",
		zap.Int("user_count", len(userIDs)),
		zap.String("subject", subject))

	// TODO: Implement bulk email sending logic
	fmt.Printf("Sending bulk email to %d users: %s - %s\n", len(userIDs), subject, body)

	return nil
}

func (j *JobQueueService) handleDataCleanup(ctx context.Context, t *asynq.Task) error {
	var payload DataCleanupPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal data cleanup payload: %w", err)
	}

	j.logger.Info("Processing data cleanup",
		zap.String("table", payload.TableName),
		zap.Time("older_than", payload.OlderThan),
		zap.Int("batch_size", payload.BatchSize))

	// TODO: Implement data cleanup logic
	// This would clean up old records from the specified table
	fmt.Printf("Cleaning up data from %s older than %v\n", payload.TableName, payload.OlderThan)

	return nil
}

func (j *JobQueueService) handleReportGeneration(ctx context.Context, t *asynq.Task) error {
	var payload ReportGenerationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal report generation payload: %w", err)
	}

	j.logger.Info("Processing report generation",
		zap.String("report_type", payload.ReportType),
		zap.Uint("user_id", payload.UserID),
		zap.String("format", payload.Format))

	// TODO: Implement report generation logic
	fmt.Printf("Generating %s report for user %d in %s format\n", payload.ReportType, payload.UserID, payload.Format)

	return nil
}

func (j *JobQueueService) handlePaymentReconciliation(ctx context.Context, t *asynq.Task) error {
	var payload PaymentReconciliationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payment reconciliation payload: %w", err)
	}

	j.logger.Info("Processing payment reconciliation",
		zap.String("provider", payload.Provider),
		zap.Time("start_date", payload.StartDate),
		zap.Time("end_date", payload.EndDate))

	// TODO: Implement payment reconciliation logic
	fmt.Printf("Reconciling payments for %s from %v to %v\n", payload.Provider, payload.StartDate, payload.EndDate)

	return nil
}

func (j *JobQueueService) handleWebhookRetry(ctx context.Context, t *asynq.Task) error {
	var payload WebhookRetryPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal webhook retry payload: %w", err)
	}

	j.logger.Info("Processing webhook retry",
		zap.Uint("webhook_id", payload.WebhookID),
		zap.Int("retry_count", payload.RetryCount),
		zap.Int("max_retries", payload.MaxRetries))

	// TODO: Implement webhook retry logic
	fmt.Printf("Retrying webhook %d (attempt %d/%d)\n", payload.WebhookID, payload.RetryCount+1, payload.MaxRetries)

	return nil
}

func (j *JobQueueService) handleCacheWarmup(ctx context.Context, t *asynq.Task) error {
	var payload CacheWarmupPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal cache warmup payload: %w", err)
	}

	j.logger.Info("Processing cache warmup",
		zap.Strings("cache_keys", payload.CacheKeys),
		zap.String("pattern", payload.Pattern))

	// TODO: Implement cache warmup logic
	fmt.Printf("Warming up cache for %d keys\n", len(payload.CacheKeys))

	return nil
}

func (j *JobQueueService) handleUserActivity(ctx context.Context, t *asynq.Task) error {
	var payload UserActivityPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal user activity payload: %w", err)
	}

	j.logger.Info("Processing user activity",
		zap.Uint("user_id", payload.UserID),
		zap.String("activity", payload.Activity),
		zap.Time("timestamp", payload.Timestamp))

	// TODO: Implement user activity tracking logic
	fmt.Printf("Tracking activity for user %d: %s\n", payload.UserID, payload.Activity)

	return nil
}

func (j *JobQueueService) handleSubscriptionReminder(ctx context.Context, t *asynq.Task) error {
	var payload SubscriptionReminderPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal subscription reminder payload: %w", err)
	}

	j.logger.Info("Processing subscription reminder",
		zap.Uint("subscription_id", payload.SubscriptionID),
		zap.String("reminder_type", payload.ReminderType),
		zap.Int("days_before", payload.DaysBefore))

	// TODO: Implement subscription reminder logic
	fmt.Printf("Sending %s reminder for subscription %d\n", payload.ReminderType, payload.SubscriptionID)

	return nil
}

func (j *JobQueueService) handleBackupTask(ctx context.Context, t *asynq.Task) error {
	var payload BackupTaskPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal backup task payload: %w", err)
	}

	j.logger.Info("Processing backup task",
		zap.String("backup_type", payload.BackupType),
		zap.Int("retention_days", payload.Retention),
		zap.Bool("compression", payload.Compression),
		zap.Bool("encryption", payload.Encryption))

	// TODO: Implement backup task logic
	fmt.Printf("Creating %s backup with %d days retention\n", payload.BackupType, payload.Retention)

	return nil
}

// GetQueueStats returns queue statistics
func (j *JobQueueService) GetQueueStats() (map[string]interface{}, error) {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: j.redisAddr})
	queueInfo, err := inspector.GetQueueInfo("default")
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"size":      queueInfo.Size,
		"processed": queueInfo.Processed,
		"failed":    queueInfo.Failed,
		"pending":   queueInfo.Pending,
		"active":    queueInfo.Active,
		"scheduled": queueInfo.Scheduled,
		"retry":     queueInfo.Retry,
		"archived":  queueInfo.Archived,
	}

	return stats, nil
}

// GetTaskInfo returns information about a specific task
func (j *JobQueueService) GetTaskInfo(queue, taskID string) (*asynq.TaskInfo, error) {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: j.redisAddr})
	return inspector.GetTaskInfo(queue, taskID)
}

// CancelTask cancels a specific task
func (j *JobQueueService) CancelTask(queue, taskID string) error {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: j.redisAddr})
	// Note: CancelTask method doesn't exist in asynq, using DeleteTask instead
	return inspector.DeleteTask(queue, taskID)
}

// DeleteTask deletes a specific task
func (j *JobQueueService) DeleteTask(queue, taskID string) error {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: j.redisAddr})
	return inspector.DeleteTask(queue, taskID)
}
