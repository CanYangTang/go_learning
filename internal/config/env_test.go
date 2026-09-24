package config

import (
	"os"
	"testing"
)

// JWTSecret must return the env value when set, and fall back to the
// development default otherwise. The default is dev-only; production overrides
// it via JWT_SECRET.
func TestJWTSecret(t *testing.T) {
	// Ensure no ambient value leaks in from the runner's environment.
	os.Unsetenv("JWT_SECRET")
	if got := JWTSecret(); got != devJWTSecret {
		t.Fatalf("JWTSecret() without env = %q, want dev default %q", got, devJWTSecret)
	}

	// t.Setenv restores the previous value at test end and forbids t.Parallel.
	t.Setenv("JWT_SECRET", "prod-signing-key")
	if got := JWTSecret(); got != "prod-signing-key" {
		t.Fatalf("JWTSecret() with env = %q, want %q", got, "prod-signing-key")
	}
}

func TestAllowedOrigins(t *testing.T) {
	t.Run("default when unset", func(t *testing.T) {
		os.Unsetenv("CORS_ALLOWED_ORIGINS")
		got := AllowedOrigins()
		if len(got) != 1 || got[0] != defaultAllowedOrigin {
			t.Fatalf("AllowedOrigins() without env = %v, want [%s]", got, defaultAllowedOrigin)
		}
	})

	t.Run("single value", func(t *testing.T) {
		t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")
		got := AllowedOrigins()
		if len(got) != 1 || got[0] != "https://app.example.com" {
			t.Fatalf("AllowedOrigins() = %v, want single element", got)
		}
	})

	t.Run("comma separated with whitespace and empties trimmed", func(t *testing.T) {
		t.Setenv("CORS_ALLOWED_ORIGINS", " https://a.com , ,https://b.com ")
		got := AllowedOrigins()
		want := []string{"https://a.com", "https://b.com"}
		if len(got) != len(want) {
			t.Fatalf("AllowedOrigins() = %v, want %v (empties dropped, each trimmed)", got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("AllowedOrigins()[%d] = %q, want %q", i, got[i], want[i])
			}
		}
	})
}
