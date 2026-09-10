package handler

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CanYangTang/go_learning/internal/middleware"
	"github.com/gin-gonic/gin"
)

// newFullRouter mirrors the global middleware stack and route layout of
// cmd/server/main.go, so the 404 path is exercised the same way it is in
// production. A bare gin.New() would not prove the request ID reaches a 404.
func newFullRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	silenceLog(t)

	router := gin.New()
	router.Use(
		middleware.RequestID(),
		middleware.Logging(),
		middleware.Recovery(),
		middleware.CORS(),
		middleware.AuthPlaceholder(),
	)
	router.GET("/api/v1/health", HealthHandler)
	router.NoRoute(NotFoundHandler)
	return router
}

// silenceLog keeps the access log out of the test output.
func silenceLog(t *testing.T) {
	t.Helper()

	original := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(original) })
}

func get(t *testing.T, router *gin.Engine, path string) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
	return recorder
}

func TestNotFoundUsesSharedErrorEnvelope(t *testing.T) {
	router := newFullRouter(t)

	recorder := get(t, router, "/api/v1/nope")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}

	body := recorder.Body.String()
	assertBodyContains(t, body, `"error":{`, `"code":"NOT_FOUND"`, `"message":`)

	// The old shape was {"error":"not found"} - error as a string. Any client
	// reading resp.error.code would break on it.
	if strings.Contains(body, `"error":"`) {
		t.Fatalf("response body %q still uses a string for the error field", body)
	}
}

// A bare /health (no /api/v1 prefix) is what docs/api/todo-api.md wrongly
// advertises, so it is the 404 a real client is most likely to hit.
func TestNotFoundHandlesBareHealthPath(t *testing.T) {
	router := newFullRouter(t)

	recorder := get(t, router, "/health")

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	assertBodyContains(t, recorder.Body.String(), `"code":"NOT_FOUND"`)
}

// NoRoute handlers do run the global middleware chain. This is what makes a
// client-reported 404 traceable in the logs.
func TestNotFoundResponseCarriesRequestIDAndCORSHeaders(t *testing.T) {
	router := newFullRouter(t)

	recorder := get(t, router, "/api/v1/nope")

	if got := strings.TrimSpace(recorder.Header().Get(middleware.RequestIDHeader)); got == "" {
		t.Fatalf("expected %s on a 404 response", middleware.RequestIDHeader)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got == "" {
		t.Fatalf("expected Access-Control-Allow-Origin on a 404 response")
	}
}

func TestKnownRouteIsUnaffectedByNoRoute(t *testing.T) {
	router := newFullRouter(t)

	recorder := get(t, router, "/api/v1/health")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	assertBodyContains(t, recorder.Body.String(), `"status":"ok"`)
}
