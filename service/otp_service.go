package service

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"otp-auth/config"
	"otp-auth/entity"
	"otp-auth/pkg/logger"
	"otp-auth/repository"
)

// OTPService interface defines OTP business operations
type OTPService interface {
	SendOTP(phoneNumber string) (*entity.OTPResponse, error)
	VerifyOTP(sessionToken, code string) (*entity.User, error)
	IsRateLimited(phoneNumber string) (bool, error)
	CleanupExpiredOTPs() error
}

// otpService implements OTPService interface
type otpService struct {
	otpRepo       repository.OTPRepository
	userRepo      repository.UserRepository
	rateLimitRepo repository.RateLimitRepository
	cfg           *config.Config
	logger        *logger.Logger
}

// NewOTPService creates a new OTP service instance
func NewOTPService(otpRepo repository.OTPRepository, userRepo repository.UserRepository, rateLimitRepo repository.RateLimitRepository, cfg *config.Config, logger *logger.Logger) OTPService {
	return &otpService{
		otpRepo:       otpRepo,
		userRepo:      userRepo,
		rateLimitRepo: rateLimitRepo,
		cfg:           cfg,
		logger:        logger,
	}
}

// SendOTP generates and sends an OTP to the provided phone number
func (s *otpService) SendOTP(phoneNumber string) (*entity.OTPResponse, error) {
	// Check rate limiting
	isLimited, err := s.IsRateLimited(phoneNumber)
	if err != nil {
		s.logger.Errorw("Failed to check rate limit", "phone_number", phoneNumber, "error", err)
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if isLimited {
		return nil, fmt.Errorf("rate limit exceeded. Maximum %d requests per %v", s.cfg.RateLimit.MaxRequests, s.cfg.RateLimit.WindowDuration)
	}

	// Generate OTP code
	code, err := s.generateOTPCode()
	if err != nil {
		s.logger.Errorw("Failed to generate OTP code", "error", err)
		return nil, fmt.Errorf("failed to generate OTP code: %w", err)
	}

	// Generate session token
	sessionToken, err := s.generateSessionToken()
	if err != nil {
		s.logger.Errorw("Failed to generate session token", "error", err)
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create OTP entity
	otp := &entity.OTP{
		PhoneNumber:  phoneNumber,
		Code:         code,
		SessionToken: sessionToken,
		ExpiresAt:    time.Now().Add(s.cfg.OTP.ExpirationTime),
	}

	// Store OTP in database
	createdOTP, err := s.otpRepo.Create(otp)
	if err != nil {
		s.logger.Errorw("Failed to create OTP", "phone_number", phoneNumber, "error", err)
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}

	// Update rate limiting
	err = s.updateRateLimit(phoneNumber)
	if err != nil {
		s.logger.Errorw("Failed to update rate limit", "phone_number", phoneNumber, "error", err)
		// Don't return error as OTP was created successfully
	}

	// Print OTP to console (as per requirements)
	fmt.Printf("🔐 OTP for %s: %s (expires at %s)\n", phoneNumber, code, createdOTP.ExpiresAt.Format("15:04:05"))
	s.logger.Infow("OTP generated", "phone_number", phoneNumber, "expires_at", createdOTP.ExpiresAt)

	return &entity.OTPResponse{
		Message:     "OTP sent successfully",
		Token:       sessionToken,
		PhoneNumber: phoneNumber,
		ExpiresAt:   createdOTP.ExpiresAt,
	}, nil
}

// VerifyOTP verifies the provided OTP code using session token
func (s *otpService) VerifyOTP(sessionToken, code string) (*entity.User, error) {
	// Get active OTP by session token
	otp, err := s.otpRepo.GetActiveBySessionTokenAndCode(sessionToken, code)
	if err != nil {
		s.logger.Errorw("Failed to get OTP", "session_token", sessionToken, "error", err)
		return nil, fmt.Errorf("failed to verify OTP: %w", err)
	}

	if otp == nil {
		s.logger.Warnw("Invalid or expired OTP", "session_token", sessionToken, "code", code)
		return nil, fmt.Errorf("invalid or expired OTP")
	}

	// Mark OTP as used
	err = s.otpRepo.MarkAsUsed(otp.ID)
	if err != nil {
		s.logger.Errorw("Failed to mark OTP as used", "otp_id", otp.ID, "error", err)
		return nil, fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	// Get or create user using phone number from OTP record
	phoneNumber := otp.PhoneNumber
	user, err := s.userRepo.GetByPhoneNumber(phoneNumber)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Errorw("Failed to get user", "phone_number", phoneNumber, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil || errors.Is(err, sql.ErrNoRows) {
		// Create new user (last_login_at will be set to registered_at in repository)
		newUser := &entity.User{
			PhoneNumber: phoneNumber,
		}
		user, err = s.userRepo.Create(newUser)
		if err != nil {
			s.logger.Errorw("Failed to create user", "phone_number", phoneNumber, "error", err)
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		s.logger.Infow("New user registered", "user_id", user.ID, "phone_number", phoneNumber)
	} else {
		// Update last login for existing user
		if err := s.userRepo.UpdateLastLogin(phoneNumber); err != nil {
			s.logger.Errorw("Failed to update last login", "phone_number", phoneNumber, "error", err)
			return nil, fmt.Errorf("failed to update last login: %w", err)
		}
		s.logger.Infow("User logged in", "user_id", user.ID, "phone_number", phoneNumber)
	}

	return user, nil
}

// IsRateLimited checks if the phone number has exceeded the rate limit
func (s *otpService) IsRateLimited(phoneNumber string) (bool, error) {
	rateLimitInfo, err := s.rateLimitRepo.GetRateLimit(phoneNumber)
	if err != nil {
		return false, fmt.Errorf("failed to get rate limit info: %w", err)
	}

	if rateLimitInfo == nil {
		// No previous requests, not rate limited
		return false, nil
	}

	now := time.Now()
	windowStart := rateLimitInfo.WindowStartAt
	windowDuration := s.cfg.RateLimit.WindowDuration

	// Check if we're still in the same window
	if now.Sub(windowStart) >= windowDuration {
		// Window has expired, reset the counter
		return false, nil
	}

	// Check if request count exceeds limit
	return rateLimitInfo.RequestCount >= s.cfg.RateLimit.MaxRequests, nil
}

// updateRateLimit updates the rate limiting information
func (s *otpService) updateRateLimit(phoneNumber string) error {
	rateLimitInfo, err := s.rateLimitRepo.GetRateLimit(phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to get rate limit info: %w", err)
	}

	now := time.Now()

	if rateLimitInfo == nil {
		// First request
		rateLimitInfo = &entity.RateLimitInfo{
			PhoneNumber:   phoneNumber,
			RequestCount:  1,
			LastRequestAt: now,
			WindowStartAt: now,
		}
	} else {
		// Check if we need to reset the window
		if now.Sub(rateLimitInfo.WindowStartAt) >= s.cfg.RateLimit.WindowDuration {
			// Reset window
			rateLimitInfo.RequestCount = 1
			rateLimitInfo.WindowStartAt = now
		} else {
			// Increment counter
			rateLimitInfo.RequestCount++
		}
		rateLimitInfo.LastRequestAt = now
	}

	return s.rateLimitRepo.UpdateRateLimit(rateLimitInfo)
}

// generateOTPCode generates a random OTP code
func (s *otpService) generateOTPCode() (string, error) {
	maxValue := big.NewInt(1)
	for i := 0; i < s.cfg.OTP.Length; i++ {
		maxValue.Mul(maxValue, big.NewInt(10))
	}

	randomNumber, err := rand.Int(rand.Reader, maxValue)
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	// Format with leading zeros
	format := fmt.Sprintf("%%0%dd", s.cfg.OTP.Length)
	return fmt.Sprintf(format, randomNumber), nil
}

// generateSessionToken generates a random session token
func (s *otpService) generateSessionToken() (string, error) {
	// Generate 32 random bytes for session token
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}

	// Encode as hex string
	return fmt.Sprintf("%x", bytes), nil
}

// CleanupExpiredOTPs removes expired OTPs and old rate limit records
func (s *otpService) CleanupExpiredOTPs() error {
	// Delete expired OTPs
	if err := s.otpRepo.DeleteExpired(); err != nil {
		s.logger.Errorw("Failed to delete expired OTPs", "error", err)
		return fmt.Errorf("failed to delete expired OTPs: %w", err)
	}

	// Cleanup old rate limit records (older than 24 hours)
	olderThan := time.Now().Add(-24 * time.Hour)
	if err := s.rateLimitRepo.CleanupRateLimits(olderThan); err != nil {
		s.logger.Errorw("Failed to cleanup rate limits", "error", err)
		return fmt.Errorf("failed to cleanup rate limits: %w", err)
	}

	return nil
}
