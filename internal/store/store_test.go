package store_test

import (
	"errors"
	"testing"

	"github.com/spikebaylor/go-react-template/internal/dbtest"
	"github.com/spikebaylor/go-react-template/internal/store"
)

func TestMain(m *testing.M) { dbtest.Main(m) }

func open(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.Open(t.Context(), dbtest.New(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestNotesRoundTrip(t *testing.T) {
	s := open(t)
	ctx := t.Context()

	first, err := s.CreateNote(ctx, "first")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateNote(ctx, "second"); err != nil {
		t.Fatal(err)
	}

	notes, err := s.ListNotes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 || notes[0].Text != "second" || notes[1].Text != "first" {
		t.Fatalf("ListNotes() = %+v, want second then first", notes)
	}

	if err := s.DeleteNote(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteNote(ctx, first.ID); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("deleting twice: err = %v, want ErrNotFound", err)
	}
}

func TestOpenIsIdempotent(t *testing.T) {
	url := dbtest.New(t)
	for range 2 {
		s, err := store.Open(t.Context(), url)
		if err != nil {
			t.Fatal(err)
		}
		s.Close()
	}
}

func TestOpenFailsWithoutADatabase(t *testing.T) {
	for name, url := range map[string]string{
		"malformed":   "postgres://%zz",
		"unreachable": "postgres://app@127.0.0.1:1/app?connect_timeout=1",
	} {
		if s, err := store.Open(t.Context(), url); err == nil {
			s.Close()
			t.Errorf("%s: Open(%q) succeeded, want an error", name, url)
		}
	}
}
