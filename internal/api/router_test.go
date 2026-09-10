package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSchemaEndpoint(t *testing.T) {
	srv := httptest.NewServer(NewRouter())
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/schema")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}

	var tables []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&tables); err != nil {
		t.Fatal(err)
	}
	if len(tables) < 10 {
		t.Fatalf("want at least 10 tables in schema, got %d", len(tables))
	}
}

func TestHealthEndpoint(t *testing.T) {
	srv := httptest.NewServer(NewRouter())
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
}
