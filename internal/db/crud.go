package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrUnknownTable — таблиці немає в реєстрі схеми.
var ErrUnknownTable = errors.New("невідома таблиця")

// tableDef дивиться спершу в динамічний реєстр (жива схема), потім у статичний.
func tableDef(name string) (Table, bool) {
	metaMu.RLock()
	for _, t := range metaDyn {
		if t.Name == name {
			metaMu.RUnlock()
			return t, true
		}
	}
	metaMu.RUnlock()
	for _, t := range schemaMeta {
		if t.Name == name {
			return t, true
		}
	}
	return Table{}, false
}

// coerce перетворює JSON-значення у Go-значення, придатне для pgx.
func coerce(typ string, v any) (any, error) {
	if v == nil {
		return nil, nil
	}
	switch {
	case strings.HasPrefix(typ, "integer"):
		if f, ok := v.(float64); ok {
			return int64(f), nil
		}
		if i, ok := v.(int64); ok {
			return i, nil
		}
		return nil, fmt.Errorf("очікувалося число (тип %s)", typ)
	case typ == "bigint", typ == "smallint":
		if f, ok := v.(float64); ok {
			return int64(f), nil
		}
		return nil, fmt.Errorf("очікувалося число (тип %s)", typ)
	case strings.HasPrefix(typ, "numeric"):
		if f, ok := v.(float64); ok {
			return f, nil
		}
		return nil, fmt.Errorf("очікувалося число (тип %s)", typ)
	case typ == "boolean":
		if b, ok := v.(bool); ok {
			return b, nil
		}
		return nil, errors.New("очікувався boolean")
	case strings.Contains(typ, "timestamp"), typ == "date":
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("очікувався рядок дати (тип %s)", typ)
		}
		if s == "" {
			return nil, nil
		}
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return nil, fmt.Errorf("некоректна дата %q: %w", s, err)
		}
		return t, nil
	case typ == "jsonb":
		switch v := v.(type) {
		case map[string]any:
			b, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			return json.RawMessage(b), nil
		case string:
			return v, nil
		default:
			return nil, errors.New("очікувався об'єкт або рядок для jsonb")
		}
	default: // varchar, text
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("очікувався рядок (тип %s)", typ)
		}
		return s, nil
	}
}

// ListRows повертає рядки таблиці (до limit записів).
func ListRows(ctx context.Context, pool *pgxpool.Pool, table string, limit int) ([]map[string]any, error) {
	if _, ok := tableDef(table); !ok {
		return nil, ErrUnknownTable
	}
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	rows, err := pool.Query(ctx, fmt.Sprintf("SELECT * FROM %s LIMIT %d", table, limit))
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToMap)
}

// InsertRow вставляє рядок у таблицю і повертає його (RETURNING *).
// Колонки валідуються за реєстром схеми, значення — параметризовані ($n).
func InsertRow(ctx context.Context, pool *pgxpool.Pool, table string, values map[string]any) (map[string]any, error) {
	def, ok := tableDef(table)
	if !ok {
		return nil, ErrUnknownTable
	}

	cols := make([]string, 0, len(values))
	args := make([]any, 0, len(values))
	for _, c := range def.Columns {
		if c.Primary {
			continue // id генерується автоматично
		}
		v, exists := values[c.Name]
		if !exists {
			continue
		}
		cv, err := coerce(c.Type, v)
		if err != nil {
			return nil, fmt.Errorf("%s.%s: %w", table, c.Name, err)
		}
		cols = append(cols, c.Name)
		args = append(args, cv)
	}
	if len(cols) == 0 {
		return nil, errors.New("немає допустимих колонок для вставки")
	}

	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	sql := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		table, strings.Join(cols, ", "), strings.Join(placeholders, ", "),
	)
	return queryRowMap(ctx, pool, sql, args...)
}

// UpdateRow оновлює рядок за id і повертає оновлений рядок.
func UpdateRow(ctx context.Context, pool *pgxpool.Pool, table string, id int64, values map[string]any) (map[string]any, error) {
	def, ok := tableDef(table)
	if !ok {
		return nil, ErrUnknownTable
	}

	sets := make([]string, 0, len(values))
	args := []any{id} // $1 — id
	for _, c := range def.Columns {
		if c.Primary {
			continue
		}
		v, exists := values[c.Name]
		if !exists {
			continue
		}
		cv, err := coerce(c.Type, v)
		if err != nil {
			return nil, fmt.Errorf("%s.%s: %w", table, c.Name, err)
		}
		sets = append(sets, fmt.Sprintf("%s = $%d", c.Name, len(args)+1))
		args = append(args, cv)
	}
	if len(sets) == 0 {
		return nil, errors.New("немає допустимих колонок для оновлення")
	}

	sql := fmt.Sprintf("UPDATE %s SET %s WHERE id = $1 RETURNING *", table, strings.Join(sets, ", "))
	return queryRowMap(ctx, pool, sql, args...)
}

// DeleteRow видаляє рядок за id.
func DeleteRow(ctx context.Context, pool *pgxpool.Pool, table string, id int64) error {
	if _, ok := tableDef(table); !ok {
		return ErrUnknownTable
	}
	ct, err := pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = $1", table), id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func queryRowMap(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (map[string]any, error) {
	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectOneRow(rows, pgx.RowToMap)
}
