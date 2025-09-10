package entity

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           int        `db:"id" json:"id"`
	PhoneNumber  string     `db:"phone_number" json:"phone_number" validate:"required,phone_number"`
	RegisteredAt time.Time  `db:"registered_at" json:"registered_at"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"last_login_at"`
	IsActive     bool       `db:"is_active" json:"is_active"`
}

// TableName returns the table name for the User entity
func (User) TableName() string {
	return "users"
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	PhoneNumber string `json:"phone_number" validate:"required,phone_number"`
}

// UserResponse represents the user response
type UserResponse struct {
	ID           int        `json:"id"`
	PhoneNumber  string     `json:"phone_number"`
	RegisteredAt time.Time  `json:"registered_at"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	IsActive     bool       `json:"is_active"`
}

// UsersListResponse represents the paginated list of users
type UsersListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// PaginatedUsersResponse is an alias for backward compatibility
type PaginatedUsersResponse = UsersListResponse

// LogoutRequest represents the logout request structure
type LogoutRequest struct {
	LogoutAll bool `json:"logout_all,omitempty"`
}

// LogoutResponse represents the logout response structure
type LogoutResponse struct {
	Message       string `json:"message"`
	TokensRevoked int    `json:"tokens_revoked,omitempty"`
}
