package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type AuthMiddleware struct {
	jwtSecret []byte
	issuer    string
	audience  string
}

// NewAuthMiddleware creates auth middleware. Issuer and audience are optional —
// if non-empty, JWT claims are validated against them.
func NewAuthMiddleware(jwtSecret, issuer, audience string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: []byte(jwtSecret),
		issuer:    issuer,
		audience:  audience,
	}
}

// parserOptions returns JWT parser options for algorithm, issuer, and audience validation.
func (m *AuthMiddleware) parserOptions() []jwt.ParserOption {
	opts := []jwt.ParserOption{jwt.WithValidMethods([]string{"HS256"})}
	if m.issuer != "" {
		opts = append(opts, jwt.WithIssuer(m.issuer))
	}
	if m.audience != "" {
		opts = append(opts, jwt.WithAudience(m.audience))
	}
	return opts
}

// OptionalAuth extracts user ID from JWT if present. Request proceeds regardless.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractBearerToken(r)
		if tokenStr != "" {
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
				return m.jwtSecret, nil
			}, m.parserOptions()...)
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if sub, ok := claims["sub"].(string); ok {
						ctx := context.WithValue(r.Context(), UserIDKey, sub)
						r = r.WithContext(ctx)
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth rejects requests without a valid JWT. Returns 401 on failure.
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractBearerToken(r)
		if tokenStr == "" {
			response.Error(w, http.StatusUnauthorized, "missing authorization token")
			return
		}
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			return m.jwtSecret, nil
		}, m.parserOptions()...)
		if err != nil || !token.Valid {
			response.Error(w, http.StatusUnauthorized, "invalid token")
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "invalid token claims")
			return
		}
		sub, ok := claims["sub"].(string)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "invalid token subject")
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, sub)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// GetUserID extracts the authenticated user ID from the request context.
func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}
