package controller

import (
	"net/http"

	"otp-auth/pkg/logger"
	"otp-auth/service"

	"github.com/labstack/echo/v4"
)

// AuthController handles authentication-related operations
type AuthController struct {
	jwtService service.JWTService
	logger     *logger.Logger
}

// NewAuthController creates a new auth controller
func NewAuthController(jwtService service.JWTService, logger *logger.Logger) *AuthController {
	return &AuthController{
		jwtService: jwtService,
		logger:     logger,
	}
}

// LogoutRequest represents the logout request body
type LogoutRequest struct {
	LogoutAll bool `json:"logout_all,omitempty"` // Optional: logout from all devices
}

// @Summary Logout user
// @Description Logout user and revoke JWT token from Redis session store
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LogoutRequest false "Logout options"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /auth/logout [post]
func (c *AuthController) Logout(ctx echo.Context) error {
	// Get token from Authorization header
	authHeader := ctx.Request().Header.Get("Authorization")
	if authHeader == "" {
		return ctx.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "Unauthorized",
			"details": "Missing Authorization header",
		})
	}

	// Extract token (remove "Bearer " prefix)
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return ctx.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "Unauthorized",
			"details": "Invalid Authorization header format",
		})
	}

	tokenString := authHeader[7:]

	// Parse request body for logout options
	var req LogoutRequest
	if err := ctx.Bind(&req); err != nil {
		// Ignore binding errors for optional request body
		req = LogoutRequest{}
	}

	// Get user from token first (before revoking)
	token, err := c.jwtService.ValidateToken(tokenString)
	if err != nil {
		c.logger.Warnw("Failed to validate token for logout", "error", err)
		return ctx.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "Unauthorized",
			"details": "Invalid token",
		})
	}

	user, err := c.jwtService.GetUserFromToken(token)
	if err != nil {
		c.logger.Errorw("Failed to get user from token", "error", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Internal Server Error",
			"details": "Failed to process logout",
		})
	}

	// Revoke tokens
	if req.LogoutAll {
		// Logout from all devices
		if err := c.jwtService.RevokeAllUserTokens(user.ID); err != nil {
			c.logger.Errorw("Failed to revoke all user tokens", "user_id", user.ID, "error", err)
			return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Internal Server Error",
				"details": "Failed to logout from all devices",
			})
		}
		c.logger.Infow("User logged out from all devices", "user_id", user.ID)
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Successfully logged out from all devices",
		})
	} else {
		// Logout from current device only
		if err := c.jwtService.RevokeToken(tokenString); err != nil {
			c.logger.Errorw("Failed to revoke token", "user_id", user.ID, "error", err)
			return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   "Internal Server Error",
				"details": "Failed to logout",
			})
		}
		c.logger.Infow("User logged out", "user_id", user.ID)
		return ctx.JSON(http.StatusOK, map[string]string{
			"message": "Successfully logged out",
		})
	}
}
