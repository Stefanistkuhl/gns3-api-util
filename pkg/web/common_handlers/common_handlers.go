package commonhandlers

import (
	"encoding/json"
	"net/http"

	"github.com/0xveya/gns3util/pkg/web/helpers"
)

func HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{
		"status": "OK",
	}
	b, _ := json.Marshal(resp)
	_, _ = w.Write(b)
}

func Handle404(w http.ResponseWriter, r *http.Request) {
	helpers.WriteAPIError(w, "no found", helpers.ErrCodeNotFound, "this route is not real twin", http.StatusNotFound)
}

func Handle405(w http.ResponseWriter, r *http.Request) {
	helpers.WriteAPIError(w, "method not allowed", helpers.ErrCodeNotFound, "method not allowed on this route", http.StatusMethodNotAllowed)
}
