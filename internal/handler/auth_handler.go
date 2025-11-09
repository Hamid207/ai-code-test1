package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Hamid207/ai-code-test1/internal/model"
	"github.com/Hamid207/ai-code-test1/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
	dbPool      *pgxpool.Pool
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *service.AuthService, dbPool *pgxpool.Pool) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		dbPool:      dbPool,
	}
}

// SignInWithApple handles Apple OAuth sign-in
// @Summary Sign in with Apple
// @Description Authenticate user using Apple ID token
// @Accept json
// @Produce json
// @Param request body model.AppleSignInRequest true "Apple Sign In Request"
// @Success 200 {object} model.AppleSignInResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 401 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /auth/apple [post]
func (h *AuthHandler) SignInWithApple(c *gin.Context) {
	var req model.AppleSignInRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Process authentication
	response, err := h.authService.SignInWithApple(c.Request.Context(), &req)
	if err != nil {
		// Log internal error for debugging (do not expose to client)
		log.Printf("Authentication failed: %v", err)

		// Return generic error message to prevent information disclosure
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{
			Error:   "authentication_failed",
			Message: "Invalid or expired token",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// HealthCheck returns the health status of the service
func (h *AuthHandler) HealthCheck(c *gin.Context) {
	// Check database connectivity with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.dbPool.Ping(ctx); err != nil {
		log.Printf("Health check failed: database ping error: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unhealthy",
			"database": "disconnected",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"database": "connected",
	})
}
