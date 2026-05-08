package handlers

import (
	"fmt"
	"net/http"
)

func writeDownloadHeaders(w http.ResponseWriter, downloadObject *DownloadObject) {
	w.Header().Set("Content-Type", downloadObject.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", downloadObject.Filename))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("ETag", fmt.Sprintf("%q", downloadObject.ETag))
	w.Header().Set("Cache-Control", "public, max-age=3600")
}
