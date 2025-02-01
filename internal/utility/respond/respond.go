package respond

import (
	"encoding/json"
	"net/http"
)

// WriteError writes an error response
func WriteError(w http.ResponseWriter, code int, err error) {
	JSON(w, code, map[string]interface{}{
		"error": err.Error(),
	})
}

// JSON writes a JSON response
func JSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error since we can't return it after headers are sent
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
