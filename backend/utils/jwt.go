package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, email string) (string, error) {
	return GenerateTokenPair(userID, email)
}

func GenerateTokenPair(userID uint, email string) (string, error) {
	accessToken, _, err := GenerateAccessAndRefreshTokens(userID, email)
	return accessToken, err
}

func GenerateAccessAndRefreshTokens(userID uint, email string) (string, string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", "", errors.New("JWT_SECRET environment variable is required")
	}

	// Generate access token (15 minutes)
	accessClaims := Claims{
		UserID: userID,
		Email:  email,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	// Generate refresh token (7 days)
	refreshClaims := Claims{
		UserID: userID,
		Email:  email,
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

func ValidateToken(tokenString string) (*Claims, error) {
	return ValidateTokenWithType(tokenString, "access")
}

func ValidateRefreshToken(tokenString string) (*Claims, error) {
	return ValidateTokenWithType(tokenString, "refresh")
}

func ValidateTokenWithType(tokenString string, expectedType string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET environment variable is required")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.Type != expectedType {
			return nil, errors.New("invalid token type")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
