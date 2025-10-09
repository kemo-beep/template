package services

import (
	"time"

	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// JobQueueMetrics handles job queue metrics collection
type JobQueueMetrics struct {
	redisAddr string
	logger    *zap.Logger

	// Prometheus metrics
	jobsProcessedTotal    *prometheus.CounterVec
	jobsFailedTotal       *prometheus.CounterVec
	jobsEnqueuedTotal     *prometheus.CounterVec
	jobProcessingDuration *prometheus.HistogramVec
	queueSize             *prometheus.GaugeVec
	activeWorkers         *prometheus.GaugeVec
	workerErrors          *prometheus.CounterVec
}

// NewJobQueueMetrics creates a new job queue metrics collector
func NewJobQueueMetrics(redisAddr string, logger *zap.Logger) *JobQueueMetrics {
	return &JobQueueMetrics{
		redisAddr: redisAddr,
		logger:    logger,

		jobsProcessedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "job_queue_jobs_processed_total",
				Help: "Total number of jobs processed",
			},
			[]string{"queue", "type", "status"},
		),

		jobsFailedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "job_queue_jobs_failed_total",
				Help: "Total number of jobs that failed",
			},
			[]string{"queue", "type", "error_type"},
		),

		jobsEnqueuedTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "job_queue_jobs_enqueued_total",
				Help: "Total number of jobs enqueued",
			},
			[]string{"queue", "type"},
		),

		jobProcessingDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "job_queue_job_processing_duration_seconds",
				Help:    "Time spent processing jobs",
				Buckets: prometheus.ExponentialBuckets(0.1, 2, 10),
			},
			[]string{"queue", "type"},
		),

		queueSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "job_queue_queue_size",
				Help: "Current size of job queues",
			},
			[]string{"queue"},
		),

		activeWorkers: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "job_queue_active_workers",
				Help: "Number of active workers",
			},
			[]string{"worker_type"},
		),

		workerErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "job_queue_worker_errors_total",
				Help: "Total number of worker errors",
			},
			[]string{"worker_id", "worker_type", "error_type"},
		),
	}
}

// Start starts the metrics collection
func (jqm *JobQueueMetrics) Start() {
	go jqm.collectMetrics()
}

// collectMetrics collects metrics periodically
func (jqm *JobQueueMetrics) collectMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		jqm.updateQueueMetrics()
	}
}

// updateQueueMetrics updates queue-related metrics
func (jqm *JobQueueMetrics) updateQueueMetrics() {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: jqm.redisAddr})

	// Get queue information
	queues := []string{"critical", "default", "low"}
	for _, queue := range queues {
		queueInfo, err := inspector.GetQueueInfo(queue)
		if err != nil {
			jqm.logger.Error("Failed to get queue info", zap.String("queue", queue), zap.Error(err))
			continue
		}

		// Update queue size
		jqm.queueSize.WithLabelValues(queue).Set(float64(queueInfo.Size))

		// Update processed jobs
		jqm.jobsProcessedTotal.WithLabelValues(queue, "all", "completed").Add(float64(queueInfo.Processed))
		jqm.jobsFailedTotal.WithLabelValues(queue, "all", "failed").Add(float64(queueInfo.Failed))
	}
}

// RecordJobProcessed records a processed job
func (jqm *JobQueueMetrics) RecordJobProcessed(queue, jobType, status string) {
	jqm.jobsProcessedTotal.WithLabelValues(queue, jobType, status).Inc()
}

// RecordJobFailed records a failed job
func (jqm *JobQueueMetrics) RecordJobFailed(queue, jobType, errorType string) {
	jqm.jobsFailedTotal.WithLabelValues(queue, jobType, errorType).Inc()
}

// RecordJobEnqueued records an enqueued job
func (jqm *JobQueueMetrics) RecordJobEnqueued(queue, jobType string) {
	jqm.jobsEnqueuedTotal.WithLabelValues(queue, jobType).Inc()
}

// RecordJobProcessingDuration records job processing duration
func (jqm *JobQueueMetrics) RecordJobProcessingDuration(queue, jobType string, duration time.Duration) {
	jqm.jobProcessingDuration.WithLabelValues(queue, jobType).Observe(duration.Seconds())
}

// RecordActiveWorkers records the number of active workers
func (jqm *JobQueueMetrics) RecordActiveWorkers(workerType string, count int) {
	jqm.activeWorkers.WithLabelValues(workerType).Set(float64(count))
}

// RecordWorkerError records a worker error
func (jqm *JobQueueMetrics) RecordWorkerError(workerID, workerType, errorType string) {
	jqm.workerErrors.WithLabelValues(workerID, workerType, errorType).Inc()
}

// GetQueueStats returns current queue statistics
func (jqm *JobQueueMetrics) GetQueueStats() (map[string]interface{}, error) {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: jqm.redisAddr})

	stats := make(map[string]interface{})
	queues := []string{"critical", "default", "low"}

	for _, queue := range queues {
		queueInfo, err := inspector.GetQueueInfo(queue)
		if err != nil {
			jqm.logger.Error("Failed to get queue info", zap.String("queue", queue), zap.Error(err))
			continue
		}

		stats[queue] = map[string]interface{}{
			"size":      queueInfo.Size,
			"processed": queueInfo.Processed,
			"failed":    queueInfo.Failed,
			"pending":   queueInfo.Pending,
			"active":    queueInfo.Active,
			"scheduled": queueInfo.Scheduled,
			"retry":     queueInfo.Retry,
			"archived":  queueInfo.Archived,
		}
	}

	return stats, nil
}

// GetWorkerStats returns current worker statistics
func (jqm *JobQueueMetrics) GetWorkerStats() (map[string]interface{}, error) {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: jqm.redisAddr})
	defer inspector.Close()

	servers, err := inspector.Servers()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_servers": len(servers),
		"servers":       servers,
	}

	// Group servers by status
	serverStatuses := make(map[string]int)
	for _, server := range servers {
		serverStatuses[server.Status]++
	}
	stats["server_statuses"] = serverStatuses

	return stats, nil
}

// GetTaskStats returns task statistics
func (jqm *JobQueueMetrics) GetTaskStats() (map[string]interface{}, error) {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: jqm.redisAddr})
	defer inspector.Close()

	// Get queues first
	queues, err := inspector.Queues()
	if err != nil {
		return nil, err
	}

	taskStats := make(map[string]interface{})
	totalTasks := 0

	for _, queue := range queues {
		queueStats := make(map[string]interface{})

		// Get tasks by state for each queue
		pendingTasks, _ := inspector.ListPendingTasks(queue)
		activeTasks, _ := inspector.ListActiveTasks(queue)
		scheduledTasks, _ := inspector.ListScheduledTasks(queue)
		retryTasks, _ := inspector.ListRetryTasks(queue)
		archivedTasks, _ := inspector.ListArchivedTasks(queue)
		completedTasks, _ := inspector.ListCompletedTasks(queue)

		queueStats["pending"] = len(pendingTasks)
		queueStats["active"] = len(activeTasks)
		queueStats["scheduled"] = len(scheduledTasks)
		queueStats["retry"] = len(retryTasks)
		queueStats["archived"] = len(archivedTasks)
		queueStats["completed"] = len(completedTasks)

		queueTotal := len(pendingTasks) + len(activeTasks) + len(scheduledTasks) + len(retryTasks) + len(archivedTasks) + len(completedTasks)
		queueStats["total"] = queueTotal
		totalTasks += queueTotal

		taskStats[queue] = queueStats
	}

	taskStats["total_tasks"] = totalTasks
	return taskStats, nil
}

// GetHealthStatus returns the health status of the job queue system
func (jqm *JobQueueMetrics) GetHealthStatus() map[string]interface{} {
	inspector := asynq.NewInspector(asynq.RedisClientOpt{Addr: jqm.redisAddr})
	defer inspector.Close()

	// Check Redis connection
	_, err := inspector.GetQueueInfo("default")
	redisHealthy := err == nil

	// Get server count
	servers, err := inspector.Servers()
	serverCount := 0
	if err == nil {
		serverCount = len(servers)
	}

	// Get queue sizes
	queueSizes := make(map[string]int)
	queues := []string{"critical", "default", "low"}
	for _, queue := range queues {
		queueInfo, err := inspector.GetQueueInfo(queue)
		if err == nil {
			queueSizes[queue] = queueInfo.Size
		}
	}

	// Determine overall health
	healthy := redisHealthy && serverCount > 0

	status := "unhealthy"
	if healthy {
		status = "healthy"
	}

	return map[string]interface{}{
		"status":        status,
		"redis_healthy": redisHealthy,
		"server_count":  serverCount,
		"queue_sizes":   queueSizes,
		"timestamp":     time.Now().Unix(),
	}
}

// ResetMetrics resets all metrics
func (jqm *JobQueueMetrics) ResetMetrics() {
	jqm.jobsProcessedTotal.Reset()
	jqm.jobsFailedTotal.Reset()
	jqm.jobsEnqueuedTotal.Reset()
	jqm.jobProcessingDuration.Reset()
	jqm.queueSize.Reset()
	jqm.activeWorkers.Reset()
	jqm.workerErrors.Reset()

	jqm.logger.Info("Job queue metrics reset")
}

// GetMetricsSummary returns a summary of all metrics
func (jqm *JobQueueMetrics) GetMetricsSummary() map[string]interface{} {
	queueStats, _ := jqm.GetQueueStats()
	workerStats, _ := jqm.GetWorkerStats()
	taskStats, _ := jqm.GetTaskStats()
	healthStatus := jqm.GetHealthStatus()

	return map[string]interface{}{
		"queue_stats":   queueStats,
		"worker_stats":  workerStats,
		"task_stats":    taskStats,
		"health_status": healthStatus,
		"timestamp":     time.Now().Unix(),
	}
}
