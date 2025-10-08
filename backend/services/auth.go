package services

import (
	"context"
	"errors"
	"time"

	"mobile-backend/models"
	"mobile-backend/utils"

	"gorm.io/gorm"
)

type AuthService struct {
	db    *gorm.DB
	cache *CacheService
}

func NewAuthService(db *gorm.DB, cache *CacheService) *AuthService {
	return &AuthService{
		db:    db,
		cache: cache,
	}
}

func (s *AuthService) RegisterUser(email, password, name string) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user already exists")
	}

	// Create new user
	user := &models.User{
		Email:    email,
		Password: password, // Will be hashed by GORM BeforeCreate hook
		Name:     name,
		IsActive: true,
	}

	// Save user to database
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) LoginUser(email, password string) (*models.User, string, error) {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", errors.New("invalid credentials")
		}
		return nil, "", err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, "", errors.New("account is deactivated")
	}

	// Verify password
	if err := user.CheckPassword(password); err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// Update last login
	now := time.Now()
	user.LastLogin = &now
	s.db.Save(&user)

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, "", err
	}

	// Store session in cache
	sessionKey := "session:" + token
	sessionData := map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"expires": time.Now().Add(72 * time.Hour).Unix(),
	}
	s.cache.Set(context.Background(), sessionKey, sessionData, 72*time.Hour)

	return &user, token, nil
}

func (s *AuthService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *AuthService) UpdateUser(userID uint, updates map[string]interface{}) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) DeleteUser(userID uint) error {
	return s.db.Delete(&models.User{}, userID).Error
}

func (s *AuthService) ValidateToken(token string) (*models.User, error) {
	claims, err := utils.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Check if session exists in cache
	sessionKey := "session:" + token
	var sessionData map[string]interface{}
	if err := s.cache.Get(context.Background(), sessionKey, &sessionData); err != nil {
		return nil, errors.New("session expired")
	}

	// Get user from database
	user, err := s.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Logout(token string) error {
	sessionKey := "session:" + token
	return s.cache.Delete(context.Background(), sessionKey)
}
