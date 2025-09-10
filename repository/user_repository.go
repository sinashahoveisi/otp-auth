package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"otp-auth/entity"

	"github.com/jmoiron/sqlx"
)

// UserRepository interface defines user data operations
type UserRepository interface {
	Create(user *entity.User) (*entity.User, error)
	GetByID(id int) (*entity.User, error)
	GetByPhoneNumber(phoneNumber string) (*entity.User, error)
	Update(user *entity.User) (*entity.User, error)
	List(page, pageSize int, search string) ([]entity.User, int, error)
	UpdateLastLogin(phoneNumber string) error
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user
func (r *userRepository) Create(user *entity.User) (*entity.User, error) {
	query := `
		INSERT INTO users (phone_number, registered_at, last_login_at, is_active)
		VALUES (:phone_number, :registered_at, :last_login_at, :is_active)
		RETURNING id, phone_number, registered_at, last_login_at, is_active
	`

	now := time.Now()
	user.RegisteredAt = now
	user.LastLoginAt = &now // Set last_login_at equal to registered_at
	user.IsActive = true

	rows, err := r.db.NamedQuery(query, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("failed to get created user")
	}

	var createdUser entity.User
	if err := rows.StructScan(&createdUser); err != nil {
		return nil, fmt.Errorf("failed to scan created user: %w", err)
	}

	return &createdUser, nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(id int) (*entity.User, error) {
	query := `
		SELECT id, phone_number, registered_at, last_login_at, is_active
		FROM users
		WHERE id = $1 AND is_active = TRUE
	`

	var user entity.User
	err := r.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// GetByPhoneNumber retrieves a user by phone number
func (r *userRepository) GetByPhoneNumber(phoneNumber string) (*entity.User, error) {
	query := `
		SELECT id, phone_number, registered_at, last_login_at, is_active
		FROM users
		WHERE phone_number = $1 AND is_active = TRUE
	`

	var user entity.User
	err := r.db.Get(&user, query, phoneNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by phone number: %w", err)
	}

	return &user, nil
}

// Update updates an existing user
func (r *userRepository) Update(user *entity.User) (*entity.User, error) {
	query := `
		UPDATE users
		SET phone_number = :phone_number, last_login_at = :last_login_at, is_active = :is_active
		WHERE id = :id
		RETURNING id, phone_number, registered_at, last_login_at, is_active
	`

	rows, err := r.db.NamedQuery(query, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user not found")
	}

	var updatedUser entity.User
	if err := rows.StructScan(&updatedUser); err != nil {
		return nil, fmt.Errorf("failed to scan updated user: %w", err)
	}

	return &updatedUser, nil
}

// List retrieves paginated users with optional search
func (r *userRepository) List(page, pageSize int, search string) ([]entity.User, int, error) {
	offset := (page - 1) * pageSize

	// Build WHERE clause for search
	whereClause := "WHERE is_active = TRUE"
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		whereClause += fmt.Sprintf(" AND phone_number ILIKE $%d", argIndex)
		args = append(args, "%"+strings.ToLower(search)+"%")
		argIndex++
	}

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get paginated results
	listQuery := fmt.Sprintf(`
		SELECT id, phone_number, registered_at, last_login_at, is_active
		FROM users
		%s
		ORDER BY registered_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, pageSize, offset)

	var users []entity.User
	err = r.db.Select(&users, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, total, nil
}

// UpdateLastLogin updates the last login timestamp for a user
func (r *userRepository) UpdateLastLogin(phoneNumber string) error {
	query := `
		UPDATE users
		SET last_login_at = CURRENT_TIMESTAMP
		WHERE phone_number = $1 AND is_active = TRUE
	`

	result, err := r.db.Exec(query, phoneNumber)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found or inactive")
	}

	return nil
}
