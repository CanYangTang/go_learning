package config

import (
	"os"
	"strings"
)

// devJWTSecret is the fallback signing key for local development only.
// Production MUST override it via the JWT_SECRET environment variable — a
// leaked signing key lets anyone forge a token for any user.
const devJWTSecret = "dev-only-insecure-jwt-secret-change-me"

// defaultAllowedOrigin is the single origin allowed for CORS by default,
// matching a typical local frontend dev server.
const defaultAllowedOrigin = "http://localhost:3000"

// JWTSecret returns the JWT signing key from the JWT_SECRET environment
// variable, falling back to a development-only default. The returned value is
// a secret: never log it.
func JWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return devJWTSecret
}

// AllowedOrigins returns the CORS whitelist from the CORS_ALLOWED_ORIGINS
// environment variable (comma-separated), falling back to a local dev origin.
func AllowedOrigins() []string {
	raw := os.Getenv("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		return []string{defaultAllowedOrigin}
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
