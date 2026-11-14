package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	domain_auth "Grippy/internal/domain/auth"
)

type contextKey string

const UserKey contextKey = "user_info"

type TokenParser interface {
	ParseAccessToken(tokenStr string) (*domain_auth.Claims, error)
}

func NewAuthMiddleware(parser TokenParser) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnauthorized(w, "Missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" || strings.TrimSpace(parts[1]) == "" {
				writeUnauthorized(w, "Invalid token format")
				return
			}

			accessToken := strings.TrimSpace(parts[1])

			claims, err := parser.ParseAccessToken(accessToken)
			if err != nil {
				writeUnauthorized(w, "Invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserClaims(ctx context.Context) (*domain_auth.Claims, bool) {
	claims, ok := ctx.Value(UserKey).(*domain_auth.Claims)
	return claims, ok
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	claims, ok := GetUserClaims(ctx)
	if !ok || claims == nil {
		return uuid.Nil, false
	}
	return claims.ID, true
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
