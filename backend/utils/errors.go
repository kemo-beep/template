package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorCode represents different types of errors
type ErrorCode string

const (
	// Authentication & Authorization
	ErrorCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrorCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrorCodeInvalidToken       ErrorCode = "INVALID_TOKEN"
	ErrorCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"

	// Validation
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
	ErrorCodeInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrorCodeMissingField     ErrorCode = "MISSING_FIELD"
	ErrorCodeInvalidFormat    ErrorCode = "INVALID_FORMAT"

	// Resource Management
	ErrorCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrorCodeAlreadyExists   ErrorCode = "ALREADY_EXISTS"
	ErrorCodeConflict        ErrorCode = "CONFLICT"
	ErrorCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"

	// Server Errors
	ErrorCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
	ErrorCodeCacheError         ErrorCode = "CACHE_ERROR"

	// Business Logic
	ErrorCodeInsufficientFunds ErrorCode = "INSUFFICIENT_FUNDS"
	ErrorCodeQuotaExceeded     ErrorCode = "QUOTA_EXCEEDED"
	ErrorCodeFeatureDisabled   ErrorCode = "FEATURE_DISABLED"
)

// APIError represents a structured API error
type APIError struct {
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id,omitempty"`
	TraceID    string                 `json:"trace_id,omitempty"`
	StatusCode int                    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Timestamp:  time.Now(),
	}
}

// WithDetails adds details to the error
func (e *APIError) WithDetails(details map[string]interface{}) *APIError {
	e.Details = details
	return e
}

// WithRequestID adds request ID to the error
func (e *APIError) WithRequestID(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// WithTraceID adds trace ID to the error
func (e *APIError) WithTraceID(traceID string) *APIError {
	e.TraceID = traceID
	return e
}

// Predefined errors
var (
	// Authentication errors
	ErrUnauthorized       = NewAPIError(ErrorCodeUnauthorized, "Authentication required", http.StatusUnauthorized)
	ErrForbidden          = NewAPIError(ErrorCodeForbidden, "Access denied", http.StatusForbidden)
	ErrTokenExpired       = NewAPIError(ErrorCodeTokenExpired, "Token has expired", http.StatusUnauthorized)
	ErrInvalidToken       = NewAPIError(ErrorCodeInvalidToken, "Invalid token", http.StatusUnauthorized)
	ErrInvalidCredentials = NewAPIError(ErrorCodeInvalidCredentials, "Invalid credentials", http.StatusUnauthorized)

	// Validation errors
	ErrValidationFailed = NewAPIError(ErrorCodeValidationFailed, "Validation failed", http.StatusBadRequest)
	ErrInvalidInput     = NewAPIError(ErrorCodeInvalidInput, "Invalid input", http.StatusBadRequest)
	ErrMissingField     = NewAPIError(ErrorCodeMissingField, "Required field is missing", http.StatusBadRequest)
	ErrInvalidFormat    = NewAPIError(ErrorCodeInvalidFormat, "Invalid format", http.StatusBadRequest)

	// Resource errors
	ErrNotFound        = NewAPIError(ErrorCodeNotFound, "Resource not found", http.StatusNotFound)
	ErrAlreadyExists   = NewAPIError(ErrorCodeAlreadyExists, "Resource already exists", http.StatusConflict)
	ErrConflict        = NewAPIError(ErrorCodeConflict, "Resource conflict", http.StatusConflict)
	ErrTooManyRequests = NewAPIError(ErrorCodeTooManyRequests, "Too many requests", http.StatusTooManyRequests)

	// Server errors
	ErrInternalError      = NewAPIError(ErrorCodeInternalError, "Internal server error", http.StatusInternalServerError)
	ErrServiceUnavailable = NewAPIError(ErrorCodeServiceUnavailable, "Service unavailable", http.StatusServiceUnavailable)
	ErrDatabaseError      = NewAPIError(ErrorCodeDatabaseError, "Database error", http.StatusInternalServerError)
	ErrCacheError         = NewAPIError(ErrorCodeCacheError, "Cache error", http.StatusInternalServerError)

	// Business logic errors
	ErrInsufficientFunds = NewAPIError(ErrorCodeInsufficientFunds, "Insufficient funds", http.StatusBadRequest)
	ErrQuotaExceeded     = NewAPIError(ErrorCodeQuotaExceeded, "Quota exceeded", http.StatusTooManyRequests)
	ErrFeatureDisabled   = NewAPIError(ErrorCodeFeatureDisabled, "Feature is disabled", http.StatusForbidden)
)

// APIErrorResponse represents the error response structure
type APIErrorResponse struct {
	Success   bool      `json:"success"`
	Error     *APIError `json:"error"`
	RequestID string    `json:"request_id,omitempty"`
	TraceID   string    `json:"trace_id,omitempty"`
}

// NewAPIErrorResponse creates a new error response
func NewAPIErrorResponse(err *APIError, requestID, traceID string) APIErrorResponse {
	return APIErrorResponse{
		Success:   false,
		Error:     err,
		RequestID: requestID,
		TraceID:   traceID,
	}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// APIValidationErrorResponse represents validation error response
type APIValidationErrorResponse struct {
	Success   bool              `json:"success"`
	Error     *APIError         `json:"error"`
	Details   []ValidationError `json:"details"`
	RequestID string            `json:"request_id,omitempty"`
	TraceID   string            `json:"trace_id,omitempty"`
}

// NewAPIValidationErrorResponse creates a new validation error response
func NewAPIValidationErrorResponse(errors []ValidationError, requestID, traceID string) APIValidationErrorResponse {
	return APIValidationErrorResponse{
		Success: false,
		Error: ErrValidationFailed.WithDetails(map[string]interface{}{
			"validation_errors": errors,
		}),
		Details:   errors,
		RequestID: requestID,
		TraceID:   traceID,
	}
}

// ErrorHandler handles panics and converts them to API errors
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, NewAPIErrorResponse(
				ErrInternalError.WithDetails(map[string]interface{}{
					"panic": err,
				}),
				"", "",
			))
		} else {
			c.JSON(http.StatusInternalServerError, NewAPIErrorResponse(
				ErrInternalError.WithDetails(map[string]interface{}{
					"panic": fmt.Sprintf("%v", recovered),
				}),
				"", "",
			))
		}
		c.Abort()
	})
}

// HandleError handles errors and returns appropriate response
func HandleError(c *gin.Context, err error) {
	var apiErr *APIError

	switch e := err.(type) {
	case *APIError:
		apiErr = e
	case error:
		apiErr = ErrInternalError.WithDetails(map[string]interface{}{
			"original_error": e.Error(),
		})
	default:
		apiErr = ErrInternalError
	}

	// Add request context
	if requestID := c.GetString("request_id"); requestID != "" {
		apiErr.RequestID = requestID
	}
	if traceID := c.GetString("trace_id"); traceID != "" {
		apiErr.TraceID = traceID
	}

	c.JSON(apiErr.StatusCode, NewAPIErrorResponse(apiErr, apiErr.RequestID, apiErr.TraceID))
	c.Abort()
}

// HandleValidationError handles validation errors
func HandleValidationError(c *gin.Context, errors []ValidationError) {
	requestID := c.GetString("request_id")
	traceID := c.GetString("trace_id")

	c.JSON(http.StatusBadRequest, NewAPIValidationErrorResponse(errors, requestID, traceID))
	c.Abort()
}
