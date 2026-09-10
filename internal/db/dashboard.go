package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DashboardStats — агрегати для головної сторінки.
type DashboardStats struct {
	ActiveIncidents   int64          `json:"activeIncidents"`
	IncidentsToday     int64          `json:"incidentsToday"`
	VehiclesAvailable  int64          `json:"vehiclesAvailable"`
	VehiclesTotal      int64          `json:"vehiclesTotal"`
	ActiveEmployees    int64          `json:"activeEmployees"`
	RecentIncidents    []IncidentCard `json:"recentIncidents"`
	VehiclesByStatus   []StatusCount  `json:"vehiclesByStatus"`
}

// IncidentCard — останній виклик для списку на дашборді.
type IncidentCard struct {
	ID         int64     `json:"id"`
	Type       string    `json:"type"`
	Address    string    `json:"address"`
	Station    string    `json:"station"`
	Status     string    `json:"status"`
	ReceivedAt time.Time `json:"receivedAt"`
}

// StatusCount — кількість одиниць техніки у статусі.
type StatusCount struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// QueryDashboard збирає агреговані показники дашборда.
func QueryDashboard(ctx context.Context, pool *pgxpool.Pool) (*DashboardStats, error) {
	stats := &DashboardStats{
		RecentIncidents:  make([]IncidentCard, 0),
		VehiclesByStatus: make([]StatusCount, 0),
	}

	counts := []struct {
		sql  string
		dest *int64
	}{
		{`SELECT COUNT(*) FROM incidents WHERE status = 'active'`, &stats.ActiveIncidents},
		{`SELECT COUNT(*) FROM incidents WHERE received_at >= CURRENT_DATE`, &stats.IncidentsToday},
		{`SELECT COUNT(*) FROM vehicles WHERE status = 'available'`, &stats.VehiclesAvailable},
		{`SELECT COUNT(*) FROM vehicles`, &stats.VehiclesTotal},
		{`SELECT COUNT(*) FROM employees WHERE is_active`, &stats.ActiveEmployees},
	}
	for _, q := range counts {
		if err := pool.QueryRow(ctx, q.sql).Scan(q.dest); err != nil {
			return nil, err
		}
	}

	rows, err := pool.Query(ctx, `
		SELECT i.id, t.title, i.address, s.name, i.status, i.received_at
		FROM incidents i
		JOIN incident_types t ON t.id = i.type_id
		JOIN fire_stations s ON s.id = i.station_id
		ORDER BY i.received_at DESC
		LIMIT 8`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var c IncidentCard
		if err := rows.Scan(&c.ID, &c.Type, &c.Address, &c.Station, &c.Status, &c.ReceivedAt); err != nil {
			return nil, err
		}
		stats.RecentIncidents = append(stats.RecentIncidents, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	vrows, err := pool.Query(ctx, `
		SELECT status, COUNT(*) FROM vehicles GROUP BY status ORDER BY status`)
	if err != nil {
		return nil, err
	}
	defer vrows.Close()
	for vrows.Next() {
		var sc StatusCount
		if err := vrows.Scan(&sc.Status, &sc.Count); err != nil {
			return nil, err
		}
		stats.VehiclesByStatus = append(stats.VehiclesByStatus, sc)
	}
	if err := vrows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}
