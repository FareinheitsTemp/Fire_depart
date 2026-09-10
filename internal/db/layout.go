package db

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetLayout повертає збережені позиції вузлів для подання view (nil, якщо ще не зберігали).
func GetLayout(ctx context.Context, pool *pgxpool.Pool, view string) (json.RawMessage, error) {
	var positions json.RawMessage
	err := pool.QueryRow(ctx,
		`SELECT positions FROM schema_layouts WHERE view_name = $1`, view,
	).Scan(&positions)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return positions, err
}

// SaveLayout зберігає (UPSERT) позиції вузлів подання view.
func SaveLayout(ctx context.Context, pool *pgxpool.Pool, view string, positions json.RawMessage) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO schema_layouts (view_name, positions, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (view_name)
		DO UPDATE SET positions = EXCLUDED.positions, updated_at = NOW()`,
		view, positions)
	return err
}
