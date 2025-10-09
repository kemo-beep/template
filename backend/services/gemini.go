package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GeminiService struct {
	db          *gorm.DB
	cache       *CacheService
	httpClient  *http.Client
	apiKey      string
	baseURL     string
	model       string
	maxTokens   int
	temperature float64
	topP        float64
	topK        int
	logger      *zap.Logger
}

type GeminiRequest struct {
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

type GeminiResponse struct {
	ID           string                 `json:"id"`
	Content      string                 `json:"content"`
	Model        string                 `json:"model"`
	Usage        *GeminiUsage           `json:"usage,omitempty"`
	FinishReason string                 `json:"finish_reason,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

type GeminiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type GeminiConversation struct {
	ID        string          `json:"id"`
	UserID    uint            `json:"user_id"`
	Title     string          `json:"title"`
	Messages  []GeminiMessage `json:"messages"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type GeminiMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func NewGeminiService(db *gorm.DB, cache *CacheService, logger *zap.Logger) (*GeminiService, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	// Parse configuration from environment
	model := getEnvOrDefault("GEMINI_MODEL", "gemini-1.5-flash")
	maxTokens, _ := strconv.Atoi(getEnvOrDefault("GEMINI_MAX_TOKENS", "8192"))
	temperature, _ := strconv.ParseFloat(getEnvOrDefault("GEMINI_TEMPERATURE", "0.7"), 64)
	topP, _ := strconv.ParseFloat(getEnvOrDefault("GEMINI_TOP_P", "0.8"), 64)
	topK, _ := strconv.Atoi(getEnvOrDefault("GEMINI_TOP_K", "40"))

	return &GeminiService{
		db:          db,
		cache:       cache,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		apiKey:      apiKey,
		baseURL:     "https://generativelanguage.googleapis.com/v1beta",
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
		topP:        topP,
		topK:        topK,
		logger:      logger,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GenerateText generates text using Gemini AI
func (s *GeminiService) GenerateText(ctx context.Context, req *GeminiRequest) (*GeminiResponse, error) {
	// Use request parameters or fall back to service defaults
	model := req.Model
	if model == "" {
		model = s.model
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = s.maxTokens
	}
	temperature := req.Temperature
	if temperature == 0 {
		temperature = s.temperature
	}
	topP := req.TopP
	if topP == 0 {
		topP = s.topP
	}
	topK := req.TopK
	if topK == 0 {
		topK = s.topK
	}

	// Create the request payload
	requestPayload := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": req.Prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"maxOutputTokens": maxTokens,
			"temperature":     temperature,
			"topP":            topP,
			"topK":            topK,
		},
	}

	// Marshal the request
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", s.baseURL, model, s.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse the response
	var apiResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract content from response
	var content string
	if candidates, ok := apiResponse["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]interface{}); ok {
			if contentData, ok := candidate["content"].(map[string]interface{}); ok {
				if parts, ok := contentData["parts"].([]interface{}); ok && len(parts) > 0 {
					if part, ok := parts[0].(map[string]interface{}); ok {
						if text, ok := part["text"].(string); ok {
							content = text
						}
					}
				}
			}
		}
	}

	// Create response
	geminiResponse := &GeminiResponse{
		ID:           generateID(),
		Content:      content,
		Model:        model,
		Usage:        nil, // Usage info not available in this API response
		FinishReason: "stop",
		Metadata:     req.Metadata,
		CreatedAt:    time.Now(),
	}

	// Cache the response if caching is enabled
	if s.cache != nil {
		cacheKey := fmt.Sprintf("gemini:response:%s", geminiResponse.ID)
		s.cache.Set(ctx, cacheKey, geminiResponse, 24*time.Hour)
	}

	// Log the request for monitoring
	s.logger.Info("Gemini text generation completed",
		zap.String("model", model),
		zap.Int("max_tokens", maxTokens),
		zap.Float64("temperature", temperature),
		zap.String("response_id", geminiResponse.ID),
	)

	return geminiResponse, nil
}

// GenerateTextWithContext generates text with additional context
func (s *GeminiService) GenerateTextWithContext(ctx context.Context, req *GeminiRequest, conversationHistory []GeminiMessage) (*GeminiResponse, error) {
	// Build context-aware prompt
	contextPrompt := s.buildContextPrompt(req.Prompt, req.Context, conversationHistory)

	// Create new request with context
	contextReq := *req
	contextReq.Prompt = contextPrompt

	return s.GenerateText(ctx, &contextReq)
}

// CreateConversation creates a new conversation
func (s *GeminiService) CreateConversation(ctx context.Context, userID uint, title string) (*GeminiConversation, error) {
	conversation := &GeminiConversation{
		ID:        generateID(),
		UserID:    userID,
		Title:     title,
		Messages:  []GeminiMessage{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store in database
	if err := s.db.Create(conversation).Error; err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	return conversation, nil
}

// AddMessage adds a message to a conversation
func (s *GeminiService) AddMessage(ctx context.Context, conversationID string, role, content string) (*GeminiMessage, error) {
	message := &GeminiMessage{
		ID:        generateID(),
		Role:      role,
		Content:   content,
		CreatedAt: time.Now(),
	}

	// Update conversation in database
	var conversation GeminiConversation
	if err := s.db.Where("id = ?", conversationID).First(&conversation).Error; err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	conversation.Messages = append(conversation.Messages, *message)
	conversation.UpdatedAt = time.Now()

	if err := s.db.Save(&conversation).Error; err != nil {
		return nil, fmt.Errorf("failed to add message: %w", err)
	}

	return message, nil
}

// GetConversation retrieves a conversation by ID
func (s *GeminiService) GetConversation(ctx context.Context, conversationID string, userID uint) (*GeminiConversation, error) {
	var conversation GeminiConversation
	if err := s.db.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	return &conversation, nil
}

// ListConversations lists conversations for a user
func (s *GeminiService) ListConversations(ctx context.Context, userID uint, limit, offset int) ([]*GeminiConversation, error) {
	var conversations []*GeminiConversation
	query := s.db.Where("user_id = ?", userID).Order("updated_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&conversations).Error; err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	return conversations, nil
}

// DeleteConversation deletes a conversation
func (s *GeminiService) DeleteConversation(ctx context.Context, conversationID string, userID uint) error {
	result := s.db.Where("id = ? AND user_id = ?", conversationID, userID).Delete(&GeminiConversation{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete conversation: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("conversation not found")
	}

	return nil
}

// buildContextPrompt builds a context-aware prompt
func (s *GeminiService) buildContextPrompt(prompt, context string, history []GeminiMessage) string {
	var contextPrompt string

	// Add system context if provided
	if context != "" {
		contextPrompt += fmt.Sprintf("Context: %s\n\n", context)
	}

	// Add conversation history
	if len(history) > 0 {
		contextPrompt += "Conversation History:\n"
		for _, msg := range history {
			contextPrompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
		}
		contextPrompt += "\n"
	}

	// Add current prompt
	contextPrompt += fmt.Sprintf("Current Request: %s", prompt)

	return contextPrompt
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("gemini_%d", time.Now().UnixNano())
}

// Health check for Gemini service
func (s *GeminiService) HealthCheck(ctx context.Context) error {
	// Test with a simple prompt
	testReq := &GeminiRequest{
		Prompt: "Hello",
		Model:  s.model,
	}

	_, err := s.GenerateText(ctx, testReq)
	return err
}

// GetAvailableModels returns available Gemini models
func (s *GeminiService) GetAvailableModels() []string {
	return []string{
		"gemini-1.5-flash",
		"gemini-1.5-pro",
		"gemini-1.0-pro",
	}
}

// GetServiceStats returns service statistics
func (s *GeminiService) GetServiceStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get conversation count
	var conversationCount int64
	s.db.Model(&GeminiConversation{}).Count(&conversationCount)
	stats["total_conversations"] = conversationCount

	// Get message count
	var messageCount int64
	s.db.Table("gemini_conversations").Select("json_array_length(messages)").Scan(&messageCount)
	stats["total_messages"] = messageCount

	// Get cache stats if available
	if s.cache != nil {
		cacheStats, _ := s.cache.GetStats(ctx)
		stats["cache_stats"] = cacheStats
	}

	stats["model"] = s.model
	stats["max_tokens"] = s.maxTokens
	stats["temperature"] = s.temperature

	return stats, nil
}
