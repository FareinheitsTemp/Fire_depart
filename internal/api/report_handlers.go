package api

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FareinheitsTemp/fire_depart/internal/db"
	"github.com/jackc/pgx/v5"
)

// GET /api/report/incidents — зведений звіт по викликах (HTML, друк/PDF).
func (s *Store) handleIncidentsReport(w http.ResponseWriter, r *http.Request) {
	rows, err := db.ReportIncidents(r.Context(), s.Pool)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var b strings.Builder
	b.WriteString(reportHeader("Звіт по викликах"))
	b.WriteString(`<p class="meta">Згенеровано: ` +
		html.EscapeString(time.Now().Format("02.01.2006 15:04")) +
		` · записів: ` + strconv.Itoa(len(rows)) + `</p>`)
	b.WriteString(`<table><thead><tr><th>№</th><th>Тип</th><th>Адреса</th><th>Частина</th><th>Отримано</th><th>Закрито</th><th>Статус</th><th>Техніка</th><th>Особи</th><th>Збитки, грн</th></tr></thead><tbody>`)
	for _, row := range rows {
		closed := "—"
		if row.ClosedAt != nil {
			closed = row.ClosedAt.Format("02.01.2006 15:04")
		}
		fmt.Fprintf(&b, `<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td>%.2f</td></tr>`,
			row.ID,
			html.EscapeString(row.Type),
			html.EscapeString(row.Address),
			html.EscapeString(row.Station),
			row.ReceivedAt.Format("02.01.2006 15:04"),
			closed,
			html.EscapeString(statusLabel(row.Status)),
			row.Vehicles,
			row.Responders,
			row.DamageUah,
		)
	}
	b.WriteString(`</tbody></table>`)
	b.WriteString(reportFooter())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}

// GET /api/report/incident/{id} — детальний звіт по одному виклику (HTML, друк/PDF).
func (s *Store) handleIncidentDetailReport(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "некоректний id", http.StatusBadRequest)
		return
	}
	d, err := db.ReportIncidentDetail(r.Context(), s.Pool, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "виклик не знайдено", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	closed := "—"
	if d.ClosedAt != nil {
		closed = d.ClosedAt.Format("02.01.2006 15:04")
	}

	var b strings.Builder
	b.WriteString(reportHeader(fmt.Sprintf("Звіт по виклику №%d", d.ID)))
	fmt.Fprintf(&b, `<dl>
<dt>Тип</dt><dd>%s</dd>
<dt>Статус</dt><dd>%s</dd>
<dt>Частина</dt><dd>%s</dd>
<dt>Адреса</dt><dd>%s</dd>
<dt>Отримано</dt><dd>%s</dd>
<dt>Закрито</dt><dd>%s</dd>
<dt>Опис</dt><dd>%s</dd>
<dt>Збитки</dt><dd>%.2f грн</dd>
<dt>Постраждалі</dt><dd>%d</dd>
</dl>`,
		html.EscapeString(d.Type),
		html.EscapeString(statusLabel(d.Status)),
		html.EscapeString(d.Station),
		html.EscapeString(d.Address),
		d.ReceivedAt.Format("02.01.2006 15:04"),
		closed,
		html.EscapeString(d.Description),
		d.DamageUah,
		d.Casualties,
	)

	b.WriteString(`<div class="section"><h2>Техніка на виклику</h2>`)
	b.WriteString(listOrDash(d.Vehicles))
	b.WriteString(`</div>`)
	b.WriteString(`<div class="section"><h2>Бригада</h2>`)
	b.WriteString(listOrDash(d.Responders))
	b.WriteString(`</div>`)
	b.WriteString(`<div class="section"><h2>Підсумок операції</h2><p>`) 
b.WriteString(html.EscapeString(d.Summary))
	b.WriteString(`</p></div>`)
	b.WriteString(reportFooter())

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}

func listOrDash(items []string) string {
	if len(items) == 0 {
		return `<p>—</p>`
	}
	var b strings.Builder
	b.WriteString(`<ul>`)
	for _, it := range items {
		b.WriteString(`<li>` + html.EscapeString(it) + `</li>`)
	}
	b.WriteString(`</ul>`)
	return b.String()
}

func statusLabel(status string) string {
	switch status {
	case "active":
		return "активний"
	case "closed":
		return "закритий"
	case "cancelled":
		return "скасований"
	case "available":
		return "доступна"
	case "on_call":
		return "на виклику"
	case "maintenance":
		return "в ремонті"
	case "retired":
		return "списана"
	default:
		return status
	}
}

func reportHeader(title string) string {
	return `<!doctype html>
<html lang="uk">
<head>
<meta charset="utf-8">
<title>` + html.EscapeString(title) + `</title>
<style>
  body { font-family: 'Segoe UI', Arial, sans-serif; color: #1f2933; margin: 32px; }
  h1 { font-size: 20px; }
  h2 { font-size: 15px; }
  .meta { color: #616e7c; font-size: 12px; }
  table { border-collapse: collapse; width: 100%; margin-top: 16px; font-size: 12px; }
  th, td { border: 1px solid #cbd2d9; padding: 6px 8px; text-align: left; }
  th { background: #f4f6f8; }
  .section { margin-top: 20px; }
  dl { display: grid; grid-template-columns: 180px 1fr; gap: 4px 12px; font-size: 13px; margin: 16px 0; }
  dt { font-weight: 600; color: #3e4c59; }
  ul { font-size: 13px; }
  .actions { margin-top: 24px; }
  .btn { padding: 8px 16px; border: none; border-radius: 6px; background: #d94640; color: #fff; cursor: pointer; font-size: 14px; }
  @media print { .actions { display: none; } }
</style>
</head>
<body>
<h1>` + html.EscapeString(title) + `</h1>`
}

func reportFooter() string {
	return `<div class="actions"><button class="btn" onclick="window.print()">Друк / Зберегти як PDF</button></div>
</body>
</html>`
}
