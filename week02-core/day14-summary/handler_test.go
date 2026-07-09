package summary

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBatchHandlerSuccess(t *testing.T) {
	body := `{"tasks":[{"id":1,"data":"x"},{"id":2,"data":"y"},{"id":3,"data":"z"}]}`
	req := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	BatchHandler(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", contentType, "application/json")
	}

	var resp BatchResponse
	if err := json.NewDecoder(recorder.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error = %v", err)
	}

	if len(resp.Results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(resp.Results))
	}

	// Results should be sorted by ID
	for i, result := range resp.Results {
		if result.ID != i+1 {
			t.Fatalf("result.ID = %d, want %d", result.ID, i+1)
		}
	}
}

func TestBatchHandlerInvalidJSON(t *testing.T) {
	body := `{"tasks":invalid}`
	req := httptest.NewRequest(http.MethodPost, "/batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	BatchHandler(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestBatchHandlerNonPostMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/batch", nil)
	recorder := httptest.NewRecorder()

	BatchHandler(recorder, req)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}
