package redis

import "fmt"

// Key prefixes for different data types
// Following naming convention: <namespace>:<entity>:<identifier>
const (
	// Token-related keys
	PrefixRefreshToken  = "refresh"      // refresh:<user_id>:<token_id>
	PrefixBlacklist     = "blacklist"    // blacklist:<token_id>
	PrefixTokenFamily   = "token_family" // token_family:<family_id>

	// Rate limiting keys
	PrefixRateLimitUser = "ratelimit:user" // ratelimit:user:<user_id>
	PrefixRateLimitIP   = "ratelimit:ip"   // ratelimit:ip:<ip_address>

	// Cache keys
	PrefixUserCache    = "cache:user"    // cache:user:<user_id>
	PrefixProfileCache = "cache:profile" // cache:profile:<user_id>
)

// KeyBuilder provides methods to build Redis keys consistently
type KeyBuilder struct{}

// NewKeyBuilder creates a new KeyBuilder instance
func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{}
}

// RefreshToken builds a key for storing refresh tokens
// Format: refresh:<user_id>:<token_id>
func (kb *KeyBuilder) RefreshToken(userID, tokenID string) string {
	return fmt.Sprintf("%s:%s:%s", PrefixRefreshToken, userID, tokenID)
}

// RefreshTokenPattern builds a pattern for scanning all refresh tokens for a user
// Format: refresh:<user_id>:*
func (kb *KeyBuilder) RefreshTokenPattern(userID string) string {
	return fmt.Sprintf("%s:%s:*", PrefixRefreshToken, userID)
}

// BlacklistToken builds a key for blacklisted tokens
// Format: blacklist:<token_id>
func (kb *KeyBuilder) BlacklistToken(tokenID string) string {
	return fmt.Sprintf("%s:%s", PrefixBlacklist, tokenID)
}

// TokenFamily builds a key for token family tracking
// Format: token_family:<family_id>
func (kb *KeyBuilder) TokenFamily(familyID string) string {
	return fmt.Sprintf("%s:%s", PrefixTokenFamily, familyID)
}

// RateLimitUser builds a key for user-based rate limiting
// Format: ratelimit:user:<user_id>
func (kb *KeyBuilder) RateLimitUser(userID string) string {
	return fmt.Sprintf("%s:%s", PrefixRateLimitUser, userID)
}

// RateLimitIP builds a key for IP-based rate limiting
// Format: ratelimit:ip:<ip_address>
func (kb *KeyBuilder) RateLimitIP(ipAddress string) string {
	return fmt.Sprintf("%s:%s", PrefixRateLimitIP, ipAddress)
}

// UserCache builds a key for caching user data
// Format: cache:user:<user_id>
func (kb *KeyBuilder) UserCache(userID string) string {
	return fmt.Sprintf("%s:%s", PrefixUserCache, userID)
}

// ProfileCache builds a key for caching user profile data
// Format: cache:profile:<user_id>
func (kb *KeyBuilder) ProfileCache(userID string) string {
	return fmt.Sprintf("%s:%s", PrefixProfileCache, userID)
}
