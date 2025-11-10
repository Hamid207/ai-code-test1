package model

import "time"

// User represents a user in the database
type User struct {
	ID        int64     `json:"id" db:"id"`
	AppleID   string    `json:"apple_id,omitempty" db:"apple_id"`
	GoogleID  string    `json:"google_id,omitempty" db:"google_id"`
	Email     string    `json:"email" db:"email"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
