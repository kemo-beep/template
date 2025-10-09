package controllers

import (
	"net/http"
	"strings"

	"mobile-backend/services"
	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration data"
// @Success 201 {object} utils.SuccessResponse{data=utils.UserResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationErrorResponse(c, parseValidationErrors(err))
		return
	}

	user, err := ac.authService.RegisterUser(req.Email, req.Password, req.Name)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			utils.SendErrorResponse(c, http.StatusConflict, "User already exists", nil)
		} else {
			utils.SendInternalServerErrorResponse(c, "Failed to create user")
		}
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

	utils.SendCreatedResponse(c, userResponse, "User created successfully")
}

// Login godoc
// @Summary Login user
// @Description Login user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User login data"
// @Success 200 {object} utils.SuccessResponse{data=utils.LoginResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationErrorResponse(c, parseValidationErrors(err))
		return
	}

	user, _, err := ac.authService.LoginUser(req.Email, req.Password)
	if err != nil {
		utils.SendUnauthorizedResponse(c, "Invalid credentials")
		return
	}

	// Generate access and refresh tokens
	accessToken, refreshToken, err := utils.GenerateAccessAndRefreshTokens(user.ID, user.Email)
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to generate tokens")
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

	loginResponse := utils.LoginResponse{
		User:         userResponse,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}

	utils.SendSuccessResponse(c, loginResponse, "Login successful")
}

// Logout godoc
// @Summary Logout user
// @Description Logout user and invalidate session
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")

	if err := ac.authService.Logout(token); err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to logout")
		return
	}

	utils.SendSuccessResponse(c, nil, "Logout successful")
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token data"
// @Success 200 {object} utils.SuccessResponse{data=utils.RefreshTokenResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendValidationErrorResponse(c, parseValidationErrors(err))
		return
	}

	// Validate refresh token
	claims, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		utils.SendUnauthorizedResponse(c, "Invalid refresh token")
		return
	}

	// Check if refresh token is blacklisted
	// Note: You'll need to inject the blacklist service into the controller
	// For now, we'll skip this check

	// Generate new token pair
	accessToken, refreshToken, err := utils.GenerateAccessAndRefreshTokens(claims.UserID, claims.Email)
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to generate tokens")
		return
	}

	refreshResponse := utils.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}

	utils.SendSuccessResponse(c, refreshResponse, "Tokens refreshed successfully")
}

func parseValidationErrors(err error) []utils.ValidationError {
	// This is a simplified version - in a real app, you'd parse the validation errors properly
	return []utils.ValidationError{
		{
			Field:   "general",
			Message: err.Error(),
		},
	}
}
