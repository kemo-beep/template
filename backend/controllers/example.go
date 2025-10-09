package controllers

import (
	"strconv"
	"time"

	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
)

type ExampleController struct {
	// Add any dependencies here
}

func NewExampleController() *ExampleController {
	return &ExampleController{}
}

// ExampleRequest represents a request with comprehensive validation
type ExampleRequest struct {
	Name       string                 `json:"name" binding:"required,min=2,max=50" validate:"username"`
	Email      string                 `json:"email" binding:"required,email" validate:"email"`
	Phone      string                 `json:"phone" binding:"required" validate:"phone"`
	Age        int                    `json:"age" binding:"required,min=18,max=120"`
	BirthDate  string                 `json:"birth_date" binding:"required" validate:"date"`
	Website    string                 `json:"website" binding:"required" validate:"url"`
	IPAddress  string                 `json:"ip_address" binding:"required" validate:"ip"`
	CreditCard string                 `json:"credit_card" binding:"required" validate:"creditcard"`
	SSN        string                 `json:"ssn" binding:"required" validate:"ssn"`
	PostalCode string                 `json:"postal_code" binding:"required" validate:"postalcode"`
	Currency   string                 `json:"currency" binding:"required" validate:"currency"`
	Latitude   float64                `json:"latitude" binding:"required" validate:"latitude"`
	Longitude  float64                `json:"longitude" binding:"required" validate:"longitude"`
	HexColor   string                 `json:"hex_color" binding:"required" validate:"hexcolor"`
	ISBN       string                 `json:"isbn" binding:"required" validate:"isbn"`
	IsActive   bool                   `json:"is_active"`
	Tags       []string               `json:"tags" binding:"required,min=1"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ExampleResponse represents a response with pagination
type ExampleResponse struct {
	ID         uint                   `json:"id"`
	Name       string                 `json:"name"`
	Email      string                 `json:"email"`
	Phone      string                 `json:"phone"`
	Age        int                    `json:"age"`
	BirthDate  string                 `json:"birth_date"`
	Website    string                 `json:"website"`
	IPAddress  string                 `json:"ip_address"`
	CreditCard string                 `json:"credit_card"`
	SSN        string                 `json:"ssn"`
	PostalCode string                 `json:"postal_code"`
	Currency   string                 `json:"currency"`
	Latitude   float64                `json:"latitude"`
	Longitude  float64                `json:"longitude"`
	HexColor   string                 `json:"hex_color"`
	ISBN       string                 `json:"isbn"`
	IsActive   bool                   `json:"is_active"`
	Tags       []string               `json:"tags"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  string                 `json:"created_at"`
	UpdatedAt  string                 `json:"updated_at"`
}

// CreateExample godoc
// @Summary Create a new example
// @Description Create a new example with comprehensive validation
// @Tags example
// @Accept json
// @Produce json
// @Param request body ExampleRequest true "Example data"
// @Success 201 {object} utils.SuccessResponse{data=ExampleResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ValidationErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/examples [post]
func (ec *ExampleController) CreateExample(c *gin.Context) {
	var req ExampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Parse validation errors
		validationErrors := utils.ValidateStruct(req)
		if len(validationErrors) > 0 {
			utils.HandleValidationError(c, validationErrors)
			return
		}

		// Handle binding errors
		utils.HandleError(c, utils.ErrValidationFailed.WithDetails(map[string]interface{}{
			"binding_error": err.Error(),
		}))
		return
	}

	// Additional custom validation
	if req.Age < 18 {
		utils.HandleValidationError(c, []utils.ValidationError{
			{
				Field:   "age",
				Message: "Age must be at least 18",
				Value:   req.Age,
			},
		})
		return
	}

	// Simulate creating the example
	example := ExampleResponse{
		ID:         1,
		Name:       req.Name,
		Email:      req.Email,
		Phone:      req.Phone,
		Age:        req.Age,
		BirthDate:  req.BirthDate,
		Website:    req.Website,
		IPAddress:  req.IPAddress,
		CreditCard: req.CreditCard,
		SSN:        req.SSN,
		PostalCode: req.PostalCode,
		Currency:   req.Currency,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		HexColor:   req.HexColor,
		ISBN:       req.ISBN,
		IsActive:   req.IsActive,
		Tags:       req.Tags,
		Metadata:   req.Metadata,
		CreatedAt:  utils.FormatTime(time.Now()),
		UpdatedAt:  utils.FormatTime(time.Now()),
	}

	utils.SendSuccessResponse(c, example, "Example created successfully")
}

// GetExamples godoc
// @Summary Get examples with pagination
// @Description Get a paginated list of examples with filtering and sorting
// @Tags example
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param sort query string false "Sort field" default(created_at)
// @Param order query string false "Sort order" Enums(asc, desc) default(desc)
// @Param search query string false "Search term"
// @Param name query string false "Filter by name"
// @Param email query string false "Filter by email"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} utils.PaginatedResponse{data=[]ExampleResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/examples [get]
func (ec *ExampleController) GetExamples(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sort := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")
	search := c.Query("search")

	// Parse filters
	filters := make(map[string]interface{})
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}
	if email := c.Query("email"); email != "" {
		filters["email"] = email
	}
	if isActive := c.Query("is_active"); isActive != "" {
		if active, err := strconv.ParseBool(isActive); err == nil {
			filters["is_active"] = active
		}
	}

	// Create pagination request
	paginationReq := utils.ParsePaginationRequest(page, limit, sort, order, search, filters)

	// Simulate database query with pagination
	// In a real implementation, you would query the database here
	total := int64(100) // Simulate total count
	examples := []ExampleResponse{
		{
			ID:         1,
			Name:       "John Doe",
			Email:      "john@example.com",
			Phone:      "+1234567890",
			Age:        30,
			BirthDate:  "1993-01-01",
			Website:    "https://johndoe.com",
			IPAddress:  "192.168.1.1",
			CreditCard: "4111111111111111",
			SSN:        "123-45-6789",
			PostalCode: "12345",
			Currency:   "USD",
			Latitude:   40.7128,
			Longitude:  -74.0060,
			HexColor:   "#FF0000",
			ISBN:       "978-0-13-468599-1",
			IsActive:   true,
			Tags:       []string{"developer", "golang"},
			Metadata:   map[string]interface{}{"role": "admin"},
			CreatedAt:  utils.FormatTime(time.Now().Add(-24 * time.Hour)),
			UpdatedAt:  utils.FormatTime(time.Now().Add(-1 * time.Hour)),
		},
		{
			ID:         2,
			Name:       "Jane Smith",
			Email:      "jane@example.com",
			Phone:      "+1987654321",
			Age:        25,
			BirthDate:  "1998-05-15",
			Website:    "https://janesmith.com",
			IPAddress:  "192.168.1.2",
			CreditCard: "5555555555554444",
			SSN:        "987-65-4321",
			PostalCode: "54321",
			Currency:   "EUR",
			Latitude:   51.5074,
			Longitude:  -0.1278,
			HexColor:   "#00FF00",
			ISBN:       "978-0-13-468599-2",
			IsActive:   false,
			Tags:       []string{"designer", "ui"},
			Metadata:   map[string]interface{}{"role": "user"},
			CreatedAt:  utils.FormatTime(time.Now().Add(-48 * time.Hour)),
			UpdatedAt:  utils.FormatTime(time.Now().Add(-2 * time.Hour)),
		},
	}

	// Calculate pagination metadata
	pagination := utils.CalculatePagination(paginationReq.Page, paginationReq.Limit, total)
	pagination.Sort = sort
	pagination.Order = order
	pagination.Search = search
	pagination.FilterCount = len(filters)

	// Send paginated response
	utils.SendPaginatedResponse(c, examples, pagination, "Examples retrieved successfully")
}

// GetExample godoc
// @Summary Get a specific example
// @Description Get a specific example by ID
// @Tags example
// @Accept json
// @Produce json
// @Param id path int true "Example ID"
// @Success 200 {object} utils.SuccessResponse{data=ExampleResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/examples/{id} [get]
func (ec *ExampleController) GetExample(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.HandleError(c, utils.ErrInvalidInput.WithDetails(map[string]interface{}{
			"field": "id",
			"value": idStr,
		}))
		return
	}

	// Simulate database query
	// In a real implementation, you would query the database here
	if id == 0 {
		utils.HandleError(c, utils.ErrNotFound.WithDetails(map[string]interface{}{
			"resource": "example",
			"id":       id,
		}))
		return
	}

	example := ExampleResponse{
		ID:         uint(id),
		Name:       "John Doe",
		Email:      "john@example.com",
		Phone:      "+1234567890",
		Age:        30,
		BirthDate:  "1993-01-01",
		Website:    "https://johndoe.com",
		IPAddress:  "192.168.1.1",
		CreditCard: "4111111111111111",
		SSN:        "123-45-6789",
		PostalCode: "12345",
		Currency:   "USD",
		Latitude:   40.7128,
		Longitude:  -74.0060,
		HexColor:   "#FF0000",
		ISBN:       "978-0-13-468599-1",
		IsActive:   true,
		Tags:       []string{"developer", "golang"},
		Metadata:   map[string]interface{}{"role": "admin"},
		CreatedAt:  utils.FormatTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:  utils.FormatTime(time.Now().Add(-1 * time.Hour)),
	}

	utils.SendSuccessResponse(c, example, "Example retrieved successfully")
}

// UpdateExample godoc
// @Summary Update an example
// @Description Update an existing example
// @Tags example
// @Accept json
// @Produce json
// @Param id path int true "Example ID"
// @Param request body ExampleRequest true "Updated example data"
// @Success 200 {object} utils.SuccessResponse{data=ExampleResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 422 {object} utils.ValidationErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/examples/{id} [put]
func (ec *ExampleController) UpdateExample(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.HandleError(c, utils.ErrInvalidInput.WithDetails(map[string]interface{}{
			"field": "id",
			"value": idStr,
		}))
		return
	}

	var req ExampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := utils.ValidateStruct(req)
		if len(validationErrors) > 0 {
			utils.HandleValidationError(c, validationErrors)
			return
		}

		utils.HandleError(c, utils.ErrValidationFailed.WithDetails(map[string]interface{}{
			"binding_error": err.Error(),
		}))
		return
	}

	// Simulate database update
	example := ExampleResponse{
		ID:         uint(id),
		Name:       req.Name,
		Email:      req.Email,
		Phone:      req.Phone,
		Age:        req.Age,
		BirthDate:  req.BirthDate,
		Website:    req.Website,
		IPAddress:  req.IPAddress,
		CreditCard: req.CreditCard,
		SSN:        req.SSN,
		PostalCode: req.PostalCode,
		Currency:   req.Currency,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		HexColor:   req.HexColor,
		ISBN:       req.ISBN,
		IsActive:   req.IsActive,
		Tags:       req.Tags,
		Metadata:   req.Metadata,
		CreatedAt:  utils.FormatTime(time.Now().Add(-24 * time.Hour)),
		UpdatedAt:  utils.FormatTime(time.Now()),
	}

	utils.SendSuccessResponse(c, example, "Example updated successfully")
}

// DeleteExample godoc
// @Summary Delete an example
// @Description Delete an existing example
// @Tags example
// @Accept json
// @Produce json
// @Param id path int true "Example ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/examples/{id} [delete]
func (ec *ExampleController) DeleteExample(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.HandleError(c, utils.ErrInvalidInput.WithDetails(map[string]interface{}{
			"field": "id",
			"value": idStr,
		}))
		return
	}

	// Simulate database deletion
	// In a real implementation, you would delete from the database here
	if id == 0 {
		utils.HandleError(c, utils.ErrNotFound.WithDetails(map[string]interface{}{
			"resource": "example",
			"id":       id,
		}))
		return
	}

	utils.SendSuccessResponse(c, nil, "Example deleted successfully")
}

// GetExampleStats godoc
// @Summary Get example statistics
// @Description Get statistics about examples
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {object} utils.SuccessResponse{data=map[string]interface{}}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/examples/stats [get]
func (ec *ExampleController) GetExampleStats(c *gin.Context) {
	// Simulate statistics calculation
	stats := map[string]interface{}{
		"total_examples":    100,
		"active_examples":   75,
		"inactive_examples": 25,
		"average_age":       32.5,
		"countries": map[string]int{
			"US": 45,
			"UK": 20,
			"CA": 15,
			"AU": 10,
			"DE": 10,
		},
		"top_tags": []map[string]interface{}{
			{"tag": "developer", "count": 30},
			{"tag": "designer", "count": 25},
			{"tag": "manager", "count": 20},
			{"tag": "analyst", "count": 15},
			{"tag": "tester", "count": 10},
		},
		"created_last_30_days": 15,
		"updated_last_7_days":  8,
	}

	utils.SendSuccessResponse(c, stats, "Example statistics retrieved successfully")
}
