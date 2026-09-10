package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/FareinheitsTemp/fire_depart/internal/db"
)

const maxLayoutBytes = 1 << 20 // 1 MB

// GET /api/layout/{view} — збережені позиції вузлів ERD-мапи.
func (s *Store) handleGetLayout(w http.ResponseWriter, r *http.Request) {
	positions, err := db.GetLayout(r.Context(), s.Pool, r.PathValue("view"))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"positions": positions})
}

// PUT /api/layout/{view} — зберегти позиції вузлів (автозбереження при перетягуванні).
func (s *Store) handleSaveLayout(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxLayoutBytes))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "не вдалося прочитати тіло запиту"})
		return
	}
	if !json.Valid(body) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "тіло запиту не є валідним JSON"})
		return
	}
	if err := db.SaveLayout(r.Context(), s.Pool, r.PathValue("view"), json.RawMessage(body)); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}
