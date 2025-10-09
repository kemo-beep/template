package controllers

import (
	"net/http"
	"strconv"

	"mobile-backend/services"
	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GeminiController struct {
	geminiService *services.GeminiService
	logger        *zap.Logger
}

func NewGeminiController(geminiService *services.GeminiService, logger *zap.Logger) *GeminiController {
	return &GeminiController{
		geminiService: geminiService,
		logger:        logger,
	}
}

type GenerateTextRequest struct {
	Prompt      string                 `json:"prompt" binding:"required"`
	Model       string                 `json:"model,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	TopK        int                    `json:"top_k,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Context     string                 `json:"context,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type CreateConversationRequest struct {
	Title string `json:"title" binding:"required"`
}

type AddMessageRequest struct {
	Role    string `json:"role" binding:"required,oneof=user assistant"`
	Content string `json:"content" binding:"required"`
}

// GenerateText godoc
// @Summary Generate text using Gemini AI
// @Description Generate text using Google Gemini AI with customizable parameters
// @Tags gemini
// @Accept json
// @Produce json
// @Param request body GenerateTextRequest true "Text generation request"
// @Success 200 {object} utils.SuccessResponse{data=services.GeminiResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/gemini/generate [post]
// @Security BearerAuth
func (gc *GeminiController) GenerateText(c *gin.Context) {
	var req GenerateTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Convert to service request
	geminiReq := &services.GeminiRequest{
		Prompt:      req.Prompt,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		TopK:        req.TopK,
		Stream:      req.Stream,
		Context:     req.Context,
		Metadata:    req.Metadata,
	}

	response, err := gc.geminiService.GenerateText(c.Request.Context(), geminiReq)
	if err != nil {
		gc.logger.Error("Failed to generate text", zap.Error(err))
		utils.SendInternalServerErrorResponse(c, "Failed to generate text")
		return
	}

	utils.SendSuccessResponse(c, response, "Text generated successfully")
}

// GenerateTextWithContext godoc
// @Summary Generate text with conversation context
// @Description Generate text using Gemini AI with conversation history context
// @Tags gemini
// @Accept json
// @Produce json
// @Param conversation_id path string true "Conversation ID"
// @Param request body GenerateTextRequest true "Text generation request with context"
// @Success 200 {object} utils.SuccessResponse{data=services.GeminiResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/gemini/conversations/{conversation_id}/generate [post]
// @Security BearerAuth
func (gc *GeminiController) GenerateTextWithContext(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	if conversationID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Conversation ID is required", nil)
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Get conversation to retrieve history
	conversation, err := gc.geminiService.GetConversation(c.Request.Context(), conversationID, userID.(uint))
	if err != nil {
		utils.SendNotFoundResponse(c, "Conversation not found")
		return
	}

	var req GenerateTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Convert to service request
	geminiReq := &services.GeminiRequest{
		Prompt:      req.Prompt,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		TopK:        req.TopK,
		Stream:      req.Stream,
		Context:     req.Context,
		Metadata:    req.Metadata,
	}

	response, err := gc.geminiService.GenerateTextWithContext(c.Request.Context(), geminiReq, conversation.Messages)
	if err != nil {
		gc.logger.Error("Failed to generate text with context", zap.Error(err))
		utils.SendInternalServerErrorResponse(c, "Failed to generate text")
		return
	}

	// Add user message to conversation
	_, err = gc.geminiService.AddMessage(c.Request.Context(), conversationID, "user", req.Prompt)
	if err != nil {
		gc.logger.Warn("Failed to add user message to conversation", zap.Error(err))
	}

	// Add assistant response to conversation
	_, err = gc.geminiService.AddMessage(c.Request.Context(), conversationID, "assistant", response.Content)
	if err != nil {
		gc.logger.Warn("Failed to add assistant message to conversation", zap.Error(err))
	}

	utils.SendSuccessResponse(c, response, "Text generated successfully")
}

// CreateConversation godoc
// @Summary Create a new conversation
// @Description Create a new Gemini AI conversation
// @Tags gemini
// @Accept json
// @Produce json
// @Param request body CreateConversationRequest true "Conversation creation request"
// @Success 201 {object} utils.SuccessResponse{data=services.GeminiConversation}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/gemini/conversations [post]
// @Security BearerAuth
func (gc *GeminiController) CreateConversation(c *gin.Context) {
	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorizedResponse(c, "User not authenticated")
		return
	}

	conversation, err := gc.geminiService.CreateConversation(c.Request.Context(), userID.(uint), req.Title)
	if err != nil {
		gc.logger.Error("Failed to create conversation", zap.Error(err))
		utils.SendInternalServerErrorResponse(c, "Failed to create conversation")
		return
	}

	utils.SendCreatedResponse(c, conversation, "Conversation created successfully")
}

// GetConversation godoc
// @Summary Get a conversation
// @Description Retrieve a specific conversation by ID
// @Tags gemini
// @Produce json
// @Param conversation_id path string true "Conversation ID"
// @Success 200 {object} utils.SuccessResponse{data=services.GeminiConversation}
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/gemini/conversations/{conversation_id} [get]
// @Security BearerAuth
func (gc *GeminiController) GetConversation(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	if conversationID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Conversation ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorizedResponse(c, "User not authenticated")
		return
	}

	conversation, err := gc.geminiService.GetConversation(c.Request.Context(), conversationID, userID.(uint))
	if err != nil {
		utils.SendNotFoundResponse(c, "Conversation not found")
		return
	}

	utils.SendSuccessResponse(c, conversation, "Conversation retrieved successfully")
}

// ListConversations godoc
// @Summary List conversations
// @Description List all conversations for the authenticated user
// @Tags gemini
// @Produce json
// @Param limit query int false "Number of conversations to return" default(10)
// @Param offset query int false "Number of conversations to skip" default(0)
// @Success 200 {object} utils.SuccessResponse{data=[]services.GeminiConversation}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/gemini/conversations [get]
// @Security BearerAuth
func (gc *GeminiController) ListConversations(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	conversations, err := gc.geminiService.ListConversations(c.Request.Context(), userID.(uint), limit, offset)
	if err != nil {
		gc.logger.Error("Failed to list conversations", zap.Error(err))
		utils.SendInternalServerErrorResponse(c, "Failed to list conversations")
		return
	}

	utils.SendSuccessResponse(c, conversations, "Conversations retrieved successfully")
}

// AddMessage godoc
// @Summary Add a message to conversation
// @Description Add a message to an existing conversation
// @Tags gemini
// @Accept json
// @Produce json
// @Param conversation_id path string true "Conversation ID"
// @Param request body AddMessageRequest true "Message to add"
// @Success 201 {object} utils.SuccessResponse{data=services.GeminiMessage}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/gemini/conversations/{conversation_id}/messages [post]
// @Security BearerAuth
func (gc *GeminiController) AddMessage(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	if conversationID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Conversation ID is required", nil)
		return
	}

	var req AddMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Verify conversation exists and belongs to user
	_, err := gc.geminiService.GetConversation(c.Request.Context(), conversationID, userID.(uint))
	if err != nil {
		utils.SendNotFoundResponse(c, "Conversation not found")
		return
	}

	message, err := gc.geminiService.AddMessage(c.Request.Context(), conversationID, req.Role, req.Content)
	if err != nil {
		gc.logger.Error("Failed to add message", zap.Error(err))
		utils.SendInternalServerErrorResponse(c, "Failed to add message")
		return
	}

	utils.SendCreatedResponse(c, message, "Message added successfully")
}

// DeleteConversation godoc
// @Summary Delete a conversation
// @Description Delete a specific conversation
// @Tags gemini
// @Param conversation_id path string true "Conversation ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/gemini/conversations/{conversation_id} [delete]
// @Security BearerAuth
func (gc *GeminiController) DeleteConversation(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	if conversationID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Conversation ID is required", nil)
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorizedResponse(c, "User not authenticated")
		return
	}

	err := gc.geminiService.DeleteConversation(c.Request.Context(), conversationID, userID.(uint))
	if err != nil {
		utils.SendNotFoundResponse(c, "Failed to delete conversation")
		return
	}

	utils.SendSuccessResponse(c, nil, "Conversation deleted successfully")
}

// GetAvailableModels godoc
// @Summary Get available models
// @Description Get list of available Gemini AI models
// @Tags gemini
// @Produce json
// @Success 200 {object} utils.SuccessResponse{data=[]string}
// @Router /api/v1/gemini/models [get]
func (gc *GeminiController) GetAvailableModels(c *gin.Context) {
	models := gc.geminiService.GetAvailableModels()
	utils.SendSuccessResponse(c, models, "Available models retrieved successfully")
}

// GetServiceStats godoc
// @Summary Get service statistics
// @Description Get Gemini AI service statistics and metrics
// @Tags gemini
// @Produce json
// @Success 200 {object} utils.SuccessResponse{data=map[string]interface{}}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/gemini/stats [get]
// @Security BearerAuth
func (gc *GeminiController) GetServiceStats(c *gin.Context) {
	stats, err := gc.geminiService.GetServiceStats(c.Request.Context())
	if err != nil {
		gc.logger.Error("Failed to get service stats", zap.Error(err))
		utils.SendInternalServerErrorResponse(c, "Failed to get service stats")
		return
	}

	utils.SendSuccessResponse(c, stats, "Service stats retrieved successfully")
}

// HealthCheck godoc
// @Summary Health check for Gemini service
// @Description Check if Gemini AI service is healthy and responsive
// @Tags gemini
// @Produce json
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/gemini/health [get]
func (gc *GeminiController) HealthCheck(c *gin.Context) {
	err := gc.geminiService.HealthCheck(c.Request.Context())
	if err != nil {
		gc.logger.Error("Gemini service health check failed", zap.Error(err))
		utils.SendInternalServerErrorResponse(c, "Gemini service is not healthy")
		return
	}

	utils.SendSuccessResponse(c, map[string]string{
		"status":  "healthy",
		"service": "gemini",
	}, "Gemini service is healthy")
}
