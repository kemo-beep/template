package services

import (
	"context"
	"fmt"
	"time"

	"mobile-backend/models"

	"gorm.io/gorm"
)

type ProductSyncService struct {
	db *gorm.DB
}

func NewProductSyncService(db *gorm.DB) *ProductSyncService {
	return &ProductSyncService{
		db: db,
	}
}

// ProductSyncData represents data needed to sync a product
type ProductSyncData struct {
	Name          string
	Description   string
	IsActive      bool
	IsRecurring   bool
	Interval      string
	IntervalCount int
	Metadata      models.JSONMap
	ExternalID    string
	Provider      string // "stripe" or "polar"
}

// PlanSyncData represents data needed to sync a plan
type PlanSyncData struct {
	Name          string
	Description   string
	ProductID     uint
	Price         int64
	Currency      string
	Interval      string
	IntervalCount int
	IsActive      bool
	Metadata      models.JSONMap
	ExternalID    string
	Provider      string // "stripe" or "polar"
}

// SyncProductFromStripe syncs a product from Stripe webhook data
func (ps *ProductSyncService) SyncProductFromStripe(ctx context.Context, stripeProductID string, data *ProductSyncData) (*models.Product, error) {
	// Check if product already exists
	var existingProduct models.Product
	err := ps.db.Where("stripe_product_id = ?", stripeProductID).First(&existingProduct).Error

	if err == gorm.ErrRecordNotFound {
		// Create new product
		product := &models.Product{
			Name:            data.Name,
			Description:     data.Description,
			IsActive:        data.IsActive,
			IsRecurring:     data.IsRecurring,
			Interval:        data.Interval,
			IntervalCount:   data.IntervalCount,
			Metadata:        data.Metadata,
			StripeProductID: stripeProductID,
		}

		if err := ps.db.Create(product).Error; err != nil {
			return nil, fmt.Errorf("failed to create product: %w", err)
		}

		return product, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check existing product: %w", err)
	}

	// Update existing product
	updates := map[string]interface{}{
		"name":           data.Name,
		"description":    data.Description,
		"is_active":      data.IsActive,
		"is_recurring":   data.IsRecurring,
		"interval":       data.Interval,
		"interval_count": data.IntervalCount,
		"metadata":       data.Metadata,
		"updated_at":     time.Now(),
	}

	if err := ps.db.Model(&existingProduct).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &existingProduct, nil
}

// SyncProductFromPolar syncs a product from Polar webhook data
func (ps *ProductSyncService) SyncProductFromPolar(ctx context.Context, polarProductID string, data *ProductSyncData) (*models.Product, error) {
	// Check if product already exists
	var existingProduct models.Product
	err := ps.db.Where("polar_product_id = ?", polarProductID).First(&existingProduct).Error

	if err == gorm.ErrRecordNotFound {
		// Create new product
		product := &models.Product{
			Name:           data.Name,
			Description:    data.Description,
			IsActive:       data.IsActive,
			IsRecurring:    data.IsRecurring,
			Interval:       data.Interval,
			IntervalCount:  data.IntervalCount,
			Metadata:       data.Metadata,
			PolarProductID: polarProductID,
		}

		if err := ps.db.Create(product).Error; err != nil {
			return nil, fmt.Errorf("failed to create product: %w", err)
		}

		return product, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check existing product: %w", err)
	}

	// Update existing product
	updates := map[string]interface{}{
		"name":           data.Name,
		"description":    data.Description,
		"is_active":      data.IsActive,
		"is_recurring":   data.IsRecurring,
		"interval":       data.Interval,
		"interval_count": data.IntervalCount,
		"metadata":       data.Metadata,
		"updated_at":     time.Now(),
	}

	if err := ps.db.Model(&existingProduct).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &existingProduct, nil
}

// SyncPlanFromStripe syncs a plan from Stripe price webhook data
func (ps *ProductSyncService) SyncPlanFromStripe(ctx context.Context, stripePriceID string, productID uint, data *PlanSyncData) (*models.Plan, error) {
	// Check if plan already exists
	var existingPlan models.Plan
	err := ps.db.Where("stripe_price_id = ?", stripePriceID).First(&existingPlan).Error

	if err == gorm.ErrRecordNotFound {
		// Create new plan
		plan := &models.Plan{
			Name:          data.Name,
			Description:   data.Description,
			ProductID:     productID,
			Price:         data.Price,
			Currency:      data.Currency,
			Interval:      data.Interval,
			IntervalCount: data.IntervalCount,
			IsActive:      data.IsActive,
			Metadata:      data.Metadata,
			StripePriceID: stripePriceID,
		}

		if err := ps.db.Create(plan).Error; err != nil {
			return nil, fmt.Errorf("failed to create plan: %w", err)
		}

		return plan, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check existing plan: %w", err)
	}

	// Update existing plan
	updates := map[string]interface{}{
		"name":           data.Name,
		"description":    data.Description,
		"price":          data.Price,
		"currency":       data.Currency,
		"interval":       data.Interval,
		"interval_count": data.IntervalCount,
		"is_active":      data.IsActive,
		"metadata":       data.Metadata,
		"updated_at":     time.Now(),
	}

	if err := ps.db.Model(&existingPlan).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}

	return &existingPlan, nil
}

// SyncPlanFromPolar syncs a plan from Polar webhook data
func (ps *ProductSyncService) SyncPlanFromPolar(ctx context.Context, polarPlanID string, productID uint, data *PlanSyncData) (*models.Plan, error) {
	// Check if plan already exists
	var existingPlan models.Plan
	err := ps.db.Where("polar_plan_id = ?", polarPlanID).First(&existingPlan).Error

	if err == gorm.ErrRecordNotFound {
		// Create new plan
		plan := &models.Plan{
			Name:          data.Name,
			Description:   data.Description,
			ProductID:     productID,
			Price:         data.Price,
			Currency:      data.Currency,
			Interval:      data.Interval,
			IntervalCount: data.IntervalCount,
			IsActive:      data.IsActive,
			Metadata:      data.Metadata,
			PolarPlanID:   polarPlanID,
		}

		if err := ps.db.Create(plan).Error; err != nil {
			return nil, fmt.Errorf("failed to create plan: %w", err)
		}

		return plan, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to check existing plan: %w", err)
	}

	// Update existing plan
	updates := map[string]interface{}{
		"name":           data.Name,
		"description":    data.Description,
		"price":          data.Price,
		"currency":       data.Currency,
		"interval":       data.Interval,
		"interval_count": data.IntervalCount,
		"is_active":      data.IsActive,
		"metadata":       data.Metadata,
		"updated_at":     time.Now(),
	}

	if err := ps.db.Model(&existingPlan).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update plan: %w", err)
	}

	return &existingPlan, nil
}

// DeactivateProduct deactivates a product by external ID
func (ps *ProductSyncService) DeactivateProduct(ctx context.Context, externalID, provider string) error {
	var query *gorm.DB

	switch provider {
	case "stripe":
		query = ps.db.Model(&models.Product{}).Where("stripe_product_id = ?", externalID)
	case "polar":
		query = ps.db.Model(&models.Product{}).Where("polar_product_id = ?", externalID)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	if err := query.Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate product: %w", err)
	}

	return nil
}

// DeactivatePlan deactivates a plan by external ID
func (ps *ProductSyncService) DeactivatePlan(ctx context.Context, externalID, provider string) error {
	var query *gorm.DB

	switch provider {
	case "stripe":
		query = ps.db.Model(&models.Plan{}).Where("stripe_price_id = ?", externalID)
	case "polar":
		query = ps.db.Model(&models.Plan{}).Where("polar_plan_id = ?", externalID)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	if err := query.Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate plan: %w", err)
	}

	return nil
}

// GetProductByExternalID retrieves a product by external ID
func (ps *ProductSyncService) GetProductByExternalID(ctx context.Context, externalID, provider string) (*models.Product, error) {
	var product models.Product
	var query *gorm.DB

	switch provider {
	case "stripe":
		query = ps.db.Where("stripe_product_id = ?", externalID)
	case "polar":
		query = ps.db.Where("polar_product_id = ?", externalID)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err := query.First(&product).Error; err != nil {
		return nil, err
	}

	return &product, nil
}

// GetPlanByExternalID retrieves a plan by external ID
func (ps *ProductSyncService) GetPlanByExternalID(ctx context.Context, externalID, provider string) (*models.Plan, error) {
	var plan models.Plan
	var query *gorm.DB

	switch provider {
	case "stripe":
		query = ps.db.Where("stripe_price_id = ?", externalID)
	case "polar":
		query = ps.db.Where("polar_plan_id = ?", externalID)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err := query.First(&plan).Error; err != nil {
		return nil, err
	}

	return &plan, nil
}

// SyncStats returns statistics about product synchronization
func (ps *ProductSyncService) SyncStats(ctx context.Context) (map[string]interface{}, error) {
	var stats = make(map[string]interface{})

	// Count products by provider
	var stripeProductCount int64
	var polarProductCount int64
	var totalProductCount int64

	ps.db.Model(&models.Product{}).Where("stripe_product_id IS NOT NULL AND stripe_product_id != ''").Count(&stripeProductCount)
	ps.db.Model(&models.Product{}).Where("polar_product_id IS NOT NULL AND polar_product_id != ''").Count(&polarProductCount)
	ps.db.Model(&models.Product{}).Count(&totalProductCount)

	// Count plans by provider
	var stripePlanCount int64
	var polarPlanCount int64
	var totalPlanCount int64

	ps.db.Model(&models.Plan{}).Where("stripe_price_id IS NOT NULL AND stripe_price_id != ''").Count(&stripePlanCount)
	ps.db.Model(&models.Plan{}).Where("polar_plan_id IS NOT NULL AND polar_plan_id != ''").Count(&polarPlanCount)
	ps.db.Model(&models.Plan{}).Count(&totalPlanCount)

	// Count active vs inactive
	var activeProductCount int64
	var activePlanCount int64

	ps.db.Model(&models.Product{}).Where("is_active = ?", true).Count(&activeProductCount)
	ps.db.Model(&models.Plan{}).Where("is_active = ?", true).Count(&activePlanCount)

	stats["products"] = map[string]interface{}{
		"total":  totalProductCount,
		"active": activeProductCount,
		"stripe": stripeProductCount,
		"polar":  polarProductCount,
	}

	stats["plans"] = map[string]interface{}{
		"total":  totalPlanCount,
		"active": activePlanCount,
		"stripe": stripePlanCount,
		"polar":  polarPlanCount,
	}

	return stats, nil
}
