package controller

import (
	"net/http"
	"strings"

	"otp-auth/entity"
	"otp-auth/pkg/logger"
	"otp-auth/service"
	"otp-auth/validator"

	"github.com/labstack/echo/v4"
)

// OTPController handles OTP-related HTTP requests
type OTPController struct {
	otpService service.OTPService
	jwtService service.JWTService
	validator  *validator.Validator
	logger     *logger.Logger
}

// NewOTPController creates a new OTP controller instance
func NewOTPController(otpService service.OTPService, jwtService service.JWTService, validator *validator.Validator, logger *logger.Logger) *OTPController {
	return &OTPController{
		otpService: otpService,
		jwtService: jwtService,
		validator:  validator,
		logger:     logger,
	}
}

// SendOTP handles OTP generation and sending
// @Summary Send OTP
// @Description Generate and send OTP to the provided phone number
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body entity.SendOTPRequest true "Send OTP Request"
// @Success 200 {object} entity.OTPResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 429 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /otp/send [post]
func (c *OTPController) SendOTP(ctx echo.Context) error {
	var req entity.SendOTPRequest

	// Bind request body
	if err := ctx.Bind(&req); err != nil {
		c.logger.Errorw("Failed to bind request", "error", err)
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
	}

	// Validate request
	if err := c.validator.ValidateStruct(&req); err != nil {
		c.logger.Warnw("Validation failed", "request", req, "error", err)
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Send OTP
	response, err := c.otpService.SendOTP(req.PhoneNumber)
	if err != nil {
		c.logger.Errorw("Failed to send OTP", "phone_number", req.PhoneNumber, "error", err)

		// Check if it's a rate limiting error
		if strings.Contains(err.Error(), "rate limit exceeded") {
			return ctx.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"error":   "Rate limit exceeded",
				"details": "Maximum 3 OTP requests per phone number within 10 minutes. Please try again later.",
			})
		}

		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to send OTP",
			"details": "Internal server error",
		})
	}

	c.logger.Infow("OTP sent successfully", "phone_number", req.PhoneNumber)
	return ctx.JSON(http.StatusOK, response)
}

// VerifyOTP handles OTP verification and authentication
// @Summary Verify OTP
// @Description Verify OTP and authenticate user
// @Tags OTP
// @Accept json
// @Produce json
// @Param request body entity.VerifyOTPRequest true "Verify OTP Request (token from send response)"
// @Success 200 {object} entity.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /otp/verify [post]
func (c *OTPController) VerifyOTP(ctx echo.Context) error {
	var req entity.VerifyOTPRequest

	// Bind request body
	if err := ctx.Bind(&req); err != nil {
		c.logger.Errorw("Failed to bind request", "error", err)
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
	}

	// Validate request
	if err := c.validator.ValidateStruct(&req); err != nil {
		c.logger.Warnw("Validation failed", "request", req, "error", err)
		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Verify OTP
	user, err := c.otpService.VerifyOTP(req.Token, req.Code)
	if err != nil {
		c.logger.Warnw("OTP verification failed", "token", req.Token, "error", err)

		if err.Error() == "invalid or expired OTP" {
			return ctx.JSON(http.StatusUnauthorized, map[string]interface{}{
				"error":   "Invalid or expired OTP",
				"details": "Please request a new OTP",
			})
		}

		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to verify OTP",
			"details": "Internal server error",
		})
	}

	// Generate JWT token
	authResponse, err := c.jwtService.GenerateToken(user)
	if err != nil {
		c.logger.Errorw("Failed to generate JWT token", "user_id", user.ID, "error", err)
		return ctx.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "Failed to generate authentication token",
			"details": "Internal server error",
		})
	}

	c.logger.Infow("OTP verified successfully", "user_id", user.ID, "phone_number", user.PhoneNumber)
	return ctx.JSON(http.StatusOK, authResponse)
}
