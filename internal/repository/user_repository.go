package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Hamid207/ai-code-test1/internal/model"
	"github.com/Hamid207/ai-code-test1/pkg/validator"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// DefaultQueryTimeout is the default timeout for database queries
	DefaultQueryTimeout = 5 * time.Second
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
	// Create context with timeout to prevent hanging queries
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	query := `
		SELECT id, apple_id, google_id, email, created_at, updated_at
		FROM users
		WHERE apple_id = $1
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, appleID).Scan(
		&user.ID,
		&user.AppleID,
		&user.GoogleID,
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

// GetByGoogleID retrieves a user by their Google ID
func (r *UserRepository) GetByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	// Create context with timeout to prevent hanging queries
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	query := `
		SELECT id, apple_id, google_id, email, created_at, updated_at
		FROM users
		WHERE google_id = $1
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, googleID).Scan(
		&user.ID,
		&user.AppleID,
		&user.GoogleID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // User not found
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by google_id: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by their email address
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	// Create context with timeout to prevent hanging queries
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	query := `
		SELECT id, apple_id, google_id, email, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.AppleID,
		&user.GoogleID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // User not found
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Create creates a new user in the database
func (r *UserRepository) Create(ctx context.Context, appleID, email string) (*model.User, error) {
	// Create context with timeout to prevent hanging queries
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	query := `
		INSERT INTO users (apple_id, email)
		VALUES ($1, $2)
		RETURNING id, apple_id, google_id, email, created_at, updated_at
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, appleID, email).Scan(
		&user.ID,
		&user.AppleID,
		&user.GoogleID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// CreateOrGet creates a new user or returns existing user (atomic upsert with account linking)
// If a user with the same email exists, links the Apple ID to that account
// This prevents duplicate accounts when users sign in with different providers
func (r *UserRepository) CreateOrGet(ctx context.Context, appleID, email string) (*model.User, error) {
	// Validate input
	if err := validator.ValidateAppleID(appleID); err != nil {
		return nil, fmt.Errorf("invalid apple_id: %w", err)
	}

	if err := validator.ValidateEmail(email); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Create context with timeout to prevent hanging queries
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	// Use PostgreSQL UPSERT with email conflict handling for account linking
	// Priority: email > apple_id (email takes precedence for linking accounts)
	query := `
		INSERT INTO users (apple_id, email)
		VALUES ($1, $2)
		ON CONFLICT (email)
		DO UPDATE SET
			apple_id = COALESCE(users.apple_id, EXCLUDED.apple_id),
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, apple_id, google_id, email, created_at, updated_at
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, appleID, email).Scan(
		&user.ID,
		&user.AppleID,
		&user.GoogleID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create or update user: %w", err)
	}

	return &user, nil
}

// CreateOrGetWithGoogle creates a new user or returns existing user for Google auth (atomic upsert with account linking)
// If a user with the same email exists, links the Google ID to that account
// This prevents duplicate accounts when users sign in with different providers
func (r *UserRepository) CreateOrGetWithGoogle(ctx context.Context, googleID, email string) (*model.User, error) {
	// Validate input
	if err := validator.ValidateGoogleID(googleID); err != nil {
		return nil, fmt.Errorf("invalid google_id: %w", err)
	}

	if err := validator.ValidateEmail(email); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Create context with timeout to prevent hanging queries
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	// Use PostgreSQL UPSERT with email conflict handling for account linking
	// Priority: email > google_id (email takes precedence for linking accounts)
	query := `
		INSERT INTO users (google_id, email)
		VALUES ($1, $2)
		ON CONFLICT (email)
		DO UPDATE SET
			google_id = COALESCE(users.google_id, EXCLUDED.google_id),
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, apple_id, google_id, email, created_at, updated_at
	`

	var user model.User
	err := r.db.QueryRow(ctx, query, googleID, email).Scan(
		&user.ID,
		&user.AppleID,
		&user.GoogleID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create or update user with google: %w", err)
	}

	return &user, nil
}
