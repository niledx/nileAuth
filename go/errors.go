package main

import (
	"encoding/json"
	"net/http"
)

// APIError represents a structured API error response
type APIError struct {
	Code    string `json:"error_code"`
	Message string `json:"error_message"`
	Details string `json:"details,omitempty"`
}

// writeError writes a structured error response
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIError{
		Code:    code,
		Message: message,
	})
}

// writeSuccess writes a success response
func writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	writeJSON(w, status, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

