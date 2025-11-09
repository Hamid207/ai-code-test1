package service

import (
	"context"
	"fmt"

	"github.com/Hamid207/ai-code-test1/internal/model"
	"github.com/Hamid207/ai-code-test1/internal/repository"
	"github.com/Hamid207/ai-code-test1/pkg/apple"
)

// AuthService handles authentication business logic
type AuthService struct {
	appleVerifier  *apple.Verifier
	userRepository *repository.UserRepository
}

// NewAuthService creates a new authentication service
func NewAuthService(clientID string, userRepo *repository.UserRepository) *AuthService {
	return &AuthService{
		appleVerifier:  apple.NewVerifier(clientID),
		userRepository: userRepo,
	}
}

// SignInWithApple verifies Apple ID token and returns user information
func (s *AuthService) SignInWithApple(ctx context.Context, req *model.AppleSignInRequest) (*model.AppleSignInResponse, error) {
	// Verify the ID token
	claims, err := s.appleVerifier.VerifyIDToken(req.IDToken, req.Nonce)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	// Verify email is confirmed (security best practice)
	if claims.EmailVerified != "true" {
		return nil, fmt.Errorf("email not verified by Apple")
	}

	// Create or get user from database
	user, err := s.userRepository.CreateOrGet(ctx, claims.Subject, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to create or get user: %w", err)
	}

	// Extract user information from claims
	response := &model.AppleSignInResponse{
		UserID: claims.Subject,
		Email:  user.Email,
		// Here you would typically generate your own JWT token
		// for subsequent API requests
		// Token: generateYourOwnJWT(claims.Subject),
	}

	return response, nil
}
