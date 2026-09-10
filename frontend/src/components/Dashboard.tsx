import { useEffect, useState } from 'react'
import { fetchDashboard, type DashboardStats } from '../api/client'

export function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchDashboard()
      .then(setStats)
      .catch((e: unknown) => setError(e instanceof Error ? e.message : 'невідома помилка'))
  }, [])

  const openIncidentsReport = () => {
    window.open('/api/report/incidents', '_blank')
  }

  return (
    <section className="dashboard">
      <div className="dashboard__header">
        <h2 className="dashboard__title">Дашборд</h2>
        <button type="button" className="dashboard__report" onClick={openIncidentsReport}>
          PDF-звіт по викликах
        </button>
      </div>

      {error && <p className="dashboard__hint dashboard__hint--error">Помилка: {error}</p>}
      {!stats && !error && <p className="dashboard__hint">Завантаження…</p>}

      {stats && (
        <>
          <div className="dashboard__cards">
            <StatCard label="Активні виклики" value={stats.activeIncidents} tone="danger" />
            <StatCard label="Виклики за сьогодні" value={stats.incidentsToday} />
            <StatCard label="Техніка доступна" value={stats.vehiclesAvailable} tone="ok" />
            <StatCard label="Техніка всього" value={stats.vehiclesTotal} />
            <StatCard label="Активний склад" value={stats.activeEmployees} />
          </div>

          <div className="dashboard__lists">
            <div className="dashboard__panel">
              <h3 className="dashboard__panel-title">Останні виклики</h3>
              <table className="data-table">
                <thead>
                  <tr className="data-table__row data-table__row--head">
                    <th className="data-table__cell data-table__cell--head">№</th>
                    <th className="data-table__cell data-table__cell--head">Тип</th>
                    <th className="data-table__cell data-table__cell--head">Адреса</th>
                    <th className="data-table__cell data-table__cell--head">Частина</th>
                    <th className="data-table__cell data-table__cell--head">Отримано</th>
                    <th className="data-table__cell data-table__cell--head">Статус</th>
                    <th className="data-table__cell data-table__cell--head" />
                  </tr>
                </thead>
                <tbody>
                  {stats.recentIncidents.map((i) => (
                    <tr key={i.id} className="data-table__row">
                      <td className="data-table__cell">{i.id}</td>
                      <td className="data-table__cell">{i.type}</td>
                      <td className="data-table__cell">{i.address}</td>
                      <td className="data-table__cell">{i.station}</td>
                      <td className="data-table__cell">{formatDateTime(i.receivedAt)}</td>
                      <td className="data-table__cell">{statusLabel(i.status)}</td>
                      <td className="data-table__cell data-table__cell--actions">
                        <button
                          type="button"
                          className="data-table__btn"
                          title="Звіт по виклику"
                          onClick={() => window.open(`/api/report/incident/${i.id}`, '_blank')}
                        >
                          📄
                        </button>
                      </td>
                    </tr>
                  ))}
                  {stats.recentIncidents.length === 0 && (
                    <tr className="data-table__row">
                      <td className="data-table__cell data-table__cell--empty" colSpan={7}>
                        Викликів ще немає — додай їх у розділі «Дані (CRUD)»
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            <div className="dashboard__panel">
              <h3 className="dashboard__panel-title">Техніка за статусами</h3>
              <ul className="dashboard__status-list">
                {stats.vehiclesByStatus.map((v) => (
                  <li key={v.status} className="dashboard__status-item">
                    <span>{statusLabel(v.status)}</span>
                    <span>{v.count}</span>
                  </li>
                ))}
                {stats.vehiclesByStatus.length === 0 && (
                  <li className="dashboard__status-item">Техніки ще немає</li>
                )}
              </ul>
            </div>
          </div>
        </>
      )}
    </section>
  )
}

function StatCard({ label, value, tone }: { label: string; value: number; tone?: 'danger' | 'ok' }) {
  return (
    <div className={`dashboard-card${tone ? ` dashboard-card--${tone}` : ''}`}>
      <span className="dashboard-card__value">{value}</span>
      <span className="dashboard-card__label">{label}</span>
    </div>
  )
}

function formatDateTime(s: string): string {
  const d = new Date(s)
  return Number.isNaN(d.getTime()) ? s : d.toLocaleString('uk-UA')
}

function statusLabel(status: string): string {
  const labels: Record<string, string> = {
    active: 'активний',
    closed: 'закритий',
    cancelled: 'скасований',
    available: 'доступна',
    on_call: 'на виклику',
    maintenance: 'в ремонті',
    retired: 'списана',
  }
  return labels[status] ?? status
}
