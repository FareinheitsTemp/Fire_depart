package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	metaMu  sync.RWMutex
	metaDyn []Table
)

// SetSchemaMeta встановлює динамічний реєстр схеми (після DDL-операцій).
func SetSchemaMeta(tables []Table) {
	metaMu.Lock()
	defer metaMu.Unlock()
	metaDyn = tables
}

// RefreshSchemaMeta перечитує схему з живої БД і оновлює динамічний реєстр.
func RefreshSchemaMeta(ctx context.Context, pool *pgxpool.Pool) error {
	tables, err := LoadSchemaMeta(ctx, pool)
	if err != nil {
		return err
	}
	SetSchemaMeta(tables)
	return nil
}

// LoadSchemaMeta читає метадані схеми з information_schema (жива БД).
func LoadSchemaMeta(ctx context.Context, pool *pgxpool.Pool) ([]Table, error) {
	var names []string
	trows, err := pool.Query(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
		ORDER BY table_name`)
	if err != nil {
		return nil, err
	}
	defer trows.Close()
	for trows.Next() {
		var name string
		if err := trows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	if err := trows.Err(); err != nil {
		return nil, err
	}

	tables := make([]Table, 0, len(names))
	byName := make(map[string]*Table, len(names))
	for _, n := range names {
		tables = append(tables, Table{Name: n, Columns: make([]Column, 0), ForeignKeys: make([]ForeignKey, 0)})
		byName[n] = &tables[len(tables)-1]
	}

	crows, err := pool.Query(ctx, `
		SELECT table_name, column_name, data_type, is_nullable,
		       COALESCE(character_maximum_length, 0)::int,
		       COALESCE(numeric_precision, 0)::int,
		       COALESCE(numeric_scale, -1)::int
		FROM information_schema.columns
		WHERE table_schema = 'public'
		ORDER BY table_name, ordinal_position`)
	if err != nil {
		return nil, err
	}
	defer crows.Close()
	for crows.Next() {
		var tbl, colName, dataType, isNullable string
		var maxLen, prec, scale int
		if err := crows.Scan(&tbl, &colName, &dataType, &isNullable, &maxLen, &prec, &scale); err != nil {
			return nil, err
		}
		t, ok := byName[tbl]
		if !ok {
			continue
		}
		t.Columns = append(t.Columns, Column{
			Name:     colName,
			Type:     formatType(dataType, maxLen, prec, scale),
			Nullable: isNullable == "YES",
		})
	}
	if err := crows.Err(); err != nil {
		return nil, err
	}

	prows, err := pool.Query(ctx, `
		SELECT kcu.table_name, kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON kcu.constraint_name = tc.constraint_name AND kcu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'PRIMARY KEY' AND tc.table_schema = 'public'`)
	if err != nil {
		return nil, err
	}
	defer prows.Close()
	for prows.Next() {
		var tbl, colName string
		if err := prows.Scan(&tbl, &colName); err != nil {
			return nil, err
		}
		if t, ok := byName[tbl]; ok {
			for i := range t.Columns {
				if t.Columns[i].Name == colName {
					t.Columns[i].Primary = true
				}
			}
		}
	}
	if err := prows.Err(); err != nil {
		return nil, err
	}

	frows, err := pool.Query(ctx, `
		SELECT tc.table_name, kcu.column_name, ccu.table_name, ccu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON kcu.constraint_name = tc.constraint_name AND kcu.table_schema = tc.table_schema
		JOIN information_schema.constraint_column_usage ccu
		  ON ccu.constraint_name = tc.constraint_name AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema = 'public'`)
	if err != nil {
		return nil, err
	}
	defer frows.Close()
	for frows.Next() {
		var tbl, colName, refTable, refColumn string
		if err := frows.Scan(&tbl, &colName, &refTable, &refColumn); err != nil {
			return nil, err
		}
		if t, ok := byName[tbl]; ok {
			t.ForeignKeys = append(t.ForeignKeys, ForeignKey{Column: colName, RefTable: refTable, RefColumn: refColumn})
		}
	}
	if err := frows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

func formatType(dataType string, maxLen, prec, scale int) string {
	switch dataType {
	case "character varying":
		if maxLen > 0 {
			return fmt.Sprintf("varchar(%d)", maxLen)
		}
		return "varchar"
	case "numeric":
		if prec > 0 {
			return fmt.Sprintf("numeric(%d,%d)", prec, scale)
		}
		return "numeric"
	case "timestamp without time zone", "timestamp with time zone":
		return "timestamp"
	default:
		return dataType
	}
}
