package model

// AppleSignInRequest represents the request body for Apple sign-in
type AppleSignInRequest struct {
	IDToken string `json:"id_token" binding:"required"`
	Nonce   string `json:"nonce" binding:"required"`
}

// AppleSignInResponse represents the response after successful authentication
type AppleSignInResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
