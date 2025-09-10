package service

import (
	"fmt"
	"math"

	"otp-auth/entity"
	"otp-auth/pkg/logger"
	"otp-auth/repository"
)

// UserService interface defines user business operations
type UserService interface {
	GetByID(id int) (*entity.UserResponse, error)
	GetList(page, pageSize int, search string) (*entity.UsersListResponse, error)
}

// userService implements UserService interface
type userService struct {
	userRepo repository.UserRepository
	logger   *logger.Logger
}

// NewUserService creates a new user service instance
func NewUserService(userRepo repository.UserRepository, logger *logger.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// GetByID retrieves a user by ID
func (s *userService) GetByID(id int) (*entity.UserResponse, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		s.logger.Errorw("Failed to get user by ID", "user_id", id, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	return s.toUserResponse(user), nil
}

// GetList retrieves paginated list of users with optional search
func (s *userService) GetList(page, pageSize int, search string) (*entity.UsersListResponse, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := s.userRepo.List(page, pageSize, search)
	if err != nil {
		s.logger.Errorw("Failed to get users list", "page", page, "page_size", pageSize, "search", search, "error", err)
		return nil, fmt.Errorf("failed to get users list: %w", err)
	}

	// Convert to response format
	userResponses := make([]entity.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = *s.toUserResponse(&user)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &entity.UsersListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// toUserResponse converts User entity to UserResponse
func (s *userService) toUserResponse(user *entity.User) *entity.UserResponse {
	return &entity.UserResponse{
		ID:           user.ID,
		PhoneNumber:  user.PhoneNumber,
		RegisteredAt: user.RegisteredAt,
		LastLoginAt:  user.LastLoginAt,
		IsActive:     user.IsActive,
	}
}
