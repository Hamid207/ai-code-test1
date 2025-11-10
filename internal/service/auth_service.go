package service

import (
	"context"
	"fmt"

	"github.com/Hamid207/ai-code-test1/internal/model"
	"github.com/Hamid207/ai-code-test1/internal/repository"
	"github.com/Hamid207/ai-code-test1/pkg/apple"
	"github.com/Hamid207/ai-code-test1/pkg/google"
	"github.com/Hamid207/ai-code-test1/pkg/jwt"
)

// AuthService handles authentication business logic
type AuthService struct {
	appleVerifier   *apple.Verifier
	googleVerifier  *google.Verifier
	userRepository  *repository.UserRepository
	tokenRepository *repository.TokenRepository
	tokenService    *jwt.TokenService
}

// NewAuthService creates a new authentication service
func NewAuthService(
	appleClientID string,
	googleClientID string,
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	tokenService *jwt.TokenService,
) *AuthService {
	return &AuthService{
		appleVerifier:   apple.NewVerifier(appleClientID),
		googleVerifier:  google.NewVerifier(googleClientID),
		userRepository:  userRepo,
		tokenRepository: tokenRepo,
		tokenService:    tokenService,
	}
}

// SignInWithApple verifies Apple ID token and returns user information with JWT tokens
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

	// Generate JWT token pair (access + refresh)
	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.AppleID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token in database
	err = s.tokenRepository.StoreRefreshToken(ctx, user.ID, tokenPair.RefreshToken, tokenPair.RefreshTokenExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Build response with tokens
	response := &model.AppleSignInResponse{
		UserID:                user.ID,
		AppleID:               user.AppleID,
		Email:                 user.Email,
		AccessToken:           tokenPair.AccessToken,
		RefreshToken:          tokenPair.RefreshToken,
		AccessTokenExpiresAt:  tokenPair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokenPair.RefreshTokenExpiresAt,
		TokenType:             "Bearer",
	}

	return response, nil
}

// SignInWithGoogle verifies Google ID token and returns user information with JWT tokens
func (s *AuthService) SignInWithGoogle(ctx context.Context, req *model.GoogleSignInRequest) (*model.GoogleSignInResponse, error) {
	// Verify the ID token
	claims, err := s.googleVerifier.VerifyIDToken(req.IDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	// Email is already verified in the verifier (EmailVerified must be true)

	// Create or get user from database
	user, err := s.userRepository.CreateOrGetWithGoogle(ctx, claims.Subject, claims.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to create or get user: %w", err)
	}

	// Generate JWT token pair (access + refresh)
	// Use GoogleID as the provider ID (AppleID field in JWT for backward compatibility)
	tokenPair, err := s.tokenService.GenerateTokenPair(user.ID, user.GoogleID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store refresh token in database
	err = s.tokenRepository.StoreRefreshToken(ctx, user.ID, tokenPair.RefreshToken, tokenPair.RefreshTokenExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Build response with tokens
	response := &model.GoogleSignInResponse{
		UserID:                user.ID,
		GoogleID:              user.GoogleID,
		Email:                 user.Email,
		AccessToken:           tokenPair.AccessToken,
		RefreshToken:          tokenPair.RefreshToken,
		AccessTokenExpiresAt:  tokenPair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokenPair.RefreshTokenExpiresAt,
		TokenType:             "Bearer",
	}

	return response, nil
}

// RefreshAccessToken generates new access AND refresh tokens (token rotation)
// This implements refresh token rotation for security - old token is revoked
func (s *AuthService) RefreshAccessToken(ctx context.Context, req *model.RefreshTokenRequest) (*model.RefreshTokenResponse, error) {
	// Validate refresh token (JWT validation)
	claims, err := s.tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Validate refresh token in database
	userID, err := s.tokenRepository.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh token validation failed: %w", err)
	}

	// Verify user ID matches
	if userID != claims.UserID {
		return nil, fmt.Errorf("user ID mismatch")
	}

	// CRITICAL: Revoke the old refresh token BEFORE generating new ones
	// This prevents reuse of stolen tokens
	err = s.tokenRepository.RevokeRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// Generate NEW token pair (access + refresh) - TOKEN ROTATION
	tokenPair, err := s.tokenService.GenerateTokenPair(
		claims.UserID,
		claims.AppleID,
		claims.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token pair: %w", err)
	}

	// Store the NEW refresh token in database
	err = s.tokenRepository.StoreRefreshToken(ctx, userID, tokenPair.RefreshToken, tokenPair.RefreshTokenExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// Return BOTH new access and refresh tokens
	response := &model.RefreshTokenResponse{
		AccessToken:           tokenPair.AccessToken,
		RefreshToken:          tokenPair.RefreshToken,
		AccessTokenExpiresAt:  tokenPair.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: tokenPair.RefreshTokenExpiresAt,
		TokenType:             "Bearer",
	}

	return response, nil
}
