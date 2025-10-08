package utils

import (
	"time"

	"github.com/gin-gonic/gin"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Success bool              `json:"success"`
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"` // seconds
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents a refresh token response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(data interface{}, message ...string) SuccessResponse {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return SuccessResponse{
		Success: true,
		Message: msg,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err string, code ...int) ErrorResponse {
	c := 500
	if len(code) > 0 {
		c = code[0]
	}
	return ErrorResponse{
		Success: false,
		Error:   err,
		Code:    c,
	}
}

// NewValidationErrorResponse creates a new validation error response
func NewValidationErrorResponse(err string, details map[string]string) ValidationErrorResponse {
	return ValidationErrorResponse{
		Success: false,
		Error:   err,
		Details: details,
	}
}

// FormatTime formats a time.Time to string
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// Response helper functions for Gin context
func SendSuccessResponse(c *gin.Context, data interface{}, message ...string) {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(200, NewSuccessResponse(data, msg))
}

func SendCreatedResponse(c *gin.Context, data interface{}, message ...string) {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(201, NewSuccessResponse(data, msg))
}

func SendErrorResponse(c *gin.Context, err string, code ...int) {
	statusCode := 500
	if len(code) > 0 {
		statusCode = code[0]
	}
	c.JSON(statusCode, NewErrorResponse(err, statusCode))
}

func SendValidationErrorResponse(c *gin.Context, details map[string]string) {
	c.JSON(400, NewValidationErrorResponse("Validation failed", details))
}

func SendUnauthorizedResponse(c *gin.Context, message ...string) {
	msg := "Unauthorized"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(401, NewErrorResponse(msg, 401))
}

func SendNotFoundResponse(c *gin.Context, message ...string) {
	msg := "Not found"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(404, NewErrorResponse(msg, 404))
}

func SendInternalServerErrorResponse(c *gin.Context, message ...string) {
	msg := "Internal server error"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(500, NewErrorResponse(msg, 500))
}

// GenerateRandomString generates a random string of specified length
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
