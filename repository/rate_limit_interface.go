package repository

import (
	"otp-auth/entity"
	"time"
)

// RateLimitRepository interface defines rate limiting operations
// This can be implemented by both PostgreSQL and MongoDB repositories
type RateLimitRepository interface {
	GetRateLimit(phoneNumber string) (*entity.RateLimitInfo, error)
	UpdateRateLimit(rateLimitInfo *entity.RateLimitInfo) error
	CleanupRateLimits(olderThan time.Time) error
}
