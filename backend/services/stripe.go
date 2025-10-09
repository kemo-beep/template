package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"mobile-backend/models"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/checkout/session"
	"github.com/stripe/stripe-go/v78/customer"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/price"
	"github.com/stripe/stripe-go/v78/product"
	"github.com/stripe/stripe-go/v78/subscription"
	"github.com/stripe/stripe-go/v78/webhook"
	"gorm.io/gorm"
)

type StripeService struct {
	db                        *gorm.DB
	cache                     *CacheService
	websocketService          *WebSocketService
	subscriptionStatusService *SubscriptionStatusService
}

func NewStripeService(db *gorm.DB, cache *CacheService, websocketService *WebSocketService, subscriptionStatusService *SubscriptionStatusService) *StripeService {
	// Initialize Stripe with API key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	return &StripeService{
		db:                        db,
		cache:                     cache,
		websocketService:          websocketService,
		subscriptionStatusService: subscriptionStatusService,
	}
}

// Product Management

// CreateProduct creates a product in Stripe and our database
func (s *StripeService) CreateProduct(ctx context.Context, productData *models.Product) (*models.Product, error) {
	// Create product in Stripe
	stripeProductParams := &stripe.ProductParams{
		Name:        stripe.String(productData.Name),
		Description: stripe.String(productData.Description),
		Active:      stripe.Bool(productData.IsActive),
		Metadata:    s.convertMetadataToStripe(productData.Metadata),
	}

	stripeProduct, err := product.New(stripeProductParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe product: %w", err)
	}

	// Update our product with Stripe ID
	productData.StripeProductID = stripeProduct.ID

	// Save to database
	if err := s.db.Create(productData).Error; err != nil {
		// Clean up Stripe product if database save fails
		product.Del(stripeProduct.ID, nil)
		return nil, fmt.Errorf("failed to save product to database: %w", err)
	}

	return productData, nil
}

// UpdateProduct updates a product in both Stripe and our database
func (s *StripeService) UpdateProduct(ctx context.Context, productID uint, updates *models.Product) (*models.Product, error) {
	var existingProduct models.Product
	if err := s.db.First(&existingProduct, productID).Error; err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Update in Stripe
	if existingProduct.StripeProductID != "" {
		stripeProductParams := &stripe.ProductParams{
			Name:        stripe.String(updates.Name),
			Description: stripe.String(updates.Description),
			Active:      stripe.Bool(updates.IsActive),
			Metadata:    s.convertMetadataToStripe(updates.Metadata),
		}

		_, err := product.Update(existingProduct.StripeProductID, stripeProductParams)
		if err != nil {
			return nil, fmt.Errorf("failed to update Stripe product: %w", err)
		}
	}

	// Update in database
	if err := s.db.Model(&existingProduct).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update product in database: %w", err)
	}

	return &existingProduct, nil
}

// CreatePrice creates a price for a product in Stripe
func (s *StripeService) CreatePrice(ctx context.Context, planData *models.Plan) (*models.Plan, error) {
	var product models.Product
	if err := s.db.First(&product, planData.ProductID).Error; err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	if product.StripeProductID == "" {
		return nil, fmt.Errorf("product does not have Stripe ID")
	}

	// Create price in Stripe
	stripePriceParams := &stripe.PriceParams{
		Product:    stripe.String(product.StripeProductID),
		UnitAmount: stripe.Int64(planData.Price),
		Currency:   stripe.String(planData.Currency),
		Recurring: &stripe.PriceRecurringParams{
			Interval:      stripe.String(planData.Interval),
			IntervalCount: stripe.Int64(int64(planData.IntervalCount)),
		},
		Active:   stripe.Bool(planData.IsActive),
		Metadata: s.convertMetadataToStripe(planData.Metadata),
	}

	stripePrice, err := price.New(stripePriceParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe price: %w", err)
	}

	// Update our plan with Stripe ID
	planData.StripePriceID = stripePrice.ID

	// Save to database
	if err := s.db.Create(planData).Error; err != nil {
		// Clean up Stripe price if database save fails
		// Note: Stripe doesn't have a direct delete method for prices, they are archived
		// price.Del(stripePrice.ID, nil)
		return nil, fmt.Errorf("failed to save plan to database: %w", err)
	}

	return planData, nil
}

// Customer Management

// CreateCustomer creates a customer in Stripe
func (s *StripeService) CreateCustomer(ctx context.Context, user *models.User) (string, error) {
	// Check if customer already exists
	cacheKey := fmt.Sprintf("stripe_customer:%d", user.ID)
	var customerID string

	if err := s.cache.Get(ctx, cacheKey, &customerID); err == nil {
		return customerID, nil
	}

	// Create customer in Stripe
	customerParams := &stripe.CustomerParams{
		Email: stripe.String(user.Email),
		Name:  stripe.String(user.Name),
		Metadata: map[string]string{
			"user_id": strconv.FormatUint(uint64(user.ID), 10),
		},
	}

	stripeCustomer, err := customer.New(customerParams)
	if err != nil {
		return "", fmt.Errorf("failed to create Stripe customer: %w", err)
	}

	// Cache the customer ID
	s.cache.Set(ctx, cacheKey, stripeCustomer.ID, 24*time.Hour)

	return stripeCustomer.ID, nil
}

// Payment Intent Management

// CreatePaymentIntent creates a payment intent for one-time payments
func (s *StripeService) CreatePaymentIntent(ctx context.Context, userID uint, productID uint, amount int64, currency string) (*models.Payment, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var product models.Product
	if err := s.db.First(&product, productID).Error; err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Create Stripe customer
	customerID, err := s.CreateCustomer(ctx, &user)
	if err != nil {
		return nil, err
	}

	// Create payment intent in Stripe
	paymentIntentParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(amount),
		Currency: stripe.String(currency),
		Customer: stripe.String(customerID),
		Metadata: map[string]string{
			"user_id":    strconv.FormatUint(uint64(userID), 10),
			"product_id": strconv.FormatUint(uint64(productID), 10),
		},
	}

	stripePaymentIntent, err := paymentintent.New(paymentIntentParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Create payment record in database
	payment := &models.Payment{
		UserID:                userID,
		ProductID:             productID,
		Amount:                amount,
		Currency:              currency,
		Status:                "pending",
		PaymentMethod:         "stripe",
		StripePaymentIntentID: stripePaymentIntent.ID,
		Description:           fmt.Sprintf("Payment for %s", product.Name),
	}

	if err := s.db.Create(payment).Error; err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	return payment, nil
}

// CreateCheckoutSession creates a Stripe checkout session
func (s *StripeService) CreateCheckoutSession(ctx context.Context, userID uint, productID uint, successURL, cancelURL string) (*stripe.CheckoutSession, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var product models.Product
	if err := s.db.First(&product, productID).Error; err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Create Stripe customer
	customerID, err := s.CreateCustomer(ctx, &user)
	if err != nil {
		return nil, err
	}

	// Create checkout session
	sessionParams := &stripe.CheckoutSessionParams{
		Customer:           stripe.String(customerID),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(product.StripeProductID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String("payment"),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata: map[string]string{
			"user_id":    strconv.FormatUint(uint64(userID), 10),
			"product_id": strconv.FormatUint(uint64(productID), 10),
		},
	}

	session, err := session.New(sessionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return session, nil
}

// Subscription Management

// CreateSubscription creates a subscription in Stripe
func (s *StripeService) CreateSubscription(ctx context.Context, userID uint, planID uint, paymentMethodID string) (*models.Subscription, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var plan models.Plan
	if err := s.db.Preload("Product").First(&plan, planID).Error; err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}

	// Create Stripe customer
	customerID, err := s.CreateCustomer(ctx, &user)
	if err != nil {
		return nil, err
	}

	// Create subscription in Stripe
	subscriptionParams := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(plan.StripePriceID),
			},
		},
		PaymentBehavior: stripe.String("default_incomplete"),
		PaymentSettings: &stripe.SubscriptionPaymentSettingsParams{
			PaymentMethodTypes:       stripe.StringSlice([]string{"card"}),
			SaveDefaultPaymentMethod: stripe.String("on_subscription"),
		},
		Expand: []*string{stripe.String("latest_invoice.payment_intent")},
	}

	if paymentMethodID != "" {
		subscriptionParams.DefaultPaymentMethod = stripe.String(paymentMethodID)
	}

	stripeSubscription, err := subscription.New(subscriptionParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create Stripe subscription: %w", err)
	}

	// Create subscription record in database
	sub := &models.Subscription{
		UserID:               userID,
		ProductID:            plan.ProductID,
		PlanID:               &planID,
		Status:               string(stripeSubscription.Status),
		CurrentPeriodStart:   time.Unix(stripeSubscription.CurrentPeriodStart, 0),
		CurrentPeriodEnd:     time.Unix(stripeSubscription.CurrentPeriodEnd, 0),
		CancelAtPeriodEnd:    stripeSubscription.CancelAtPeriodEnd,
		StripeSubscriptionID: stripeSubscription.ID,
		Quantity:             1,
	}

	if stripeSubscription.TrialStart != 0 {
		trialStart := time.Unix(stripeSubscription.TrialStart, 0)
		sub.TrialStart = &trialStart
	}
	if stripeSubscription.TrialEnd != 0 {
		trialEnd := time.Unix(stripeSubscription.TrialEnd, 0)
		sub.TrialEnd = &trialEnd
	}

	if err := s.db.Create(sub).Error; err != nil {
		// Clean up Stripe subscription if database save fails
		subscription.Cancel(stripeSubscription.ID, &stripe.SubscriptionCancelParams{})
		return nil, fmt.Errorf("failed to create subscription record: %w", err)
	}

	return sub, nil
}

// CancelSubscription cancels a subscription
func (s *StripeService) CancelSubscription(ctx context.Context, subscriptionID uint, immediately bool) error {
	var sub models.Subscription
	if err := s.db.First(&sub, subscriptionID).Error; err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	if sub.StripeSubscriptionID == "" {
		return fmt.Errorf("subscription does not have Stripe ID")
	}

	// Cancel in Stripe
	var params *stripe.SubscriptionParams
	if immediately {
		params = &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(false),
		}
	} else {
		params = &stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		}
	}

	_, err := subscription.Update(sub.StripeSubscriptionID, params)
	if err != nil {
		return fmt.Errorf("failed to cancel Stripe subscription: %w", err)
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

	if err := s.db.Model(&sub).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update subscription in database: %w", err)
	}

	return nil
}

// Webhook Handling

// HandleWebhook processes Stripe webhook events
func (s *StripeService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, signature, os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if err != nil {
		return fmt.Errorf("invalid webhook signature: %w", err)
	}

	// Check if event already processed
	var existingEvent models.WebhookEvent
	if err := s.db.Where("event_id = ?", event.ID).First(&existingEvent).Error; err == nil {
		if existingEvent.Processed {
			return nil // Already processed
		}
	}

	// Create webhook event record
	var data models.JSONMap
	if err := json.Unmarshal(event.Data.Raw, &data); err != nil {
		return fmt.Errorf("failed to unmarshal webhook data: %w", err)
	}

	webhookEvent := &models.WebhookEvent{
		Provider:  "stripe",
		EventType: string(event.Type),
		EventID:   event.ID,
		Data:      data,
	}

	if err := s.db.Create(webhookEvent).Error; err != nil {
		return fmt.Errorf("failed to create webhook event record: %w", err)
	}

	// Process the event
	switch event.Type {
	case "payment_intent.succeeded":
		return s.handlePaymentIntentSucceeded(ctx, event)
	case "payment_intent.payment_failed":
		return s.handlePaymentIntentFailed(ctx, event)
	case "customer.subscription.created":
		return s.handleSubscriptionCreated(ctx, event)
	case "customer.subscription.updated":
		return s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return s.handleSubscriptionDeleted(ctx, event)
	case "invoice.payment_succeeded":
		return s.handleInvoicePaymentSucceeded(ctx, event)
	case "invoice.payment_failed":
		return s.handleInvoicePaymentFailed(ctx, event)
	case "product.created":
		return s.handleProductCreated(ctx, event)
	case "product.updated":
		return s.handleProductUpdated(ctx, event)
	case "product.deleted":
		return s.handleProductDeleted(ctx, event)
	case "price.created":
		return s.handlePriceCreated(ctx, event)
	case "price.updated":
		return s.handlePriceUpdated(ctx, event)
	case "price.deleted":
		return s.handlePriceDeleted(ctx, event)
	default:
		// Mark as processed even if we don't handle it
		s.db.Model(webhookEvent).Updates(map[string]interface{}{
			"processed":    true,
			"processed_at": time.Now(),
		})
		return nil
	}
}

// Event Handlers

func (s *StripeService) handlePaymentIntentSucceeded(ctx context.Context, event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to unmarshal payment intent: %w", err)
	}

	// Get payment record to find user ID
	var payment models.Payment
	if err := s.db.Where("stripe_payment_intent_id = ?", paymentIntent.ID).First(&payment).Error; err != nil {
		return fmt.Errorf("failed to find payment record: %w", err)
	}

	// Update payment status
	if err := s.db.Model(&models.Payment{}).
		Where("stripe_payment_intent_id = ?", paymentIntent.ID).
		Updates(map[string]interface{}{
			"status": "succeeded",
		}).Error; err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Send WebSocket notification
	if s.websocketService != nil {
		paymentData := map[string]interface{}{
			"payment_id": payment.ID,
			"amount":     paymentIntent.Amount,
			"currency":   paymentIntent.Currency,
			"status":     "succeeded",
		}
		s.websocketService.SendPaymentNotification(ctx, payment.UserID, "payment_succeeded", paymentData)
	}

	return s.markWebhookProcessed(ctx, event.ID)
}

func (s *StripeService) handlePaymentIntentFailed(ctx context.Context, event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("failed to unmarshal payment intent: %w", err)
	}

	// Get payment record to find user ID
	var payment models.Payment
	if err := s.db.Where("stripe_payment_intent_id = ?", paymentIntent.ID).First(&payment).Error; err != nil {
		return fmt.Errorf("failed to find payment record: %w", err)
	}

	// Update payment status
	if err := s.db.Model(&models.Payment{}).
		Where("stripe_payment_intent_id = ?", paymentIntent.ID).
		Updates(map[string]interface{}{
			"status": "failed",
		}).Error; err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	// Send WebSocket notification
	if s.websocketService != nil {
		paymentData := map[string]interface{}{
			"payment_id": payment.ID,
			"amount":     paymentIntent.Amount,
			"currency":   paymentIntent.Currency,
			"status":     "failed",
			"error":      paymentIntent.LastPaymentError,
		}
		s.websocketService.SendPaymentNotification(ctx, payment.UserID, "payment_failed", paymentData)
	}

	return s.markWebhookProcessed(ctx, event.ID)
}

func (s *StripeService) handleSubscriptionCreated(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	// Get subscription from database
	var dbSubscription models.Subscription
	if err := s.db.Where("stripe_subscription_id = ?", subscription.ID).First(&dbSubscription).Error; err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	// Update subscription status
	updates := map[string]interface{}{
		"status":               string(subscription.Status),
		"current_period_start": time.Unix(subscription.CurrentPeriodStart, 0),
		"current_period_end":   time.Unix(subscription.CurrentPeriodEnd, 0),
	}

	if subscription.TrialStart != 0 {
		trialStart := time.Unix(subscription.TrialStart, 0)
		updates["trial_start"] = &trialStart
	}
	if subscription.TrialEnd != 0 {
		trialEnd := time.Unix(subscription.TrialEnd, 0)
		updates["trial_end"] = &trialEnd
	}

	if err := s.db.Model(&dbSubscription).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Update user subscription status
	if s.subscriptionStatusService != nil {
		if err := s.subscriptionStatusService.UpdateUserSubscriptionStatus(ctx, dbSubscription.UserID, &dbSubscription); err != nil {
			return fmt.Errorf("failed to update user subscription status: %w", err)
		}
	}

	return s.markWebhookProcessed(ctx, event.ID)
}

func (s *StripeService) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	// Get subscription from database
	var dbSubscription models.Subscription
	if err := s.db.Where("stripe_subscription_id = ?", subscription.ID).First(&dbSubscription).Error; err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	// Update subscription
	updates := map[string]interface{}{
		"status":               string(subscription.Status),
		"current_period_start": time.Unix(subscription.CurrentPeriodStart, 0),
		"current_period_end":   time.Unix(subscription.CurrentPeriodEnd, 0),
		"cancel_at_period_end": subscription.CancelAtPeriodEnd,
	}

	if subscription.CanceledAt != 0 {
		canceledAt := time.Unix(subscription.CanceledAt, 0)
		updates["canceled_at"] = &canceledAt
	}

	if subscription.TrialStart != 0 {
		trialStart := time.Unix(subscription.TrialStart, 0)
		updates["trial_start"] = &trialStart
	}
	if subscription.TrialEnd != 0 {
		trialEnd := time.Unix(subscription.TrialEnd, 0)
		updates["trial_end"] = &trialEnd
	}

	if err := s.db.Model(&dbSubscription).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Update user subscription status
	if s.subscriptionStatusService != nil {
		if err := s.subscriptionStatusService.UpdateUserSubscriptionStatus(ctx, dbSubscription.UserID, &dbSubscription); err != nil {
			return fmt.Errorf("failed to update user subscription status: %w", err)
		}
	}

	return s.markWebhookProcessed(ctx, event.ID)
}

func (s *StripeService) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	// Get subscription from database
	var dbSubscription models.Subscription
	if err := s.db.Where("stripe_subscription_id = ?", subscription.ID).First(&dbSubscription).Error; err != nil {
		return fmt.Errorf("failed to find subscription: %w", err)
	}

	// Update subscription status
	now := time.Now()
	updates := map[string]interface{}{
		"status":      "canceled",
		"canceled_at": &now,
	}

	if err := s.db.Model(&dbSubscription).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	// Update user subscription status (downgrade to free)
	if s.subscriptionStatusService != nil {
		if err := s.subscriptionStatusService.UpdateUserSubscriptionStatus(ctx, dbSubscription.UserID, &dbSubscription); err != nil {
			return fmt.Errorf("failed to update user subscription status: %w", err)
		}
	}

	return s.markWebhookProcessed(ctx, event.ID)
}

func (s *StripeService) handleInvoicePaymentSucceeded(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	// Create payment record for subscription
	if invoice.Subscription != nil {
		var subscription models.Subscription
		if err := s.db.Where("stripe_subscription_id = ?", invoice.Subscription.ID).First(&subscription).Error; err != nil {
			return fmt.Errorf("subscription not found: %w", err)
		}

		payment := &models.Payment{
			UserID:         subscription.UserID,
			ProductID:      subscription.ProductID,
			SubscriptionID: &subscription.ID,
			Amount:         invoice.AmountPaid,
			Currency:       string(invoice.Currency),
			Status:         "succeeded",
			PaymentMethod:  "stripe",
			StripeChargeID: invoice.Charge.ID,
			Description:    fmt.Sprintf("Subscription payment for %s", invoice.Subscription.ID),
		}

		if err := s.db.Create(payment).Error; err != nil {
			return fmt.Errorf("failed to create payment record: %w", err)
		}
	}

	return s.markWebhookProcessed(ctx, event.ID)
}

func (s *StripeService) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("failed to unmarshal invoice: %w", err)
	}

	// Update subscription status to past_due
	if invoice.Subscription != nil {
		if err := s.db.Model(&models.Subscription{}).
			Where("stripe_subscription_id = ?", invoice.Subscription.ID).
			Updates(map[string]interface{}{
				"status": "past_due",
			}).Error; err != nil {
			return fmt.Errorf("failed to update subscription status: %w", err)
		}
	}

	return s.markWebhookProcessed(ctx, event.ID)
}

// Helper Methods

func (s *StripeService) markWebhookProcessed(ctx context.Context, eventID string) error {
	return s.db.Model(&models.WebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"processed":    true,
			"processed_at": time.Now(),
		}).Error
}

func (s *StripeService) convertMetadataToStripe(metadata models.JSONMap) map[string]string {
	result := make(map[string]string)
	for k, v := range metadata {
		if str, ok := v.(string); ok {
			result[k] = str
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}

// GetSubscription retrieves a subscription by ID
func (s *StripeService) GetSubscription(ctx context.Context, subscriptionID uint) (*models.Subscription, error) {
	var subscription models.Subscription
	if err := s.db.Preload("User").Preload("Product").Preload("Plan").First(&subscription, subscriptionID).Error; err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}
	return &subscription, nil
}

// GetUserSubscriptions retrieves all subscriptions for a user
func (s *StripeService) GetUserSubscriptions(ctx context.Context, userID uint) ([]models.Subscription, error) {
	var subscriptions []models.Subscription
	if err := s.db.Preload("Product").Preload("Plan").
		Where("user_id = ?", userID).
		Find(&subscriptions).Error; err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}
	return subscriptions, nil
}

// GetPayment retrieves a payment by ID
func (s *StripeService) GetPayment(ctx context.Context, paymentID uint) (*models.Payment, error) {
	var payment models.Payment
	if err := s.db.Preload("User").Preload("Product").Preload("Subscription").
		First(&payment, paymentID).Error; err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}
	return &payment, nil
}

// GetUserPayments retrieves all payments for a user
func (s *StripeService) GetUserPayments(ctx context.Context, userID uint) ([]models.Payment, error) {
	var payments []models.Payment
	if err := s.db.Preload("Product").Preload("Subscription").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get user payments: %w", err)
	}
	return payments, nil
}

// Product Webhook Handlers

func (s *StripeService) handleProductCreated(ctx context.Context, event stripe.Event) error {
	var stripeProduct stripe.Product
	if err := json.Unmarshal(event.Data.Raw, &stripeProduct); err != nil {
		return fmt.Errorf("failed to unmarshal product: %w", err)
	}

	// Check if product already exists
	var existingProduct models.Product
	if err := s.db.Where("stripe_product_id = ?", stripeProduct.ID).First(&existingProduct).Error; err == nil {
		// Product already exists, update it
		return s.updateProductFromStripe(ctx, &existingProduct, &stripeProduct)
	}

	// Create new product
	product := &models.Product{
		Name:            stripeProduct.Name,
		Description:     stripeProduct.Description,
		IsActive:        stripeProduct.Active,
		StripeProductID: stripeProduct.ID,
		Metadata:        s.convertStripeMetadataToMap(stripeProduct.Metadata),
	}

	// Set recurring status based on Stripe product type
	if stripeProduct.Type == "service" {
		product.IsRecurring = true
	}

	if err := s.db.Create(product).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return s.markWebhookProcessed(ctx, event.ID)
}

func (s *StripeService) handleProductUpdated(ctx context.Context, event stripe.Event) error {
	var stripeProduct stripe.Product
	if err := json.Unmarshal(event.Data.Raw, &stripeProduct); err != nil {
		return fmt.Errorf("failed to unmarshal product: %w", err)
	}

	// Find existing product
	var existingProduct models.Product
	if err := s.db.Where("stripe_product_id = ?", stripeProduct.ID).First(&existingProduct).Error; err != nil {
		// Product doesn't exist, create it
		return s.handleProductCreated(ctx, event)
	}

	// Update existing product
	return s.updateProductFromStripe(ctx, &existingProduct, &stripeProduct)
}

func (s *StripeService) handleProductDeleted(ctx context.Context, event stripe.Event) error {
	var stripeProduct stripe.Product
	if err := json.Unmarshal(event.Data.Raw, &stripeProduct); err != nil {
		return fmt.Errorf("failed to unmarshal product: %w", err)
	}

	// Find and deactivate product
	var existingProduct models.Product
	if err := s.db.Where("stripe_product_id = ?", stripeProduct.ID).First(&existingProduct).Error; err != nil {
		// Product doesn't exist, mark as processed
		return s.markWebhookProcessed(ctx, event.ID)
	}

	// Deactivate product instead of deleting to preserve data integrity
	if err := s.db.Model(&existingProduct).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate product: %w", err)
	}

	return s.markWebhookProcessed(ctx, event.ID)
}

func (s *StripeService) handlePriceCreated(ctx context.Context, event stripe.Event) error {
	var stripePrice stripe.Price
	if err := json.Unmarshal(event.Data.Raw, &stripePrice); err != nil {
		return fmt.Errorf("failed to unmarshal price: %w", err)
	}

	// Find the product
	var product models.Product
	if err := s.db.Where("stripe_product_id = ?", stripePrice.Product.ID).First(&product).Error; err != nil {
		return fmt.Errorf("failed to find product for price: %w", err)
	}

	// Check if plan already exists
	var existingPlan models.Plan
	if err := s.db.Where("stripe_price_id = ?", stripePrice.ID).First(&existingPlan).Error; err == nil {
		// Plan already exists, update it
		return s.updatePlanFromStripePrice(ctx, &existingPlan, &stripePrice)
	}

	// Create new plan
	plan := &models.Plan{
		Name:          fmt.Sprintf("%s - %s", product.Name, s.formatPriceInterval(&stripePrice)),
		Description:   fmt.Sprintf("Price for %s", product.Name),
		ProductID:     product.ID,
		Price:         stripePrice.UnitAmount,
		Currency:      string(stripePrice.Currency),
		Interval:      string(stripePrice.Recurring.Interval),
		IntervalCount: int(stripePrice.Recurring.IntervalCount),
		IsActive:      stripePrice.Active,
		StripePriceID: stripePrice.ID,
		Metadata:      s.convertStripeMetadataToMap(stripePrice.Metadata),
	}

	if err := s.db.Create(plan).Error; err != nil {
		return fmt.Errorf("failed to create plan: %w", err)
	}

	return s.markWebhookProcessed(ctx, event.ID)
}

func (s *StripeService) handlePriceUpdated(ctx context.Context, event stripe.Event) error {
	var stripePrice stripe.Price
	if err := json.Unmarshal(event.Data.Raw, &stripePrice); err != nil {
		return fmt.Errorf("failed to unmarshal price: %w", err)
	}

	// Find existing plan
	var existingPlan models.Plan
	if err := s.db.Where("stripe_price_id = ?", stripePrice.ID).First(&existingPlan).Error; err != nil {
		// Plan doesn't exist, create it
		return s.handlePriceCreated(ctx, event)
	}

	// Update existing plan
	return s.updatePlanFromStripePrice(ctx, &existingPlan, &stripePrice)
}

func (s *StripeService) handlePriceDeleted(ctx context.Context, event stripe.Event) error {
	var stripePrice stripe.Price
	if err := json.Unmarshal(event.Data.Raw, &stripePrice); err != nil {
		return fmt.Errorf("failed to unmarshal price: %w", err)
	}

	// Find and deactivate plan
	var existingPlan models.Plan
	if err := s.db.Where("stripe_price_id = ?", stripePrice.ID).First(&existingPlan).Error; err != nil {
		// Plan doesn't exist, mark as processed
		return s.markWebhookProcessed(ctx, event.ID)
	}

	// Deactivate plan instead of deleting to preserve data integrity
	if err := s.db.Model(&existingPlan).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate plan: %w", err)
	}

	return s.markWebhookProcessed(ctx, event.ID)
}

// Helper methods for product/plan synchronization

func (s *StripeService) updateProductFromStripe(ctx context.Context, product *models.Product, stripeProduct *stripe.Product) error {
	updates := map[string]interface{}{
		"name":        stripeProduct.Name,
		"description": stripeProduct.Description,
		"is_active":   stripeProduct.Active,
		"metadata":    s.convertStripeMetadataToMap(stripeProduct.Metadata),
	}

	// Update recurring status
	if stripeProduct.Type == "service" {
		updates["is_recurring"] = true
	}

	if err := s.db.Model(product).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return s.markWebhookProcessed(ctx, stripeProduct.ID)
}

func (s *StripeService) updatePlanFromStripePrice(ctx context.Context, plan *models.Plan, stripePrice *stripe.Price) error {
	updates := map[string]interface{}{
		"price":     stripePrice.UnitAmount,
		"currency":  string(stripePrice.Currency),
		"is_active": stripePrice.Active,
		"metadata":  s.convertStripeMetadataToMap(stripePrice.Metadata),
	}

	// Update recurring fields if present
	if stripePrice.Recurring != nil {
		updates["interval"] = string(stripePrice.Recurring.Interval)
		updates["interval_count"] = int(stripePrice.Recurring.IntervalCount)
	}

	if err := s.db.Model(plan).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update plan: %w", err)
	}

	return s.markWebhookProcessed(ctx, stripePrice.ID)
}

func (s *StripeService) formatPriceInterval(price *stripe.Price) string {
	if price.Recurring == nil {
		return "one-time"
	}

	interval := string(price.Recurring.Interval)
	count := price.Recurring.IntervalCount

	if count == 1 {
		return interval
	}
	return fmt.Sprintf("every %d %ss", count, interval)
}

func (s *StripeService) convertStripeMetadataToMap(metadata map[string]string) models.JSONMap {
	result := make(models.JSONMap)
	for k, v := range metadata {
		result[k] = v
	}
	return result
}
