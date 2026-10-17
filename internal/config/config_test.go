package config

import "testing"

func TestLoad(t *testing.T) {
	t.Setenv("APP_DATABASE_URL", "postgres://x")
	t.Setenv("APP_ADDR", ":9090")

	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if c.Addr != ":9090" || c.DatabaseURL != "postgres://x" {
		t.Errorf("Load() = %+v", c)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("APP_DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Error("Load() succeeded without APP_DATABASE_URL")
	}
}
