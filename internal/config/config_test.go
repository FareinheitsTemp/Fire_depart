package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.HTTPAddr == "" {
		t.Fatal("HTTPAddr must have a default")
	}
	if cfg.DatabaseDSN == "" {
		t.Fatal("DatabaseDSN must have a default")
	}
}

func TestGetenvFallback(t *testing.T) {
	if got := getenv("FIRE_DEPART_DEFINITELY_MISSING", "x"); got != "x" {
		t.Fatalf("want fallback, got %q", got)
	}
}
