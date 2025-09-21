package server

import (
	"encoding/json"
	"fmt"
	"github.com/stefanistkuhl/gns3util/pkg/api/schemas"
	"github.com/stefanistkuhl/gns3util/pkg/utils"
	"github.com/stefanistkuhl/gns3util/pkg/utils/messageUtils"
	"net/http"
)

func StartInteractiveServer(host string, port int) (schemas.Class, error) {
	var classData schemas.Class
	classDataChan := make(chan schemas.Class, 1)

	mux := http.NewServeMux()

	mux.HandleFunc("/", handleIndex())
	mux.HandleFunc("/style.css", handleCSS)
	mux.HandleFunc("/script.js", handleJS(port))
	mux.HandleFunc("/favicon.ico", handleFavicon)
	mux.HandleFunc("/data", handleDataSubmission(classDataChan))

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Error starting server: %v\n", err)
		}
	}()

	url := fmt.Sprintf("http://%s:%d", host, port)
	fmt.Printf("%v at %v\n",
		messageUtils.InfoMsg("Opening interactive class creation interface"),
		messageUtils.Highlight(url))

	classData = <-classDataChan

	return classData, nil
}

func handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(utils.GetEmbeddedHTML())
	}
}

func handleCSS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	_, _ = w.Write(utils.GetEmbeddedCSS())
}

func handleJS(port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")

		jsContent := string(utils.GetEmbeddedJS())

		jsContent = fmt.Sprintf(`
// Injected port configuration
const SERVER_PORT = %d;

%s
`, port, jsContent)

		_, _ = w.Write([]byte(jsContent))
	}
}

func handleFavicon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	_, _ = w.Write(utils.GetEmbeddedFavicon())
}

func handleDataSubmission(classDataChan chan<- schemas.Class) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var classData schemas.Class

		if err := json.NewDecoder(r.Body).Decode(&classData); err != nil {
			http.Error(w, "Invalid JSON data", http.StatusBadRequest)
			return
		}

		classDataChan <- classData

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}
