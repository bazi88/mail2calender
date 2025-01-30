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
	json.NewEncoder(w).Encode(data)
}
