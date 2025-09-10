package controller

import (
	"net/http"
	"strconv"

	"otp-auth/pkg/logger"
	"otp-auth/service"

	"github.com/labstack/echo/v4"
)

// UserController handles user-related HTTP requests
type UserController struct {
	userService service.UserService
	logger      *logger.Logger
}

// NewUserController creates a new user controller instance
func NewUserController(userService service.UserService, logger *logger.Logger) *UserController {
	return &UserController{
		userService: userService,
		logger:      logger,
	}
}

// GetUser retrieves a single user by ID
// @Summary Get User
// @Description Get user details by ID
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Success 200 {object} entity.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/{id} [get]
func (c *UserController) GetUser(ctx echo.Context) error {
	// Parse user ID from path parameter
	idParam := ctx.Param("id")
	userID, err := strconv.Atoi(idParam)
	if err != nil {
		c.logger.Warnw("Invalid user ID", "id", idParam, "error", err)
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid user ID",
			"details": "User ID must be a valid integer",
		})
	}

	// Get user from service
	user, err := c.userService.GetByID(userID)
	if err != nil {
		if err.Error() == "user not found" {
			c.logger.Infow("User not found", "user_id", userID)
			return ctx.JSON(http.StatusNotFound, map[string]interface{}{
				"error":   "User not found",
				"details": "The requested user does not exist",
			})
		}

		c.logger.Errorw("Failed to get user", "user_id", userID, "error", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to retrieve user",
			"details": "Internal server error",
		})
	}

	c.logger.Infow("User retrieved successfully", "user_id", userID)
	return ctx.JSON(http.StatusOK, user)
}

// ListUsers retrieves paginated list of users with optional search
// @Summary List Users
// @Description Get paginated list of users with optional search
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param search query string false "Search by phone number"
// @Success 200 {object} entity.UsersListResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users [get]
func (c *UserController) ListUsers(ctx echo.Context) error {
	// Parse query parameters
	page := 1
	if pageParam := ctx.QueryParam("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeParam := ctx.QueryParam("page_size"); pageSizeParam != "" {
		if ps, err := strconv.Atoi(pageSizeParam); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	search := ctx.QueryParam("search")

	// Get users list from service
	response, err := c.userService.GetList(page, pageSize, search)
	if err != nil {
		c.logger.Errorw("Failed to get users list", "page", page, "page_size", pageSize, "search", search, "error", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to retrieve users list",
			"details": "Internal server error",
		})
	}

	c.logger.Infow("Users list retrieved successfully", "page", page, "page_size", pageSize, "search", search, "total", response.Total)
	return ctx.JSON(http.StatusOK, response)
}
