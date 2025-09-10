package repository

import (
	"database/sql"
	"fmt"
	"time"

	"otp-auth/entity"

	"github.com/jmoiron/sqlx"
)

// OTPRepository interface defines OTP data operations
type OTPRepository interface {
	Create(otp *entity.OTP) (*entity.OTP, error)
	GetActiveByPhoneNumberAndCode(phoneNumber, code string) (*entity.OTP, error)
	GetActiveBySessionTokenAndCode(sessionToken, code string) (*entity.OTP, error)
	MarkAsUsed(id int) error
	DeleteExpired() error
}

// otpRepository implements OTPRepository interface
type otpRepository struct {
	db *sqlx.DB
}

// NewOTPRepository creates a new OTP repository instance
func NewOTPRepository(db *sqlx.DB) OTPRepository {
	return &otpRepository{
		db: db,
	}
}

// Create creates a new OTP
func (r *otpRepository) Create(otp *entity.OTP) (*entity.OTP, error) {
	query := `
		INSERT INTO otps (phone_number, code, session_token, expires_at, is_used, created_at)
		VALUES (:phone_number, :code, :session_token, :expires_at, :is_used, :created_at)
		RETURNING id, phone_number, code, session_token, expires_at, is_used, created_at, used_at
	`

	otp.CreatedAt = time.Now()
	otp.IsUsed = false

	rows, err := r.db.NamedQuery(query, otp)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTP: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("failed to get created OTP")
	}

	var createdOTP entity.OTP
	if err := rows.StructScan(&createdOTP); err != nil {
		return nil, fmt.Errorf("failed to scan created OTP: %w", err)
	}

	return &createdOTP, nil
}

// GetActiveByPhoneNumberAndCode retrieves an active OTP by phone number and code
func (r *otpRepository) GetActiveByPhoneNumberAndCode(phoneNumber, code string) (*entity.OTP, error) {
	query := `
		SELECT id, phone_number, code, session_token, expires_at, is_used, created_at, used_at
		FROM otps
		WHERE phone_number = $1 AND code = $2 AND is_used = FALSE AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC
		LIMIT 1
	`

	var otp entity.OTP
	err := r.db.Get(&otp, query, phoneNumber, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get OTP: %w", err)
	}

	return &otp, nil
}

// GetActiveBySessionTokenAndCode retrieves an active OTP by session token and code
func (r *otpRepository) GetActiveBySessionTokenAndCode(sessionToken, code string) (*entity.OTP, error) {
	query := `
		SELECT id, phone_number, code, session_token, expires_at, is_used, created_at, used_at
		FROM otps
		WHERE session_token = $1 AND code = $2 AND is_used = FALSE AND expires_at > CURRENT_TIMESTAMP
		ORDER BY created_at DESC
		LIMIT 1
	`

	var otp entity.OTP
	err := r.db.Get(&otp, query, sessionToken, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get OTP by session token: %w", err)
	}

	return &otp, nil
}

// MarkAsUsed marks an OTP as used
func (r *otpRepository) MarkAsUsed(id int) error {
	query := `
		UPDATE otps
		SET is_used = TRUE, used_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND is_used = FALSE
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to mark OTP as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("OTP not found or already used")
	}

	return nil
}

// DeleteExpired deletes expired OTPs
func (r *otpRepository) DeleteExpired() error {
	query := `DELETE FROM otps WHERE expires_at < CURRENT_TIMESTAMP`

	result, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete expired OTPs: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected > 0 {
		fmt.Printf("Deleted %d expired OTPs\n", rowsAffected)
	}

	return nil
}

// GetRateLimit retrieves rate limit information for a phone number
func (r *otpRepository) GetRateLimit(phoneNumber string) (*entity.RateLimitInfo, error) {
	query := `
		SELECT phone_number, request_count, last_request_at, window_start_at
		FROM otp_rate_limits
		WHERE phone_number = $1
	`

	var rateLimitInfo entity.RateLimitInfo
	err := r.db.Get(&rateLimitInfo, query, phoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get rate limit info: %w", err)
	}

	return &rateLimitInfo, nil
}

// UpdateRateLimit updates or creates rate limit information
func (r *otpRepository) UpdateRateLimit(rateLimitInfo *entity.RateLimitInfo) error {
	query := `
		INSERT INTO otp_rate_limits (phone_number, request_count, last_request_at, window_start_at)
		VALUES (:phone_number, :request_count, :last_request_at, :window_start_at)
		ON CONFLICT (phone_number)
		DO UPDATE SET
			request_count = EXCLUDED.request_count,
			last_request_at = EXCLUDED.last_request_at,
			window_start_at = EXCLUDED.window_start_at
	`

	_, err := r.db.NamedExec(query, rateLimitInfo)
	if err != nil {
		return fmt.Errorf("failed to update rate limit: %w", err)
	}

	return nil
}

// CleanupRateLimits removes old rate limit records
func (r *otpRepository) CleanupRateLimits(olderThan time.Time) error {
	query := `DELETE FROM otp_rate_limits WHERE window_start_at < $1`

	result, err := r.db.Exec(query, olderThan)
	if err != nil {
		return fmt.Errorf("failed to cleanup rate limits: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected > 0 {
		fmt.Printf("Cleaned up %d old rate limit records\n", rowsAffected)
	}

	return nil
}
