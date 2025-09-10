package handler

import (
	"otp-auth/config"
	"otp-auth/controller"
	_ "otp-auth/docs" // Import for swagger docs
	"otp-auth/pkg/logger"
	"otp-auth/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// RegisterRoutes registers all HTTP routes and middleware
func RegisterRoutes(
	e *echo.Echo,
	otpController *controller.OTPController,
	userController *controller.UserController,
	authController *controller.AuthController,
	healthController *controller.HealthController,
	jwtService service.JWTService,
	cfg *config.Config,
	logger *logger.Logger,
) {
	// Add common middleware
	e.Use(middleware.Recover())
	e.Use(CORSMiddleware())
	e.Use(RequestLoggerMiddleware(logger))
	e.Use(JWTMiddleware(jwtService, logger))

	// System endpoints
	e.GET("/health", healthController.HealthCheck)
	e.GET("/", healthController.ServiceInfo)

	// Swagger documentation
	if cfg.Swagger.Enabled {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
		e.GET("/docs/*", echoSwagger.WrapHandler)
	}

	// API v1 group
	v1 := e.Group("/api/v1")

	// OTP routes (public)
	otpGroup := v1.Group("/otp")
	otpGroup.POST("/send", otpController.SendOTP)
	otpGroup.POST("/verify", otpController.VerifyOTP)

	// User routes (protected)
	userGroup := v1.Group("/users")
	userGroup.GET("/:id", userController.GetUser)
	userGroup.GET("", userController.ListUsers)

	// Auth routes (protected)
	authGroup := v1.Group("/auth")
	authGroup.POST("/logout", authController.Logout)
}
