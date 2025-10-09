package routes

import (
	"mobile-backend/controllers"

	"github.com/gin-gonic/gin"
)

// SetupJobQueueRoutes sets up job queue related routes
func SetupJobQueueRoutes(r *gin.Engine, jobQueueController *controllers.JobQueueController, metricsController *controllers.JobQueueMetricsController) {
	// Job queue API group
	jobGroup := r.Group("/api/v1/jobs")
	{
		// Queue statistics
		jobGroup.GET("/stats", jobQueueController.GetQueueStats)

		// Worker management
		jobGroup.GET("/workers", jobQueueController.GetWorkerStatus)
		jobGroup.GET("/workers/stats", jobQueueController.GetWorkerStats)
		jobGroup.POST("/workers/:worker_id/restart", jobQueueController.RestartWorker)
		jobGroup.POST("/workers/:worker_id/pause", jobQueueController.PauseWorker)
		jobGroup.POST("/workers/:worker_id/resume", jobQueueController.ResumeWorker)

		// Task management
		jobGroup.GET("/tasks/:queue/:task_id", jobQueueController.GetTaskInfo)
		jobGroup.POST("/tasks/:queue/:task_id/cancel", jobQueueController.CancelTask)
		jobGroup.DELETE("/tasks/:queue/:task_id", jobQueueController.DeleteTask)

		// Email jobs
		emailGroup := jobGroup.Group("/email")
		{
			emailGroup.POST("/notification", jobQueueController.EnqueueEmailNotification)
			emailGroup.POST("/bulk", jobQueueController.EnqueueBulkEmail)
		}

		// Data jobs
		dataGroup := jobGroup.Group("/data")
		{
			dataGroup.POST("/cleanup", jobQueueController.EnqueueDataCleanup)
		}

		// Report jobs
		reportGroup := jobGroup.Group("/reports")
		{
			reportGroup.POST("/generate", jobQueueController.EnqueueReportGeneration)
		}

		// Backup jobs
		jobGroup.POST("/backup", jobQueueController.EnqueueBackupTask)

		// User activity jobs
		userGroup := jobGroup.Group("/user")
		{
			userGroup.POST("/activity", jobQueueController.EnqueueUserActivity)
		}

		// Subscription jobs
		subscriptionGroup := jobGroup.Group("/subscription")
		{
			subscriptionGroup.POST("/reminder", jobQueueController.EnqueueSubscriptionReminder)
		}

		// Metrics endpoints
		metricsGroup := jobGroup.Group("/metrics")
		{
			metricsGroup.GET("/queues", metricsController.GetQueueStats)
			metricsGroup.GET("/workers", metricsController.GetWorkerStats)
			metricsGroup.GET("/tasks", metricsController.GetTaskStats)
			metricsGroup.GET("/health", metricsController.GetHealthStatus)
			metricsGroup.GET("/summary", metricsController.GetMetricsSummary)
			metricsGroup.POST("/reset", metricsController.ResetMetrics)
			metricsGroup.GET("/prometheus", metricsController.GetPrometheusMetrics)
		}
	}
}
