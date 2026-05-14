package server

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/squall-chua/go-grpc-auth/web"
)

func ServeUI(mux *http.ServeMux) error {
	distFS, err := fs.Sub(web.Assets, ".output/public")
	if err != nil {
		return nil
	}

	fileServer := http.FileServer(http.FS(distFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Skip API, OAuth2, Well-known, and Swagger
		if strings.HasPrefix(path, "/v1") ||
			strings.HasPrefix(path, "/oauth2") ||
			strings.HasPrefix(path, "/.well-known") ||
			strings.HasPrefix(path, "/swagger") {
			return
		}

		f, err := distFS.Open(strings.TrimPrefix(path, "/"))
		if err != nil {
			// File not found, serve index.html for SPA routing
			index, err := distFS.Open("index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer index.Close()

			content, err := io.ReadAll(index)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			stat, _ := index.Stat()
			http.ServeContent(w, r, "index.html", stat.ModTime(), bytes.NewReader(content))
			return
		}
		defer f.Close()

		fileServer.ServeHTTP(w, r)
	})

	return nil
}
