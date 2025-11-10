package google

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	googlePublicKeyURL  = "https://www.googleapis.com/oauth2/v3/certs"
	googleIssuer1       = "https://accounts.google.com"
	googleIssuer2       = "accounts.google.com"
	maxResponseBodySize = 1024 * 1024 // 1MB limit for response body
)

// GooglePublicKeys represents Google's public keys response
type GooglePublicKeys struct {
	Keys []GooglePublicKey `json:"keys"`
}

// GooglePublicKey represents a single Google public key
type GooglePublicKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// GoogleClaims represents the claims in Google ID token
type GoogleClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Locale        string `json:"locale"`
}

// Verifier handles Google ID token verification
type Verifier struct {
	clientID   string
	httpClient *http.Client
	mu         sync.RWMutex
	keys       map[string]*rsa.PublicKey
	lastFetch  time.Time
}

// NewVerifier creates a new Google token verifier
func NewVerifier(clientID string) *Verifier {
	return &Verifier{
		clientID: clientID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // Prevent hanging requests
		},
		keys: make(map[string]*rsa.PublicKey),
	}
}

// VerifyIDToken verifies a Google ID token and returns the claims
func (v *Verifier) VerifyIDToken(idToken string) (*GoogleClaims, error) {
	// Parse the token without verification first to get the kid
	token, err := jwt.ParseWithClaims(idToken, &GoogleClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("kid not found in token header")
		}

		// Fetch public keys if not cached or expired
		if err := v.refreshKeysIfNeeded(); err != nil {
			return nil, fmt.Errorf("failed to fetch public keys: %w", err)
		}

		// Get the public key for this kid (with read lock)
		v.mu.RLock()
		publicKey, ok := v.keys[kid]
		v.mu.RUnlock()

		if !ok {
			return nil, fmt.Errorf("public key not found for kid: %s", kid)
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract and validate claims
	claims, ok := token.Claims.(*GoogleClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Validate issuer (Google has two valid issuers)
	if claims.Issuer != googleIssuer1 && claims.Issuer != googleIssuer2 {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	// Validate audience (client ID)
	validAudience := false
	for _, aud := range claims.Audience {
		if aud == v.clientID {
			validAudience = true
			break
		}
	}
	if !validAudience {
		return nil, errors.New("invalid audience")
	}

	// Validate expiration
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	// Validate email is verified (security best practice)
	if !claims.EmailVerified {
		return nil, errors.New("email not verified by Google")
	}

	return claims, nil
}

// refreshKeysIfNeeded fetches Google's public keys if cache is stale
// Uses double-checked locking to prevent multiple concurrent fetches
func (v *Verifier) refreshKeysIfNeeded() error {
	// First check (with read lock) - fast path
	v.mu.RLock()
	needsRefresh := time.Since(v.lastFetch) >= 24*time.Hour || len(v.keys) == 0
	v.mu.RUnlock()

	if !needsRefresh {
		return nil
	}

	// Acquire write lock for second check
	v.mu.Lock()
	defer v.mu.Unlock()

	// Second check (with write lock) - prevents race condition
	// Another goroutine may have already fetched keys while we were waiting for the lock
	needsRefresh = time.Since(v.lastFetch) >= 24*time.Hour || len(v.keys) == 0
	if !needsRefresh {
		return nil
	}

	// Only one goroutine reaches here and fetches keys
	return v.fetchPublicKeysLocked()
}

// fetchPublicKeysLocked retrieves Google's public keys
// IMPORTANT: Caller must hold write lock (v.mu.Lock) before calling this method
func (v *Verifier) fetchPublicKeysLocked() error {
	resp, err := v.httpClient.Get(googlePublicKeyURL)
	if err != nil {
		return fmt.Errorf("failed to fetch keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Limit response body size to prevent memory exhaustion attacks
	limitedReader := io.LimitReader(resp.Body, maxResponseBodySize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var googleKeys GooglePublicKeys
	if err := json.Unmarshal(body, &googleKeys); err != nil {
		return fmt.Errorf("failed to unmarshal keys: %w", err)
	}

	// Convert Google's JWK to RSA public keys (with validation)
	newKeys := make(map[string]*rsa.PublicKey)
	for _, key := range googleKeys.Keys {
		// Validate key type and algorithm
		if key.Kty != "RSA" {
			// Skip non-RSA keys
			continue
		}
		if key.Alg != "RS256" {
			// Skip keys with unsupported algorithm
			continue
		}

		publicKey, err := jwkToRSAPublicKey(key)
		if err != nil {
			return fmt.Errorf("failed to convert key %s: %w", key.Kid, err)
		}
		newKeys[key.Kid] = publicKey
	}

	// Update keys and lastFetch (caller already holds write lock)
	v.keys = newKeys
	v.lastFetch = time.Now()

	return nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func jwkToRSAPublicKey(key GooglePublicKey) (*rsa.PublicKey, error) {
	// Decode the modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert bytes to big.Int
	n := new(big.Int).SetBytes(nBytes)

	// Convert exponent bytes to int
	var e int
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}
