package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestNotFoundCarriesCodeAndStatus(t *testing.T) {
	err := NotFound("route not found")

	if err.Code != "NOT_FOUND" {
		t.Fatalf("code = %q, want %q", err.Code, "NOT_FOUND")
	}
	if err.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", err.StatusCode, http.StatusNotFound)
	}
	if err.Message != "route not found" {
		t.Fatalf("message = %q, want %q", err.Message, "route not found")
	}
	if err.Error() != "route not found" {
		t.Fatalf("Error() = %q, want %q", err.Error(), "route not found")
	}
}

func TestConstructorsUseExpectedCodeAndStatus(t *testing.T) {
	cases := []struct {
		name       string
		err        Error
		wantCode   string
		wantStatus int
	}{
		{"validation", Validation("bad input"), "VALIDATION_ERROR", http.StatusBadRequest},
		{"internal", Internal("boom"), "INTERNAL_ERROR", http.StatusInternalServerError},
		{"notfound", NotFound("missing"), "NOT_FOUND", http.StatusNotFound},
		{"unauthorized", Unauthorized("nope"), "UNAUTHORIZED", http.StatusUnauthorized},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", tc.err.Code, tc.wantCode)
			}
			if tc.err.StatusCode != tc.wantStatus {
				t.Fatalf("status = %d, want %d", tc.err.StatusCode, tc.wantStatus)
			}
		})
	}
}

// errors.As is how handler.writeError picks the status code, so it has to work
// on the value type these constructors return.
func TestErrorIsMatchableWithErrorsAs(t *testing.T) {
	var err error = NotFound("missing")

	var appErr Error
	if !errors.As(err, &appErr) {
		t.Fatalf("errors.As failed to match %v as an apperror.Error", err)
	}
	if appErr.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", appErr.StatusCode, http.StatusNotFound)
	}
}
