package repository

import (
	"context"
	"fmt"

	"github.com/Hamid207/ai-code-test1/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetByAppleID retrieves a user by their Apple ID
func (r *UserRepository) GetByAppleID(ctx context.Context, appleID string) (*model.User, error) {
	query := `
		SELECT id, apple_id, email, created_at, updated_at
		FROM users
		WHERE apple_id = $1
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, appleID).Scan(
		&user.ID,
		&user.AppleID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // User not found
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by apple_id: %w", err)
	}

	return &user, nil
}

// Create creates a new user in the database
func (r *UserRepository) Create(ctx context.Context, appleID, email string) (*model.User, error) {
	query := `
		INSERT INTO users (apple_id, email)
		VALUES ($1, $2)
		RETURNING id, apple_id, email, created_at, updated_at
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, appleID, email).Scan(
		&user.ID,
		&user.AppleID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// CreateOrGet creates a new user or returns existing user (upsert pattern)
func (r *UserRepository) CreateOrGet(ctx context.Context, appleID, email string) (*model.User, error) {
	// First, try to get the existing user
	user, err := r.GetByAppleID(ctx, appleID)
	if err != nil {
		return nil, err
	}

	// If user exists, return it
	if user != nil {
		// Update email if it changed
		if user.Email != email {
			query := `
				UPDATE users
				SET email = $1
				WHERE apple_id = $2
				RETURNING id, apple_id, email, created_at, updated_at
			`
			err := r.db.QueryRow(ctx, query, email, appleID).Scan(
				&user.ID,
				&user.AppleID,
				&user.Email,
				&user.CreatedAt,
				&user.UpdatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to update user email: %w", err)
			}
		}
		return user, nil
	}

	// User doesn't exist, create new one
	return r.Create(ctx, appleID, email)
}
