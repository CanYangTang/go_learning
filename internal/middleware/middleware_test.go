package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddlewareGeneratesAndExposesID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, RequestIDFromContext(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	requestID := strings.TrimSpace(recorder.Header().Get(RequestIDHeader))
	if requestID == "" {
		t.Fatalf("expected %s header to be set", RequestIDHeader)
	}

	if body := strings.TrimSpace(recorder.Body.String()); body != requestID {
		t.Fatalf("body = %q, want %q", body, requestID)
	}
}

func TestRequestIDMiddlewarePreservesIncomingID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, RequestIDFromContext(c))
	})

	const incomingID = "client-request-id"
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(RequestIDHeader, incomingID)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if got := strings.TrimSpace(recorder.Header().Get(RequestIDHeader)); got != incomingID {
		t.Fatalf("header = %q, want %q", got, incomingID)
	}

	if body := strings.TrimSpace(recorder.Body.String()); body != incomingID {
		t.Fatalf("body = %q, want %q", body, incomingID)
	}
}

func TestLoggingMiddlewareWritesRequestDetails(t *testing.T) {
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
	router.Use(RequestID(), Logging())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping?foo=bar", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	logLine := buf.String()
	if !strings.Contains(logLine, "request_id=") {
		t.Fatalf("log output = %q, want request_id field", logLine)
	}
	if !strings.Contains(logLine, "method=GET") {
		t.Fatalf("log output = %q, want method=GET", logLine)
	}
	if !strings.Contains(logLine, "path=/ping?foo=bar") {
		t.Fatalf("log output = %q, want path=/ping?foo=bar", logLine)
	}
	if !strings.Contains(logLine, "status=200") {
		t.Fatalf("log output = %q, want status=200", logLine)
	}
}

func TestCORSMiddlewareHandlesPreflight(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "http://example.com")
	}

	if got := recorder.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, http.MethodGet) {
		t.Fatalf("Access-Control-Allow-Methods = %q, want %q", got, http.MethodGet)
	}

	if got := recorder.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, RequestIDHeader) {
		t.Fatalf("Access-Control-Allow-Headers = %q, want %q", got, RequestIDHeader)
	}
}

func TestAuthPlaceholderPassesThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(AuthPlaceholder())
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
		t.Fatalf("body = %q, want %q", body, "pong")
	}
}
