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

const panicMessage = "boom: connection to 10.0.0.7 refused"

// newPanicRouter mounts the middleware in the same order as cmd/server/main.go
// and exposes one route that panics.
func newPanicRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID(), Logging(), Recovery())
	router.GET("/boom", func(c *gin.Context) {
		panic(panicMessage)
	})
	return router
}

// serveExpectingNoPanic fails the test cleanly if the panic escapes the
// middleware chain. Letting it escape would abort the whole test binary and
// hide every other failure in this package.
func serveExpectingNoPanic(t *testing.T, router *gin.Engine, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()

	recorder := httptest.NewRecorder()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic escaped the middleware chain (%v) - Recovery did not handle it", r)
		}
	}()

	router.ServeHTTP(recorder, req)
	return recorder
}

// captureLog redirects the standard logger into a buffer for the duration of
// the test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	buf := &bytes.Buffer{}
	originalOutput := log.Writer()
	originalFlags := log.Flags()
	log.SetOutput(buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(originalOutput)
		log.SetFlags(originalFlags)
	})
	return buf
}

func TestRecoveryReturnsErrorEnvelopeOnPanic(t *testing.T) {
	captureLog(t)
	router := newPanicRouter()

	recorder := serveExpectingNoPanic(t, router, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}

	body := recorder.Body.String()
	for _, want := range []string{`"error":{`, `"code":"INTERNAL_ERROR"`, `"message":`} {
		if !strings.Contains(body, want) {
			t.Fatalf("response body %q does not contain %q", body, want)
		}
	}
}

// The panic value can contain hostnames, paths or SQL fragments. It belongs in
// the log, not in the response.
func TestRecoveryDoesNotLeakPanicValueToClient(t *testing.T) {
	captureLog(t)
	router := newPanicRouter()

	recorder := serveExpectingNoPanic(t, router, httptest.NewRequest(http.MethodGet, "/boom", nil))

	if strings.Contains(recorder.Body.String(), "10.0.0.7") {
		t.Fatalf("response body %q leaks the panic value", recorder.Body.String())
	}
}

// A panic with no log line is worse than a crash: there is nothing to
// investigate afterwards.
func TestRecoveryLogsPanicWithRequestID(t *testing.T) {
	buf := captureLog(t)
	router := newPanicRouter()

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	req.Header.Set(RequestIDHeader, "test-request-id")

	serveExpectingNoPanic(t, router, req)

	logged := buf.String()
	if !strings.Contains(logged, "test-request-id") {
		t.Fatalf("log output %q does not contain the request ID", logged)
	}
	if !strings.Contains(logged, "10.0.0.7") {
		t.Fatalf("log output %q does not contain the panic value", logged)
	}
}

// Recovery must sit inside Logging, otherwise the panic unwinds past Logging's
// post-Next() code and the request never appears in the access log.
func TestRecoveryLetsLoggingRecordThePanickedRequest(t *testing.T) {
	buf := captureLog(t)
	router := newPanicRouter()

	serveExpectingNoPanic(t, router, httptest.NewRequest(http.MethodGet, "/boom", nil))

	logged := buf.String()
	if !strings.Contains(logged, "status=500") {
		t.Fatalf("log output %q has no access log line with status=500", logged)
	}
}

func TestRecoveryLeavesNormalRequestsAlone(t *testing.T) {
	captureLog(t)
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestID(), Logging(), Recovery())
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	recorder := serveExpectingNoPanic(t, router, httptest.NewRequest(http.MethodGet, "/ping", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); body != "pong" {
		t.Fatalf("body = %q, want %q", body, "pong")
	}
}
