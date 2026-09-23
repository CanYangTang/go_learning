package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/CanYangTang/go_learning/pkg/apperror"
	"github.com/gin-gonic/gin"
)

// newTestContext returns a gin.Context wired to a fresh recorder, so a call to
// WriteError/WriteSuccess can be inspected without a full router.
func newTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, rec
}

func TestWriteErrorUsesAppErrorStatusAndCode(t *testing.T) {
	c, rec := newTestContext()

	WriteError(c, apperror.NotFound("todo not found"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	var body ErrorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Error.Code != "NOT_FOUND" || body.Error.Message != "todo not found" {
		t.Fatalf("body = %+v, want code=NOT_FOUND message=todo not found", body.Error)
	}
	if !c.IsAborted() {
		t.Fatalf("expected context to be aborted after WriteError")
	}
}

func TestWriteErrorWrappedAppErrorIsUnwrapped(t *testing.T) {
	c, rec := newTestContext()

	// A wrapped apperror.Error must still be matched by errors.As.
	wrapped := fmt.Errorf("service failed: %w", apperror.Validation("invalid id"))
	WriteError(c, wrapped)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	var body ErrorBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "VALIDATION_ERROR" || body.Error.Message != "invalid id" {
		t.Fatalf("body = %+v, want code=VALIDATION_ERROR message=invalid id", body.Error)
	}
}

func TestWriteErrorNonAppErrorFallsBackTo500(t *testing.T) {
	c, rec := newTestContext()

	// A plain error carries no status/code: the client must get a fixed 500,
	// never the raw error text.
	WriteError(c, errors.New("connection refused: 10.0.0.5:3306"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var body ErrorBody
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != "INTERNAL_ERROR" || body.Error.Message != "internal server error" {
		t.Fatalf("body = %+v, want generic 500 envelope", body.Error)
	}
	if got := rec.Body.String(); strings.Contains(got, "10.0.0.5") {
		t.Fatalf("raw error leaked into response body: %q", got)
	}
}

func TestWriteSuccessWrapsDataInEnvelope(t *testing.T) {
	c, rec := newTestContext()

	WriteSuccess(c, http.StatusCreated, map[string]any{"id": 7})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	var body Body
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Message != "ok" {
		t.Fatalf("message = %q, want ok", body.Message)
	}
	if body.Data == nil {
		t.Fatalf("data should be present")
	}
}

func TestWriteSuccessNilDataIsOmitted(t *testing.T) {
	c, rec := newTestContext()

	WriteSuccess(c, http.StatusOK, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	// omitempty means a nil payload disappears entirely: {"message":"ok"}.
	got := rec.Body.String()
	if strings.Contains(got, `"data"`) {
		t.Fatalf("nil data should be omitted, got %q", got)
	}
	if !strings.Contains(got, `"message":"ok"`) {
		t.Fatalf("body = %q, want message ok", got)
	}
}
