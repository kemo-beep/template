package controllers

import (
	"net/http"
	"strconv"

	"mobile-backend/models"
	"mobile-backend/services"
	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaymentController struct {
	stripeService *services.StripeService
	polarService  *services.PolarService
	db            *gorm.DB
}

func NewPaymentController(stripeService *services.StripeService, polarService *services.PolarService, db *gorm.DB) *PaymentController {
	return &PaymentController{
		stripeService: stripeService,
		polarService:  polarService,
		db:            db,
	}
}

// Product Management

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product for sale
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body models.Product true "Product data"
// @Success 201 {object} utils.SuccessResponse{data=models.Product}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/products [post]
func (pc *PaymentController) CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid product data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Validate required fields
	if product.Name == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Product name is required", nil)
		return
	}

	if product.Price <= 0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Product price must be greater than 0", nil)
		return
	}

	// Create product using Stripe service
	createdProduct, err := pc.stripeService.CreateProduct(c.Request.Context(), &product)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create product", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, createdProduct, "Product created successfully")
}

// GetProducts godoc
// @Summary Get all products
// @Description Get a list of all products
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param active query bool false "Filter by active status"
// @Success 200 {object} utils.SuccessResponse{data=[]models.Product}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/products [get]
func (pc *PaymentController) GetProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	active := c.Query("active")

	offset := (page - 1) * limit

	var products []models.Product
	query := pc.db.Model(&models.Product{})

	if active != "" {
		query = query.Where("is_active = ?", active == "true")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to count products", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get products", map[string]interface{}{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"products": products,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	utils.SendSuccessResponse(c, response, "Products retrieved successfully")
}

// GetProduct godoc
// @Summary Get a product by ID
// @Description Get a specific product by its ID
// @Tags products
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} utils.SuccessResponse{data=models.Product}
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/products/{id} [get]
func (pc *PaymentController) GetProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid product ID", map[string]interface{}{"error": err.Error()})
		return
	}

	var product models.Product
	if err := pc.db.First(&product, uint(id)).Error; err != nil {
		utils.SendNotFoundResponse(c, "Product not found")
		return
	}

	utils.SendSuccessResponse(c, product, "Product retrieved successfully")
}

// UpdateProduct godoc
// @Summary Update a product
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param product body models.Product true "Updated product data"
// @Success 200 {object} utils.SuccessResponse{data=models.Product}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/products/{id} [put]
func (pc *PaymentController) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid product ID", map[string]interface{}{"error": err.Error()})
		return
	}

	var updates models.Product
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid product data", map[string]interface{}{"error": err.Error()})
		return
	}

	updatedProduct, err := pc.stripeService.UpdateProduct(c.Request.Context(), uint(id), &updates)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update product", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, updatedProduct, "Product updated successfully")
}

// Plan Management

// CreatePlan godoc
// @Summary Create a new plan
// @Description Create a new subscription plan
// @Tags plans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param plan body models.Plan true "Plan data"
// @Success 201 {object} utils.SuccessResponse{data=models.Plan}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/plans [post]
func (pc *PaymentController) CreatePlan(c *gin.Context) {
	var plan models.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid plan data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Validate required fields
	if plan.Name == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Plan name is required", nil)
		return
	}

	if plan.ProductID == 0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Product ID is required", nil)
		return
	}

	if plan.Price <= 0 {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Plan price must be greater than 0", nil)
		return
	}

	if plan.Interval == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Plan interval is required", nil)
		return
	}

	// Create plan using Stripe service
	createdPlan, err := pc.stripeService.CreatePrice(c.Request.Context(), &plan)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create plan", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, createdPlan, "Plan created successfully")
}

// GetPlans godoc
// @Summary Get all plans
// @Description Get a list of all subscription plans
// @Tags plans
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param product_id query int false "Filter by product ID"
// @Success 200 {object} utils.SuccessResponse{data=[]models.Plan}
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/plans [get]
func (pc *PaymentController) GetPlans(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	productID := c.Query("product_id")

	offset := (page - 1) * limit

	var plans []models.Plan
	query := pc.db.Model(&models.Plan{}).Preload("Product")

	if productID != "" {
		query = query.Where("product_id = ?", productID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to count plans", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := query.Offset(offset).Limit(limit).Find(&plans).Error; err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get plans", map[string]interface{}{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"plans": plans,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	}

	utils.SendSuccessResponse(c, response, "Plans retrieved successfully")
}

// Payment Management

// CreatePayment godoc
// @Summary Create a payment
// @Description Create a new payment for a product
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment body CreatePaymentRequest true "Payment data"
// @Success 201 {object} utils.SuccessResponse{data=models.Payment}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/payments [post]
func (pc *PaymentController) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payment data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var payment *models.Payment
	var err error

	switch req.PaymentMethod {
	case "stripe":
		payment, err = pc.stripeService.CreatePaymentIntent(c.Request.Context(), userID.(uint), req.ProductID, req.Amount, req.Currency)
	case "polar":
		payment, err = pc.polarService.CreatePayment(c.Request.Context(), userID.(uint), req.ProductID, req.Amount, req.Currency)
	default:
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payment method", nil)
		return
	}

	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create payment", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, payment, "Payment created successfully")
}

// CreateCheckoutSession godoc
// @Summary Create a checkout session
// @Description Create a Stripe checkout session for a product
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param checkout body CreateCheckoutRequest true "Checkout data"
// @Success 200 {object} utils.SuccessResponse{data=map[string]interface{}}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/payments/checkout [post]
func (pc *PaymentController) CreateCheckoutSession(c *gin.Context) {
	var req CreateCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid checkout data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	session, err := pc.stripeService.CreateCheckoutSession(c.Request.Context(), userID.(uint), req.ProductID, req.SuccessURL, req.CancelURL)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create checkout session", map[string]interface{}{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"session_id": session.ID,
		"url":        session.URL,
	}

	utils.SendSuccessResponse(c, response, "Checkout session created successfully")
}

// GetPayments godoc
// @Summary Get user payments
// @Description Get all payments for the authenticated user
// @Tags payments
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} utils.SuccessResponse{data=[]models.Payment}
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/payments [get]
func (pc *PaymentController) GetPayments(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	payments, err := pc.stripeService.GetUserPayments(c.Request.Context(), userID.(uint))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get payments", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, payments, "Payments retrieved successfully")
}

// Subscription Management

// CreateSubscription godoc
// @Summary Create a subscription
// @Description Create a new subscription for a plan
// @Tags subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param subscription body CreateSubscriptionRequest true "Subscription data"
// @Success 201 {object} utils.SuccessResponse{data=models.Subscription}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/subscriptions [post]
func (pc *PaymentController) CreateSubscription(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid subscription data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var subscription *models.Subscription
	var err error

	switch req.PaymentMethod {
	case "stripe":
		subscription, err = pc.stripeService.CreateSubscription(c.Request.Context(), userID.(uint), req.PlanID, req.PaymentMethodID)
	case "polar":
		subscription, err = pc.polarService.CreateSubscription(c.Request.Context(), userID.(uint), req.PlanID)
	default:
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payment method", nil)
		return
	}

	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create subscription", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, subscription, "Subscription created successfully")
}

// GetSubscriptions godoc
// @Summary Get user subscriptions
// @Description Get all subscriptions for the authenticated user
// @Tags subscriptions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=[]models.Subscription}
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/subscriptions [get]
func (pc *PaymentController) GetSubscriptions(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	subscriptions, err := pc.stripeService.GetUserSubscriptions(c.Request.Context(), userID.(uint))
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get subscriptions", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, subscriptions, "Subscriptions retrieved successfully")
}

// CancelSubscription godoc
// @Summary Cancel a subscription
// @Description Cancel an existing subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Subscription ID"
// @Param cancel body CancelSubscriptionRequest true "Cancel data"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/subscriptions/{id}/cancel [post]
func (pc *PaymentController) CancelSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid subscription ID", map[string]interface{}{"error": err.Error()})
		return
	}

	var req CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid cancel data", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Verify subscription belongs to user
	var subscription models.Subscription
	if err := pc.db.Where("id = ? AND user_id = ?", uint(id), userID).First(&subscription).Error; err != nil {
		utils.SendNotFoundResponse(c, "Subscription not found")
		return
	}

	// Cancel subscription based on payment method
	switch subscription.PaymentMethod {
	case "stripe":
		err = pc.stripeService.CancelSubscription(c.Request.Context(), uint(id), req.Immediately)
	case "polar":
		err = pc.polarService.CancelSubscription(c.Request.Context(), uint(id), req.Immediately)
	default:
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payment method", nil)
		return
	}

	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to cancel subscription", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, nil, "Subscription canceled successfully")
}

// Webhook Handlers

// StripeWebhook godoc
// @Summary Handle Stripe webhook
// @Description Process Stripe webhook events
// @Tags webhooks
// @Accept application/json
// @Produce json
// @Param X-Stripe-Signature header string true "Stripe signature"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/webhooks/stripe [post]
func (pc *PaymentController) StripeWebhook(c *gin.Context) {
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Missing Stripe signature", nil)
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Failed to read request body", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := pc.stripeService.HandleWebhook(c.Request.Context(), body, signature); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to process webhook", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, nil, "Webhook processed successfully")
}

// PolarWebhook godoc
// @Summary Handle Polar webhook
// @Description Process Polar webhook events
// @Tags webhooks
// @Accept application/json
// @Produce json
// @Param X-Polar-Signature header string true "Polar signature"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/webhooks/polar [post]
func (pc *PaymentController) PolarWebhook(c *gin.Context) {
	signature := c.GetHeader("X-Polar-Signature")
	if signature == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Missing Polar signature", nil)
		return
	}

	body, err := c.GetRawData()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Failed to read request body", map[string]interface{}{"error": err.Error()})
		return
	}

	if err := pc.polarService.HandleWebhook(c.Request.Context(), body, signature); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to process webhook", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, nil, "Webhook processed successfully")
}

// Request Types

type CreatePaymentRequest struct {
	ProductID     uint   `json:"product_id" binding:"required"`
	Amount        int64  `json:"amount" binding:"required,min=1"`
	Currency      string `json:"currency" binding:"required,len=3"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=stripe polar"`
}

type CreateCheckoutRequest struct {
	ProductID  uint   `json:"product_id" binding:"required"`
	SuccessURL string `json:"success_url" binding:"required,url"`
	CancelURL  string `json:"cancel_url" binding:"required,url"`
}

type CreateSubscriptionRequest struct {
	PlanID          uint   `json:"plan_id" binding:"required"`
	PaymentMethod   string `json:"payment_method" binding:"required,oneof=stripe polar"`
	PaymentMethodID string `json:"payment_method_id,omitempty"`
}

type CancelSubscriptionRequest struct {
	Immediately bool `json:"immediately"`
}
