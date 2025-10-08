package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"mobile-backend/models"
	"mobile-backend/utils"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type OAuth2Provider string

const (
	GoogleProvider OAuth2Provider = "google"
	GitHubProvider OAuth2Provider = "github"
)

type OAuth2Service struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewOAuth2Service(db *gorm.DB, redis *redis.Client) *OAuth2Service {
	return &OAuth2Service{db: db, redis: redis}
}

type OAuth2User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Provider string `json:"provider"`
}

type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

func (o *OAuth2Service) GetConfig(provider OAuth2Provider) (*OAuth2Config, error) {
	switch provider {
	case GoogleProvider:
		return &OAuth2Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			AuthURL:      "https://accounts.google.com/o/oauth2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserInfoURL:  "https://www.googleapis.com/oauth2/v2/userinfo",
		}, nil
	case GitHubProvider:
		return &OAuth2Config{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			UserInfoURL:  "https://api.github.com/user",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported OAuth2 provider: %s", provider)
	}
}

func (o *OAuth2Service) GetAuthURL(provider OAuth2Provider, state string) (string, error) {
	config, err := o.GetConfig(provider)
	if err != nil {
		return "", err
	}

	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=email profile&state=%s",
		config.AuthURL, config.ClientID, config.RedirectURL, state)

	return authURL, nil
}

func (o *OAuth2Service) ExchangeCodeForToken(provider OAuth2Provider, code string) (string, error) {
	config, err := o.GetConfig(provider)
	if err != nil {
		return "", err
	}

	// Exchange code for access token
	tokenReq := map[string]string{
		"client_id":     config.ClientID,
		"client_secret": config.ClientSecret,
		"code":          code,
		"redirect_uri":  config.RedirectURL,
		"grant_type":    "authorization_code",
	}

	// This is a simplified implementation
	// In production, you'd use a proper OAuth2 library
	req, err := http.NewRequest("POST", config.TokenURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	return tokenResp.AccessToken, nil
}

func (o *OAuth2Service) GetUserInfo(provider OAuth2Provider, accessToken string) (*OAuth2User, error) {
	config, err := o.GetConfig(provider)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", config.UserInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo OAuth2User
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	userInfo.Provider = string(provider)
	return &userInfo, nil
}

func (o *OAuth2Service) CreateOrUpdateUser(ctx context.Context, oauthUser *OAuth2User) (*models.User, error) {
	var user models.User

	// Try to find existing user by email
	err := o.db.Where("email = ?", oauthUser.Email).First(&user).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			Email:    oauthUser.Email,
			Name:     oauthUser.Name,
			IsActive: true,
		}
		// Generate a random password for OAuth users
		randomPassword := utils.GenerateRandomString(32)
		if err := user.HashPassword(randomPassword); err != nil {
			return nil, err
		}

		if err := o.db.Create(&user).Error; err != nil {
			return nil, err
		}
	} else {
		// Update existing user
		user.Name = oauthUser.Name
		user.IsActive = true
		if err := o.db.Save(&user).Error; err != nil {
			return nil, err
		}
	}

	return &user, nil
}

func (o *OAuth2Service) StoreOAuthState(ctx context.Context, state string, provider OAuth2Provider) error {
	return o.redis.Set(ctx, "oauth_state:"+state, string(provider), 10*time.Minute).Err()
}

func (o *OAuth2Service) ValidateOAuthState(ctx context.Context, state string) (OAuth2Provider, error) {
	result := o.redis.Get(ctx, "oauth_state:"+state)
	if result.Err() == redis.Nil {
		return "", fmt.Errorf("invalid or expired state")
	}
	if result.Err() != nil {
		return "", result.Err()
	}

	provider := OAuth2Provider(result.Val())
	o.redis.Del(ctx, "oauth_state:"+state) // Remove used state
	return provider, nil
}
