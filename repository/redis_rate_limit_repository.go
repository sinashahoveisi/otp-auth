package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"otp-auth/config"
	"otp-auth/entity"
	"otp-auth/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimitRepository implements rate limiting using Redis
type RedisRateLimitRepository struct {
	client *redis.Client
	ctx    context.Context
	config *config.Config
	logger *logger.Logger
}

// NewRedisRateLimitRepository creates a new Redis rate limit repository
func NewRedisRateLimitRepository(client *redis.Client, cfg *config.Config, logger *logger.Logger) RateLimitRepository {
	return &RedisRateLimitRepository{
		client: client,
		ctx:    context.Background(),
		config: cfg,
		logger: logger,
	}
}

// GetRateLimit retrieves rate limit information for a phone number
func (r *RedisRateLimitRepository) GetRateLimit(phoneNumber string) (*entity.RateLimitInfo, error) {
	key := fmt.Sprintf("rate_limit:%s", phoneNumber)

	// Use pipeline to get both data and TTL in one round trip
	pipe := r.client.Pipeline()
	dataCmd := pipe.Get(r.ctx, key)
	ttlCmd := pipe.TTL(r.ctx, key)
	_, err := pipe.Exec(r.ctx)

	data, err := dataCmd.Result()
	if err == redis.Nil {
		// No existing rate limit record
		r.logger.Debugw("No rate limit record found", "phone_number", phoneNumber)
		return &entity.RateLimitInfo{
			PhoneNumber:   phoneNumber,
			RequestCount:  0,
			LastRequestAt: time.Time{},
			WindowStartAt: time.Time{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rate limit info: %w", err)
	}

	var rateLimitInfo entity.RateLimitInfo
	if err := json.Unmarshal([]byte(data), &rateLimitInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rate limit info: %w", err)
	}

	// Get TTL for logging
	ttl, _ := ttlCmd.Result()
	r.logger.Debugw("Rate limit retrieved",
		"phone_number", phoneNumber,
		"request_count", rateLimitInfo.RequestCount,
		"ttl_seconds", int(ttl.Seconds()))

	return &rateLimitInfo, nil
}

// UpdateRateLimit updates rate limit information for a phone number
func (r *RedisRateLimitRepository) UpdateRateLimit(rateLimitInfo *entity.RateLimitInfo) error {
	key := fmt.Sprintf("rate_limit:%s", rateLimitInfo.PhoneNumber)

	// Calculate TTL based on rate limit window duration from config
	now := time.Now()
	windowDuration := r.config.RateLimit.WindowDuration

	// If this is a new rate limit window, set WindowStartAt
	if rateLimitInfo.WindowStartAt.IsZero() {
		rateLimitInfo.WindowStartAt = now
	}

	// Calculate remaining time in the current window
	windowEnd := rateLimitInfo.WindowStartAt.Add(windowDuration)
	ttl := windowEnd.Sub(now)

	// Ensure TTL is positive and not less than 1 minute
	if ttl <= 0 {
		// Start a new window if current window has expired
		rateLimitInfo.WindowStartAt = now
		rateLimitInfo.RequestCount = 1
		ttl = windowDuration
		r.logger.Debugw("Starting new rate limit window",
			"phone_number", rateLimitInfo.PhoneNumber,
			"window_start", rateLimitInfo.WindowStartAt,
			"ttl_seconds", int(ttl.Seconds()))
	} else if ttl < time.Minute {
		// Extend TTL to at least 1 minute for better cleanup
		ttl = time.Minute
	}

	data, err := json.Marshal(rateLimitInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal rate limit info: %w", err)
	}

	// Set the key with calculated TTL
	err = r.client.Set(r.ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to update rate limit info: %w", err)
	}

	r.logger.Debugw("Rate limit updated with TTL",
		"phone_number", rateLimitInfo.PhoneNumber,
		"request_count", rateLimitInfo.RequestCount,
		"ttl_seconds", int(ttl.Seconds()),
		"expires_at", now.Add(ttl).Format(time.RFC3339))

	return nil
}

// CleanupRateLimits cleans up expired rate limits (Redis handles this automatically with TTL)
func (r *RedisRateLimitRepository) CleanupRateLimits(olderThan time.Time) error {
	// Redis automatically handles cleanup with TTL, so this is mostly a no-op
	// But we can use this to check for any keys that might not have TTL set

	pattern := "rate_limit:*"
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get rate limit keys: %w", err)
	}

	cleanedCount := 0
	for _, key := range keys {
		ttl, err := r.client.TTL(r.ctx, key).Result()
		if err != nil {
			r.logger.Warnw("Failed to get TTL for key", "key", key, "error", err)
			continue
		}

		// If TTL is -1, it means the key has no expiration set
		if ttl == -1 {
			// Set a default TTL for keys without expiration
			defaultTTL := r.config.RateLimit.WindowDuration
			err := r.client.Expire(r.ctx, key, defaultTTL).Err()
			if err != nil {
				r.logger.Warnw("Failed to set TTL for key", "key", key, "error", err)
			} else {
				r.logger.Infow("Set missing TTL for rate limit key", "key", key, "ttl_seconds", int(defaultTTL.Seconds()))
				cleanedCount++
			}
		}
	}

	if cleanedCount > 0 {
		r.logger.Infow("Rate limit cleanup completed", "keys_with_ttl_added", cleanedCount)
	}

	return nil
}
