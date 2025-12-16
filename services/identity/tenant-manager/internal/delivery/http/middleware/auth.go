package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/cotai/tenant-manager/internal/delivery/http/dto"
	"go.uber.org/zap"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtValidator JWTValidator
	logger       *zap.Logger
}

// JWTValidator interface for validating JWT tokens
type JWTValidator interface {
	ValidateToken(tokenString string) (*TokenClaims, error)
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	Subject  string
	Email    string
	Roles    []string
	TenantID string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtValidator JWTValidator, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtValidator: jwtValidator,
		logger:       logger,
	}
}

// Handler returns the authentication middleware handler
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondError(w, http.StatusUnauthorized, "MISSING_AUTH_HEADER", "Authorization header is required")
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.respondError(w, http.StatusUnauthorized, "INVALID_AUTH_HEADER", "Authorization header must be Bearer token")
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := m.jwtValidator.ValidateToken(tokenString)
		if err != nil {
			m.logger.Warn("JWT validation failed",
				zap.Error(err),
				zap.String("path", r.URL.Path),
			)
			m.respondError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired token")
			return
		}

		// Store claims in context
		ctx := context.WithValue(r.Context(), "claims", claims)
		ctx = context.WithValue(ctx, "user_id", claims.Subject)

		m.logger.Debug("Request authenticated",
			zap.String("user_id", claims.Subject),
			zap.String("path", r.URL.Path),
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns a middleware that checks for specific role
func (m *AuthMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("claims").(*TokenClaims)
			if !ok {
				m.respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range claims.Roles {
				if role == requiredRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.logger.Warn("Insufficient permissions",
					zap.String("user_id", claims.Subject),
					zap.String("required_role", requiredRole),
					zap.Strings("user_roles", claims.Roles),
				)
				m.respondError(w, http.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// respondError sends an error response
func (m *AuthMiddleware) respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = dto.ErrorResponse{
		Error: dto.ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	// Note: In production, use proper JSON encoding
	w.Write([]byte(`{"error":{"code":"` + code + `","message":"` + message + `"}}`))
}
