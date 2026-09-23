package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthHandler(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.New()
	router.GET("/api/v1/health", HealthHandler)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	recorder := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(recorder, req)

	// Check status code
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	// Check Content-Type
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want %q", contentType, "application/json; charset=utf-8")
	}

	// Check response body contains expected fields
	body := recorder.Body.String()
	if body == "" {
		t.Fatalf("response body is empty")
	}

	// Verify JSON contains expected fields
	expectedFields := []string{`"status":"ok"`, `"service":"go-learning"`, `"version":"0.1.0"`}
	for _, field := range expectedFields {
		if !strings.Contains(body, field) {
			t.Fatalf("response body %q does not contain %q", body, field)
		}
	}
}

func TestHealthHandlerMethodNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/api/v1/health", HealthHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/health", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
