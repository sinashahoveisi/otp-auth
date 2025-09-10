package entity

import (
	"time"
)

// OTP represents an OTP code in the system
type OTP struct {
	ID           int        `db:"id" json:"id"`
	PhoneNumber  string     `db:"phone_number" json:"phone_number" validate:"required,phone_number"`
	Code         string     `db:"code" json:"code"`
	SessionToken string     `db:"session_token" json:"session_token"`
	ExpiresAt    time.Time  `db:"expires_at" json:"expires_at"`
	IsUsed       bool       `db:"is_used" json:"is_used"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UsedAt       *time.Time `db:"used_at" json:"used_at"`
}

// TableName returns the table name for the OTP entity
func (OTP) TableName() string {
	return "otps"
}

// SendOTPRequest represents the request to send an OTP
type SendOTPRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,phone_number"`
}

// VerifyOTPRequest represents the request to verify an OTP
type VerifyOTPRequest struct {
	Token string `json:"token" validate:"required"`
	Code  string `json:"code" validate:"required,len=6"`
}

// OTPResponse represents the OTP response
type OTPResponse struct {
	Message     string    `json:"message"`
	Token       string    `json:"token"` // Session token for verification
	PhoneNumber string    `json:"phone_number"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// AuthResponse represents the authentication response with JWT token
type AuthResponse struct {
	Token     string       `json:"token"`
	User      UserResponse `json:"user"`
	ExpiresAt time.Time    `json:"expires_at"`
	Message   string       `json:"message"`
}

// RateLimitInfo represents rate limiting information for OTP requests
type RateLimitInfo struct {
	PhoneNumber   string    `db:"phone_number" bson:"phone_number" json:"phone_number"`
	RequestCount  int       `db:"request_count" bson:"request_count" json:"request_count"`
	LastRequestAt time.Time `db:"last_request_at" bson:"last_request_at" json:"last_request_at"`
	WindowStartAt time.Time `db:"window_start_at" bson:"window_start_at" json:"window_start_at"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updated_at"`
	ExpiresAt     time.Time `bson:"expires_at" json:"expires_at"`
}

// TableName returns the table name for the RateLimitInfo entity
func (RateLimitInfo) TableName() string {
	return "otp_rate_limits"
}
