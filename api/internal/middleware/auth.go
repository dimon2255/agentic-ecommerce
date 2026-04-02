package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"

	"github.com/dimon2255/agentic-ecommerce/api/pkg/response"
)

type contextKey string

const UserIDKey contextKey = "user_id"

type AuthMiddleware struct {
	jwtSecret  []byte
	issuer     string
	audience   string
	jwksURL    string
	jwksKeys   map[string]*ecdsa.PublicKey
	jwksMu     sync.RWMutex
}

// NewAuthMiddleware creates auth middleware. Issuer and audience are optional —
// if non-empty, JWT claims are validated against them. supabaseURL is used to
// derive the JWKS endpoint for ES256 token verification.
func NewAuthMiddleware(jwtSecret, issuer, audience, supabaseURL string) *AuthMiddleware {
	m := &AuthMiddleware{
		jwtSecret: []byte(jwtSecret),
		issuer:    issuer,
		audience:  audience,
		jwksKeys:  make(map[string]*ecdsa.PublicKey),
	}
	// Derive JWKS URL from Supabase URL or issuer
	if issuer != "" {
		m.jwksURL = strings.TrimSuffix(issuer, "/") + "/.well-known/jwks.json"
	} else if supabaseURL != "" {
		m.jwksURL = strings.TrimSuffix(supabaseURL, "/") + "/auth/v1/.well-known/jwks.json"
	}
	return m
}

// parserOptions returns JWT parser options for algorithm, issuer, and audience validation.
func (m *AuthMiddleware) parserOptions() []jwt.ParserOption {
	opts := []jwt.ParserOption{jwt.WithValidMethods([]string{"HS256", "ES256"})}
	if m.issuer != "" {
		opts = append(opts, jwt.WithIssuer(m.issuer))
	}
	if m.audience != "" {
		opts = append(opts, jwt.WithAudience(m.audience))
	}
	return opts
}

// keyFunc returns the appropriate verification key based on the JWT signing algorithm.
func (m *AuthMiddleware) keyFunc(token *jwt.Token) (any, error) {
	switch token.Method.Alg() {
	case "HS256":
		return m.jwtSecret, nil
	case "ES256":
		kid, _ := token.Header["kid"].(string)
		return m.getECDSAKey(kid)
	default:
		return nil, fmt.Errorf("unsupported signing method: %s", token.Method.Alg())
	}
}

// getECDSAKey returns the ECDSA public key for the given key ID, fetching JWKS if needed.
func (m *AuthMiddleware) getECDSAKey(kid string) (*ecdsa.PublicKey, error) {
	m.jwksMu.RLock()
	key, ok := m.jwksKeys[kid]
	m.jwksMu.RUnlock()
	if ok {
		return key, nil
	}

	if err := m.fetchJWKS(); err != nil {
		return nil, fmt.Errorf("fetch JWKS: %w", err)
	}

	m.jwksMu.RLock()
	key, ok = m.jwksKeys[kid]
	m.jwksMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("key %q not found in JWKS", kid)
	}
	return key, nil
}

// fetchJWKS fetches and caches ECDSA public keys from the JWKS endpoint.
func (m *AuthMiddleware) fetchJWKS() error {
	if m.jwksURL == "" {
		return fmt.Errorf("JWKS URL not configured")
	}

	resp, err := http.Get(m.jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kty string `json:"kty"`
			Crv string `json:"crv"`
			X   string `json:"x"`
			Y   string `json:"y"`
			Kid string `json:"kid"`
			Alg string `json:"alg"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	m.jwksMu.Lock()
	defer m.jwksMu.Unlock()
	for _, k := range jwks.Keys {
		if k.Kty != "EC" || k.Crv != "P-256" {
			continue
		}
		xBytes, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil {
			continue
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(k.Y)
		if err != nil {
			continue
		}
		m.jwksKeys[k.Kid] = &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}
	}
	return nil
}

// OptionalAuth extracts user ID from JWT if present. Request proceeds regardless.
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := extractBearerToken(r)
		if tokenStr != "" {
			token, err := jwt.Parse(tokenStr, m.keyFunc, m.parserOptions()...)
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
		token, err := jwt.Parse(tokenStr, m.keyFunc, m.parserOptions()...)
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
