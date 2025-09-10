package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"otp-auth/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// TokenInfo stores token metadata in Redis
type TokenInfo struct {
	UserID    int       `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
	LastUsed  time.Time `json:"last_used"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// TokenService handles token storage and management in Redis
type TokenService struct {
	redis  *redis.Client
	logger *logger.Logger
	ctx    context.Context
}

// NewTokenService creates a new token service
func NewTokenService(redis *redis.Client, logger *logger.Logger) *TokenService {
	return &TokenService{
		redis:  redis,
		logger: logger,
		ctx:    context.Background(),
	}
}

// StoreToken stores token information in Redis
func (s *TokenService) StoreToken(tokenHash string, tokenInfo *TokenInfo, expiration time.Duration) error {
	key := fmt.Sprintf("token:%s", tokenHash)

	data, err := json.Marshal(tokenInfo)
	if err != nil {
		s.logger.Errorw("Failed to marshal token info", "error", err)
		return fmt.Errorf("failed to marshal token info: %w", err)
	}

	err = s.redis.Set(s.ctx, key, data, expiration).Err()
	if err != nil {
		s.logger.Errorw("Failed to store token in Redis", "token_hash", tokenHash, "error", err)
		return fmt.Errorf("failed to store token in Redis: %w", err)
	}

	// Also store user's active tokens list
	userKey := fmt.Sprintf("user_tokens:%d", tokenInfo.UserID)
	err = s.redis.SAdd(s.ctx, userKey, tokenHash).Err()
	if err != nil {
		s.logger.Warnw("Failed to add token to user's active tokens list", "user_id", tokenInfo.UserID, "error", err)
	}

	// Set expiration for user tokens list (a bit longer than token expiration)
	s.redis.Expire(s.ctx, userKey, expiration+time.Hour)

	s.logger.Infow("Token stored successfully", "user_id", tokenInfo.UserID, "token_hash", tokenHash[:8]+"...")
	return nil
}

// ValidateToken checks if token exists and is valid in Redis
func (s *TokenService) ValidateToken(tokenHash string) (*TokenInfo, error) {
	key := fmt.Sprintf("token:%s", tokenHash)

	data, err := s.redis.Get(s.ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("token not found or expired")
	}
	if err != nil {
		s.logger.Errorw("Failed to get token from Redis", "token_hash", tokenHash, "error", err)
		return nil, fmt.Errorf("failed to get token from Redis: %w", err)
	}

	var tokenInfo TokenInfo
	if err := json.Unmarshal([]byte(data), &tokenInfo); err != nil {
		s.logger.Errorw("Failed to unmarshal token info", "error", err)
		return nil, fmt.Errorf("failed to unmarshal token info: %w", err)
	}

	// Update last used timestamp
	tokenInfo.LastUsed = time.Now()
	s.updateTokenLastUsed(tokenHash, &tokenInfo)

	return &tokenInfo, nil
}

// updateTokenLastUsed updates the last used timestamp (async)
func (s *TokenService) updateTokenLastUsed(tokenHash string, tokenInfo *TokenInfo) {
	go func() {
		key := fmt.Sprintf("token:%s", tokenHash)
		data, err := json.Marshal(tokenInfo)
		if err != nil {
			s.logger.Warnw("Failed to marshal token info for last used update", "error", err)
			return
		}

		// Get current TTL and preserve it
		ttl := s.redis.TTL(s.ctx, key).Val()
		if ttl > 0 {
			s.redis.Set(s.ctx, key, data, ttl)
		}
	}()
}

// RevokeToken removes token from Redis (logout)
func (s *TokenService) RevokeToken(tokenHash string) error {
	key := fmt.Sprintf("token:%s", tokenHash)

	// Get token info to find user ID
	tokenInfo, err := s.ValidateToken(tokenHash)
	if err == nil {
		// Remove from user's active tokens
		userKey := fmt.Sprintf("user_tokens:%d", tokenInfo.UserID)
		s.redis.SRem(s.ctx, userKey, tokenHash)
	}

	err = s.redis.Del(s.ctx, key).Err()
	if err != nil {
		s.logger.Errorw("Failed to revoke token", "token_hash", tokenHash, "error", err)
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	s.logger.Infow("Token revoked successfully", "token_hash", tokenHash[:8]+"...")
	return nil
}

// RevokeAllUserTokens revokes all tokens for a user
func (s *TokenService) RevokeAllUserTokens(userID int) error {
	userKey := fmt.Sprintf("user_tokens:%d", userID)

	// Get all user's tokens
	tokenHashes, err := s.redis.SMembers(s.ctx, userKey).Result()
	if err != nil {
		s.logger.Errorw("Failed to get user tokens", "user_id", userID, "error", err)
		return fmt.Errorf("failed to get user tokens: %w", err)
	}

	// Delete each token
	pipe := s.redis.Pipeline()
	for _, tokenHash := range tokenHashes {
		key := fmt.Sprintf("token:%s", tokenHash)
		pipe.Del(s.ctx, key)
	}

	// Delete user tokens set
	pipe.Del(s.ctx, userKey)

	_, err = pipe.Exec(s.ctx)
	if err != nil {
		s.logger.Errorw("Failed to revoke all user tokens", "user_id", userID, "error", err)
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}

	s.logger.Infow("All user tokens revoked", "user_id", userID, "token_count", len(tokenHashes))
	return nil
}

// GetUserActiveTokens returns list of active tokens for a user
func (s *TokenService) GetUserActiveTokens(userID int) ([]TokenInfo, error) {
	userKey := fmt.Sprintf("user_tokens:%d", userID)

	tokenHashes, err := s.redis.SMembers(s.ctx, userKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get user tokens: %w", err)
	}

	var tokens []TokenInfo
	for _, tokenHash := range tokenHashes {
		tokenInfo, err := s.ValidateToken(tokenHash)
		if err == nil {
			tokens = append(tokens, *tokenInfo)
		}
	}

	return tokens, nil
}
