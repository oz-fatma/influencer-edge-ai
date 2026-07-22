package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

// JSON writes a JSON response with the given status code and payload.
func JSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

// Error writes a JSON error response, mapping domain errors to HTTP codes.
func Error(w http.ResponseWriter, err error) {
	code := domainErr.HTTPStatusCode(err)
	msg := err.Error()
	if code >= http.StatusInternalServerError {
		slog.Error("request failed", "code", code, "error", err)
		msg = "an internal error occurred"
	}
	JSON(w, code, domainErr.ErrorResponse{
		Error:   http.StatusText(code),
		Message: msg,
		Code:    code,
	})
}

// Created writes a 201 Created JSON response.
func Created(w http.ResponseWriter, payload interface{}) {
	JSON(w, http.StatusCreated, payload)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}
