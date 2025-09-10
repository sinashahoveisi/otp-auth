package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"otp-auth/config"
	"otp-auth/entity"
	"otp-auth/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService interface defines JWT operations
type JWTService interface {
	GenerateToken(user *entity.User) (*entity.AuthResponse, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	GetUserFromToken(token *jwt.Token) (*entity.User, error)
	RevokeToken(tokenString string) error
	RevokeAllUserTokens(userID int) error
}

// jwtService implements JWTService interface
type jwtService struct {
	cfg          *config.Config
	logger       *logger.Logger
	tokenService *TokenService
}

// JWTClaims represents the JWT claims
type JWTClaims struct {
	UserID      int    `json:"user_id"`
	PhoneNumber string `json:"phone_number"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service instance
func NewJWTService(cfg *config.Config, logger *logger.Logger, tokenService *TokenService) JWTService {
	return &jwtService{
		cfg:          cfg,
		logger:       logger,
		tokenService: tokenService,
	}
}

// GenerateToken generates a JWT token for the user
func (s *jwtService) GenerateToken(user *entity.User) (*entity.AuthResponse, error) {
	expiresAt := time.Now().Add(s.cfg.JWT.ExpirationTime)

	claims := JWTClaims{
		UserID:      user.ID,
		PhoneNumber: user.PhoneNumber,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "otp-auth-service",
			Subject:   fmt.Sprintf("user:%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		s.logger.Errorw("Failed to sign JWT token", "user_id", user.ID, "error", err)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Store token in Redis if token service is available
	if s.tokenService != nil {
		tokenHash := s.hashToken(tokenString)
		tokenInfo := &TokenInfo{
			UserID:    user.ID,
			TokenHash: tokenHash,
			IssuedAt:  time.Now(),
			ExpiresAt: expiresAt,
			LastUsed:  time.Now(),
		}

		if err := s.tokenService.StoreToken(tokenHash, tokenInfo, s.cfg.JWT.ExpirationTime); err != nil {
			s.logger.Warnw("Failed to store token in Redis", "user_id", user.ID, "error", err)
			// Don't fail token generation if Redis storage fails
		}
	}

	s.logger.Infow("JWT token generated", "user_id", user.ID, "expires_at", expiresAt)

	userResponse := &entity.UserResponse{
		ID:           user.ID,
		PhoneNumber:  user.PhoneNumber,
		RegisteredAt: user.RegisteredAt,
		LastLoginAt:  user.LastLoginAt,
		IsActive:     user.IsActive,
	}

	return &entity.AuthResponse{
		Token:     tokenString,
		User:      *userResponse,
		ExpiresAt: expiresAt,
		Message:   "Authentication successful",
	}, nil
}

// ValidateToken validates a JWT token
func (s *jwtService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		s.logger.Warnw("Failed to validate JWT token", "error", err)
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Verify token exists in Redis if token service is available
	if s.tokenService != nil {
		tokenHash := s.hashToken(tokenString)
		_, err := s.tokenService.ValidateToken(tokenHash)
		if err != nil {
			s.logger.Warnw("Token not found in Redis or expired", "error", err)
			return nil, fmt.Errorf("token session expired")
		}
	}

	return token, nil
}

// GetUserFromToken extracts user information from a validated JWT token
func (s *jwtService) GetUserFromToken(token *jwt.Token) (*entity.User, error) {
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	user := &entity.User{
		ID:          claims.UserID,
		PhoneNumber: claims.PhoneNumber,
		IsActive:    true, // Assuming active if token is valid
	}

	return user, nil
}

// RevokeToken revokes a specific token (logout)
func (s *jwtService) RevokeToken(tokenString string) error {
	if s.tokenService == nil {
		return fmt.Errorf("token service not available")
	}

	tokenHash := s.hashToken(tokenString)
	return s.tokenService.RevokeToken(tokenHash)
}

// RevokeAllUserTokens revokes all tokens for a user (logout from all devices)
func (s *jwtService) RevokeAllUserTokens(userID int) error {
	if s.tokenService == nil {
		return fmt.Errorf("token service not available")
	}

	return s.tokenService.RevokeAllUserTokens(userID)
}

// hashToken creates a hash of the token for storage in Redis
func (s *jwtService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
