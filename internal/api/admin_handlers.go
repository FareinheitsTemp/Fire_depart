package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/FareinheitsTemp/fire_depart/internal/db"
	"github.com/jackc/pgx/v5"
)

type columnRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

// POST /api/admin/tables — створити таблицю (DDL з валідацією).
func (s *Store) handleCreateTable(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string          `json:"name"`
		Columns []columnRequest `json:"columns"`
	}
	if !decodeStruct(w, r, &req) {
		return
	}
	cols := make([]db.ColumnInput, 0, len(req.Columns))
	for _, c := range req.Columns {
		cols = append(cols, db.ColumnInput{Name: c.Name, Type: c.Type, Nullable: c.Nullable})
	}
	if err := db.CreateTable(r.Context(), s.Pool, req.Name, cols); err != nil {
		respondDDLError(w, err)
		return
	}
	if err := db.RefreshSchemaMeta(r.Context(), s.Pool); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

// DELETE /api/admin/tables/{table} — видалити таблицю.
func (s *Store) handleDropTable(w http.ResponseWriter, r *http.Request) {
	if err := db.DropTable(r.Context(), s.Pool, r.PathValue("table")); err != nil {
		respondDDLError(w, err)
		return
	}
	_ = db.RefreshSchemaMeta(r.Context(), s.Pool)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// POST /api/admin/tables/{table}/columns — додати колонку.
func (s *Store) handleAddColumn(w http.ResponseWriter, r *http.Request) {
	var req columnRequest
	if !decodeStruct(w, r, &req) {
		return
	}
	if err := db.AddColumn(r.Context(), s.Pool, r.PathValue("table"), req.Name, req.Type, req.Nullable); err != nil {
		respondDDLError(w, err)
		return
	}
	_ = db.RefreshSchemaMeta(r.Context(), s.Pool)
	writeJSON(w, http.StatusCreated, map[string]string{"status": "added"})
}

// DELETE /api/admin/tables/{table}/columns/{column} — видалити колонку.
func (s *Store) handleDropColumn(w http.ResponseWriter, r *http.Request) {
	if err := db.DropColumn(r.Context(), s.Pool, r.PathValue("table"), r.PathValue("column")); err != nil {
		respondDDLError(w, err)
		return
	}
	_ = db.RefreshSchemaMeta(r.Context(), s.Pool)
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func decodeStruct(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return false
	}
	return true
}

func respondDDLError(w http.ResponseWriter, err error) {
	if errors.Is(err, db.ErrUnknownTable) || errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
}
