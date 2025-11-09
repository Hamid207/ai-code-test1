package apple

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	applePublicKeyURL = "https://appleid.apple.com/auth/keys"
	appleIssuer       = "https://appleid.apple.com"
)

// ApplePublicKeys represents Apple's public keys response
type ApplePublicKeys struct {
	Keys []ApplePublicKey `json:"keys"`
}

// ApplePublicKey represents a single Apple public key
type ApplePublicKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// AppleClaims represents the claims in Apple ID token
type AppleClaims struct {
	jwt.RegisteredClaims
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Nonce         string `json:"nonce"`
	NonceSupported bool  `json:"nonce_supported"`
}

// Verifier handles Apple ID token verification
type Verifier struct {
	clientID string
	keys     map[string]*rsa.PublicKey
	lastFetch time.Time
}

// NewVerifier creates a new Apple token verifier
func NewVerifier(clientID string) *Verifier {
	return &Verifier{
		clientID: clientID,
		keys:     make(map[string]*rsa.PublicKey),
	}
}

// VerifyIDToken verifies an Apple ID token and returns the claims
func (v *Verifier) VerifyIDToken(idToken, expectedNonce string) (*AppleClaims, error) {
	// Parse the token without verification first to get the kid
	token, err := jwt.ParseWithClaims(idToken, &AppleClaims{}, func(token *jwt.Token) (interface{}, error) {
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

		// Get the public key for this kid
		publicKey, ok := v.keys[kid]
		if !ok {
			return nil, fmt.Errorf("public key not found for kid: %s", kid)
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract and validate claims
	claims, ok := token.Claims.(*AppleClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Validate issuer
	if claims.Issuer != appleIssuer {
		return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
	}

	// Validate audience (client ID)
	if !claims.VerifyAudience(v.clientID, true) {
		return nil, errors.New("invalid audience")
	}

	// Validate expiration
	if !claims.VerifyExpiresAt(time.Now(), true) {
		return nil, errors.New("token expired")
	}

	// Validate nonce
	if expectedNonce != "" && claims.Nonce != expectedNonce {
		return nil, errors.New("nonce mismatch")
	}

	return claims, nil
}

// refreshKeysIfNeeded fetches Apple's public keys if cache is stale
func (v *Verifier) refreshKeysIfNeeded() error {
	// Refresh keys every 24 hours or if not yet fetched
	if time.Since(v.lastFetch) < 24*time.Hour && len(v.keys) > 0 {
		return nil
	}

	return v.fetchPublicKeys()
}

// fetchPublicKeys retrieves Apple's public keys
func (v *Verifier) fetchPublicKeys() error {
	resp, err := http.Get(applePublicKeyURL)
	if err != nil {
		return fmt.Errorf("failed to fetch keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var appleKeys ApplePublicKeys
	if err := json.Unmarshal(body, &appleKeys); err != nil {
		return fmt.Errorf("failed to unmarshal keys: %w", err)
	}

	// Convert Apple's JWK to RSA public keys
	newKeys := make(map[string]*rsa.PublicKey)
	for _, key := range appleKeys.Keys {
		publicKey, err := jwkToRSAPublicKey(key)
		if err != nil {
			return fmt.Errorf("failed to convert key %s: %w", key.Kid, err)
		}
		newKeys[key.Kid] = publicKey
	}

	v.keys = newKeys
	v.lastFetch = time.Now()

	return nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func jwkToRSAPublicKey(key ApplePublicKey) (*rsa.PublicKey, error) {
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
