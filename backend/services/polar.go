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

	"mobile-backend/models"

	"gorm.io/gorm"
)

type PolarService struct {
	db                        *gorm.DB
	cache                     *CacheService
	apiKey                    string
	baseURL                   string
	webhookSecret             string
	subscriptionStatusService *SubscriptionStatusService
}

func NewPolarService(db *gorm.DB, cache *CacheService, subscriptionStatusService *SubscriptionStatusService) *PolarService {
	return &PolarService{
		db:                        db,
		cache:                     cache,
		apiKey:                    os.Getenv("POLAR_API_KEY"),
		baseURL:                   os.Getenv("POLAR_BASE_URL"),
		webhookSecret:             os.Getenv("POLAR_WEBHOOK_SECRET"),
		subscriptionStatusService: subscriptionStatusService,
	}
}

// Polar API Types
type PolarProduct struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Active      bool                   `json:"active"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

type PolarPrice struct {
	ID        string                 `json:"id"`
	ProductID string                 `json:"product_id"`
	Amount    int64                  `json:"amount"`
	Currency  string                 `json:"currency"`
	Recurring *PolarRecurring        `json:"recurring"`
	Active    bool                   `json:"active"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type PolarRecurring struct {
	Interval      string `json:"interval"`
	IntervalCount int    `json:"interval_count"`
}

type PolarCustomer struct {
	ID        string                 `json:"id"`
	Email     string                 `json:"email"`
	Name      string                 `json:"name"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

type PolarSubscription struct {
	ID                 string                 `json:"id"`
	CustomerID         string                 `json:"customer_id"`
	ProductID          string                 `json:"product_id"`
	PriceID            string                 `json:"price_id"`
	Status             string                 `json:"status"`
	CurrentPeriodStart string                 `json:"current_period_start"`
	CurrentPeriodEnd   string                 `json:"current_period_end"`
	CancelAtPeriodEnd  bool                   `json:"cancel_at_period_end"`
	CanceledAt         *string                `json:"canceled_at"`
	TrialStart         *string                `json:"trial_start"`
	TrialEnd           *string                `json:"trial_end"`
	Quantity           int                    `json:"quantity"`
	Metadata           map[string]interface{} `json:"metadata"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
}

type PolarPayment struct {
	ID          string                 `json:"id"`
	CustomerID  string                 `json:"customer_id"`
	Amount      int64                  `json:"amount"`
	Currency    string                 `json:"currency"`
	Status      string                 `json:"status"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

type PolarWebhookEvent struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt string                 `json:"created_at"`
}

// Product Management

// CreateProduct creates a product in Polar and our database
func (p *PolarService) CreateProduct(ctx context.Context, productData *models.Product) (*models.Product, error) {
	// Create product in Polar
	polarProduct := PolarProduct{
		Name:        productData.Name,
		Description: productData.Description,
		Active:      productData.IsActive,
		Metadata:    p.convertMetadataToPolar(productData.Metadata),
	}

	createdProduct, err := p.createPolarProduct(ctx, polarProduct)
	if err != nil {
		return nil, fmt.Errorf("failed to create Polar product: %w", err)
	}

	// Update our product with Polar ID
	productData.PolarProductID = createdProduct.ID

	// Save to database
	if err := p.db.Create(productData).Error; err != nil {
		// Clean up Polar product if database save fails
		p.deletePolarProduct(ctx, createdProduct.ID)
		return nil, fmt.Errorf("failed to save product to database: %w", err)
	}

	return productData, nil
}

// UpdateProduct updates a product in both Polar and our database
func (p *PolarService) UpdateProduct(ctx context.Context, productID uint, updates *models.Product) (*models.Product, error) {
	var existingProduct models.Product
	if err := p.db.First(&existingProduct, productID).Error; err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Update in Polar
	if existingProduct.PolarProductID != "" {
		polarProduct := PolarProduct{
			Name:        updates.Name,
			Description: updates.Description,
			Active:      updates.IsActive,
			Metadata:    p.convertMetadataToPolar(updates.Metadata),
		}

		_, err := p.updatePolarProduct(ctx, existingProduct.PolarProductID, polarProduct)
		if err != nil {
			return nil, fmt.Errorf("failed to update Polar product: %w", err)
		}
	}

	// Update in database
	if err := p.db.Model(&existingProduct).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update product in database: %w", err)
	}

	return &existingProduct, nil
}

// CreatePrice creates a price for a product in Polar
func (p *PolarService) CreatePrice(ctx context.Context, planData *models.Plan) (*models.Plan, error) {
	var product models.Product
	if err := p.db.First(&product, planData.ProductID).Error; err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	if product.PolarProductID == "" {
		return nil, fmt.Errorf("product does not have Polar ID")
	}

	// Create price in Polar
	polarPrice := PolarPrice{
		ProductID: product.PolarProductID,
		Amount:    planData.Price,
		Currency:  planData.Currency,
		Recurring: &PolarRecurring{
			Interval:      planData.Interval,
			IntervalCount: planData.IntervalCount,
		},
		Active:   planData.IsActive,
		Metadata: p.convertMetadataToPolar(planData.Metadata),
	}

	createdPrice, err := p.createPolarPrice(ctx, polarPrice)
	if err != nil {
		return nil, fmt.Errorf("failed to create Polar price: %w", err)
	}

	// Update our plan with Polar ID
	planData.PolarPlanID = createdPrice.ID

	// Save to database
	if err := p.db.Create(planData).Error; err != nil {
		// Clean up Polar price if database save fails
		p.deletePolarPrice(ctx, createdPrice.ID)
		return nil, fmt.Errorf("failed to save plan to database: %w", err)
	}

	return planData, nil
}

// Customer Management

// CreateCustomer creates a customer in Polar
func (p *PolarService) CreateCustomer(ctx context.Context, user *models.User) (string, error) {
	// Check if customer already exists
	cacheKey := fmt.Sprintf("polar_customer:%d", user.ID)
	var customerID string

	if err := p.cache.Get(ctx, cacheKey, &customerID); err == nil {
		return customerID, nil
	}

	// Create customer in Polar
	polarCustomer := PolarCustomer{
		Email: user.Email,
		Name:  user.Name,
		Metadata: map[string]interface{}{
			"user_id": strconv.FormatUint(uint64(user.ID), 10),
		},
	}

	createdCustomer, err := p.createPolarCustomer(ctx, polarCustomer)
	if err != nil {
		return "", fmt.Errorf("failed to create Polar customer: %w", err)
	}

	// Cache the customer ID
	p.cache.Set(ctx, cacheKey, createdCustomer.ID, 24*time.Hour)

	return createdCustomer.ID, nil
}

// Payment Management

// CreatePayment creates a payment in Polar
func (p *PolarService) CreatePayment(ctx context.Context, userID uint, productID uint, amount int64, currency string) (*models.Payment, error) {
	var user models.User
	if err := p.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var product models.Product
	if err := p.db.First(&product, productID).Error; err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Create Polar customer
	customerID, err := p.CreateCustomer(ctx, &user)
	if err != nil {
		return nil, err
	}

	// Create payment in Polar
	polarPayment := PolarPayment{
		CustomerID:  customerID,
		Amount:      amount,
		Currency:    currency,
		Status:      "pending",
		Description: fmt.Sprintf("Payment for %s", product.Name),
		Metadata: map[string]interface{}{
			"user_id":    strconv.FormatUint(uint64(userID), 10),
			"product_id": strconv.FormatUint(uint64(productID), 10),
		},
	}

	createdPayment, err := p.createPolarPayment(ctx, polarPayment)
	if err != nil {
		return nil, fmt.Errorf("failed to create Polar payment: %w", err)
	}

	// Create payment record in database
	payment := &models.Payment{
		UserID:         userID,
		ProductID:      productID,
		Amount:         amount,
		Currency:       currency,
		Status:         "pending",
		PaymentMethod:  "polar",
		PolarPaymentID: createdPayment.ID,
		Description:    fmt.Sprintf("Payment for %s", product.Name),
	}

	if err := p.db.Create(payment).Error; err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	return payment, nil
}

// Subscription Management

// CreateSubscription creates a subscription in Polar
func (p *PolarService) CreateSubscription(ctx context.Context, userID uint, planID uint) (*models.Subscription, error) {
	var user models.User
	if err := p.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var plan models.Plan
	if err := p.db.Preload("Product").First(&plan, planID).Error; err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// Create Polar customer
	customerID, err := p.CreateCustomer(ctx, &user)
	if err != nil {
		return nil, err
	}

	// Create subscription in Polar
	polarSubscription := PolarSubscription{
		CustomerID:        customerID,
		ProductID:         plan.Product.PolarProductID,
		PriceID:           plan.PolarPlanID,
		Status:            "active",
		Quantity:          1,
		CancelAtPeriodEnd: false,
		Metadata: map[string]interface{}{
			"user_id": strconv.FormatUint(uint64(userID), 10),
			"plan_id": strconv.FormatUint(uint64(planID), 10),
		},
	}

	createdSubscription, err := p.createPolarSubscription(ctx, polarSubscription)
	if err != nil {
		return nil, fmt.Errorf("failed to create Polar subscription: %w", err)
	}

	// Parse dates
	currentPeriodStart, _ := time.Parse(time.RFC3339, createdSubscription.CurrentPeriodStart)
	currentPeriodEnd, _ := time.Parse(time.RFC3339, createdSubscription.CurrentPeriodEnd)

	// Create subscription record in database
	subscription := &models.Subscription{
		UserID:              userID,
		ProductID:           plan.ProductID,
		PlanID:              &planID,
		Status:              createdSubscription.Status,
		CurrentPeriodStart:  currentPeriodStart,
		CurrentPeriodEnd:    currentPeriodEnd,
		CancelAtPeriodEnd:   createdSubscription.CancelAtPeriodEnd,
		PolarSubscriptionID: createdSubscription.ID,
		Quantity:            1,
	}

	// Handle trial period
	if createdSubscription.TrialStart != nil && createdSubscription.TrialEnd != nil {
		trialStart, _ := time.Parse(time.RFC3339, *createdSubscription.TrialStart)
		trialEnd, _ := time.Parse(time.RFC3339, *createdSubscription.TrialEnd)
		subscription.TrialStart = &trialStart
		subscription.TrialEnd = &trialEnd
	}

	// Handle cancellation
	if createdSubscription.CanceledAt != nil {
		canceledAt, _ := time.Parse(time.RFC3339, *createdSubscription.CanceledAt)
		subscription.CanceledAt = &canceledAt
	}

	if err := p.db.Create(subscription).Error; err != nil {
		// Clean up Polar subscription if database save fails
		p.cancelPolarSubscription(ctx, createdSubscription.ID)
		return nil, fmt.Errorf("failed to create subscription record: %w", err)
	}

	return subscription, nil
}

// CancelSubscription cancels a subscription
func (p *PolarService) CancelSubscription(ctx context.Context, subscriptionID uint, immediately bool) error {
	var subscription models.Subscription
	if err := p.db.First(&subscription, subscriptionID).Error; err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	if subscription.PolarSubscriptionID == "" {
		return fmt.Errorf("subscription does not have Polar ID")
	}

	// Cancel in Polar
	if immediately {
		err := p.cancelPolarSubscription(ctx, subscription.PolarSubscriptionID)
		if err != nil {
			return fmt.Errorf("failed to cancel Polar subscription: %w", err)
		}
	} else {
		err := p.updatePolarSubscription(ctx, subscription.PolarSubscriptionID, map[string]interface{}{
			"cancel_at_period_end": true,
		})
		if err != nil {
			return fmt.Errorf("failed to update Polar subscription: %w", err)
		}
	}

	// Update in database
	updates := map[string]interface{}{
		"cancel_at_period_end": !immediately,
	}
	if immediately {
		updates["status"] = "canceled"
		now := time.Now()
		updates["canceled_at"] = &now
	}

	if err := p.db.Model(&subscription).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update subscription in database: %w", err)
	}

	return nil
}

// Webhook Handling

// HandleWebhook processes Polar webhook events
func (p *PolarService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	// Verify webhook signature (implement signature verification)
	if !p.verifyWebhookSignature(payload, signature) {
		return fmt.Errorf("invalid webhook signature")
	}

	var event PolarWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("failed to unmarshal webhook event: %w", err)
	}

	// Check if event already processed
	var existingEvent models.WebhookEvent
	if err := p.db.Where("event_id = ?", event.ID).First(&existingEvent).Error; err == nil {
		if existingEvent.Processed {
			return nil // Already processed
		}
	}

	// Create webhook event record
	webhookEvent := &models.WebhookEvent{
		Provider:  "polar",
		EventType: event.Type,
		EventID:   event.ID,
		Data:      event.Data,
	}

	if err := p.db.Create(webhookEvent).Error; err != nil {
		return fmt.Errorf("failed to create webhook event record: %w", err)
	}

	// Process the event
	switch event.Type {
	case "payment.succeeded":
		return p.handlePaymentSucceeded(ctx, event)
	case "payment.failed":
		return p.handlePaymentFailed(ctx, event)
	case "subscription.created":
		return p.handleSubscriptionCreated(ctx, event)
	case "subscription.updated":
		return p.handleSubscriptionUpdated(ctx, event)
	case "subscription.canceled":
		return p.handleSubscriptionCanceled(ctx, event)
	case "product.created":
		return p.handleProductCreated(ctx, event)
	case "product.updated":
		return p.handleProductUpdated(ctx, event)
	case "product.deleted":
		return p.handleProductDeleted(ctx, event)
	case "plan.created":
		return p.handlePlanCreated(ctx, event)
	case "plan.updated":
		return p.handlePlanUpdated(ctx, event)
	case "plan.deleted":
		return p.handlePlanDeleted(ctx, event)
	default:
		// Mark as processed even if we don't handle it
		p.db.Model(webhookEvent).Updates(map[string]interface{}{
			"processed":    true,
			"processed_at": time.Now(),
		})
		return nil
	}
}

// Event Handlers

func (p *PolarService) handlePaymentSucceeded(ctx context.Context, event PolarWebhookEvent) error {
	paymentData, ok := event.Data["payment"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment data in webhook")
	}

	paymentID, ok := paymentData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid payment ID in webhook")
	}

	// Update payment status
	if err := p.db.Model(&models.Payment{}).
		Where("polar_payment_id = ?", paymentID).
		Updates(map[string]interface{}{
			"status": "succeeded",
		}).Error; err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return p.markWebhookProcessed(ctx, event.ID)
}

func (p *PolarService) handlePaymentFailed(ctx context.Context, event PolarWebhookEvent) error {
	paymentData, ok := event.Data["payment"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment data in webhook")
	}

	paymentID, ok := paymentData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid payment ID in webhook")
	}

	// Update payment status
	if err := p.db.Model(&models.Payment{}).
		Where("polar_payment_id = ?", paymentID).
		Updates(map[string]interface{}{
			"status": "failed",
		}).Error; err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	return p.markWebhookProcessed(ctx, event.ID)
}

func (p *PolarService) handleSubscriptionCreated(ctx context.Context, event PolarWebhookEvent) error {
	subscriptionData, ok := event.Data["subscription"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription data in webhook")
	}

	subscriptionID, ok := subscriptionData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid subscription ID in webhook")
	}

	// Get subscription from database
	var dbSubscription models.Subscription
	if err := p.db.Where("polar_subscription_id = ?", subscriptionID).First(&dbSubscription).Error; err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	status, _ := subscriptionData["status"].(string)
	currentPeriodStart, _ := subscriptionData["current_period_start"].(string)
	currentPeriodEnd, _ := subscriptionData["current_period_end"].(string)

	// Parse dates
	startTime, _ := time.Parse(time.RFC3339, currentPeriodStart)
	endTime, _ := time.Parse(time.RFC3339, currentPeriodEnd)

	// Update subscription status
	updates := map[string]interface{}{
		"status":               status,
		"current_period_start": startTime,
		"current_period_end":   endTime,
	}

	// Handle trial period
	if trialStart, ok := subscriptionData["trial_start"].(string); ok && trialStart != "" {
		if trialStartTime, err := time.Parse(time.RFC3339, trialStart); err == nil {
			updates["trial_start"] = &trialStartTime
		}
	}
	if trialEnd, ok := subscriptionData["trial_end"].(string); ok && trialEnd != "" {
		if trialEndTime, err := time.Parse(time.RFC3339, trialEnd); err == nil {
			updates["trial_end"] = &trialEndTime
		}
	}

	if err := p.db.Model(&dbSubscription).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Update user subscription status
	if p.subscriptionStatusService != nil {
		if err := p.subscriptionStatusService.UpdateUserSubscriptionStatus(ctx, dbSubscription.UserID, &dbSubscription); err != nil {
			return fmt.Errorf("failed to update user subscription status: %w", err)
		}
	}

	return p.markWebhookProcessed(ctx, event.ID)
}

func (p *PolarService) handleSubscriptionUpdated(ctx context.Context, event PolarWebhookEvent) error {
	subscriptionData, ok := event.Data["subscription"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription data in webhook")
	}

	subscriptionID, ok := subscriptionData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid subscription ID in webhook")
	}

	// Get subscription from database
	var dbSubscription models.Subscription
	if err := p.db.Where("polar_subscription_id = ?", subscriptionID).First(&dbSubscription).Error; err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	status, _ := subscriptionData["status"].(string)
	currentPeriodStart, _ := subscriptionData["current_period_start"].(string)
	currentPeriodEnd, _ := subscriptionData["current_period_end"].(string)
	cancelAtPeriodEnd, _ := subscriptionData["cancel_at_period_end"].(bool)

	// Parse dates
	startTime, _ := time.Parse(time.RFC3339, currentPeriodStart)
	endTime, _ := time.Parse(time.RFC3339, currentPeriodEnd)

	// Update subscription
	updates := map[string]interface{}{
		"status":               status,
		"current_period_start": startTime,
		"current_period_end":   endTime,
		"cancel_at_period_end": cancelAtPeriodEnd,
	}

	if canceledAt, ok := subscriptionData["canceled_at"].(string); ok && canceledAt != "" {
		canceledAtTime, _ := time.Parse(time.RFC3339, canceledAt)
		updates["canceled_at"] = &canceledAtTime
	}

	// Handle trial period
	if trialStart, ok := subscriptionData["trial_start"].(string); ok && trialStart != "" {
		if trialStartTime, err := time.Parse(time.RFC3339, trialStart); err == nil {
			updates["trial_start"] = &trialStartTime
		}
	}
	if trialEnd, ok := subscriptionData["trial_end"].(string); ok && trialEnd != "" {
		if trialEndTime, err := time.Parse(time.RFC3339, trialEnd); err == nil {
			updates["trial_end"] = &trialEndTime
		}
	}

	if err := p.db.Model(&dbSubscription).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Update user subscription status
	if p.subscriptionStatusService != nil {
		if err := p.subscriptionStatusService.UpdateUserSubscriptionStatus(ctx, dbSubscription.UserID, &dbSubscription); err != nil {
			return fmt.Errorf("failed to update user subscription status: %w", err)
		}
	}

	return p.markWebhookProcessed(ctx, event.ID)
}

func (p *PolarService) handleSubscriptionCanceled(ctx context.Context, event PolarWebhookEvent) error {
	subscriptionData, ok := event.Data["subscription"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid subscription data in webhook")
	}

	subscriptionID, ok := subscriptionData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid subscription ID in webhook")
	}

	// Get subscription from database
	var dbSubscription models.Subscription
	if err := p.db.Where("polar_subscription_id = ?", subscriptionID).First(&dbSubscription).Error; err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	// Update subscription status
	now := time.Now()
	updates := map[string]interface{}{
		"status":      "canceled",
		"canceled_at": &now,
	}

	if err := p.db.Model(&dbSubscription).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update subscription status: %w", err)
	}

	// Update user subscription status (downgrade to free)
	if p.subscriptionStatusService != nil {
		if err := p.subscriptionStatusService.UpdateUserSubscriptionStatus(ctx, dbSubscription.UserID, &dbSubscription); err != nil {
			return fmt.Errorf("failed to update user subscription status: %w", err)
		}
	}

	return p.markWebhookProcessed(ctx, event.ID)
}

// API Methods

func (p *PolarService) createPolarProduct(ctx context.Context, product PolarProduct) (*PolarProduct, error) {
	result, err := p.makePolarRequest(ctx, "POST", "/products", product)
	if err != nil {
		return nil, err
	}
	return result.(*PolarProduct), nil
}

func (p *PolarService) updatePolarProduct(ctx context.Context, productID string, product PolarProduct) (*PolarProduct, error) {
	result, err := p.makePolarRequest(ctx, "PUT", "/products/"+productID, product)
	if err != nil {
		return nil, err
	}
	return result.(*PolarProduct), nil
}

func (p *PolarService) deletePolarProduct(ctx context.Context, productID string) error {
	_, err := p.makePolarRequest(ctx, "DELETE", "/products/"+productID, nil)
	return err
}

func (p *PolarService) createPolarPrice(ctx context.Context, price PolarPrice) (*PolarPrice, error) {
	result, err := p.makePolarRequest(ctx, "POST", "/prices", price)
	if err != nil {
		return nil, err
	}
	return result.(*PolarPrice), nil
}

func (p *PolarService) deletePolarPrice(ctx context.Context, priceID string) error {
	_, err := p.makePolarRequest(ctx, "DELETE", "/prices/"+priceID, nil)
	return err
}

func (p *PolarService) createPolarCustomer(ctx context.Context, customer PolarCustomer) (*PolarCustomer, error) {
	result, err := p.makePolarRequest(ctx, "POST", "/customers", customer)
	if err != nil {
		return nil, err
	}
	return result.(*PolarCustomer), nil
}

func (p *PolarService) createPolarPayment(ctx context.Context, payment PolarPayment) (*PolarPayment, error) {
	result, err := p.makePolarRequest(ctx, "POST", "/payments", payment)
	if err != nil {
		return nil, err
	}
	return result.(*PolarPayment), nil
}

func (p *PolarService) createPolarSubscription(ctx context.Context, subscription PolarSubscription) (*PolarSubscription, error) {
	result, err := p.makePolarRequest(ctx, "POST", "/subscriptions", subscription)
	if err != nil {
		return nil, err
	}
	return result.(*PolarSubscription), nil
}

func (p *PolarService) updatePolarSubscription(ctx context.Context, subscriptionID string, updates map[string]interface{}) error {
	_, err := p.makePolarRequest(ctx, "PUT", "/subscriptions/"+subscriptionID, updates)
	return err
}

func (p *PolarService) cancelPolarSubscription(ctx context.Context, subscriptionID string) error {
	_, err := p.makePolarRequest(ctx, "DELETE", "/subscriptions/"+subscriptionID, nil)
	return err
}

func (p *PolarService) makePolarRequest(ctx context.Context, method, endpoint string, data interface{}) (interface{}, error) {
	url := p.baseURL + endpoint

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result, nil
}

// Helper Methods

func (p *PolarService) markWebhookProcessed(ctx context.Context, eventID string) error {
	return p.db.Model(&models.WebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"processed":    true,
			"processed_at": time.Now(),
		}).Error
}

func (p *PolarService) convertMetadataToPolar(metadata models.JSONMap) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range metadata {
		result[k] = v
	}
	return result
}

func (p *PolarService) verifyWebhookSignature(payload []byte, signature string) bool {
	// Implement webhook signature verification
	// This is a placeholder - implement actual signature verification
	return true
}

// GetSubscription retrieves a subscription by ID
func (p *PolarService) GetSubscription(ctx context.Context, subscriptionID uint) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := p.db.Preload("User").Preload("Product").Preload("Plan").First(&subscription, subscriptionID).Error; err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}
	return &subscription, nil
}

// GetUserSubscriptions retrieves all subscriptions for a user
func (p *PolarService) GetUserSubscriptions(ctx context.Context, userID uint) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	if err := p.db.Preload("Product").Preload("Plan").
		Where("user_id = ?", userID).
		Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}
	return subscriptions, nil
}

// GetPayment retrieves a payment by ID
func (p *PolarService) GetPayment(ctx context.Context, paymentID uint) (*models.Payment, error) {
	var payment models.Payment
	if err := p.db.Preload("User").Preload("Product").Preload("Subscription").
		First(&payment, paymentID).Error; err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}
	return &payment, nil
}

// GetUserPayments retrieves all payments for a user
func (p *PolarService) GetUserPayments(ctx context.Context, userID uint) ([]models.Payment, error) {
	var payments []models.Payment
	if err := p.db.Preload("Product").Preload("Subscription").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get user payments: %w", err)
	}
	return payments, nil
}

// Product Webhook Handlers

func (p *PolarService) handleProductCreated(ctx context.Context, event PolarWebhookEvent) error {
	productData, ok := event.Data["product"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid product data in webhook")
	}

	productID, ok := productData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid product ID in webhook")
	}

	// Check if product already exists
	var existingProduct models.Product
	if err := p.db.Where("polar_product_id = ?", productID).First(&existingProduct).Error; err == nil {
		// Product already exists, update it
		return p.updateProductFromPolar(ctx, &existingProduct, productData)
	}

	// Create new product
	product := &models.Product{
		Name:           p.getStringFromMap(productData, "name"),
		Description:    p.getStringFromMap(productData, "description"),
		IsActive:       p.getBoolFromMap(productData, "active", true),
		PolarProductID: productID,
		Metadata:       p.convertPolarMetadataToMap(productData),
	}

	// Set recurring status based on Polar product type
	if productType, ok := productData["type"].(string); ok && productType == "recurring" {
		product.IsRecurring = true
	}

	if err := p.db.Create(product).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return p.markWebhookProcessed(ctx, event.ID)
}

func (p *PolarService) handleProductUpdated(ctx context.Context, event PolarWebhookEvent) error {
	productData, ok := event.Data["product"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid product data in webhook")
	}

	productID, ok := productData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid product ID in webhook")
	}

	// Find existing product
	var existingProduct models.Product
	if err := p.db.Where("polar_product_id = ?", productID).First(&existingProduct).Error; err != nil {
		// Product doesn't exist, create it
		return p.handleProductCreated(ctx, event)
	}

	// Update existing product
	return p.updateProductFromPolar(ctx, &existingProduct, productData)
}

func (p *PolarService) handleProductDeleted(ctx context.Context, event PolarWebhookEvent) error {
	productData, ok := event.Data["product"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid product data in webhook")
	}

	productID, ok := productData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid product ID in webhook")
	}

	// Find and deactivate product
	var existingProduct models.Product
	if err := p.db.Where("polar_product_id = ?", productID).First(&existingProduct).Error; err != nil {
		// Product doesn't exist, mark as processed
		return p.markWebhookProcessed(ctx, event.ID)
	}

	// Deactivate product instead of deleting to preserve data integrity
	if err := p.db.Model(&existingProduct).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate product: %w", err)
	}

	return p.markWebhookProcessed(ctx, event.ID)
}

func (p *PolarService) handlePlanCreated(ctx context.Context, event PolarWebhookEvent) error {
	planData, ok := event.Data["plan"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid plan data in webhook")
	}

	planID, ok := planData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid plan ID in webhook")
	}

	// Find the product
	productID, ok := planData["product_id"].(string)
	if !ok {
		return fmt.Errorf("invalid product ID in plan data")
	}

	var product models.Product
	if err := p.db.Where("polar_product_id = ?", productID).First(&product).Error; err != nil {
		return fmt.Errorf("failed to find product for plan: %w", err)
	}

	// Check if plan already exists
	var existingPlan models.Plan
	if err := p.db.Where("polar_plan_id = ?", planID).First(&existingPlan).Error; err == nil {
		// Plan already exists, update it
		return p.updatePlanFromPolar(ctx, &existingPlan, planData)
	}

	// Create new plan
	plan := &models.Plan{
		Name:          p.getStringFromMap(planData, "name"),
		Description:   p.getStringFromMap(planData, "description"),
		ProductID:     product.ID,
		Price:         p.getInt64FromMap(planData, "price"),
		Currency:      p.getStringFromMap(planData, "currency"),
		Interval:      p.getStringFromMap(planData, "interval"),
		IntervalCount: p.getIntFromMap(planData, "interval_count", 1),
		IsActive:      p.getBoolFromMap(planData, "active", true),
		PolarPlanID:   planID,
		Metadata:      p.convertPolarMetadataToMap(planData),
	}

	if err := p.db.Create(plan).Error; err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	return p.markWebhookProcessed(ctx, event.ID)
}

func (p *PolarService) handlePlanUpdated(ctx context.Context, event PolarWebhookEvent) error {
	planData, ok := event.Data["plan"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid plan data in webhook")
	}

	planID, ok := planData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid plan ID in webhook")
	}

	// Find existing plan
	var existingPlan models.Plan
	if err := p.db.Where("polar_plan_id = ?", planID).First(&existingPlan).Error; err != nil {
		// Plan doesn't exist, create it
		return p.handlePlanCreated(ctx, event)
	}

	// Update existing plan
	return p.updatePlanFromPolar(ctx, &existingPlan, planData)
}

func (p *PolarService) handlePlanDeleted(ctx context.Context, event PolarWebhookEvent) error {
	planData, ok := event.Data["plan"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid plan data in webhook")
	}

	planID, ok := planData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid plan ID in webhook")
	}

	// Find and deactivate plan
	var existingPlan models.Plan
	if err := p.db.Where("polar_plan_id = ?", planID).First(&existingPlan).Error; err != nil {
		// Plan doesn't exist, mark as processed
		return p.markWebhookProcessed(ctx, event.ID)
	}

	// Deactivate plan instead of deleting to preserve data integrity
	if err := p.db.Model(&existingPlan).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate plan: %w", err)
	}

	return p.markWebhookProcessed(ctx, event.ID)
}

// Helper methods for product/plan synchronization

func (p *PolarService) updateProductFromPolar(ctx context.Context, product *models.Product, productData map[string]interface{}) error {
	updates := map[string]interface{}{
		"name":        p.getStringFromMap(productData, "name"),
		"description": p.getStringFromMap(productData, "description"),
		"is_active":   p.getBoolFromMap(productData, "active", true),
		"metadata":    p.convertPolarMetadataToMap(productData),
	}

	// Update recurring status
	if productType, ok := productData["type"].(string); ok && productType == "recurring" {
		updates["is_recurring"] = true
	}

	if err := p.db.Model(product).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return p.markWebhookProcessed(ctx, p.getStringFromMap(productData, "id"))
}

func (p *PolarService) updatePlanFromPolar(ctx context.Context, plan *models.Plan, planData map[string]interface{}) error {
	updates := map[string]interface{}{
		"name":           p.getStringFromMap(planData, "name"),
		"description":    p.getStringFromMap(planData, "description"),
		"price":          p.getInt64FromMap(planData, "price"),
		"currency":       p.getStringFromMap(planData, "currency"),
		"interval":       p.getStringFromMap(planData, "interval"),
		"interval_count": p.getIntFromMap(planData, "interval_count", 1),
		"is_active":      p.getBoolFromMap(planData, "active", true),
		"metadata":       p.convertPolarMetadataToMap(planData),
	}

	if err := p.db.Model(plan).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	return p.markWebhookProcessed(ctx, p.getStringFromMap(planData, "id"))
}

// Helper methods for data extraction

func (p *PolarService) getStringFromMap(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func (p *PolarService) getInt64FromMap(data map[string]interface{}, key string) int64 {
	if val, ok := data[key].(float64); ok {
		return int64(val)
	}
	if val, ok := data[key].(int64); ok {
		return val
	}
	if val, ok := data[key].(int); ok {
		return int64(val)
	}
	return 0
}

func (p *PolarService) getIntFromMap(data map[string]interface{}, key string, defaultValue int) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	if val, ok := data[key].(int64); ok {
		return int(val)
	}
	if val, ok := data[key].(int); ok {
		return val
	}
	return defaultValue
}

func (p *PolarService) getBoolFromMap(data map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := data[key].(bool); ok {
		return val
	}
	return defaultValue
}

func (p *PolarService) convertPolarMetadataToMap(data map[string]interface{}) models.JSONMap {
	result := make(models.JSONMap)
	for k, v := range data {
		// Skip known fields that aren't metadata
		if k == "id" || k == "name" || k == "description" || k == "active" || k == "type" ||
			k == "price" || k == "currency" || k == "interval" || k == "interval_count" ||
			k == "product_id" || k == "created_at" || k == "updated_at" {
			continue
		}
		result[k] = v
	}
	return result
}
