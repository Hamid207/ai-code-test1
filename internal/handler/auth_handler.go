package handler

import (
	"log"
	"net/http"

	"github.com/Hamid207/ai-code-test1/internal/model"
	"github.com/Hamid207/ai-code-test1/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
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
	response, err := h.authService.SignInWithApple(&req)
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
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
