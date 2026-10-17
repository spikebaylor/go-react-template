package webui

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestSPAFallsBackToIndex(t *testing.T) {
	h := spa(fstest.MapFS{
		"index.html":    {Data: []byte("index")},
		"assets/app.js": {Data: []byte("js")},
	})

	for path, want := range map[string]string{
		"/":              "index",
		"/assets/app.js": "js",
		"/notes/42":      "index",
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
		if body, _ := io.ReadAll(rec.Body); string(body) != want {
			t.Errorf("GET %s = %q, want %q", path, body, want)
		}
	}
}
