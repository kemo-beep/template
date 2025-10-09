package controllers

import (
	"net/http"
	"strconv"

	"mobile-backend/services"
	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	authService *services.AuthService
}

func NewUserController(authService *services.AuthService) *UserController {
	return &UserController{authService: authService}
}

type UpdateUserRequest struct {
	Name string `json:"name" binding:"omitempty,min=2"`
}

func (uc *UserController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorizedResponse(c, "User not authenticated")
		return
	}

	user, err := uc.authService.GetUserByID(userID.(uint))
	if err != nil {
		utils.SendNotFoundResponse(c, "User not found")
		return
	}

	userResponse := utils.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		// Add subscription status fields
		SubscriptionStatus:   user.SubscriptionStatus,
		IsPro:                user.IsPro,
		SubscriptionEndsAt:   user.SubscriptionEndsAt,
		TrialEndsAt:          user.TrialEndsAt,
		StatusDisplay:        user.GetSubscriptionStatus(),
		HasTrialAccess:       user.HasTrialAccess(),
		IsSubscriptionActive: user.IsSubscriptionActive(),
	}

	utils.SendSuccessResponse(c, userResponse, "Profile retrieved successfully")
}

func (uc *UserController) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorizedResponse(c, "User not authenticated")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationErrorResponse(c, parseValidationErrors(err))
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}

	user, err := uc.authService.UpdateUser(userID.(uint), updates)
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to update profile")
		return
	}

	userResponse := utils.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	utils.SendSuccessResponse(c, userResponse, "Profile updated successfully")
}

func (uc *UserController) DeleteProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.SendUnauthorizedResponse(c, "User not authenticated")
		return
	}

	if err := uc.authService.DeleteUser(userID.(uint)); err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to delete profile")
		return
	}

	utils.SendSuccessResponse(c, nil, "Profile deleted successfully")
}

func (uc *UserController) GetUserByID(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	user, err := uc.authService.GetUserByID(uint(userID))
	if err != nil {
		utils.SendNotFoundResponse(c, "User not found")
		return
	}

	userResponse := utils.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	utils.SendSuccessResponse(c, userResponse, "User retrieved successfully")
}
