package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func parseOptionalInt64(value string, defaultValue int64) (int64, error) {
	if value == "" {
		return defaultValue, nil
	}
	return strconv.ParseInt(value, 10, 64)
}

func parseOptionalInt(value string, defaultValue int) (int, error) {
	if value == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(value)
}
