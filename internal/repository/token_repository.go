package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TokenRepository handles database operations for refresh tokens
type TokenRepository struct {
	db *pgxpool.Pool
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

// hashToken creates a SHA256 hash of the token for storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// StoreRefreshToken stores a refresh token in the database
func (r *TokenRepository) StoreRefreshToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	tokenHash := hashToken(token)

	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to store refresh token: %w", err)
	}

	return nil
}

// ValidateRefreshToken validates a refresh token and returns the associated user ID
func (r *TokenRepository) ValidateRefreshToken(ctx context.Context, token string) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	tokenHash := hashToken(token)

	query := `
		SELECT user_id, expires_at, revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var userID int64
	var expiresAt time.Time
	var revokedAt *time.Time

	err := r.db.QueryRow(ctx, query, tokenHash).Scan(&userID, &expiresAt, &revokedAt)
	if err == pgx.ErrNoRows {
		return 0, fmt.Errorf("refresh token not found")
	}
	if err != nil {
		return 0, fmt.Errorf("failed to validate refresh token: %w", err)
	}

	// Check if token is revoked
	if revokedAt != nil {
		return 0, fmt.Errorf("refresh token has been revoked")
	}

	// Check if token is expired
	if time.Now().After(expiresAt) {
		return 0, fmt.Errorf("refresh token has expired")
	}

	// Update last_used_at
	updateQuery := `
		UPDATE refresh_tokens
		SET last_used_at = CURRENT_TIMESTAMP
		WHERE token_hash = $1
	`
	_, _ = r.db.Exec(ctx, updateQuery, tokenHash)

	return userID, nil
}

// RevokeRefreshToken revokes a specific refresh token
func (r *TokenRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	tokenHash := hashToken(token)

	query := `
		UPDATE refresh_tokens
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE token_hash = $1 AND revoked_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("refresh token not found or already revoked")
	}

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a specific user
func (r *TokenRepository) RevokeAllUserTokens(ctx context.Context, userID int64) error {
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	query := `
		UPDATE refresh_tokens
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	return nil
}

// CleanupExpiredTokens removes expired refresh tokens from the database
func (r *TokenRepository) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < CURRENT_TIMESTAMP
	`

	result, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}

	return result.RowsAffected(), nil
}
