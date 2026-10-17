// Package config reads every setting from the environment into one struct.
// Each setting is listed in README.md's Configuration table; add a row there
// with each new one.
package config

import (
	"errors"
	"os"
)

type Config struct {
	// Addr is the address the HTTP server listens on.
	Addr string
	// DatabaseURL is a Postgres connection string, e.g.
	// postgres://user:pass@host:5432/db?sslmode=disable.
	DatabaseURL string
}

// Load reads the configuration from the environment.
func Load() (Config, error) {
	c := Config{
		Addr:        getenv("APP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("APP_DATABASE_URL"),
	}
	if c.DatabaseURL == "" {
		return Config{}, errors.New("APP_DATABASE_URL is required")
	}
	return c, nil
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
