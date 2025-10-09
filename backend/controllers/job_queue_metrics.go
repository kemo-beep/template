package controllers

import (
	"net/http"

	"mobile-backend/services"
	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// JobQueueMetricsController handles job queue metrics operations
type JobQueueMetricsController struct {
	metrics *services.JobQueueMetrics
	logger  *zap.Logger
}

// NewJobQueueMetricsController creates a new job queue metrics controller
func NewJobQueueMetricsController(metrics *services.JobQueueMetrics, logger *zap.Logger) *JobQueueMetricsController {
	return &JobQueueMetricsController{
		metrics: metrics,
		logger:  logger,
	}
}

// GetQueueStats returns detailed queue statistics
// @Summary Get queue statistics
// @Description Get detailed statistics about job queues
// @Tags job-queue-metrics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/metrics/queues [get]
func (jqmc *JobQueueMetricsController) GetQueueStats(c *gin.Context) {
	stats, err := jqmc.metrics.GetQueueStats()
	if err != nil {
		jqmc.logger.Error("Failed to get queue stats", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get queue statistics", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, stats, "Queue statistics retrieved successfully")
}

// GetWorkerStats returns worker statistics
// @Summary Get worker statistics
// @Description Get detailed statistics about workers
// @Tags job-queue-metrics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/metrics/workers [get]
func (jqmc *JobQueueMetricsController) GetWorkerStats(c *gin.Context) {
	stats, err := jqmc.metrics.GetWorkerStats()
	if err != nil {
		jqmc.logger.Error("Failed to get worker stats", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get worker statistics", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, stats, "Worker statistics retrieved successfully")
}

// GetTaskStats returns task statistics
// @Summary Get task statistics
// @Description Get detailed statistics about tasks
// @Tags job-queue-metrics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/metrics/tasks [get]
func (jqmc *JobQueueMetricsController) GetTaskStats(c *gin.Context) {
	stats, err := jqmc.metrics.GetTaskStats()
	if err != nil {
		jqmc.logger.Error("Failed to get task stats", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get task statistics", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, stats, "Task statistics retrieved successfully")
}

// GetHealthStatus returns the health status of the job queue system
// @Summary Get job queue health status
// @Description Get the health status of the job queue system
// @Tags job-queue-metrics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/metrics/health [get]
func (jqmc *JobQueueMetricsController) GetHealthStatus(c *gin.Context) {
	status := jqmc.metrics.GetHealthStatus()

	utils.SendSuccessResponse(c, status, "Health status retrieved successfully")
}

// GetMetricsSummary returns a comprehensive metrics summary
// @Summary Get metrics summary
// @Description Get a comprehensive summary of all job queue metrics
// @Tags job-queue-metrics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/metrics/summary [get]
func (jqmc *JobQueueMetricsController) GetMetricsSummary(c *gin.Context) {
	summary := jqmc.metrics.GetMetricsSummary()

	utils.SendSuccessResponse(c, summary, "Metrics summary retrieved successfully")
}

// ResetMetrics resets all job queue metrics
// @Summary Reset metrics
// @Description Reset all job queue metrics
// @Tags job-queue-metrics
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/metrics/reset [post]
func (jqmc *JobQueueMetricsController) ResetMetrics(c *gin.Context) {
	jqmc.metrics.ResetMetrics()

	utils.SendSuccessResponse(c, map[string]string{
		"status": "reset",
	}, "Metrics reset successfully")
}

// GetPrometheusMetrics returns Prometheus-formatted metrics
// @Summary Get Prometheus metrics
// @Description Get job queue metrics in Prometheus format
// @Tags job-queue-metrics
// @Accept json
// @Produce text/plain
// @Success 200 {string} string
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/metrics/prometheus [get]
func (jqmc *JobQueueMetricsController) GetPrometheusMetrics(c *gin.Context) {
	// This would return Prometheus-formatted metrics
	// For now, we'll return a simple response
	c.String(http.StatusOK, "# Job Queue Metrics\n# This endpoint would return Prometheus-formatted metrics\n")
}
