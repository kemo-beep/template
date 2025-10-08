package controllers

import (
	"net/http"
	"strconv"
	"mobile-backend/models"
	"mobile-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductController struct {
	db *gorm.DB
}

func NewProductController(db *gorm.DB) *ProductController {
	return &ProductController{db: db}
}

// GetProductList retrieves all Product records
// @Summary Get Product list
// @Description Get list of all Product records
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /product [get]
func (c *ProductController) GetProductList(ctx *gin.Context) {
	var products []models.Product
	
	if err := c.db.Find(&products).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to fetch products", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, products, "Products retrieved successfully")
}

// GetProduct retrieves a Product by ID
// @Summary Get Product by ID
// @Description Get Product record by ID
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /product/{id} [get]
func (c *ProductController) GetProduct(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var product models.Product
	if err := c.db.First(&product, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFoundResponse(ctx, "Product not found")
			return
		}
		utils.SendErrorResponse(ctx, "Failed to fetch product", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, product, "Product retrieved successfully")
}

// CreateProduct creates a new Product
// @Summary Create Product
// @Description Create a new Product record
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product body models.Product true "Product data"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /product [post]
func (c *ProductController) CreateProduct(ctx *gin.Context) {
	var product models.Product
	
	if err := ctx.ShouldBindJSON(&product); err != nil {
		utils.SendValidationErrorResponse(ctx, map[string]string{"input": "Invalid input"})
		return
	}
	
	if err := c.db.Create(&product).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to create product", http.StatusInternalServerError)
		return
	}
	
	utils.SendCreatedResponse(ctx, product, "Product created successfully")
}

// UpdateProduct updates a Product by ID
// @Summary Update Product
// @Description Update Product record by ID
// @Tags Product
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Product ID"
// @Param product body models.Product true "Product data"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /product/{id} [put]
func (c *ProductController) UpdateProduct(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var product models.Product
	if err := c.db.First(&product, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFoundResponse(ctx, "Product not found")
			return
		}
		utils.SendErrorResponse(ctx, "Failed to fetch product", http.StatusInternalServerError)
		return
	}
	
	if err := ctx.ShouldBindJSON(&product); err != nil {
		utils.SendValidationErrorResponse(ctx, map[string]string{"input": "Invalid input"})
		return
	}
	
	if err := c.db.Save(&product).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to update product", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, product, "Product updated successfully")
}

// DeleteProduct deletes a Product by ID
// @Summary Delete Product
// @Description Delete Product record by ID
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /product/{id} [delete]
func (c *ProductController) DeleteProduct(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	if err := c.db.Delete(&models.Product{}, uint(id)).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to delete product", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, nil, "Product deleted successfully")
}
