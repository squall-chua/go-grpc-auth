package server

import (
	"bytes"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/squall-chua/go-grpc-auth/web"
	"go.uber.org/zap"
)

type UIConfig struct {
	ApiBase string
	AppName string
}

var placeholders = map[string]func(UIConfig) string{
	"__API_BASE_PLACEHOLDER__":  func(c UIConfig) string { return c.ApiBase },
	"__APP_NAME_PLACEHOLDER__": func(c UIConfig) string { return c.AppName },
}

func patchedHTML(distFS fs.FS, cfg UIConfig) ([]byte, error) {
	f, err := distFS.Open("index.html")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	for placeholder, valueFn := range placeholders {
		content = bytes.ReplaceAll(content, []byte(placeholder), []byte(valueFn(cfg)))
	}
	return content, nil
}

func ServeUI(mux *http.ServeMux, cfg UIConfig) error {
	distFS, err := fs.Sub(web.Assets, ".output/public")
	if err != nil {
		return nil
	}

	indexHTML, err := patchedHTML(distFS, cfg)
	if err != nil {
		zap.L().Warn("No embedded index.html; UI will not be served", zap.Error(err))
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
			// File not found, serve patched index.html for SPA routing
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(indexHTML))
			return
		}
		defer f.Close()

		fileServer.ServeHTTP(w, r)
	})

	return nil
}
