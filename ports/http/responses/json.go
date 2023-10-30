package responses

import (
	"encoding/json"
	"net/http"
)

func SendJSON(w http.ResponseWriter, statusCode int, v any) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
