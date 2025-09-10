package handler

import (
	"net/http"
	"strings"

	"otp-auth/pkg/logger"
	"otp-auth/service"

	"github.com/labstack/echo/v4"
)

// JWTMiddleware creates a JWT authentication middleware
func JWTMiddleware(jwtService service.JWTService, logger *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip authentication for public endpoints
			path := c.Request().URL.Path
			if strings.HasPrefix(path, "/api/v1/otp/") ||
				strings.HasPrefix(path, "/swagger") ||
				strings.HasPrefix(path, "/docs") ||
				path == "/" ||
				path == "/health" {
				return next(c)
			}

			// Get Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				logger.Warnw("Missing Authorization header", "path", path)
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "Unauthorized",
					"details": "Missing Authorization header",
				})
			}

			// Check Bearer token format
			if !strings.HasPrefix(authHeader, "Bearer ") {
				logger.Warnw("Invalid Authorization header format", "path", path)
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "Unauthorized",
					"details": "Invalid Authorization header format",
				})
			}

			// Extract token
			tokenString := authHeader[7:] // Remove "Bearer " prefix

			// Validate token
			token, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				logger.Warnw("Invalid JWT token", "path", path, "error", err)
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "Unauthorized",
					"details": "Invalid or expired token",
				})
			}

			// Extract user information from token
			user, err := jwtService.GetUserFromToken(token)
			if err != nil {
				logger.Errorw("Failed to extract user from token", "path", path, "error", err)
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "Unauthorized",
					"details": "Invalid token claims",
				})
			}

			// Store user in context
			c.Set("user", user)

			logger.Debugw("JWT authentication successful", "user_id", user.ID, "path", path)
			return next(c)
		}
	}
}

// CORSMiddleware creates a CORS middleware
func CORSMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusNoContent)
			}

			return next(c)
		}
	}
}

// RequestLoggerMiddleware creates a request logging middleware
func RequestLoggerMiddleware(logger *logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := c.Request().Context().Value("start_time")

			logger.Infow("HTTP Request",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"remote_addr", c.RealIP(),
				"user_agent", c.Request().UserAgent(),
			)

			err := next(c)

			logger.Infow("HTTP Response",
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"status", c.Response().Status,
				"start_time", start,
			)

			return err
		}
	}
}
