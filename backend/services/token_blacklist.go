package services

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type TokenBlacklistService struct {
	redis *redis.Client
}

func NewTokenBlacklistService(redis *redis.Client) *TokenBlacklistService {
	return &TokenBlacklistService{redis: redis}
}

// AddToBlacklist adds a token to the blacklist
func (t *TokenBlacklistService) AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error {
	return t.redis.Set(ctx, "blacklist:"+token, "1", expiration).Err()
}

// IsBlacklisted checks if a token is blacklisted
func (t *TokenBlacklistService) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	result := t.redis.Get(ctx, "blacklist:"+token)
	if result.Err() == redis.Nil {
		return false, nil
	}
	if result.Err() != nil {
		return false, result.Err()
	}
	return true, nil
}

// RemoveFromBlacklist removes a token from the blacklist
func (t *TokenBlacklistService) RemoveFromBlacklist(ctx context.Context, token string) error {
	return t.redis.Del(ctx, "blacklist:"+token).Err()
}

// CleanupExpiredTokens removes expired tokens from blacklist
func (t *TokenBlacklistService) CleanupExpiredTokens(ctx context.Context) error {
	// Redis automatically removes expired keys, but we can add custom cleanup logic here
	return nil
}
