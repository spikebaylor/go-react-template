// Package store owns the database: opening it, migrating it, and every query.
package store

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

//go:embed migrations/*.sql
var migrations embed.FS

// ErrNotFound is returned when a row the caller named doesn't exist.
var ErrNotFound = errors.New("not found")

type Store struct {
	db *sql.DB
}

type Note struct {
	ID        int64
	Text      string
	CreatedAt time.Time
}

// Open connects to Postgres and migrates it to the latest schema. Migrations
// run under goose's Postgres advisory lock, so replicas starting together
// apply each one once.
func Open(ctx context.Context, url string) (*Store, error) {
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	if err := migrate(ctx, db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}
	return &Store{db: db}, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return err
	}
	p, err := goose.NewProvider(goose.DialectPostgres, db, mustSub(migrations, "migrations"),
		goose.WithSessionLocker(locker))
	if err != nil {
		return err
	}
	_, err = p.Up(ctx)
	return err
}

func (s *Store) Close() error { return s.db.Close() }

// Ping reports whether the database is reachable.
func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

func (s *Store) ListNotes(ctx context.Context) ([]Note, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, text, created_at FROM notes ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []Note{}
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Text, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (s *Store) CreateNote(ctx context.Context, text string) (Note, error) {
	n := Note{Text: text}
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO notes (text) VALUES ($1) RETURNING id, created_at`, text,
	).Scan(&n.ID, &n.CreatedAt)
	return n, err
}

// DeleteNote deletes the note, or returns ErrNotFound if there isn't one.
func (s *Store) DeleteNote(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func mustSub(f fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
