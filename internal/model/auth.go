package model

import "time"

// AppleSignInRequest represents the request body for Apple sign-in
type AppleSignInRequest struct {
	IDToken string `json:"id_token" binding:"required"`
	Nonce   string `json:"nonce" binding:"required"`
}

// AppleSignInResponse represents the response after successful authentication
type AppleSignInResponse struct {
	UserID                int64     `json:"user_id"`
	AppleID               string    `json:"apple_id"`
	Email                 string    `json:"email"`
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"` // Always "Bearer"
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents the response after successful token refresh
type RefreshTokenResponse struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"` // New refresh token (rotation)
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"` // Always "Bearer"
}

// GoogleSignInRequest represents the request body for Google sign-in
type GoogleSignInRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// GoogleSignInResponse represents the response after successful Google authentication
type GoogleSignInResponse struct {
	UserID                int64     `json:"user_id"`
	GoogleID              string    `json:"google_id"`
	Email                 string    `json:"email"`
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
	TokenType             string    `json:"token_type"` // Always "Bearer"
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
