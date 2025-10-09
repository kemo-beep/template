package controllers

import (
	"mobile-backend/models"
	"mobile-backend/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CategoryController struct {
	db *gorm.DB
}

func NewCategoryController(db *gorm.DB) *CategoryController {
	return &CategoryController{db: db}
}

// GetCategoryList retrieves all Category records
// @Summary Get Category list
// @Description Get list of all Category records
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /category [get]
func (c *CategoryController) GetCategoryList(ctx *gin.Context) {
	var categorys []models.Category

	if err := c.db.Find(&categorys).Error; err != nil {
		utils.SendErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch categorys", nil)
		return
	}

	utils.SendSuccessResponse(ctx, categorys, "Categorys retrieved successfully")
}

// GetCategory retrieves a Category by ID
// @Summary Get Category by ID
// @Description Get Category record by ID
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /category/{id} [get]
func (c *CategoryController) GetCategory(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "Invalid ID", nil)
		return
	}

	var category models.Category
	if err := c.db.First(&category, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFoundResponse(ctx, "Category not found")
			return
		}
		utils.SendErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch category", nil)
		return
	}

	utils.SendSuccessResponse(ctx, category, "Category retrieved successfully")
}

// CreateCategory creates a new Category
// @Summary Create Category
// @Description Create a new Category record
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category body models.Category true "Category data"
// @Success 201 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /category [post]
func (c *CategoryController) CreateCategory(ctx *gin.Context) {
	var category models.Category

	if err := ctx.ShouldBindJSON(&category); err != nil {
		utils.SendValidationErrorResponse(ctx, []utils.ValidationError{{Field: "input", Message: "Invalid input"}})
		return
	}

	if err := c.db.Create(&category).Error; err != nil {
		utils.SendErrorResponse(ctx, http.StatusInternalServerError, "Failed to create category", nil)
		return
	}

	utils.SendCreatedResponse(ctx, category, "Category created successfully")
}

// UpdateCategory updates a Category by ID
// @Summary Update Category
// @Description Update Category record by ID
// @Tags Category
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Category ID"
// @Param category body models.Category true "Category data"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /category/{id} [put]
func (c *CategoryController) UpdateCategory(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "Invalid ID", nil)
		return
	}

	var category models.Category
	if err := c.db.First(&category, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.SendNotFoundResponse(ctx, "Category not found")
			return
		}
		utils.SendErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch category", nil)
		return
	}

	if err := ctx.ShouldBindJSON(&category); err != nil {
		utils.SendValidationErrorResponse(ctx, []utils.ValidationError{{Field: "input", Message: "Invalid input"}})
		return
	}

	if err := c.db.Save(&category).Error; err != nil {
		utils.SendErrorResponse(ctx, http.StatusInternalServerError, "Failed to update category", nil)
		return
	}

	utils.SendSuccessResponse(ctx, category, "Category updated successfully")
}

// DeleteCategory deletes a Category by ID
// @Summary Delete Category
// @Description Delete Category record by ID
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /category/{id} [delete]
func (c *CategoryController) DeleteCategory(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendErrorResponse(ctx, http.StatusBadRequest, "Invalid ID", nil)
		return
	}

	if err := c.db.Delete(&models.Category{}, uint(id)).Error; err != nil {
		utils.SendErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete category", nil)
		return
	}

	utils.SendSuccessResponse(ctx, nil, "Category deleted successfully")
}
