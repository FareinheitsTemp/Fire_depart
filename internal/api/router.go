// Package api містить REST-хендлери бекенду.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/FareinheitsTemp/fire_depart/internal/db"
)

// NewRouter будує маршрутизатор API (патерни Go 1.22).
func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/schema", handleSchema)
	return mux
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleSchema(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, db.SchemaMeta())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
