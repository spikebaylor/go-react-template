// Package server serves the JSON API defined in api/openapi.yaml and, in the
// combined build, the web UI.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/spikebaylor/go-react-template/internal/api"
	"github.com/spikebaylor/go-react-template/internal/store"
)

const maxNoteLength = 500

// New returns the app's HTTP handler. ui serves everything outside /api/; pass
// nil to serve the API alone, as the split api image does.
func New(s *store.Store, version string, ui http.Handler) http.Handler {
	mux := http.NewServeMux()
	api.HandlerWithOptions(
		api.NewStrictHandler(&handlers{store: s, version: version}, nil),
		api.StdHTTPServerOptions{BaseRouter: mux},
	)
	if ui != nil {
		mux.Handle("/", ui)
	}
	return mux
}

// handlers implements the generated api.StrictServerInterface, so the compiler
// fails when it and the contract drift apart.
type handlers struct {
	store   *store.Store
	version string
}

var _ api.StrictServerInterface = (*handlers)(nil)

func (h *handlers) GetHealth(ctx context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	if err := h.store.Ping(ctx); err != nil {
		return nil, err
	}
	return api.GetHealth200JSONResponse{Status: "ok", Version: h.version}, nil
}

func (h *handlers) ListNotes(ctx context.Context, _ api.ListNotesRequestObject) (api.ListNotesResponseObject, error) {
	notes, err := h.store.ListNotes(ctx)
	if err != nil {
		return nil, err
	}
	resp := make(api.ListNotes200JSONResponse, len(notes))
	for i, n := range notes {
		resp[i] = toAPI(n)
	}
	return resp, nil
}

func (h *handlers) CreateNote(ctx context.Context, req api.CreateNoteRequestObject) (api.CreateNoteResponseObject, error) {
	text := strings.TrimSpace(req.Body.Text)
	if text == "" || len([]rune(text)) > maxNoteLength {
		return api.CreateNote400JSONResponse{ProblemJSONResponse: api.ProblemJSONResponse{
			Message: "text must be 1 to 500 characters",
		}}, nil
	}
	n, err := h.store.CreateNote(ctx, text)
	if err != nil {
		return nil, err
	}
	return api.CreateNote201JSONResponse(toAPI(n)), nil
}

func (h *handlers) DeleteNote(ctx context.Context, req api.DeleteNoteRequestObject) (api.DeleteNoteResponseObject, error) {
	err := h.store.DeleteNote(ctx, req.Id)
	if errors.Is(err, store.ErrNotFound) {
		return api.DeleteNote404JSONResponse{ProblemJSONResponse: api.ProblemJSONResponse{
			Message: "no such note",
		}}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.DeleteNote204Response{}, nil
}

func toAPI(n store.Note) api.Note {
	return api.Note{Id: n.ID, Text: n.Text, CreatedAt: n.CreatedAt}
}

// Logged wraps h to log each request, never its body.
func Logged(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		h.ServeHTTP(rec, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "status", rec.status)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
