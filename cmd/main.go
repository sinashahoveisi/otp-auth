package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"otp-auth/config"
	"otp-auth/controller"
	_ "otp-auth/docs" // Import for swagger
	"otp-auth/handler"
	"otp-auth/migrations"
	"otp-auth/pkg/logger"
	"otp-auth/repository"
	"otp-auth/service"
	"otp-auth/validator"

	"github.com/redis/go-redis/v9"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

// @title OTP Authentication Service API
// @version 1.0
// @description A backend service for OTP-based authentication and user management with Redis session management
// @contact.name API Support
// @contact.email support@example.com
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
// @description Enter JWT Bearer token in format: Bearer {token}
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.New(cfg.Logger.Level, cfg.Logger.Mode)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Infow("Starting OTP Authentication Service",
		"version", "1.0.0",
		"port", cfg.HTTPServer.Port,
		"log_level", cfg.Logger.Level,
		"log_mode", cfg.Logger.Mode,
	)

	// Connect to database
	db, err := connectDB(cfg)
	if err != nil {
		log.Fatalw("Failed to connect to database", "error", err)
	}
	defer db.Close()

	log.Infow("Database connected successfully",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"database", cfg.Database.Name,
	)

	// Run migrations
	if err := migrations.RunMigrations(db.DB, "./migrations"); err != nil {
		log.Fatalw("Failed to run database migrations", "error", err)
	}

	log.Infow("Database migrations completed successfully")

	// Connect to Redis for rate limiting
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalw("Failed to connect to Redis", "error", err)
	}

	log.Infow("Redis connected successfully", "host", cfg.Redis.Host, "port", cfg.Redis.Port)

	// Initialize validator
	v := validator.New()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	otpRepo := repository.NewOTPRepository(db)
	rateLimitRepo := repository.NewRedisRateLimitRepository(redisClient, cfg, log)

	// Initialize services
	userService := service.NewUserService(userRepo, log)
	tokenService := service.NewTokenService(redisClient, log)
	jwtService := service.NewJWTService(cfg, log, tokenService)
	otpService := service.NewOTPService(otpRepo, userRepo, rateLimitRepo, cfg, log)

	// Initialize controllers
	userController := controller.NewUserController(userService, log)
	otpController := controller.NewOTPController(otpService, jwtService, v, log)
	authController := controller.NewAuthController(jwtService, log)

	// Initialize Echo server
	e := echo.New()
	e.HideBanner = true

	// Register routes
	handler.RegisterRoutes(e, otpController, userController, authController, jwtService, cfg, log)

	// Start cleanup routine in background
	go startCleanupRoutine(otpService, log)

	// Start server in a goroutine
	serverAddr := fmt.Sprintf(":%d", cfg.HTTPServer.Port)
	go func() {
		log.Infow("Starting HTTP server", "address", serverAddr)
		if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
			log.Fatalw("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Infow("Shutting down server gracefully...")

	// Create a deadline for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Application.GracefulShutdownTimeout)
	defer shutdownCancel()

	// Attempt graceful shutdown
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Errorw("Failed to shutdown server gracefully", "error", err)
		os.Exit(1)
	}

	log.Infow("Server shutdown completed successfully")
}

func connectDB(cfg *config.Config) (*sqlx.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	var db *sqlx.DB
	var err error

	// Retry connection up to 30 times with 1 second delay
	for i := 0; i < 30; i++ {
		db, err = sqlx.Connect("postgres", connStr)
		if err == nil {
			// Test connection
			if err = db.Ping(); err == nil {
				break
			}
			db.Close()
		}

		if i == 0 {
			fmt.Printf("Waiting for database to be ready...\n")
		}
		fmt.Printf("Database connection attempt %d/30 failed: %v\n", i+1, err)
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after 30 attempts: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// startCleanupRoutine runs periodic cleanup of expired OTPs and rate limit records
func startCleanupRoutine(otpService service.OTPService, logger *logger.Logger) {
	ticker := time.NewTicker(5 * time.Minute) // Run cleanup every 5 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := otpService.CleanupExpiredOTPs(); err != nil {
				logger.Errorw("Failed to cleanup expired OTPs", "error", err)
			} else {
				logger.Debugw("Cleanup routine completed successfully")
			}
		}
	}
}
