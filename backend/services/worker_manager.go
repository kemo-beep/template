package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WorkerManager manages background workers
type WorkerManager struct {
	jobQueue      *JobQueueService
	cronScheduler *CronScheduler
	db            *gorm.DB
	logger        *zap.Logger
	workers       []Worker
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.RWMutex
	running       bool
}

// Worker represents a background worker
type Worker struct {
	ID           string
	Name         string
	Type         string
	Status       string
	LastActivity time.Time
	ErrorCount   int
	MaxErrors    int
	Handler      func(ctx context.Context) error
}

// WorkerStatus represents the status of a worker
type WorkerStatus struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Status       string    `json:"status"`
	LastActivity time.Time `json:"last_activity"`
	ErrorCount   int       `json:"error_count"`
	MaxErrors    int       `json:"max_errors"`
	Uptime       string    `json:"uptime"`
}

const (
	WorkerStatusRunning = "running"
	WorkerStatusStopped = "stopped"
	WorkerStatusError   = "error"
	WorkerStatusPaused  = "paused"
)

// NewWorkerManager creates a new worker manager
func NewWorkerManager(jobQueue *JobQueueService, cronScheduler *CronScheduler, db *gorm.DB, logger *zap.Logger) *WorkerManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerManager{
		jobQueue:      jobQueue,
		cronScheduler: cronScheduler,
		db:            db,
		logger:        logger,
		workers:       make([]Worker, 0),
		ctx:           ctx,
		cancel:        cancel,
		running:       false,
	}
}

// Start starts all workers
func (wm *WorkerManager) Start() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.running {
		return fmt.Errorf("worker manager is already running")
	}

	wm.logger.Info("Starting worker manager...")

	// Start job queue server
	go func() {
		if err := wm.jobQueue.Start(); err != nil {
			wm.logger.Error("Failed to start job queue", zap.Error(err))
		}
	}()

	// Start cron scheduler
	go func() {
		if err := wm.cronScheduler.Start(); err != nil {
			wm.logger.Error("Failed to start cron scheduler", zap.Error(err))
		}
	}()

	// Register default workers
	wm.registerDefaultWorkers()

	// Start all workers
	for i := range wm.workers {
		wm.startWorker(&wm.workers[i])
	}

	wm.running = true
	wm.logger.Info("Worker manager started successfully")

	return nil
}

// Stop stops all workers
func (wm *WorkerManager) Stop() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if !wm.running {
		return
	}

	wm.logger.Info("Stopping worker manager...")

	// Cancel context to signal all workers to stop
	wm.cancel()

	// Stop job queue and cron scheduler
	wm.jobQueue.Stop()
	wm.cronScheduler.Stop()

	// Wait for all workers to finish
	wm.wg.Wait()

	wm.running = false
	wm.logger.Info("Worker manager stopped")
}

// registerDefaultWorkers registers default background workers
func (wm *WorkerManager) registerDefaultWorkers() {
	// Email worker
	wm.AddWorker(Worker{
		ID:        "email-worker",
		Name:      "Email Processing Worker",
		Type:      "email",
		Status:    WorkerStatusStopped,
		MaxErrors: 5,
		Handler:   wm.emailWorker,
	})

	// Data processing worker
	wm.AddWorker(Worker{
		ID:        "data-worker",
		Name:      "Data Processing Worker",
		Type:      "data",
		Status:    WorkerStatusStopped,
		MaxErrors: 3,
		Handler:   wm.dataWorker,
	})

	// Payment worker
	wm.AddWorker(Worker{
		ID:        "payment-worker",
		Name:      "Payment Processing Worker",
		Type:      "payment",
		Status:    WorkerStatusStopped,
		MaxErrors: 3,
		Handler:   wm.paymentWorker,
	})

	// System monitoring worker
	wm.AddWorker(Worker{
		ID:        "monitoring-worker",
		Name:      "System Monitoring Worker",
		Type:      "monitoring",
		Status:    WorkerStatusStopped,
		MaxErrors: 10,
		Handler:   wm.monitoringWorker,
	})

	// Cache worker
	wm.AddWorker(Worker{
		ID:        "cache-worker",
		Name:      "Cache Management Worker",
		Type:      "cache",
		Status:    WorkerStatusStopped,
		MaxErrors: 5,
		Handler:   wm.cacheWorker,
	})
}

// AddWorker adds a new worker
func (wm *WorkerManager) AddWorker(worker Worker) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.workers = append(wm.workers, worker)
	wm.logger.Info("Added worker", zap.String("worker_id", worker.ID), zap.String("name", worker.Name))
}

// startWorker starts a specific worker
func (wm *WorkerManager) startWorker(worker *Worker) {
	wm.wg.Add(1)

	go func(w *Worker) {
		defer wm.wg.Done()

		w.Status = WorkerStatusRunning
		w.LastActivity = time.Now()

		wm.logger.Info("Starting worker", zap.String("worker_id", w.ID), zap.String("name", w.Name))

		ticker := time.NewTicker(30 * time.Second) // Health check every 30 seconds
		defer ticker.Stop()

		for {
			select {
			case <-wm.ctx.Done():
				wm.logger.Info("Stopping worker", zap.String("worker_id", w.ID))
				w.Status = WorkerStatusStopped
				return

			case <-ticker.C:
				// Health check - update last activity
				w.LastActivity = time.Now()

				// Execute worker handler
				if err := w.Handler(wm.ctx); err != nil {
					w.ErrorCount++
					wm.logger.Error("Worker error",
						zap.String("worker_id", w.ID),
						zap.Error(err),
						zap.Int("error_count", w.ErrorCount))

					// Check if worker should be stopped due to too many errors
					if w.ErrorCount >= w.MaxErrors {
						wm.logger.Error("Worker stopped due to too many errors",
							zap.String("worker_id", w.ID),
							zap.Int("error_count", w.ErrorCount),
							zap.Int("max_errors", w.MaxErrors))
						w.Status = WorkerStatusError
						return
					}
				} else {
					// Reset error count on successful execution
					w.ErrorCount = 0
				}
			}
		}
	}(worker)
}

// Worker Handlers

func (wm *WorkerManager) emailWorker(ctx context.Context) error {
	// This worker would process email-related background tasks
	// For now, we'll just log it
	wm.logger.Debug("Email worker executing...")

	// In a real implementation, this would:
	// - Process email queues
	// - Send bulk emails
	// - Handle email bounces
	// - Process email templates

	return nil
}

func (wm *WorkerManager) dataWorker(ctx context.Context) error {
	// This worker would process data-related background tasks
	wm.logger.Debug("Data worker executing...")

	// In a real implementation, this would:
	// - Process data imports
	// - Clean up old data
	// - Generate reports
	// - Handle data migrations

	return nil
}

func (wm *WorkerManager) paymentWorker(ctx context.Context) error {
	// This worker would process payment-related background tasks
	wm.logger.Debug("Payment worker executing...")

	// In a real implementation, this would:
	// - Process payment webhooks
	// - Handle payment retries
	// - Reconcile payments
	// - Process refunds

	return nil
}

func (wm *WorkerManager) monitoringWorker(ctx context.Context) error {
	// This worker would handle system monitoring
	wm.logger.Debug("Monitoring worker executing...")

	// In a real implementation, this would:
	// - Collect system metrics
	// - Check system health
	// - Send alerts
	// - Update dashboards

	return nil
}

func (wm *WorkerManager) cacheWorker(ctx context.Context) error {
	// This worker would handle cache management
	wm.logger.Debug("Cache worker executing...")

	// In a real implementation, this would:
	// - Warm up caches
	// - Clean expired entries
	// - Optimize cache performance
	// - Monitor cache hit rates

	return nil
}

// GetWorkerStatus returns the status of all workers
func (wm *WorkerManager) GetWorkerStatus() []WorkerStatus {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	statuses := make([]WorkerStatus, len(wm.workers))
	for i, worker := range wm.workers {
		uptime := time.Since(worker.LastActivity)
		statuses[i] = WorkerStatus{
			ID:           worker.ID,
			Name:         worker.Name,
			Type:         worker.Type,
			Status:       worker.Status,
			LastActivity: worker.LastActivity,
			ErrorCount:   worker.ErrorCount,
			MaxErrors:    worker.MaxErrors,
			Uptime:       uptime.String(),
		}
	}

	return statuses
}

// GetWorkerByID returns a specific worker by ID
func (wm *WorkerManager) GetWorkerByID(id string) *Worker {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	for i := range wm.workers {
		if wm.workers[i].ID == id {
			return &wm.workers[i]
		}
	}

	return nil
}

// RestartWorker restarts a specific worker
func (wm *WorkerManager) RestartWorker(id string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for i := range wm.workers {
		if wm.workers[i].ID == id {
			wm.logger.Info("Restarting worker", zap.String("worker_id", id))

			// Reset error count and status
			wm.workers[i].ErrorCount = 0
			wm.workers[i].Status = WorkerStatusStopped

			// Start the worker again
			wm.startWorker(&wm.workers[i])

			return nil
		}
	}

	return fmt.Errorf("worker not found: %s", id)
}

// PauseWorker pauses a specific worker
func (wm *WorkerManager) PauseWorker(id string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for i := range wm.workers {
		if wm.workers[i].ID == id {
			wm.workers[i].Status = WorkerStatusPaused
			wm.logger.Info("Paused worker", zap.String("worker_id", id))
			return nil
		}
	}

	return fmt.Errorf("worker not found: %s", id)
}

// ResumeWorker resumes a specific worker
func (wm *WorkerManager) ResumeWorker(id string) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	for i := range wm.workers {
		if wm.workers[i].ID == id {
			if wm.workers[i].Status == WorkerStatusPaused {
				wm.workers[i].Status = WorkerStatusRunning
				wm.logger.Info("Resumed worker", zap.String("worker_id", id))
				return nil
			}
			return fmt.Errorf("worker is not paused: %s", id)
		}
	}

	return fmt.Errorf("worker not found: %s", id)
}

// IsRunning returns whether the worker manager is running
func (wm *WorkerManager) IsRunning() bool {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	return wm.running
}

// GetStats returns worker manager statistics
func (wm *WorkerManager) GetStats() map[string]interface{} {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	stats := map[string]interface{}{
		"running":      wm.running,
		"worker_count": len(wm.workers),
		"workers":      wm.GetWorkerStatus(),
	}

	// Add job queue stats if available
	if queueStats, err := wm.jobQueue.GetQueueStats(); err == nil {
		stats["queue_stats"] = queueStats
	}

	return stats
}
