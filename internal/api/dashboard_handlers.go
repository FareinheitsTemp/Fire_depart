package api

import (
	"net/http"

	"github.com/FareinheitsTemp/fire_depart/internal/db"
)

// GET /api/dashboard — агреговані показники для головної сторінки.
func (s *Store) handleDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := db.QueryDashboard(r.Context(), s.Pool)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
