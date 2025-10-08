package controllers

import (
	"net/http"
	"strconv"
	"mobile-backend/models"
	"mobile-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderController struct {
	db *gorm.DB
}

func NewOrderController(db *gorm.DB) *OrderController {
	return &OrderController{db: db}
}

// GetOrderList retrieves all Order records
// @Summary Get Order list
// @Description Get list of all Order records
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /order [get]
func (c *OrderController) GetOrderList(ctx *gin.Context) {
	var orders []models.Order
	
	if err := c.db.Find(&orders).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, orders, "Orders retrieved successfully")
}

// GetOrder retrieves a Order by ID
// @Summary Get Order by ID
// @Description Get Order record by ID
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /order/{id} [get]
func (c *OrderController) GetOrder(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var order models.Order
	if err := c.db.First(&order, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFoundResponse(ctx, "Order not found")
			return
		}
		utils.SendErrorResponse(ctx, "Failed to fetch order", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, order, "Order retrieved successfully")
}

// CreateOrder creates a new Order
// @Summary Create Order
// @Description Create a new Order record
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param order body models.Order true "Order data"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /order [post]
func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var order models.Order
	
	if err := ctx.ShouldBindJSON(&order); err != nil {
		utils.SendValidationErrorResponse(ctx, map[string]string{"input": "Invalid input"})
		return
	}
	
	if err := c.db.Create(&order).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to create order", http.StatusInternalServerError)
		return
	}
	
	utils.SendCreatedResponse(ctx, order, "Order created successfully")
}

// UpdateOrder updates a Order by ID
// @Summary Update Order
// @Description Update Order record by ID
// @Tags Order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param order body models.Order true "Order data"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /order/{id} [put]
func (c *OrderController) UpdateOrder(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	var order models.Order
	if err := c.db.First(&order, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFoundResponse(ctx, "Order not found")
			return
		}
		utils.SendErrorResponse(ctx, "Failed to fetch order", http.StatusInternalServerError)
		return
	}
	
	if err := ctx.ShouldBindJSON(&order); err != nil {
		utils.SendValidationErrorResponse(ctx, map[string]string{"input": "Invalid input"})
		return
	}
	
	if err := c.db.Save(&order).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to update order", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, order, "Order updated successfully")
}

// DeleteOrder deletes a Order by ID
// @Summary Delete Order
// @Description Delete Order record by ID
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /order/{id} [delete]
func (c *OrderController) DeleteOrder(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, "Invalid ID", http.StatusBadRequest)
		return
	}
	
	if err := c.db.Delete(&models.Order{}, uint(id)).Error; err != nil {
		utils.SendErrorResponse(ctx, "Failed to delete order", http.StatusInternalServerError)
		return
	}
	
	utils.SendSuccessResponse(ctx, nil, "Order deleted successfully")
}
