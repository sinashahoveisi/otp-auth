package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status" example:"healthy"`
	Service string `json:"service" example:"otp-auth-service"`
	Version string `json:"version" example:"1.0.0"`
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Returns the health status of the service
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthController) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{
		Status:  "healthy",
		Service: "otp-auth-service",
		Version: "1.0.0",
	})
}

// ServiceInfoResponse represents the service info response
type ServiceInfoResponse struct {
	Message string `json:"message" example:"OTP Authentication Service"`
	Version string `json:"version" example:"1.0.0"`
	Docs    string `json:"docs" example:"/swagger/index.html"`
}

// ServiceInfo godoc
// @Summary Service information
// @Description Returns basic service information and documentation links
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} ServiceInfoResponse
// @Router / [get]
func (h *HealthController) ServiceInfo(c echo.Context) error {
	return c.JSON(http.StatusOK, ServiceInfoResponse{
		Message: "OTP Authentication Service",
		Version: "1.0.0",
		Docs:    "/swagger/index.html",
	})
}
