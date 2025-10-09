package controllers

import (
	"net/http"

	"mobile-backend/services"
	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// JobQueueController handles job queue operations
type JobQueueController struct {
	jobQueue      *services.JobQueueService
	workerManager *services.WorkerManager
	logger        *zap.Logger
}

// NewJobQueueController creates a new job queue controller
func NewJobQueueController(jobQueue *services.JobQueueService, workerManager *services.WorkerManager, logger *zap.Logger) *JobQueueController {
	return &JobQueueController{
		jobQueue:      jobQueue,
		workerManager: workerManager,
		logger:        logger,
	}
}

// GetQueueStats returns queue statistics
// @Summary Get queue statistics
// @Description Get statistics about the job queue
// @Tags job-queue
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/stats [get]
func (jqc *JobQueueController) GetQueueStats(c *gin.Context) {
	stats, err := jqc.jobQueue.GetQueueStats()
	if err != nil {
		jqc.logger.Error("Failed to get queue stats", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get queue stats", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, stats, "Queue statistics retrieved successfully")
}

// GetWorkerStatus returns worker status information
// @Summary Get worker status
// @Description Get status information about all workers
// @Tags job-queue
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/workers [get]
func (jqc *JobQueueController) GetWorkerStatus(c *gin.Context) {
	statuses := jqc.workerManager.GetWorkerStatus()
	stats := jqc.workerManager.GetStats()

	response := map[string]interface{}{
		"workers": statuses,
		"stats":   stats,
	}

	utils.SendSuccessResponse(c, response, "Worker status retrieved successfully")
}

// EnqueueEmailNotification enqueues an email notification job
// @Summary Enqueue email notification
// @Description Enqueue an email notification job
// @Tags job-queue
// @Accept json
// @Produce json
// @Param payload body services.EmailNotificationPayload true "Email notification payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/email/notification [post]
func (jqc *JobQueueController) EnqueueEmailNotification(c *gin.Context) {
	var payload services.EmailNotificationPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payload", map[string]interface{}{"error": err.Error()})
		return
	}

	// Set default priority if not provided
	if payload.Priority == 0 {
		payload.Priority = 1
	}

	taskInfo, err := jqc.jobQueue.EnqueueEmailNotification(payload, asynq.Queue("default"))
	if err != nil {
		jqc.logger.Error("Failed to enqueue email notification", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to enqueue email notification", map[string]interface{}{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"task_id": taskInfo.ID,
		"queue":   taskInfo.Queue,
		"type":    taskInfo.Type,
		"state":   taskInfo.State,
	}

	utils.SendSuccessResponse(c, response, "Email notification job enqueued successfully")
}

// EnqueueBulkEmail enqueues a bulk email job
// @Summary Enqueue bulk email
// @Description Enqueue a bulk email job
// @Tags job-queue
// @Accept json
// @Produce json
// @Param payload body map[string]interface{} true "Bulk email payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/email/bulk [post]
func (jqc *JobQueueController) EnqueueBulkEmail(c *gin.Context) {
	var payload struct {
		UserIDs []uint `json:"user_ids" binding:"required"`
		Subject string `json:"subject" binding:"required"`
		Body    string `json:"body" binding:"required"`
		Queue   string `json:"queue,omitempty"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payload", map[string]interface{}{"error": err.Error()})
		return
	}

	queue := "low"
	if payload.Queue != "" {
		queue = payload.Queue
	}

	taskInfo, err := jqc.jobQueue.EnqueueEmailBulk(payload.UserIDs, payload.Subject, payload.Body, asynq.Queue(queue))
	if err != nil {
		jqc.logger.Error("Failed to enqueue bulk email", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to enqueue bulk email", map[string]interface{}{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"task_id":    taskInfo.ID,
		"queue":      taskInfo.Queue,
		"type":       taskInfo.Type,
		"state":      taskInfo.State,
		"user_count": len(payload.UserIDs),
	}

	utils.SendSuccessResponse(c, response, "Bulk email job enqueued successfully")
}

// EnqueueDataCleanup enqueues a data cleanup job
// @Summary Enqueue data cleanup
// @Description Enqueue a data cleanup job
// @Tags job-queue
// @Accept json
// @Produce json
// @Param payload body services.DataCleanupPayload true "Data cleanup payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/data/cleanup [post]
func (jqc *JobQueueController) EnqueueDataCleanup(c *gin.Context) {
	var payload services.DataCleanupPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payload", map[string]interface{}{"error": err.Error()})
		return
	}

	// Set default batch size if not provided
	if payload.BatchSize == 0 {
		payload.BatchSize = 1000
	}

	taskInfo, err := jqc.jobQueue.EnqueueDataCleanup(payload, asynq.Queue("low"))
	if err != nil {
		jqc.logger.Error("Failed to enqueue data cleanup", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to enqueue data cleanup", map[string]interface{}{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"task_id":    taskInfo.ID,
		"queue":      taskInfo.Queue,
		"type":       taskInfo.Type,
		"state":      taskInfo.State,
		"table_name": payload.TableName,
	}

	utils.SendSuccessResponse(c, response, "Data cleanup job enqueued successfully")
}

// EnqueueReportGeneration enqueues a report generation job
// @Summary Enqueue report generation
// @Description Enqueue a report generation job
// @Tags job-queue
// @Accept json
// @Produce json
// @Param payload body services.ReportGenerationPayload true "Report generation payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/reports/generate [post]
func (jqc *JobQueueController) EnqueueReportGeneration(c *gin.Context) {
	var payload services.ReportGenerationPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payload", map[string]interface{}{"error": err.Error()})
		return
	}

	// Set default format if not provided
	if payload.Format == "" {
		payload.Format = "pdf"
	}

	taskInfo, err := jqc.jobQueue.EnqueueReportGeneration(payload, asynq.Queue("default"))
	if err != nil {
		jqc.logger.Error("Failed to enqueue report generation", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to enqueue report generation", map[string]interface{}{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"task_id":     taskInfo.ID,
		"queue":       taskInfo.Queue,
		"type":        taskInfo.Type,
		"state":       taskInfo.State,
		"report_type": payload.ReportType,
		"format":      payload.Format,
	}

	utils.SendSuccessResponse(c, response, "Report generation job enqueued successfully")
}

// EnqueueBackupTask enqueues a backup task job
// @Summary Enqueue backup task
// @Description Enqueue a backup task job
// @Tags job-queue
// @Accept json
// @Produce json
// @Param payload body services.BackupTaskPayload true "Backup task payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/backup [post]
func (jqc *JobQueueController) EnqueueBackupTask(c *gin.Context) {
	var payload services.BackupTaskPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payload", map[string]interface{}{"error": err.Error()})
		return
	}

	// Set default values
	if payload.Retention == 0 {
		payload.Retention = 30
	}

	taskInfo, err := jqc.jobQueue.EnqueueBackupTask(payload, asynq.Queue("low"))
	if err != nil {
		jqc.logger.Error("Failed to enqueue backup task", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to enqueue backup task", map[string]interface{}{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"task_id":     taskInfo.ID,
		"queue":       taskInfo.Queue,
		"type":        taskInfo.Type,
		"state":       taskInfo.State,
		"backup_type": payload.BackupType,
	}

	utils.SendSuccessResponse(c, response, "Backup task job enqueued successfully")
}

// GetTaskInfo returns information about a specific task
// @Summary Get task information
// @Description Get information about a specific task
// @Tags job-queue
// @Accept json
// @Produce json
// @Param queue path string true "Queue name"
// @Param task_id path string true "Task ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/tasks/{queue}/{task_id} [get]
func (jqc *JobQueueController) GetTaskInfo(c *gin.Context) {
	queue := c.Param("queue")
	taskID := c.Param("task_id")

	if queue == "" || taskID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Queue and task ID are required", map[string]interface{}{})
		return
	}

	taskInfo, err := jqc.jobQueue.GetTaskInfo(queue, taskID)
	if err != nil {
		jqc.logger.Error("Failed to get task info", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusNotFound, "Task not found", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, taskInfo, "Task information retrieved successfully")
}

// CancelTask cancels a specific task
// @Summary Cancel task
// @Description Cancel a specific task
// @Tags job-queue
// @Accept json
// @Produce json
// @Param queue path string true "Queue name"
// @Param task_id path string true "Task ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/tasks/{queue}/{task_id}/cancel [post]
func (jqc *JobQueueController) CancelTask(c *gin.Context) {
	queue := c.Param("queue")
	taskID := c.Param("task_id")

	if queue == "" || taskID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Queue and task ID are required", map[string]interface{}{})
		return
	}

	err := jqc.jobQueue.CancelTask(queue, taskID)
	if err != nil {
		jqc.logger.Error("Failed to cancel task", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to cancel task", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, map[string]string{
		"task_id": taskID,
		"queue":   queue,
		"status":  "cancelled",
	}, "Task cancelled successfully")
}

// DeleteTask deletes a specific task
// @Summary Delete task
// @Description Delete a specific task
// @Tags job-queue
// @Accept json
// @Produce json
// @Param queue path string true "Queue name"
// @Param task_id path string true "Task ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/tasks/{queue}/{task_id} [delete]
func (jqc *JobQueueController) DeleteTask(c *gin.Context) {
	queue := c.Param("queue")
	taskID := c.Param("task_id")

	if queue == "" || taskID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Queue and task ID are required", map[string]interface{}{})
		return
	}

	err := jqc.jobQueue.DeleteTask(queue, taskID)
	if err != nil {
		jqc.logger.Error("Failed to delete task", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete task", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, map[string]string{
		"task_id": taskID,
		"queue":   queue,
		"status":  "deleted",
	}, "Task deleted successfully")
}

// RestartWorker restarts a specific worker
// @Summary Restart worker
// @Description Restart a specific worker
// @Tags job-queue
// @Accept json
// @Produce json
// @Param worker_id path string true "Worker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/workers/{worker_id}/restart [post]
func (jqc *JobQueueController) RestartWorker(c *gin.Context) {
	workerID := c.Param("worker_id")

	if workerID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Worker ID is required", map[string]interface{}{})
		return
	}

	err := jqc.workerManager.RestartWorker(workerID)
	if err != nil {
		jqc.logger.Error("Failed to restart worker", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusNotFound, "Worker not found", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, map[string]string{
		"worker_id": workerID,
		"status":    "restarted",
	}, "Worker restarted successfully")
}

// PauseWorker pauses a specific worker
// @Summary Pause worker
// @Description Pause a specific worker
// @Tags job-queue
// @Accept json
// @Produce json
// @Param worker_id path string true "Worker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/workers/{worker_id}/pause [post]
func (jqc *JobQueueController) PauseWorker(c *gin.Context) {
	workerID := c.Param("worker_id")

	if workerID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Worker ID is required", map[string]interface{}{})
		return
	}

	err := jqc.workerManager.PauseWorker(workerID)
	if err != nil {
		jqc.logger.Error("Failed to pause worker", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusNotFound, "Worker not found", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, map[string]string{
		"worker_id": workerID,
		"status":    "paused",
	}, "Worker paused successfully")
}

// ResumeWorker resumes a specific worker
// @Summary Resume worker
// @Description Resume a specific worker
// @Tags job-queue
// @Accept json
// @Produce json
// @Param worker_id path string true "Worker ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/workers/{worker_id}/resume [post]
func (jqc *JobQueueController) ResumeWorker(c *gin.Context) {
	workerID := c.Param("worker_id")

	if workerID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Worker ID is required", map[string]interface{}{})
		return
	}

	err := jqc.workerManager.ResumeWorker(workerID)
	if err != nil {
		jqc.logger.Error("Failed to resume worker", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusNotFound, "Worker not found", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, map[string]string{
		"worker_id": workerID,
		"status":    "resumed",
	}, "Worker resumed successfully")
}

// GetWorkerStats returns detailed worker statistics
// @Summary Get worker statistics
// @Description Get detailed statistics about workers
// @Tags job-queue
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/workers/stats [get]
func (jqc *JobQueueController) GetWorkerStats(c *gin.Context) {
	stats := jqc.workerManager.GetStats()

	utils.SendSuccessResponse(c, stats, "Worker statistics retrieved successfully")
}

// EnqueueUserActivity enqueues a user activity tracking job
// @Summary Enqueue user activity
// @Description Enqueue a user activity tracking job
// @Tags job-queue
// @Accept json
// @Produce json
// @Param payload body services.UserActivityPayload true "User activity payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/user/activity [post]
func (jqc *JobQueueController) EnqueueUserActivity(c *gin.Context) {
	var payload services.UserActivityPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payload", map[string]interface{}{"error": err.Error()})
		return
	}

	taskInfo, err := jqc.jobQueue.EnqueueUserActivity(payload, asynq.Queue("low"))
	if err != nil {
		jqc.logger.Error("Failed to enqueue user activity", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to enqueue user activity", map[string]interface{}{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"task_id":  taskInfo.ID,
		"queue":    taskInfo.Queue,
		"type":     taskInfo.Type,
		"state":    taskInfo.State,
		"user_id":  payload.UserID,
		"activity": payload.Activity,
	}

	utils.SendSuccessResponse(c, response, "User activity job enqueued successfully")
}

// EnqueueSubscriptionReminder enqueues a subscription reminder job
// @Summary Enqueue subscription reminder
// @Description Enqueue a subscription reminder job
// @Tags job-queue
// @Accept json
// @Produce json
// @Param payload body services.SubscriptionReminderPayload true "Subscription reminder payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/jobs/subscription/reminder [post]
func (jqc *JobQueueController) EnqueueSubscriptionReminder(c *gin.Context) {
	var payload services.SubscriptionReminderPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payload", map[string]interface{}{"error": err.Error()})
		return
	}

	taskInfo, err := jqc.jobQueue.EnqueueSubscriptionReminder(payload, asynq.Queue("default"))
	if err != nil {
		jqc.logger.Error("Failed to enqueue subscription reminder", zap.Error(err))
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to enqueue subscription reminder", map[string]interface{}{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"task_id":         taskInfo.ID,
		"queue":           taskInfo.Queue,
		"type":            taskInfo.Type,
		"state":           taskInfo.State,
		"subscription_id": payload.SubscriptionID,
		"reminder_type":   payload.ReminderType,
	}

	utils.SendSuccessResponse(c, response, "Subscription reminder job enqueued successfully")
}
