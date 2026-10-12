package middleware

import (
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"Fhub/auth"
	"Fhub/models"
	"Fhub/utils"
)

type TokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

func NewTokenBlacklist() *TokenBlacklist {
	return &TokenBlacklist{tokens: make(map[string]time.Time)}
}

func (b *TokenBlacklist) Revoke(token string, expiresAt time.Time) {
	if !time.Now().Before(expiresAt) {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens[token] = expiresAt
}

func (b *TokenBlacklist) IsRevoked(token string) bool {
	b.mu.RLock()
	expiresAt, exists := b.tokens[token]
	b.mu.RUnlock()
	if !exists {
		return false
	}
	if !time.Now().Before(expiresAt) {
		b.mu.Lock()
		delete(b.tokens, token)
		b.mu.Unlock()
		return false
	}
	return true
}

// Authenticate rejects requests without a valid, non-revoked bearer token.
func Authenticate(tokens *auth.JWTManager, blacklist *TokenBlacklist) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || tokenString == "" {
				utils.Error(w, http.StatusUnauthorized, "missing or malformed authorization header")
				return
			}
			claims, err := tokens.Verify(tokenString)
			if err != nil {
				utils.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			if blacklist.IsRevoked(tokenString) {
				utils.Error(w, http.StatusUnauthorized, "token has been revoked")
				return
			}
			next.ServeHTTP(w, r.WithContext(auth.WithClaims(r.Context(), claims)))
		})
	}
}

// RequireRole must be used after Authenticate.
func RequireRole(roles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := auth.ClaimsFromContext(r.Context())
			if err != nil {
				utils.Error(w, http.StatusUnauthorized, "unauthenticated")
				return
			}
			if !slices.Contains(roles, claims.Role) {
				utils.Error(w, http.StatusForbidden, "you do not have permission to perform this action")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
