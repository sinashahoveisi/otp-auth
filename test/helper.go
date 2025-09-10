package test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"otp-auth/entity"
	"otp-auth/migrations"
	"otp-auth/pkg/logger"
	"otp-auth/repository"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// TestDB wraps a test database connection
type TestDB struct {
	DB *sqlx.DB
}

// SetupTestDB creates a test database and runs migrations
func SetupTestDB(t *testing.T) *TestDB {
	// Use environment variables or defaults for test database
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5432")
	user := getEnvOrDefault("TEST_DB_USER", "otp_auth")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "otp_auth")

	// Get base database name and add _test suffix
	baseDBName := getEnvOrDefault("POSTGRES_DB", "otp_auth")
	dbName := getEnvOrDefault("TEST_DB_NAME", baseDBName+"_test")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	db, err := sqlx.Connect("postgres", connStr)
	require.NoError(t, err, "Failed to connect to test database")

	// Run migrations - check multiple possible paths
	migrationPaths := []string{"./migrations", "../migrations", "/app/migrations"}
	for _, path := range migrationPaths {
		err = migrations.RunMigrations(db.DB, path)
		if err == nil {
			break
		}
	}
	require.NoError(t, err, "Failed to run test migrations")

	return &TestDB{DB: db}
}

// Close closes the test database connection
func (tdb *TestDB) Close() {
	if tdb.DB != nil {
		tdb.DB.Close()
	}
}

// CleanTables removes all data from tables (for test isolation)
func (tdb *TestDB) CleanTables(t *testing.T) {
	_, err := tdb.DB.Exec("TRUNCATE TABLE otp_rate_limits, otps, users RESTART IDENTITY CASCADE")
	require.NoError(t, err, "Failed to clean test tables")
}

// CreateTestUser creates a test user in the database
func (tdb *TestDB) CreateTestUser(t *testing.T, phoneNumber string) *entity.User {
	user := &entity.User{
		PhoneNumber: phoneNumber,
		IsActive:    true,
	}

	userRepo := repository.NewUserRepository(tdb.DB)
	createdUser, err := userRepo.Create(user)
	require.NoError(t, err, "Failed to create test user")

	return createdUser
}

// CreateTestOTP creates a test OTP in the database
func (tdb *TestDB) CreateTestOTP(t *testing.T, phoneNumber, code string, expiresAt time.Time) *entity.OTP {
	otp := &entity.OTP{
		PhoneNumber: phoneNumber,
		Code:        code,
		ExpiresAt:   expiresAt,
		IsUsed:      false,
		CreatedAt:   time.Now(),
	}

	otpRepo := repository.NewOTPRepository(tdb.DB)
	createdOTP, err := otpRepo.Create(otp)
	require.NoError(t, err, "Failed to create test OTP")

	return createdOTP
}

// CreateExpiredOTP creates an expired OTP for testing
func (tdb *TestDB) CreateExpiredOTP(t *testing.T, phoneNumber, code string) *entity.OTP {
	return tdb.CreateTestOTP(t, phoneNumber, code, time.Now().Add(-5*time.Minute))
}

// CreateValidOTP creates a valid OTP that expires in 2 minutes
func (tdb *TestDB) CreateValidOTP(t *testing.T, phoneNumber, code string) *entity.OTP {
	return tdb.CreateTestOTP(t, phoneNumber, code, time.Now().Add(2*time.Minute))
}

// CreateTestRateLimit creates a rate limit record for testing
func (tdb *TestDB) CreateTestRateLimit(t *testing.T, phoneNumber string, requestCount int, windowStart time.Time) *entity.RateLimitInfo {
	rateLimitInfo := &entity.RateLimitInfo{
		PhoneNumber:   phoneNumber,
		RequestCount:  requestCount,
		LastRequestAt: time.Now(),
		WindowStartAt: windowStart,
	}

	// Note: Rate limiting is now handled by Redis, so this is just a mock for tests
	t.Logf("Mock rate limit created: phone=%s, count=%d", phoneNumber, requestCount)

	return rateLimitInfo
}

// GetTestLogger creates a test logger
func GetTestLogger() *logger.Logger {
	log, err := logger.New("debug", "development")
	if err != nil {
		panic(fmt.Sprintf("Failed to create test logger: %v", err))
	}
	return log
}

// AssertUserExists asserts that a user exists with the given phone number
func (tdb *TestDB) AssertUserExists(t *testing.T, phoneNumber string) *entity.User {
	userRepo := repository.NewUserRepository(tdb.DB)
	user, err := userRepo.GetByPhoneNumber(phoneNumber)
	require.NoError(t, err, "Failed to get user")
	require.NotNil(t, user, "User should exist")
	return user
}

// AssertUserCount asserts the total number of users in the database
func (tdb *TestDB) AssertUserCount(t *testing.T, expectedCount int) {
	var count int
	err := tdb.DB.Get(&count, "SELECT COUNT(*) FROM users")
	require.NoError(t, err, "Failed to count users")
	require.Equal(t, expectedCount, count, "User count mismatch")
}

// AssertOTPUsed asserts that an OTP is marked as used
func (tdb *TestDB) AssertOTPUsed(t *testing.T, otpID int) {
	var isUsed bool
	var usedAt *time.Time
	err := tdb.DB.QueryRow("SELECT is_used, used_at FROM otps WHERE id = $1", otpID).Scan(&isUsed, &usedAt)
	require.NoError(t, err, "Failed to get OTP status")
	require.True(t, isUsed, "OTP should be marked as used")
	require.NotNil(t, usedAt, "OTP should have used_at timestamp")
}

// AssertOTPNotUsed asserts that an OTP is not marked as used
func (tdb *TestDB) AssertOTPNotUsed(t *testing.T, otpID int) {
	var isUsed bool
	err := tdb.DB.Get(&isUsed, "SELECT is_used FROM otps WHERE id = $1", otpID)
	require.NoError(t, err, "Failed to get OTP status")
	require.False(t, isUsed, "OTP should not be marked as used")
}

// AssertRateLimitExists asserts that a rate limit record exists for a phone number
func (tdb *TestDB) AssertRateLimitExists(t *testing.T, phoneNumber string, expectedCount int) {
	var requestCount int
	err := tdb.DB.Get(&requestCount, "SELECT request_count FROM otp_rate_limits WHERE phone_number = $1", phoneNumber)
	require.NoError(t, err, "Failed to get rate limit")
	require.Equal(t, expectedCount, requestCount, "Rate limit count mismatch")
}

// AssertLastLoginUpdated asserts that the user's last login timestamp was recently updated
func (tdb *TestDB) AssertLastLoginUpdated(t *testing.T, phoneNumber string, within time.Duration) {
	var lastLoginAt *time.Time
	err := tdb.DB.Get(&lastLoginAt, "SELECT last_login_at FROM users WHERE phone_number = $1", phoneNumber)
	require.NoError(t, err, "Failed to get last login time")
	require.NotNil(t, lastLoginAt, "Last login should be set")

	timeSinceLogin := time.Since(*lastLoginAt)
	require.True(t, timeSinceLogin <= within,
		"Last login should be within %v, but was %v ago", within, timeSinceLogin)
}

// GetActiveOTPCount returns the number of active (unused, non-expired) OTPs for a phone number
func (tdb *TestDB) GetActiveOTPCount(t *testing.T, phoneNumber string) int {
	var count int
	err := tdb.DB.Get(&count,
		"SELECT COUNT(*) FROM otps WHERE phone_number = $1 AND is_used = FALSE AND expires_at > NOW()",
		phoneNumber)
	require.NoError(t, err, "Failed to count active OTPs")
	return count
}

// GetTotalOTPCount returns the total number of OTPs for a phone number
func (tdb *TestDB) GetTotalOTPCount(t *testing.T, phoneNumber string) int {
	var count int
	err := tdb.DB.Get(&count, "SELECT COUNT(*) FROM otps WHERE phone_number = $1", phoneNumber)
	require.NoError(t, err, "Failed to count total OTPs")
	return count
}

// WaitForCleanup waits a short period to allow cleanup routines to complete
func WaitForCleanup() {
	time.Sleep(100 * time.Millisecond)
}

// AssertPhoneNumberFormat asserts that a phone number matches expected format
func AssertPhoneNumberFormat(t *testing.T, phoneNumber string) {
	require.NotEmpty(t, phoneNumber, "Phone number should not be empty")
	require.True(t, len(phoneNumber) >= 10, "Phone number should be at least 10 digits")
	require.True(t, len(phoneNumber) <= 15, "Phone number should be at most 15 digits")
}

// GenerateTestPhoneNumber generates a test phone number with optional suffix
func GenerateTestPhoneNumber(suffix string) string {
	if suffix == "" {
		return "+1234567890"
	}
	return fmt.Sprintf("+12345678%s", suffix)
}

// GenerateTestOTPCode generates a test OTP code
func GenerateTestOTPCode(suffix string) string {
	if suffix == "" {
		return "123456"
	}
	return fmt.Sprintf("12345%s", suffix)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
