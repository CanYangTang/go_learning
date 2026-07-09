package summary

import (
	"encoding/json"
	"net/http"
)

// BatchHandler handles POST requests to process a batch of tasks concurrently.
func BatchHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	// 1. Check method is POST, return 405 otherwise
	// 2. Decode JSON request body into BatchRequest
	// 3. If JSON invalid, return 400 with error message
	// 4. defer r.Body.Close()
	// 5. Call ProcessBatch with req.Tasks and workerCount 3
	// 6. Set Content-Type: application/json
	// 7. Return 200 with BatchResponse containing results
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
		return
	}
	results := ProcessBatch(req.Tasks, 3)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(BatchResponse{Results: results})
}
