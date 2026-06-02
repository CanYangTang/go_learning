package apperror

import "net/http"

type Error struct {
	Code       string
	Message    string
	StatusCode int
}

func (e Error) Error() string {
	return e.Message
}

func New(code, message string, statusCode int) Error {
	return Error{Code: code, Message: message, StatusCode: statusCode}
}

func Internal(message string) Error {
	return New("INTERNAL_ERROR", message, http.StatusInternalServerError)
}

func Validation(message string) Error {
	return New("VALIDATION_ERROR", message, http.StatusBadRequest)
}
