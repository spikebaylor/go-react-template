// Package dbtest gives each test its own freshly migrated Postgres database.
//
// Call Main from a package's TestMain. It uses the server in TEST_DATABASE_URL
// when that's set (CI systems that provide Postgres as a service or sidecar),
// and otherwise starts a postgres:18 container with Testcontainers, which needs
// a Docker daemon. Either way it migrates a template database once, and New
// hands each test a copy of it, dropped when the test ends.
//
// With neither a TEST_DATABASE_URL nor Docker, the package fails, saying so. It
// never skips: a green run that tested nothing is worse than a red one.
package dbtest

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" driver
	"github.com/testcontainers/testcontainers-go"
	tclog "github.com/testcontainers/testcontainers-go/log"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/spikebaylor/go-react-template/internal/store"
)

var (
	// Per process: go test runs packages in parallel, and on a shared server
	// (TEST_DATABASE_URL) they would otherwise fight over one template.
	template = fmt.Sprintf("app_test_template_%d", os.Getpid())

	serverURL string  // the server's maintenance database
	admin     *sql.DB // for CREATE/DROP DATABASE
	next      atomic.Int64
)

// Main runs the package's tests and exits.
func Main(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	serverURL = os.Getenv("TEST_DATABASE_URL")
	if serverURL == "" {
		// Testcontainers logs a banner and every step; a failure is reported below.
		tclog.SetDefault(quiet{})
		ctr, err := postgres.Run(ctx, "postgres:18",
			postgres.WithDatabase("postgres"),
			postgres.WithUsername("app"),
			postgres.WithPassword("app"),
			postgres.BasicWaitStrategies(),
		)
		// Registered before the error check: Run can fail after the container exists.
		defer func() { _ = testcontainers.TerminateContainer(ctr) }()
		if err != nil {
			fmt.Fprintf(os.Stderr, "dbtest: no TEST_DATABASE_URL, and starting postgres:18 with Docker failed: %v\n"+
				"dbtest: start the Docker daemon, or set TEST_DATABASE_URL to a Postgres you can create databases on.\n", err)
			return 1
		}
		if serverURL, err = ctr.ConnectionString(ctx, "sslmode=disable"); err != nil {
			fmt.Fprintf(os.Stderr, "dbtest: container connection string: %v\n", err)
			return 1
		}
	}

	var err error
	if admin, err = sql.Open("pgx", serverURL); err != nil {
		fmt.Fprintf(os.Stderr, "dbtest: %v\n", err)
		return 1
	}
	defer admin.Close()

	if _, err := admin.ExecContext(ctx, `CREATE DATABASE `+template); err != nil {
		fmt.Fprintf(os.Stderr, "dbtest: create template: %v\n", err)
		return 1
	}
	defer func() {
		_, _ = admin.ExecContext(context.Background(), `DROP DATABASE IF EXISTS `+template+` WITH (FORCE)`)
	}()

	// Migrate through store.Open, the path the binary takes, then close it:
	// Postgres won't copy a template anything is still connected to.
	s, err := store.Open(ctx, withDatabase(serverURL, template))
	if err != nil {
		fmt.Fprintf(os.Stderr, "dbtest: migrate template: %v\n", err)
		return 1
	}
	s.Close()

	return m.Run()
}

// New returns the URL of a migrated database of the test's own.
func New(t *testing.T) string {
	t.Helper()

	name := fmt.Sprintf("app_test_%d_%d", os.Getpid(), next.Add(1))
	if _, err := admin.ExecContext(t.Context(), `CREATE DATABASE `+name+` TEMPLATE `+template); err != nil {
		t.Fatalf("create test database: %v", err)
	}
	t.Cleanup(func() {
		if _, err := admin.ExecContext(context.Background(), `DROP DATABASE `+name+` WITH (FORCE)`); err != nil {
			t.Errorf("drop test database %s: %v", name, err)
		}
	})
	return withDatabase(serverURL, name)
}

func withDatabase(raw, name string) string {
	u, err := url.Parse(raw)
	if err != nil {
		panic(fmt.Sprintf("dbtest: bad database URL: %v", err))
	}
	u.Path = "/" + name
	return u.String()
}

type quiet struct{}

func (quiet) Printf(string, ...any) {}
