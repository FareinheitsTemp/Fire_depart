package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IncidentReportRow — рядок зведеного звіту по викликах.
type IncidentReportRow struct {
	ID         int64
	Type       string
	Address    string
	Station    string
	ReceivedAt time.Time
	ClosedAt   *time.Time
	Status     string
	Vehicles   int64
	Responders int64
	DamageUah  float64
}

// ReportIncidents повертає дані для зведеного звіту по всіх викликах.
func ReportIncidents(ctx context.Context, pool *pgxpool.Pool) ([]IncidentReportRow, error) {
	rows, err := pool.Query(ctx, `
		SELECT i.id, t.title, i.address, s.name, i.received_at, i.closed_at, i.status,
		       (SELECT COUNT(*) FROM incident_vehicles iv WHERE iv.incident_id = i.id),
		       (SELECT COUNT(*) FROM incident_responders ir WHERE ir.incident_id = i.id),
		       (SELECT COALESCE(SUM(rep.damage_uah), 0)::float8 FROM incident_reports rep WHERE rep.incident_id = i.id)
		FROM incidents i
		JOIN incident_types t ON t.id = i.type_id
		JOIN fire_stations s ON s.id = i.station_id
		ORDER BY i.received_at DESC
		LIMIT 200`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]IncidentReportRow, 0)
	for rows.Next() {
		var r IncidentReportRow
		if err := rows.Scan(
			&r.ID, &r.Type, &r.Address, &r.Station, &r.ReceivedAt, &r.ClosedAt, &r.Status,
			&r.Vehicles, &r.Responders, &r.DamageUah,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// IncidentDetail — повні дані одного виклику для детального звіту.
type IncidentDetail struct {
	ID          int64
	Type        string
	Address     string
	Station     string
	Description string
	ReceivedAt  time.Time
	ClosedAt    *time.Time
	Status      string
	Summary     string
	DamageUah   float64
	Casualties  int64
	Vehicles    []string
	Responders  []string
}

// ReportIncidentDetail повертає деталі виклику (техніка, бригада, підсумок).
// Повертає pgx.ErrNoRows, якщо виклику не існує.
func ReportIncidentDetail(ctx context.Context, pool *pgxpool.Pool, id int64) (*IncidentDetail, error) {
	d := &IncidentDetail{
		Vehicles:   make([]string, 0),
		Responders: make([]string, 0),
	}
	err := pool.QueryRow(ctx, `
		SELECT i.id, t.title, i.address, s.name, i.description, i.received_at, i.closed_at, i.status,
		       COALESCE(rep.summary, ''), COALESCE(rep.damage_uah, 0)::float8, COALESCE(rep.casualties, 0)
		FROM incidents i
		JOIN incident_types t ON t.id = i.type_id
		JOIN fire_stations s ON s.id = i.station_id
		LEFT JOIN incident_reports rep ON rep.incident_id = i.id
		WHERE i.id = $1`, id).Scan(
		&d.ID, &d.Type, &d.Address, &d.Station, &d.Description, &d.ReceivedAt, &d.ClosedAt, &d.Status,
		&d.Summary, &d.DamageUah, &d.Casualties,
	)
	if err != nil {
		return nil, err
	}

	vrows, err := pool.Query(ctx, `
		SELECT v.call_sign FROM incident_vehicles iv
		JOIN vehicles v ON v.id = iv.vehicle_id
		WHERE iv.incident_id = $1 ORDER BY v.call_sign`, id)
	if err != nil {
		return nil, err
	}
	defer vrows.Close()
	for vrows.Next() {
		var callSign string
		if err := vrows.Scan(&callSign); err != nil {
			return nil, err
		}
		d.Vehicles = append(d.Vehicles, callSign)
	}
	if err := vrows.Err(); err != nil {
		return nil, err
	}

	rrows, err := pool.Query(ctx, `
		SELECT e.full_name FROM incident_responders ir
		JOIN employees e ON e.id = ir.employee_id
		WHERE ir.incident_id = $1 ORDER BY e.full_name`, id)
	if err != nil {
		return nil, err
	}
	defer rrows.Close()
	for rrows.Next() {
		var name string
		if err := rrows.Scan(&name); err != nil {
			return nil, err
		}
		d.Responders = append(d.Responders, name)
	}
	if err := rrows.Err(); err != nil {
		return nil, err
	}
	return d, nil
}
