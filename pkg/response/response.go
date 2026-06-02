package response

import (
	"encoding/json"
	"net/http"
)

type Body struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message"`
}

type ErrorBody struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, statusCode int, data any) {
	writeJSON(w, statusCode, Body{Data: data, Message: "ok"})
}

func Error(w http.ResponseWriter, statusCode int, code, message string) {
	writeJSON(w, statusCode, ErrorBody{Error: ErrorPayload{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, statusCode int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}
