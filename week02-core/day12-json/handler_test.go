package jsonhandler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEchoHandlerSuccess(t *testing.T) {
	body := `{"message":"hello world"}`
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	EchoHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", contentType, "application/json")
	}

	var resp EchoResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response error = %v", err)
	}

	if resp.Message != "hello world" {
		t.Fatalf("message = %q, want %q", resp.Message, "hello world")
	}
}

func TestEchoHandlerInvalidJSON(t *testing.T) {
	body := `{"message":invalid}`
	req := httptest.NewRequest(http.MethodPost, "/echo", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	EchoHandler(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response error = %v", err)
	}

	if resp.Error == "" {
		t.Fatalf("error message should not be empty")
	}
}

func TestEchoHandlerNonPostMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/echo", nil)
	recorder := httptest.NewRecorder()

	EchoHandler(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}
