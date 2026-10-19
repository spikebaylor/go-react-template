package server_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spikebaylor/go-react-template/internal/api"
	"github.com/spikebaylor/go-react-template/internal/dbtest"
	"github.com/spikebaylor/go-react-template/internal/server"
	"github.com/spikebaylor/go-react-template/internal/store"
)

func TestMain(m *testing.M) { dbtest.Main(m) }

func newServer(t *testing.T, ui http.Handler) *httptest.Server {
	t.Helper()
	s, err := store.Open(t.Context(), dbtest.New(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	srv := httptest.NewServer(server.New(s, "1.2.3-test", ui))
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, method, url, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func decode[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestHealthReportsVersion(t *testing.T) {
	srv := newServer(t, nil)

	resp := do(t, http.MethodGet, srv.URL+"/api/health", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if h := decode[api.Health](t, resp); h.Version != "1.2.3-test" || h.Status != "ok" {
		t.Errorf("health = %+v", h)
	}
}

func TestNotes(t *testing.T) {
	srv := newServer(t, nil)

	resp := do(t, http.MethodPost, srv.URL+"/api/notes", `{"text":"  hello  "}`)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: status = %d", resp.StatusCode)
	}
	created := decode[api.Note](t, resp)
	if created.Text != "hello" {
		t.Errorf("created text = %q, want it trimmed to %q", created.Text, "hello")
	}

	notes := decode[[]api.Note](t, do(t, http.MethodGet, srv.URL+"/api/notes", ""))
	if len(notes) != 1 || notes[0].Id != created.Id {
		t.Errorf("list = %+v", notes)
	}

	url := srv.URL + "/api/notes/" + jsonInt(created.Id)
	if resp := do(t, http.MethodDelete, url, ""); resp.StatusCode != http.StatusNoContent {
		t.Errorf("delete: status = %d", resp.StatusCode)
	}
	if resp := do(t, http.MethodDelete, url, ""); resp.StatusCode != http.StatusNotFound {
		t.Errorf("delete again: status = %d, want 404", resp.StatusCode)
	}
}

func TestCreateNoteRejectsBadText(t *testing.T) {
	srv := newServer(t, nil)

	for _, body := range []string{`{"text":"   "}`, `{"text":"` + strings.Repeat("x", 501) + `"}`} {
		if resp := do(t, http.MethodPost, srv.URL+"/api/notes", body); resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, want 400 for a %d-byte body", resp.StatusCode, len(body))
		}
	}
}

func TestUIServedOutsideAPI(t *testing.T) {
	ui := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ui")) })
	srv := newServer(t, ui)

	if resp := do(t, http.MethodGet, srv.URL+"/some/page", ""); resp.StatusCode != http.StatusOK {
		t.Errorf("ui: status = %d", resp.StatusCode)
	}
	if resp := do(t, http.MethodGet, srv.URL+"/api/health", ""); decode[api.Health](t, resp).Status != "ok" {
		t.Error("the UI handler shadowed /api/health")
	}
}

func jsonInt(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func TestDatabaseFailureIsA500(t *testing.T) {
	s, err := store.Open(t.Context(), dbtest.New(t))
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(server.New(s, "1.2.3-test", nil))
	t.Cleanup(srv.Close)
	s.Close()

	for _, r := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/health", ""},
		{http.MethodGet, "/api/notes", ""},
		{http.MethodPost, "/api/notes", `{"text":"hello"}`},
		{http.MethodDelete, "/api/notes/1", ""},
	} {
		if resp := do(t, r.method, srv.URL+r.path, r.body); resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("%s %s: status = %d, want 500", r.method, r.path, resp.StatusCode)
		}
	}
}

func TestLoggedRecordsStatus(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	h := server.Logged(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/teapot" {
			w.WriteHeader(http.StatusTeapot)
		}
	}))
	for _, path := range []string{"/teapot", "/ok"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
	}

	var got []int
	for line := range strings.Lines(buf.String()) {
		var entry struct{ Status int }
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatal(err)
		}
		got = append(got, entry.Status)
	}
	if len(got) != 2 || got[0] != http.StatusTeapot || got[1] != http.StatusOK {
		t.Errorf("logged statuses = %v, want [418 200]", got)
	}
}
