package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// errorBody is the JSON shape returned for error responses.
type errorBody struct {
	Error string `json:"error"`
}

// writeJSON serializes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// The header is already sent; just record the failure.
		slog.Error("encode response", "error", err)
	}
}

// writeError writes a JSON error body with the given status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorBody{Error: msg})
}
