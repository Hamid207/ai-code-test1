package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// Email validation regex (simplified but covers most cases)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Trim spaces
	email = strings.TrimSpace(email)

	// Check length
	if len(email) > 255 {
		return fmt.Errorf("email is too long (max 255 characters)")
	}

	if len(email) < 3 {
		return fmt.Errorf("email is too short")
	}

	// Validate format
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateAppleID validates Apple ID format
func ValidateAppleID(appleID string) error {
	if appleID == "" {
		return fmt.Errorf("apple_id is required")
	}

	// Trim spaces
	appleID = strings.TrimSpace(appleID)

	// Check length
	if len(appleID) > 255 {
		return fmt.Errorf("apple_id is too long (max 255 characters)")
	}

	if len(appleID) < 10 {
		return fmt.Errorf("apple_id is too short")
	}

	return nil
}
