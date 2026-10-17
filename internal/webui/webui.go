// Package webui serves the React app when it's built into the binary (the
// embedui build tag) and nothing otherwise.
package webui

import (
	"errors"
	"io/fs"
	"net/http"
	"strings"
)

// Handler serves the built UI, falling back to index.html for client-side
// routes. It returns nil when the binary was built without the UI.
func Handler() http.Handler {
	if f := files(); f != nil {
		return spa(f)
	}
	return nil
}

func spa(f fs.FS) http.Handler {
	fileServer := http.FileServerFS(f)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" {
			if _, err := fs.Stat(f, name); errors.Is(err, fs.ErrNotExist) {
				// A client-side route: let the app's router handle it.
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}
