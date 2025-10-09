package utils

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	TraceID   string      `json:"trace_id,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     *APIError `json:"error"`
	RequestID string    `json:"request_id,omitempty"`
	TraceID   string    `json:"trace_id,omitempty"`
}

// ValidationErrorResponse represents a validation error response
type ValidationErrorResponse struct {
	Success   bool              `json:"success"`
	Error     *APIError         `json:"error"`
	Details   []ValidationError `json:"details"`
	RequestID string            `json:"request_id,omitempty"`
	TraceID   string            `json:"trace_id,omitempty"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(data interface{}, message string) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// SendSuccessResponse sends a success response
func SendSuccessResponse(c *gin.Context, data interface{}, message string) {
	requestID := c.GetString("request_id")
	traceID := c.GetString("trace_id")

	response := SuccessResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: requestID,
		TraceID:   traceID,
	}

	c.JSON(http.StatusOK, response)
}

// SendCreatedResponse sends a created response
func SendCreatedResponse(c *gin.Context, data interface{}, message string) {
	requestID := c.GetString("request_id")
	traceID := c.GetString("trace_id")

	response := SuccessResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: requestID,
		TraceID:   traceID,
	}

	c.JSON(http.StatusCreated, response)
}

// FormatTime formats time for API responses
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// ParseInt parses string to int with default value
func ParseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}
	return defaultValue
}

// ParseBool parses string to bool with default value
func ParseBool(s string, defaultValue bool) bool {
	if s == "" {
		return defaultValue
	}
	if val, err := strconv.ParseBool(s); err == nil {
		return val
	}
	return defaultValue
}

// ParseFloat64 parses string to float64 with default value
func ParseFloat64(s string, defaultValue float64) float64 {
	if s == "" {
		return defaultValue
	}
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	return defaultValue
}

// GetPaginationFromQuery extracts pagination parameters from query
func GetPaginationFromQuery(c *gin.Context) PaginationRequest {
	page := ParseInt(c.DefaultQuery("page", "1"), 1)
	limit := ParseInt(c.DefaultQuery("limit", "20"), 20)
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")
	search := c.Query("search")

	// Extract filters from query parameters
	filters := make(map[string]interface{})
	for key, values := range c.Request.URL.Query() {
		if key != "page" && key != "limit" && key != "sort" && key != "order" && key != "search" {
			if len(values) > 0 {
				filters[key] = values[0]
			}
		}
	}

	return ParsePaginationRequest(page, limit, sort, order, search, filters)
}

// SendErrorResponse sends an error response
func SendErrorResponse(c *gin.Context, statusCode int, message string, details map[string]interface{}) {
	requestID := c.GetString("request_id")
	traceID := c.GetString("trace_id")

	response := map[string]interface{}{
		"success":    false,
		"message":    message,
		"request_id": requestID,
		"trace_id":   traceID,
	}

	if details != nil {
		response["details"] = details
	}

	c.JSON(statusCode, response)
}

// SendValidationErrorResponse sends a validation error response
func SendValidationErrorResponse(c *gin.Context, errors []ValidationError) {
	requestID := c.GetString("request_id")
	traceID := c.GetString("trace_id")

	response := map[string]interface{}{
		"success":    false,
		"message":    "Validation failed",
		"errors":     errors,
		"request_id": requestID,
		"trace_id":   traceID,
	}

	c.JSON(http.StatusBadRequest, response)
}

// SendUnauthorizedResponse sends an unauthorized response
func SendUnauthorizedResponse(c *gin.Context, message string) {
	requestID := c.GetString("request_id")
	traceID := c.GetString("trace_id")

	response := map[string]interface{}{
		"success":    false,
		"message":    message,
		"request_id": requestID,
		"trace_id":   traceID,
	}

	c.JSON(http.StatusUnauthorized, response)
}

// SendNotFoundResponse sends a not found response
func SendNotFoundResponse(c *gin.Context, message string) {
	requestID := c.GetString("request_id")
	traceID := c.GetString("trace_id")

	response := map[string]interface{}{
		"success":    false,
		"message":    message,
		"request_id": requestID,
		"trace_id":   traceID,
	}

	c.JSON(http.StatusNotFound, response)
}

// SendInternalServerErrorResponse sends an internal server error response
func SendInternalServerErrorResponse(c *gin.Context, message string) {
	requestID := c.GetString("request_id")
	traceID := c.GetString("trace_id")

	response := map[string]interface{}{
		"success":    false,
		"message":    message,
		"request_id": requestID,
		"trace_id":   traceID,
	}

	c.JSON(http.StatusInternalServerError, response)
}
