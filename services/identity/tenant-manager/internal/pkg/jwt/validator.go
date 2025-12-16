package jwt

import (
	"fmt"

	"github.com/cotai/tenant-manager/internal/delivery/http/middleware"
)

// Validator validates JWT tokens
type Validator struct {
	publicKeyURL string
	issuer       string
}

// NewValidator creates a new JWT validator
func NewValidator(publicKeyURL, issuer string) *Validator {
	return &Validator{
		publicKeyURL: publicKeyURL,
		issuer:       issuer,
	}
}

// ValidateToken validates a JWT token and returns claims
// TODO: Implement full JWT validation with Keycloak JWKS
func (v *Validator) ValidateToken(tokenString string) (*middleware.TokenClaims, error) {
	// TODO: Implement actual JWT validation:
	// 1. Fetch public key from Keycloak JWKS endpoint
	// 2. Parse and verify token signature
	// 3. Validate issuer, expiry, etc.
	// 4. Extract claims

	// For now, return error to indicate not implemented
	return nil, fmt.Errorf("JWT validation not yet implemented - needs Keycloak integration")
}
