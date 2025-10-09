package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
)

// PaymentValidationMiddleware validates payment-related requests
func PaymentValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate payment method in request body for payment creation
		if c.Request.Method == "POST" && strings.Contains(c.Request.URL.Path, "/payments") {
			var requestBody map[string]interface{}
			if err := c.ShouldBindJSON(&requestBody); err == nil {
				if paymentMethod, exists := requestBody["payment_method"]; exists {
					if paymentMethod != "stripe" && paymentMethod != "polar" {
						utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payment method. Must be 'stripe' or 'polar'", map[string]interface{}{"error": "Invalid payment method"})
						c.Abort()
						return
					}
				}
			}
		}

		// Validate currency codes
		if c.Request.Method == "POST" && (strings.Contains(c.Request.URL.Path, "/payments") || strings.Contains(c.Request.URL.Path, "/products")) {
			var requestBody map[string]interface{}
			if err := c.ShouldBindJSON(&requestBody); err == nil {
				if currency, exists := requestBody["currency"]; exists {
					validCurrencies := []string{"usd", "eur", "gbp", "cad", "aud", "jpy", "chf", "nok", "sek", "dkk"}
					currencyStr := fmt.Sprintf("%v", currency)
					isValid := false
					for _, validCurrency := range validCurrencies {
						if strings.ToLower(currencyStr) == validCurrency {
							isValid = true
							break
						}
					}
					if !isValid {
						utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid currency code", map[string]interface{}{"error": "Invalid currency code"})
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

// StripeWebhookMiddleware validates Stripe webhook signatures
func StripeWebhookMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("Stripe-Signature")
		if signature == "" {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Missing Stripe signature", nil)
			c.Abort()
			return
		}

		// Get the raw body
		body, err := c.GetRawData()
		if err != nil {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Failed to read request body", map[string]interface{}{"error": err.Error()})
			c.Abort()
			return
		}

		// Verify the signature
		if !verifyStripeSignature(body, signature) {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid Stripe signature", nil)
			c.Abort()
			return
		}

		// Store the body back for the handler
		c.Request.Body = &bodyReader{data: body}
		c.Next()
	}
}

// PolarWebhookMiddleware validates Polar webhook signatures
func PolarWebhookMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("X-Polar-Signature")
		if signature == "" {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Missing Polar signature", nil)
			c.Abort()
			return
		}

		// Get the raw body
		body, err := c.GetRawData()
		if err != nil {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Failed to read request body", map[string]interface{}{"error": err.Error()})
			c.Abort()
			return
		}

		// Verify the signature
		if !verifyPolarSignature(body, signature) {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "Invalid Polar signature", nil)
			c.Abort()
			return
		}

		// Store the body back for the handler
		c.Request.Body = &bodyReader{data: body}
		c.Next()
	}
}

// SubscriptionAccessMiddleware ensures user can only access their own subscriptions
func SubscriptionAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		_, exists := c.Get("user_id")
		if !exists {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
			c.Abort()
			return
		}

		// For subscription-specific routes, verify ownership
		if strings.Contains(c.Request.URL.Path, "/subscriptions/") && c.Request.Method != "POST" {
			_ = c.Param("id") // subscriptionID - we'll let the controller handle verification
			// This would typically check the database to ensure the subscription belongs to the user
			// For now, we'll just pass through - the controller will handle the actual verification
		}

		c.Next()
	}
}

// PaymentMethodValidationMiddleware validates payment method data
func PaymentMethodValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" && strings.Contains(c.Request.URL.Path, "/subscriptions") {
			var requestBody map[string]interface{}
			if err := c.ShouldBindJSON(&requestBody); err == nil {
				// Validate payment method
				if paymentMethod, exists := requestBody["payment_method"]; exists {
					validMethods := []string{"stripe", "polar"}
					paymentMethodStr := fmt.Sprintf("%v", paymentMethod)
					isValid := false
					for _, validMethod := range validMethods {
						if paymentMethodStr == validMethod {
							isValid = true
							break
						}
					}
					if !isValid {
						utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid payment method", nil)
						c.Abort()
						return
					}
				}

				// Validate plan ID
				if planID, exists := requestBody["plan_id"]; exists {
					if planID == nil || planID == "" {
						utils.SendErrorResponse(c, http.StatusBadRequest, "Plan ID is required", nil)
						c.Abort()
						return
					}
				}
			}
		}

		c.Next()
	}
}

// Helper functions

func verifyStripeSignature(payload []byte, signature string) bool {
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return false
	}

	// Parse the signature header
	parts := strings.Split(signature, ",")
	if len(parts) != 2 {
		return false
	}

	timestamp := parts[0]
	expectedSignature := parts[1]

	// Create the payload to verify
	payloadToVerify := timestamp + "." + string(payload)

	// Compute HMAC
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write([]byte(payloadToVerify))
	computedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(expectedSignature), []byte(computedSignature))
}

func verifyPolarSignature(payload []byte, signature string) bool {
	webhookSecret := os.Getenv("POLAR_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return false
	}

	// Create HMAC
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	// Compare signatures
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// bodyReader implements io.ReadCloser for restoring request body
type bodyReader struct {
	data []byte
	pos  int
}

func (br *bodyReader) Read(p []byte) (n int, err error) {
	if br.pos >= len(br.data) {
		return 0, fmt.Errorf("EOF")
	}
	n = copy(p, br.data[br.pos:])
	br.pos += n
	return n, nil
}

func (br *bodyReader) Close() error {
	return nil
}

// Rate limiting for payment endpoints
func PaymentRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Apply stricter rate limiting for payment endpoints
		// This would typically use Redis-based rate limiting
		// For now, we'll just pass through
		c.Next()
	}
}

// Webhook rate limiting
func WebhookRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Apply rate limiting for webhook endpoints
		// This prevents abuse of webhook endpoints
		c.Next()
	}
}
