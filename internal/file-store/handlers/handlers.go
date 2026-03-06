package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	writeErr := writeJSON(w, map[string]any{
		"status": "ok",
	})
	if writeErr != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", writeErr), http.StatusInternalServerError)
		return
	}
}

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	writeErr := writeJSON(w, map[string]any{
		"status": "ok",
	})
	if writeErr != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", writeErr), http.StatusInternalServerError)
		return
	}
}

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	writeErr := writeJSON(w, map[string]any{
		"status": "ok",
	})
	if writeErr != nil {
		http.Error(w, fmt.Sprintf("Failed to write response: %v", writeErr), http.StatusInternalServerError)
		return
	}
}

func writeJSON(w http.ResponseWriter, data any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}
