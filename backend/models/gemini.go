package models

import (
	"time"
)

// GeminiConversation represents a conversation with Gemini AI
type GeminiConversation struct {
	BaseModel
	ID       string  `json:"id" gorm:"type:varchar(255);uniqueIndex;not null"`
	UserID   uint    `json:"user_id" gorm:"not null;index"`
	Title    string  `json:"title" gorm:"type:varchar(255);not null"`
	Messages JSONMap `json:"messages" gorm:"type:jsonb"`
	User     User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// GeminiMessage represents a message in a Gemini conversation
type GeminiMessage struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// GeminiRequest represents a request to Gemini AI
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

// GeminiResponse represents a response from Gemini AI
type GeminiResponse struct {
	ID           string                 `json:"id"`
	Content      string                 `json:"content"`
	Model        string                 `json:"model"`
	Usage        *GeminiUsage           `json:"usage,omitempty"`
	FinishReason string                 `json:"finish_reason,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// GeminiUsage represents token usage information
type GeminiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// TableName returns the table name for GeminiConversation
func (GeminiConversation) TableName() string {
	return "gemini_conversations"
}
