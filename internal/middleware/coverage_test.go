package middleware

import (
	"bytes"
	"encoding/hex"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// newRequestID returns 16 random bytes hex-encoded (32 chars). The rand.Read
// error branch that falls back to fallbackRequestID is effectively unreachable
// (a CSPRNG failure means the OS is broken), so we cover the happy path here
// and fallbackRequestID directly below rather than injecting a fake rand.
func TestNewRequestIDIsHex32(t *testing.T) {
	id := newRequestID()
	if len(id) != 32 {
		t.Fatalf("newRequestID() = %q (len %d), want 32 hex chars", id, len(id))
	}
	if _, err := hex.DecodeString(id); err != nil {
		t.Fatalf("newRequestID() = %q, not valid hex: %v", id, err)
	}
}

func TestFallbackRequestID(t *testing.T) {
	if got := fallbackRequestID(); got != "request-id" {
		t.Fatalf("fallbackRequestID() = %q, want %q", got, "request-id")
	}
}

// UserIDFromContext has three outcomes: key absent, key set to a uint, and key
// set to a non-uint (failed type assertion). All three must be exercised.
func TestUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("absent key", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		if id, ok := UserIDFromContext(c); ok || id != 0 {
			t.Fatalf("UserIDFromContext() = (%d,%v), want (0,false) when unset", id, ok)
		}
	})

	t.Run("uint value", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(userIDContextKey, uint(7))
		if id, ok := UserIDFromContext(c); !ok || id != 7 {
			t.Fatalf("UserIDFromContext() = (%d,%v), want (7,true)", id, ok)
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(userIDContextKey, "not-a-uint")
		if id, ok := UserIDFromContext(c); ok || id != 0 {
			t.Fatalf("UserIDFromContext() = (%d,%v), want (0,false) on type mismatch", id, ok)
		}
	})
}

// Logging has two branches: with and without a request id. The existing suite
// covers the with-id path (RequestID runs first); here we mount Logging alone
// so RequestIDFromContext is empty and the no-id branch is taken. The assertion
// is that Logging never alters the response.
func TestLoggingMiddlewareWithoutRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	buf := &bytes.Buffer{}
	originalOutput := log.Writer()
	originalFlags := log.Flags()
	originalPrefix := log.Prefix()
	log.SetOutput(buf)
	log.SetFlags(0)
	log.SetPrefix("")
	t.Cleanup(func() {
		log.SetOutput(originalOutput)
		log.SetFlags(originalFlags)
		log.SetPrefix(originalPrefix)
	})

	router := gin.New()
	router.Use(Logging()) // no RequestID() -> RequestIDFromContext is empty
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); body != "pong" {
		t.Fatalf("body = %q, Logging must not alter the response", body)
	}

	logLine := buf.String()
	if strings.Contains(logLine, "request_id=") {
		t.Fatalf("log output = %q, want no request_id field on the no-id branch", logLine)
	}
	if !strings.Contains(logLine, "method=GET") || !strings.Contains(logLine, "status=200") {
		t.Fatalf("log output = %q, want method/status fields", logLine)
	}
}
