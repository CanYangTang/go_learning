package jsonhandler

import (
	"encoding/json"
	"net/http"
)

type EchoRequest struct {
	Message string `json:"message"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement the handler
	// 1. Only accept POST method, return 405 for other methods
	// 2. Decode JSON request body into EchoRequest
	// 3. If JSON is invalid, return 400 with error message
	// 4. Return the same message in EchoResponse with 200
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req EchoRequest
	eres := ErrorResponse{
		Error: "invalid json body",
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(eres)
		return
	}
	res := EchoResponse{Message: req.Message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
