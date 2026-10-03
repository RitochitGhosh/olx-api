package httpx

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Code string

const (
	CodeInvalidID        Code = "invalid_id"        // 400
	CodeInvalidJSON      Code = "invalid_json"      // 400
	CodeUnauthenticated  Code = "unauthenticated"   // 401
	CodeInvalidToken     Code = "invalid_token"     // 401
	CodeExpiredToken     Code = "expired_token"     // 401
	CodeForbidden        Code = "forbidden"         // 403
	CodeNotFound         Code = "not_found"         // 404
	CodeConflict         Code = "conflict"          // 409
	CodeValidationFailed Code = "validation_failed" // 422
	CodeRateLimited      Code = "rate_limited"      // 429
	CodeInternalError    Code = "internal_error"    // 500
)

type errorDefinition struct {
	Status  int
	Message string
}

var errorDefinitions = map[Code]errorDefinition{
	CodeInvalidID: {
		Status:  http.StatusBadRequest,
		Message: "Invalid ID",
	},
	CodeInvalidJSON: {
		Status:  http.StatusBadRequest,
		Message: "Invalid JSON",
	},
	CodeUnauthenticated: {
		Status:  http.StatusUnauthorized,
		Message: "Authentication Required",
	},
	CodeInvalidToken: {
		Status: http.StatusUnauthorized,
		Message: "Invalid Token",
	},
	CodeExpiredToken: {
		Status: http.StatusUnauthorized,
		Message: "Expired Token",
	},
	CodeForbidden: {
		Status:  http.StatusForbidden,
		Message: "You do not have permission to perform this action",
	},
	CodeNotFound: {
		Status:  http.StatusNotFound,
		Message: "Resource not found",
	},
	CodeConflict: {
		Status:  http.StatusConflict,
		Message: "Resource conflict",
	},
	CodeValidationFailed: {
		Status:  http.StatusUnprocessableEntity,
		Message: "Validation failed",
	},
	CodeRateLimited: {
		Status:  http.StatusTooManyRequests,
		Message: "Too many requests",
	},
	CodeInternalError: {
		Status:  http.StatusInternalServerError,
		Message: "Something went wrong",
	},
}

type errorEnvelope struct {
	Success   bool         `json:"success"`
	Error     errorPayload `json:"error"`
	RequestID string       `json:"request_id,omitempty"`
}

type errorPayload struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

type successEnvelope struct {
	Success   bool   `json:"success"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}

func Success(
	w http.ResponseWriter,
	status int,
	data any,
	requestID string,
) {
	writeJSON(w, status, successEnvelope{
		Success:   true,
		Data:      data,
		RequestID: requestID,
	})
}

func Error(w http.ResponseWriter, code Code, requestID string, customMessage ...string) {
	def, ok := errorDefinitions[code]

	// Defensive fallback in case unknown code is passed
	if !ok {
		code = CodeInternalError
		def = errorDefinitions[CodeInternalError]
	}

	message := def.Message

	// Override default only if a non-empty custom message is provided
	if len(customMessage) > 0 {
		if msg := strings.TrimSpace(customMessage[0]); msg != "" {
			message = msg
		}
	}

	writeJSON(w, def.Status, errorEnvelope{
		Success: false,
		Error: errorPayload{
			Code:    code,
			Message: message,
		},
		RequestID: requestID,
	})
}
