package services

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CronScheduler handles scheduled tasks
type CronScheduler struct {
	cron     *cron.Cron
	jobQueue *JobQueueService
	db       *gorm.DB
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

// ScheduledJob represents a scheduled job configuration
type ScheduledJob struct {
	Name        string
	Schedule    string
	Description string
	Enabled     bool
	Handler     func() error
}

// NewCronScheduler creates a new cron scheduler
func NewCronScheduler(jobQueue *JobQueueService, db *gorm.DB, logger *zap.Logger) *CronScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	// Create cron with timezone support
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithLogger(cron.VerbosePrintfLogger(nil)),
		cron.WithChain(cron.Recover(cron.DefaultLogger)),
	)

	return &CronScheduler{
		cron:     c,
		jobQueue: jobQueue,
		db:       db,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start starts the cron scheduler
func (cs *CronScheduler) Start() error {
	cs.logger.Info("Starting cron scheduler...")

	// Register all scheduled jobs
	cs.registerScheduledJobs()

	// Start the cron scheduler
	cs.cron.Start()

	// Wait for context cancellation
	<-cs.ctx.Done()

	return nil
}

// Stop stops the cron scheduler
func (cs *CronScheduler) Stop() {
	cs.logger.Info("Stopping cron scheduler...")
	cs.cron.Stop()
	cs.cancel()
}

// registerScheduledJobs registers all scheduled jobs
func (cs *CronScheduler) registerScheduledJobs() {
	// Daily cleanup jobs
	cs.addJob("daily-cleanup", "0 2 * * *", "Daily data cleanup", true, cs.dailyCleanup)
	cs.addJob("session-cleanup", "0 3 * * *", "Clean expired sessions", true, cs.sessionCleanup)
	cs.addJob("cache-cleanup", "0 4 * * *", "Clean expired cache entries", true, cs.cacheCleanup)

	// Weekly jobs
	cs.addJob("weekly-backup", "0 1 * * 0", "Weekly database backup", true, cs.weeklyBackup)
	cs.addJob("weekly-reports", "0 9 * * 1", "Generate weekly reports", true, cs.weeklyReports)

	// Monthly jobs
	cs.addJob("monthly-cleanup", "0 0 1 * *", "Monthly deep cleanup", true, cs.monthlyCleanup)
	cs.addJob("monthly-reports", "0 10 1 * *", "Generate monthly reports", true, cs.monthlyReports)

	// Hourly jobs
	cs.addJob("hourly-metrics", "0 * * * *", "Collect hourly metrics", true, cs.hourlyMetrics)
	cs.addJob("payment-reconciliation", "30 * * * *", "Hourly payment reconciliation", true, cs.paymentReconciliation)

	// Every 15 minutes
	cs.addJob("cache-warmup", "*/15 * * * *", "Cache warmup", true, cs.cacheWarmup)
	cs.addJob("health-check", "*/15 * * * *", "System health check", true, cs.healthCheck)

	// Every 5 minutes
	cs.addJob("subscription-reminders", "*/5 * * * *", "Check subscription reminders", true, cs.subscriptionReminders)
	cs.addJob("webhook-retry", "*/5 * * * *", "Retry failed webhooks", true, cs.webhookRetry)

	// Every minute
	cs.addJob("user-activity", "* * * * *", "Process user activity", true, cs.userActivity)
}

// addJob adds a scheduled job
func (cs *CronScheduler) addJob(name, schedule, description string, enabled bool, handler func() error) {
	if !enabled {
		cs.logger.Info("Skipping disabled job", zap.String("job", name))
		return
	}

	_ = &ScheduledJob{
		Name:        name,
		Schedule:    schedule,
		Description: description,
		Enabled:     enabled,
		Handler:     handler,
	}

	_, err := cs.cron.AddFunc(schedule, func() {
		cs.logger.Info("Starting scheduled job", zap.String("job", name))
		start := time.Now()

		if err := handler(); err != nil {
			cs.logger.Error("Scheduled job failed",
				zap.String("job", name),
				zap.Error(err),
				zap.Duration("duration", time.Since(start)))
		} else {
			cs.logger.Info("Scheduled job completed",
				zap.String("job", name),
				zap.Duration("duration", time.Since(start)))
		}
	})

	if err != nil {
		cs.logger.Error("Failed to add scheduled job",
			zap.String("job", name),
			zap.String("schedule", schedule),
			zap.Error(err))
	} else {
		cs.logger.Info("Added scheduled job",
			zap.String("job", name),
			zap.String("schedule", schedule),
			zap.String("description", description))
	}
}

// Scheduled Job Handlers

func (cs *CronScheduler) dailyCleanup() error {
	cs.logger.Info("Running daily cleanup...")

	// Clean up old webhook events (older than 30 days)
	payload := DataCleanupPayload{
		TableName:  "webhook_events",
		OlderThan:  time.Now().AddDate(0, 0, -30),
		BatchSize:  1000,
		SoftDelete: false,
	}

	_, err := cs.jobQueue.EnqueueDataCleanup(payload, asynq.Queue("low"))
	if err != nil {
		return fmt.Errorf("failed to enqueue webhook cleanup: %w", err)
	}

	// Clean up old sessions (older than 7 days)
	payload = DataCleanupPayload{
		TableName:  "sessions",
		OlderThan:  time.Now().AddDate(0, 0, -7),
		BatchSize:  1000,
		SoftDelete: false,
	}

	_, err = cs.jobQueue.EnqueueDataCleanup(payload, asynq.Queue("low"))
	if err != nil {
		return fmt.Errorf("failed to enqueue session cleanup: %w", err)
	}

	return nil
}

func (cs *CronScheduler) sessionCleanup() error {
	cs.logger.Info("Running session cleanup...")

	// This would be implemented to clean expired sessions
	// For now, we'll just log it
	cs.logger.Info("Session cleanup completed")
	return nil
}

func (cs *CronScheduler) cacheCleanup() error {
	cs.logger.Info("Running cache cleanup...")

	// This would be implemented to clean expired cache entries
	// For now, we'll just log it
	cs.logger.Info("Cache cleanup completed")
	return nil
}

func (cs *CronScheduler) weeklyBackup() error {
	cs.logger.Info("Running weekly backup...")

	payload := BackupTaskPayload{
		BackupType:  "database",
		Retention:   30, // Keep for 30 days
		Compression: true,
		Encryption:  true,
		Metadata: map[string]interface{}{
			"type": "weekly",
			"auto": true,
		},
	}

	_, err := cs.jobQueue.EnqueueBackupTask(payload, asynq.Queue("low"))
	if err != nil {
		return fmt.Errorf("failed to enqueue weekly backup: %w", err)
	}

	return nil
}

func (cs *CronScheduler) weeklyReports() error {
	cs.logger.Info("Running weekly reports...")

	// Generate system reports
	payload := ReportGenerationPayload{
		ReportType: "system_weekly",
		UserID:     0, // System report
		Parameters: map[string]interface{}{
			"start_date": time.Now().AddDate(0, 0, -7),
			"end_date":   time.Now(),
		},
		Format: "pdf",
	}

	_, err := cs.jobQueue.EnqueueReportGeneration(payload, asynq.Queue("low"))
	if err != nil {
		return fmt.Errorf("failed to enqueue weekly reports: %w", err)
	}

	return nil
}

func (cs *CronScheduler) monthlyCleanup() error {
	cs.logger.Info("Running monthly cleanup...")

	// Clean up old logs, temporary files, etc.
	cs.logger.Info("Monthly cleanup completed")
	return nil
}

func (cs *CronScheduler) monthlyReports() error {
	cs.logger.Info("Running monthly reports...")

	// Generate monthly reports
	payload := ReportGenerationPayload{
		ReportType: "system_monthly",
		UserID:     0, // System report
		Parameters: map[string]interface{}{
			"start_date": time.Now().AddDate(0, -1, 0),
			"end_date":   time.Now(),
		},
		Format: "pdf",
	}

	_, err := cs.jobQueue.EnqueueReportGeneration(payload, asynq.Queue("low"))
	if err != nil {
		return fmt.Errorf("failed to enqueue monthly reports: %w", err)
	}

	return nil
}

func (cs *CronScheduler) hourlyMetrics() error {
	cs.logger.Info("Collecting hourly metrics...")

	// This would collect and store system metrics
	cs.logger.Info("Hourly metrics collection completed")
	return nil
}

func (cs *CronScheduler) paymentReconciliation() error {
	cs.logger.Info("Running payment reconciliation...")

	payload := PaymentReconciliationPayload{
		Provider:  "stripe",
		StartDate: time.Now().Add(-1 * time.Hour),
		EndDate:   time.Now(),
		BatchSize: 100,
	}

	_, err := cs.jobQueue.EnqueuePaymentReconciliation(payload, asynq.Queue("default"))
	if err != nil {
		return fmt.Errorf("failed to enqueue payment reconciliation: %w", err)
	}

	return nil
}

func (cs *CronScheduler) cacheWarmup() error {
	cs.logger.Info("Running cache warmup...")

	payload := CacheWarmupPayload{
		CacheKeys:  []string{"user:active", "product:featured", "stats:daily"},
		Pattern:    "cache:*",
		Expiration: 3600, // 1 hour
	}

	_, err := cs.jobQueue.EnqueueCacheWarmup(payload, asynq.Queue("low"))
	if err != nil {
		return fmt.Errorf("failed to enqueue cache warmup: %w", err)
	}

	return nil
}

func (cs *CronScheduler) healthCheck() error {
	cs.logger.Info("Running health check...")

	// Check database connection
	if err := cs.db.Exec("SELECT 1").Error; err != nil {
		cs.logger.Error("Database health check failed", zap.Error(err))
		return err
	}

	// Check Redis connection (through job queue)
	// This would be implemented to check Redis connectivity

	cs.logger.Info("Health check completed")
	return nil
}

func (cs *CronScheduler) subscriptionReminders() error {
	cs.logger.Info("Checking subscription reminders...")

	// This would check for subscriptions that need reminders
	// For now, we'll just log it
	cs.logger.Info("Subscription reminders check completed")
	return nil
}

func (cs *CronScheduler) webhookRetry() error {
	cs.logger.Info("Retrying failed webhooks...")

	// This would retry failed webhook deliveries
	// For now, we'll just log it
	cs.logger.Info("Webhook retry completed")
	return nil
}

func (cs *CronScheduler) userActivity() error {
	cs.logger.Info("Processing user activity...")

	// This would process user activity data
	// For now, we'll just log it
	cs.logger.Info("User activity processing completed")
	return nil
}

// GetScheduledJobs returns information about all scheduled jobs
func (cs *CronScheduler) GetScheduledJobs() []cron.Entry {
	return cs.cron.Entries()
}

// AddCustomJob adds a custom scheduled job
func (cs *CronScheduler) AddCustomJob(name, schedule, description string, handler func() error) error {
	cs.addJob(name, schedule, description, true, handler)
	return nil
}

// RemoveJob removes a scheduled job by name
func (cs *CronScheduler) RemoveJob(name string) error {
	entries := cs.cron.Entries()
	for range entries {
		// Note: This is a simplified approach. In practice, you'd need to track job IDs
		// to properly remove specific jobs
		cs.logger.Info("Removing job", zap.String("job", name))
		// entry.Remove() // This would remove the job
	}
	return nil
}
