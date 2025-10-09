package controllers

import (
	"net/http"

	"mobile-backend/services"
	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
)

type ProductWebhookController struct {
	stripeService      *services.StripeService
	polarService       *services.PolarService
	productSyncService *services.ProductSyncService
}

func NewProductWebhookController(
	stripeService *services.StripeService,
	polarService *services.PolarService,
	productSyncService *services.ProductSyncService,
) *ProductWebhookController {
	return &ProductWebhookController{
		stripeService:      stripeService,
		polarService:       polarService,
		productSyncService: productSyncService,
	}
}

// HandleStripeWebhook handles Stripe product webhooks
func (pwc *ProductWebhookController) HandleStripeWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	// Read the raw body
	body, err := c.GetRawData()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Failed to read request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get the signature header
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Missing Stripe signature", map[string]interface{}{"error": "Stripe-Signature header is required"})
		return
	}

	// Process the webhook
	if err := pwc.stripeService.HandleWebhook(ctx, body, signature); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Webhook processing failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, nil, "Webhook processed successfully")
}

// HandlePolarWebhook handles Polar product webhooks
func (pwc *ProductWebhookController) HandlePolarWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	// Read the raw body
	body, err := c.GetRawData()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Failed to read request body", map[string]interface{}{"error": err.Error()})
		return
	}

	// Get the signature header
	signature := c.GetHeader("X-Polar-Signature")
	if signature == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Missing Polar signature", map[string]interface{}{"error": "X-Polar-Signature header is required"})
		return
	}

	// Process the webhook
	if err := pwc.polarService.HandleWebhook(ctx, body, signature); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Webhook processing failed", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, nil, "Webhook processed successfully")
}

// GetProductSyncStats returns product synchronization statistics
func (pwc *ProductWebhookController) GetProductSyncStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := pwc.productSyncService.SyncStats(ctx)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get sync stats", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, stats, "Sync statistics retrieved successfully")
}

// SyncProductFromProvider manually syncs a product from external provider
func (pwc *ProductWebhookController) SyncProductFromProvider(c *gin.Context) {
	provider := c.Param("provider")

	if provider != "stripe" && provider != "polar" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid provider", map[string]interface{}{"error": "Provider must be 'stripe' or 'polar'"})
		return
	}

	// This would require implementing a method to fetch product from external API
	utils.SendErrorResponse(c, http.StatusNotImplemented, "Manual sync not implemented", map[string]interface{}{"error": "Manual sync from external API not yet implemented"})
}

// GetProductByExternalID retrieves a product by external ID
func (pwc *ProductWebhookController) GetProductByExternalID(c *gin.Context) {
	ctx := c.Request.Context()

	provider := c.Param("provider")
	externalID := c.Param("external_id")

	if provider != "stripe" && provider != "polar" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid provider", map[string]interface{}{"error": "Provider must be 'stripe' or 'polar'"})
		return
	}

	product, err := pwc.productSyncService.GetProductByExternalID(ctx, externalID, provider)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Product not found", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, product, "Product retrieved successfully")
}

// GetPlanByExternalID retrieves a plan by external ID
func (pwc *ProductWebhookController) GetPlanByExternalID(c *gin.Context) {
	ctx := c.Request.Context()

	provider := c.Param("provider")
	externalID := c.Param("external_id")

	if provider != "stripe" && provider != "polar" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid provider", map[string]interface{}{"error": "Provider must be 'stripe' or 'polar'"})
		return
	}

	plan, err := pwc.productSyncService.GetPlanByExternalID(ctx, externalID, provider)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Plan not found", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, plan, "Plan retrieved successfully")
}

// DeactivateProductByExternalID deactivates a product by external ID
func (pwc *ProductWebhookController) DeactivateProductByExternalID(c *gin.Context) {
	ctx := c.Request.Context()

	provider := c.Param("provider")
	externalID := c.Param("external_id")

	if provider != "stripe" && provider != "polar" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid provider", map[string]interface{}{"error": "Provider must be 'stripe' or 'polar'"})
		return
	}

	if err := pwc.productSyncService.DeactivateProduct(ctx, externalID, provider); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to deactivate product", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, nil, "Product deactivated successfully")
}

// DeactivatePlanByExternalID deactivates a plan by external ID
func (pwc *ProductWebhookController) DeactivatePlanByExternalID(c *gin.Context) {
	ctx := c.Request.Context()

	provider := c.Param("provider")
	externalID := c.Param("external_id")

	if provider != "stripe" && provider != "polar" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid provider", map[string]interface{}{"error": "Provider must be 'stripe' or 'polar'"})
		return
	}

	if err := pwc.productSyncService.DeactivatePlan(ctx, externalID, provider); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to deactivate plan", map[string]interface{}{"error": err.Error()})
		return
	}

	utils.SendSuccessResponse(c, nil, "Plan deactivated successfully")
}

// ListProductsByProvider lists products by provider
func (pwc *ProductWebhookController) ListProductsByProvider(c *gin.Context) {
	provider := c.Param("provider")

	if provider != "stripe" && provider != "polar" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid provider", map[string]interface{}{"error": "Provider must be 'stripe' or 'polar'"})
		return
	}

	// This would require implementing a method to list products by provider
	utils.SendErrorResponse(c, http.StatusNotImplemented, "List by provider not implemented", map[string]interface{}{"error": "Listing products by provider not yet implemented"})
}

// ListPlansByProvider lists plans by provider
func (pwc *ProductWebhookController) ListPlansByProvider(c *gin.Context) {
	provider := c.Param("provider")

	if provider != "stripe" && provider != "polar" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid provider", map[string]interface{}{"error": "Provider must be 'stripe' or 'polar'"})
		return
	}

	// This would require implementing a method to list plans by provider
	utils.SendErrorResponse(c, http.StatusNotImplemented, "List by provider not implemented", map[string]interface{}{"error": "Listing plans by provider not yet implemented"})
}
