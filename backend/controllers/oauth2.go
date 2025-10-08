package controllers

import (
	"net/http"

	"mobile-backend/services"
	"mobile-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OAuth2Controller struct {
	oauth2Service *services.OAuth2Service
	authService   *services.AuthService
}

func NewOAuth2Controller(oauth2Service *services.OAuth2Service, authService *services.AuthService) *OAuth2Controller {
	return &OAuth2Controller{
		oauth2Service: oauth2Service,
		authService:   authService,
	}
}

// OAuth2Login godoc
// @Summary OAuth2 login
// @Description Initiate OAuth2 login with specified provider
// @Tags auth
// @Accept json
// @Produce json
// @Param provider path string true "OAuth2 provider" Enums(google, github)
// @Success 200 {object} map[string]string
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/auth/oauth2/{provider} [get]
func (oc *OAuth2Controller) OAuth2Login(c *gin.Context) {
	provider := services.OAuth2Provider(c.Param("provider"))

	// Validate provider
	if provider != services.GoogleProvider && provider != services.GitHubProvider {
		utils.SendErrorResponse(c, "Unsupported OAuth2 provider", http.StatusBadRequest)
		return
	}

	// Generate state parameter for security
	state := uuid.New().String()

	// Store state in Redis
	if err := oc.oauth2Service.StoreOAuthState(c.Request.Context(), state, provider); err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to store OAuth state")
		return
	}

	// Get authorization URL
	authURL, err := oc.oauth2Service.GetAuthURL(provider, state)
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to generate OAuth URL")
		return
	}

	utils.SendSuccessResponse(c, gin.H{
		"auth_url": authURL,
		"state":    state,
	}, "OAuth2 login initiated")
}

// OAuth2Callback godoc
// @Summary OAuth2 callback
// @Description Handle OAuth2 callback and complete authentication
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "Authorization code"
// @Param state query string true "State parameter"
// @Success 200 {object} utils.SuccessResponse{data=utils.LoginResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/auth/oauth2/callback [get]
func (oc *OAuth2Controller) OAuth2Callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		utils.SendErrorResponse(c, "Missing code or state parameter", http.StatusBadRequest)
		return
	}

	// Validate state
	provider, err := oc.oauth2Service.ValidateOAuthState(c.Request.Context(), state)
	if err != nil {
		utils.SendErrorResponse(c, "Invalid or expired state", http.StatusBadRequest)
		return
	}

	// Exchange code for access token
	accessToken, err := oc.oauth2Service.ExchangeCodeForToken(provider, code)
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to exchange code for token")
		return
	}

	// Get user info from OAuth2 provider
	oauthUser, err := oc.oauth2Service.GetUserInfo(provider, accessToken)
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to get user info")
		return
	}

	// Create or update user in database
	user, err := oc.oauth2Service.CreateOrUpdateUser(c.Request.Context(), oauthUser)
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to create/update user")
		return
	}

	// Generate JWT tokens
	accessTokenJWT, refreshToken, err := utils.GenerateAccessAndRefreshTokens(user.ID, user.Email)
	if err != nil {
		utils.SendInternalServerErrorResponse(c, "Failed to generate tokens")
		return
	}

	userResponse := utils.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IsActive:  user.IsActive,
		CreatedAt: utils.FormatTime(user.CreatedAt),
		UpdatedAt: utils.FormatTime(user.UpdatedAt),
	}

	loginResponse := utils.LoginResponse{
		User:         userResponse,
		AccessToken:  accessTokenJWT,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}

	utils.SendSuccessResponse(c, loginResponse, "OAuth2 login successful")
}

// OAuth2Providers godoc
// @Summary Get available OAuth2 providers
// @Description Get list of available OAuth2 providers
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} utils.SuccessResponse{data=[]string}
// @Router /api/v1/auth/oauth2/providers [get]
func (oc *OAuth2Controller) OAuth2Providers(c *gin.Context) {
	providers := []string{
		string(services.GoogleProvider),
		string(services.GitHubProvider),
	}

	utils.SendSuccessResponse(c, providers, "Available OAuth2 providers")
}
