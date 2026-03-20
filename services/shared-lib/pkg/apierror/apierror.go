package apierror

import (
	"encoding/json"
	"net/http"
)

// ErrorCode is a machine-readable error identifier.
// Use SCREAMING_SNAKE_CASE to match the pattern used by Stripe, Google APIs, etc.
type ErrorCode string

const (
	ErrBadRequest          ErrorCode = "BAD_REQUEST"
	ErrUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrForbidden           ErrorCode = "FORBIDDEN"
	ErrNotFound            ErrorCode = "NOT_FOUND"
	ErrConflict            ErrorCode = "CONFLICT"
	ErrTooManyRequests     ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
	ErrValidation          ErrorCode = "VALIDATION_ERROR"
)

// APIError is the canonical error envelope used by every microservice.
// All error responses MUST use this structure to ensure a consistent client experience.
//
// Wire format:
//
//	{
//	    "error": {
//	        "code":       "ORDER_NOT_FOUND",
//	        "message":    "The order with id 'abc-123' was not found.",
//	        "request_id": "req-9f8e7d6c..."
//	    }
//	}
type APIError struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	RequestID string    `json:"request_id,omitempty"`
}

type envelope struct {
	Error APIError `json:"error"`
}

// Write sends a structured JSON error response to the client.
// Always use this instead of http.Error to ensure consistent error format.
func Write(w http.ResponseWriter, status int, code ErrorCode, message string, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	resp := envelope{
		Error: APIError{
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// BadRequest writes a 400 response with the BAD_REQUEST error code.
func BadRequest(w http.ResponseWriter, message string, requestID string) {
	Write(w, http.StatusBadRequest, ErrBadRequest, message, requestID)
}

// NotFound writes a 404 response.
func NotFound(w http.ResponseWriter, resource string, id string, requestID string) {
	Write(w, http.StatusNotFound, ErrNotFound,
		"The "+resource+" '"+id+"' was not found.", requestID)
}

// InternalServerError writes a 500 response without leaking internal details.
func InternalServerError(w http.ResponseWriter, requestID string) {
	Write(w, http.StatusInternalServerError, ErrInternalServerError,
		"An unexpected error occurred. Please try again later.", requestID)
}

// TooManyRequests writes a 429 response.
func TooManyRequests(w http.ResponseWriter, requestID string) {
	Write(w, http.StatusTooManyRequests, ErrTooManyRequests,
		"Rate limit exceeded. Please retry after a short delay.", requestID)
}

// Conflict writes a 409 response (e.g. idempotency key already used).
func Conflict(w http.ResponseWriter, message string, requestID string) {
	Write(w, http.StatusConflict, ErrConflict, message, requestID)
}

// ValidationError writes a 422 Unprocessable Content response.
func ValidationError(w http.ResponseWriter, message string, requestID string) {
	Write(w, http.StatusUnprocessableEntity, ErrValidation, message, requestID)
}
