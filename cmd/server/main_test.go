package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	healthHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", contentType, "application/json")
	}

	var body struct {
		Message string `json:"message"`
		Data    struct {
			Status  string `json:"status"`
			Service string `json:"service"`
			Version string `json:"version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("Decode response body error = %v", err)
	}

	if body.Message != "ok" {
		t.Fatalf("message = %q, want %q", body.Message, "ok")
	}

	if body.Data.Status != "ok" {
		t.Fatalf("status = %q, want %q", body.Data.Status, "ok")
	}

	if body.Data.Service != "go-learning" {
		t.Fatalf("service = %q, want %q", body.Data.Service, "go-learning")
	}

	if body.Data.Version != "0.1.0" {
		t.Fatalf("version = %q, want %q", body.Data.Version, "0.1.0")
	}
}

func TestHealthRouteRejectsPost(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	recorder := httptest.NewRecorder()

	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}
