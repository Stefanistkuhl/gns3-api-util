package commonhandlers

import (
	"encoding/json"
	"net/http"
)

func HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{
		"status": "OK",
	}
	b, _ := json.Marshal(resp)
	_, _ = w.Write(b)
}
