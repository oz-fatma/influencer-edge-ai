package middleware

import (
	"net/http"
	"strings"
)

// ExtractBearerToken returns a JWT from the query ?token= parameter or Authorization header.
// WebSocket browser clients typically cannot set Authorization headers, so query param is supported.
func ExtractBearerToken(r *http.Request) string {
	if token := strings.TrimSpace(r.URL.Query().Get("token")); token != "" {
		return token
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
