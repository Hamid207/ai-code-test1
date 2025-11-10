package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Hamid207/ai-code-test1/internal/handler"
	"github.com/Hamid207/ai-code-test1/internal/repository"
	"github.com/Hamid207/ai-code-test1/internal/service"
	"github.com/Hamid207/ai-code-test1/pkg/config"
	"github.com/Hamid207/ai-code-test1/pkg/database"
	"github.com/Hamid207/ai-code-test1/pkg/jwt"
	"github.com/Hamid207/ai-code-test1/pkg/logger"
	redispkg "github.com/Hamid207/ai-code-test1/pkg/redis"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func main() {
	// Initialize structured logger
	if err := logger.Init(false); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection pool
	// Use a timeout context for initial connection
	initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
	poolConfig := database.PoolConfig{
		MaxConns: cfg.DBMaxConns,
		MinConns: cfg.DBMinConns,
	}
	dbPool, err := database.NewPool(initCtx, cfg.DatabaseURL, poolConfig)
	initCancel() // Cancel after connection is established
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Printf("Database connection established successfully (Max: %d, Min: %d)", cfg.DBMaxConns, cfg.DBMinConns)

	// Initialize Redis connection
	redisConfig := redispkg.Config{
		Host:         cfg.RedisHost,
		Port:         cfg.RedisPort,
		DB:           cfg.RedisDB,
		Password:     cfg.RedisPassword,
		MaxConns:     cfg.RedisMaxConns,
		MinIdleConns: cfg.RedisMinIdleConns,
	}
	redisClient, err := redispkg.NewClient(redisConfig)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Printf("Redis connection established successfully (Host: %s:%s, DB: %d)", cfg.RedisHost, cfg.RedisPort, cfg.RedisDB)

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbPool)
	tokenRepo := repository.NewTokenRepository(dbPool)

	// Initialize Redis repositories
	// TODO: Integrate Redis repositories with services
	_ = redispkg.NewTokenRepository(redisClient)
	_ = redispkg.NewBlacklistRepository(redisClient)
	_ = redispkg.NewRateLimitRepository(redisClient)
	_ = redispkg.NewCacheRepository(redisClient)

	// Initialize JWT token service
	tokenService := jwt.NewTokenService(cfg.JWTSecret)

	// Initialize services
	authService := service.NewAuthService(cfg.AppleClientID, userRepo, tokenRepo, tokenService)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, dbPool)

	// Setup router
	router := setupRouter(authHandler, cfg)

	// Create HTTP server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close database pool after server shutdown
	log.Println("Closing database connections...")
	dbPool.Close()

	// Close Redis connection
	log.Println("Closing Redis connection...")
	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("Server exited gracefully")
}

// setupRouter configures all routes and middleware
func setupRouter(authHandler *handler.AuthHandler, cfg *config.Config) *gin.Engine {
	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(requestBodyLimitMiddleware(1024 * 1024)) // 1MB limit
	router.Use(corsMiddleware(cfg.AllowedOrigins))

	// Health check
	router.GET("/health", authHandler.HealthCheck)

	// API routes with rate limiting
	api := router.Group("/api/v1")
	{
		// Rate limiter: 10 requests per minute per IP
		rate := limiter.Rate{
			Period: 1 * time.Minute,
			Limit:  10,
		}
		store := memory.NewStore()
		rateLimiter := limiter.New(store, rate)
		rateLimitMiddleware := mgin.NewMiddleware(rateLimiter)

		auth := api.Group("/auth")
		auth.Use(rateLimitMiddleware)
		{
			auth.POST("/apple", authHandler.SignInWithApple)
			auth.POST("/refresh", authHandler.RefreshToken)
		}
	}

	return router
}

// requestBodyLimitMiddleware limits request body size to prevent memory attacks
func requestBodyLimitMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// corsMiddleware adds CORS headers to responses (secure implementation)
func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is in allowed list
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		// Only set CORS headers if origin is allowed
		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		}

		if c.Request.Method == "OPTIONS" {
			if allowed {
				c.AbortWithStatus(204)
			} else {
				c.AbortWithStatus(403)
			}
			return
		}

		c.Next()
	}
}
