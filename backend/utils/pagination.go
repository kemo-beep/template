package utils

import (
	"math"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// PaginationRequest represents pagination parameters from request
type PaginationRequest struct {
	Page    int                    `json:"page" form:"page" binding:"min=1"`
	Limit   int                    `json:"limit" form:"limit" binding:"min=1,max=100"`
	Sort    string                 `json:"sort" form:"sort"`
	Order   string                 `json:"order" form:"order" binding:"oneof=asc desc"`
	Search  string                 `json:"search" form:"search"`
	Filters map[string]interface{} `json:"filters" form:"filters"`
}

// PaginationResponse represents paginated response metadata
type PaginationResponse struct {
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
	Total       int64  `json:"total"`
	TotalPages  int    `json:"total_pages"`
	HasNext     bool   `json:"has_next"`
	HasPrev     bool   `json:"has_prev"`
	NextPage    *int   `json:"next_page,omitempty"`
	PrevPage    *int   `json:"prev_page,omitempty"`
	FirstPage   int    `json:"first_page"`
	LastPage    int    `json:"last_page"`
	From        int    `json:"from"`
	To          int    `json:"to"`
	Sort        string `json:"sort,omitempty"`
	Order       string `json:"order,omitempty"`
	Search      string `json:"search,omitempty"`
	FilterCount int    `json:"filter_count"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success    bool               `json:"success"`
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
	Message    string             `json:"message"`
	RequestID  string             `json:"request_id,omitempty"`
	TraceID    string             `json:"trace_id,omitempty"`
}

// SortField represents a sortable field
type SortField struct {
	Field     string `json:"field"`
	Direction string `json:"direction"` // "asc" or "desc"
}

// FilterField represents a filterable field
type FilterField struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"` // "eq", "ne", "gt", "gte", "lt", "lte", "like", "in", "nin", "between"
	Value    interface{} `json:"value"`
}

// Default pagination values
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// ParsePaginationRequest parses pagination parameters from request
func ParsePaginationRequest(page, limit int, sort, order, search string, filters map[string]interface{}) PaginationRequest {
	// Set defaults
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 || limit > MaxLimit {
		limit = DefaultLimit
	}
	if order == "" {
		order = "desc"
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return PaginationRequest{
		Page:    page,
		Limit:   limit,
		Sort:    sort,
		Order:   order,
		Search:  search,
		Filters: filters,
	}
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, limit int, total int64) PaginationResponse {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Ensure page is within bounds
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}
	if page < 1 {
		page = 1
	}

	hasNext := page < totalPages
	hasPrev := page > 1

	var nextPage, prevPage *int
	if hasNext {
		next := page + 1
		nextPage = &next
	}
	if hasPrev {
		prev := page - 1
		prevPage = &prev
	}

	firstPage := 1
	lastPage := totalPages
	if lastPage == 0 {
		lastPage = 1
	}

	from := (page-1)*limit + 1
	to := page * limit
	if to > int(total) {
		to = int(total)
	}
	if total == 0 {
		from = 0
		to = 0
	}

	return PaginationResponse{
		Page:        page,
		Limit:       limit,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrev:     hasPrev,
		NextPage:    nextPage,
		PrevPage:    prevPage,
		FirstPage:   firstPage,
		LastPage:    lastPage,
		From:        from,
		To:          to,
		FilterCount: 0, // Will be set by caller
	}
}

// ParseSortString parses sort string (e.g., "name:asc,created_at:desc")
func ParseSortString(sortStr string) []SortField {
	if sortStr == "" {
		return nil
	}

	var sortFields []SortField
	parts := strings.Split(sortStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by colon
		fieldParts := strings.Split(part, ":")
		field := fieldParts[0]
		direction := "asc"

		if len(fieldParts) > 1 {
			direction = strings.ToLower(fieldParts[1])
			if direction != "asc" && direction != "desc" {
				direction = "asc"
			}
		}

		sortFields = append(sortFields, SortField{
			Field:     field,
			Direction: direction,
		})
	}

	return sortFields
}

// ParseFilterString parses filter string (e.g., "name:eq:John,age:gte:18")
func ParseFilterString(filterStr string) []FilterField {
	if filterStr == "" {
		return nil
	}

	var filterFields []FilterField
	parts := strings.Split(filterStr, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by colon
		fieldParts := strings.Split(part, ":")
		if len(fieldParts) < 2 {
			continue
		}

		field := fieldParts[0]
		operator := "eq"
		value := fieldParts[1]

		if len(fieldParts) > 2 {
			operator = fieldParts[1]
			value = fieldParts[2]
		}

		// Convert value to appropriate type
		var convertedValue interface{} = value
		if intVal, err := strconv.Atoi(value); err == nil {
			convertedValue = intVal
		} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			convertedValue = floatVal
		} else if boolVal, err := strconv.ParseBool(value); err == nil {
			convertedValue = boolVal
		}

		filterFields = append(filterFields, FilterField{
			Field:    field,
			Operator: operator,
			Value:    convertedValue,
		})
	}

	return filterFields
}

// ValidateSortFields validates sort fields against allowed fields
func ValidateSortFields(sortFields []SortField, allowedFields []string) []SortField {
	if len(allowedFields) == 0 {
		return sortFields
	}

	var validFields []SortField
	allowedMap := make(map[string]bool)
	for _, field := range allowedFields {
		allowedMap[field] = true
	}

	for _, field := range sortFields {
		if allowedMap[field.Field] {
			validFields = append(validFields, field)
		}
	}

	return validFields
}

// ValidateFilterFields validates filter fields against allowed fields
func ValidateFilterFields(filterFields []FilterField, allowedFields []string) []FilterField {
	if len(allowedFields) == 0 {
		return filterFields
	}

	var validFields []FilterField
	allowedMap := make(map[string]bool)
	for _, field := range allowedFields {
		allowedMap[field] = true
	}

	for _, field := range filterFields {
		if allowedMap[field.Field] {
			validFields = append(validFields, field)
		}
	}

	return validFields
}

// BuildSortClause builds SQL ORDER BY clause from sort fields
func BuildSortClause(sortFields []SortField) string {
	if len(sortFields) == 0 {
		return ""
	}

	var clauses []string
	for _, field := range sortFields {
		clauses = append(clauses, field.Field+" "+strings.ToUpper(field.Direction))
	}

	return strings.Join(clauses, ", ")
}

// BuildWhereClause builds SQL WHERE clause from filter fields
func BuildWhereClause(filterFields []FilterField) (string, []interface{}) {
	if len(filterFields) == 0 {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	for _, field := range filterFields {
		placeholder := "?"
		condition := ""

		switch field.Operator {
		case "eq":
			condition = field.Field + " = " + placeholder
		case "ne":
			condition = field.Field + " != " + placeholder
		case "gt":
			condition = field.Field + " > " + placeholder
		case "gte":
			condition = field.Field + " >= " + placeholder
		case "lt":
			condition = field.Field + " < " + placeholder
		case "lte":
			condition = field.Field + " <= " + placeholder
		case "like":
			condition = field.Field + " LIKE " + placeholder
		case "in":
			// Handle array values
			if arr, ok := field.Value.([]interface{}); ok {
				placeholders := make([]string, len(arr))
				for j := range placeholders {
					placeholders[j] = "?"
					args = append(args, arr[j])
				}
				condition = field.Field + " IN (" + strings.Join(placeholders, ",") + ")"
				conditions = append(conditions, condition)
				continue
			}
			condition = field.Field + " = " + placeholder
		case "nin":
			// Handle array values
			if arr, ok := field.Value.([]interface{}); ok {
				placeholders := make([]string, len(arr))
				for j := range placeholders {
					placeholders[j] = "?"
					args = append(args, arr[j])
				}
				condition = field.Field + " NOT IN (" + strings.Join(placeholders, ",") + ")"
				conditions = append(conditions, condition)
				continue
			}
			condition = field.Field + " != " + placeholder
		case "between":
			// Handle range values
			if arr, ok := field.Value.([]interface{}); ok && len(arr) == 2 {
				condition = field.Field + " BETWEEN " + placeholder + " AND " + placeholder
				args = append(args, arr[0], arr[1])
				conditions = append(conditions, condition)
				continue
			}
			condition = field.Field + " = " + placeholder
		default:
			condition = field.Field + " = " + placeholder
		}

		conditions = append(conditions, condition)
		if field.Operator != "in" && field.Operator != "nin" && field.Operator != "between" {
			args = append(args, field.Value)
		}
	}

	return strings.Join(conditions, " AND "), args
}

// BuildSearchClause builds SQL search clause for text search
func BuildSearchClause(search string, searchFields []string) (string, []interface{}) {
	if search == "" || len(searchFields) == 0 {
		return "", nil
	}

	var conditions []string
	var args []interface{}

	searchPattern := "%" + search + "%"

	for _, field := range searchFields {
		conditions = append(conditions, field+" LIKE ?")
		args = append(args, searchPattern)
	}

	return "(" + strings.Join(conditions, " OR ") + ")", args
}

// SendPaginatedResponse sends a paginated response
func SendPaginatedResponse(c *gin.Context, data interface{}, pagination PaginationResponse, message string) {
	response := PaginatedResponse{
		Success:    true,
		Data:       data,
		Pagination: pagination,
		Message:    message,
		RequestID:  c.GetString("request_id"),
		TraceID:    c.GetString("trace_id"),
	}

	c.JSON(200, response)
}

// SendPaginatedErrorResponse sends a paginated error response
func SendPaginatedErrorResponse(c *gin.Context, err *APIError, pagination PaginationResponse) {
	response := PaginatedResponse{
		Success:    false,
		Data:       nil,
		Pagination: pagination,
		Message:    err.Message,
		RequestID:  c.GetString("request_id"),
		TraceID:    c.GetString("trace_id"),
	}

	c.JSON(err.StatusCode, response)
}
