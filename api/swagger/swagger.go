package swagger

import (
	_ "embed"
	"net/http"
)

//go:embed index.html
var uiHTML []byte

//go:embed auth.swagger.json
var specJSON []byte

// RegisterRoutes registers the swagger UI and spec handlers on the provided mux.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/swagger/", uiHandler)
	mux.HandleFunc("/swagger/spec", specHandler)
}

func uiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(uiHTML)
}

func specHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(specJSON)
}
