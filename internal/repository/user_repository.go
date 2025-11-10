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

// CreateOrGet creates a new user or returns existing user (atomic upsert)
// This uses PostgreSQL's ON CONFLICT to prevent race conditions
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

	// Use PostgreSQL UPSERT to handle concurrent inserts atomically
	query := `
		INSERT INTO users (apple_id, email)
		VALUES ($1, $2)
		ON CONFLICT (apple_id)
		DO UPDATE SET
			email = EXCLUDED.email,
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

// CreateOrGetWithGoogle creates a new user or returns existing user for Google auth (atomic upsert)
// This uses PostgreSQL's ON CONFLICT to prevent race conditions
func (r *UserRepository) CreateOrGetWithGoogle(ctx context.Context, googleID, email string) (*model.User, error) {
	// Validate input
	if googleID == "" {
		return nil, fmt.Errorf("google_id cannot be empty")
	}

	if err := validator.ValidateEmail(email); err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	// Create context with timeout to prevent hanging queries
	ctx, cancel := context.WithTimeout(ctx, DefaultQueryTimeout)
	defer cancel()

	// Use PostgreSQL UPSERT to handle concurrent inserts atomically
	query := `
		INSERT INTO users (google_id, email)
		VALUES ($1, $2)
		ON CONFLICT (google_id)
		DO UPDATE SET
			email = EXCLUDED.email,
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
