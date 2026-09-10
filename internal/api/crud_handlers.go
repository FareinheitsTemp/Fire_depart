package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/FareinheitsTemp/fire_depart/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store — доступ до БД для хендлерів.
type Store struct {
	Pool *pgxpool.Pool
}

// GET /api/tables/{table} — список рядків таблиці (до 500).
func (s *Store) handleListRows(w http.ResponseWriter, r *http.Request) {
	rows, err := db.ListRows(r.Context(), s.Pool, r.PathValue("table"), 500)
	if errors.Is(err, db.ErrUnknownTable) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// POST /api/tables/{table} — створити запис.
func (s *Store) handleInsertRow(w http.ResponseWriter, r *http.Request) {
	values, err := decodeJSONBody(w, r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	row, err := db.InsertRow(r.Context(), s.Pool, r.PathValue("table"), values)
	respondRow(w, err, row, http.StatusCreated)
}

// PATCH/PUT /api/tables/{table}/{id} — оновити запис.
func (s *Store) handleUpdateRow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некоректний id"})
		return
	}
	values, err := decodeJSONBody(w, r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	row, err := db.UpdateRow(r.Context(), s.Pool, r.PathValue("table"), id, values)
	respondRow(w, err, row, http.StatusOK)
}

// DELETE /api/tables/{table}/{id} — видалити запис.
func (s *Store) handleDeleteRow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "некоректний id"})
		return
	}
	if err := db.DeleteRow(r.Context(), s.Pool, r.PathValue("table"), id); err != nil {
		respondRow(w, err, nil, 0)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request) (map[string]any, error) {
	var values map[string]any
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&values); err != nil {
		return nil, err
	}
	return values, nil
}

func respondRow(w http.ResponseWriter, err error, row map[string]any, okStatus int) {
	switch {
	case err == nil:
		writeJSON(w, okStatus, row)
	case errors.Is(err, db.ErrUnknownTable):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, pgx.ErrNoRows):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "запис не знайдено"})
	default:
		// помилки валідації або порушення обмежень БД (CHECK/FK/UNIQUE)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
}
