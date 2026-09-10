import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  deleteRow,
  fetchRows,
  fetchSchema,
  insertRow,
  updateRow,
  type Row,
  type TableDef,
} from '../api/client'
import { RecordForm } from './RecordForm'

export function DataBrowser() {
  const [tables, setTables] = useState<TableDef[]>([])
  const [table, setTable] = useState<string | null>(null)
  const [rows, setRows] = useState<Row[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [editing, setEditing] = useState<Row | null>(null)
  const [creating, setCreating] = useState(false)

  useEffect(() => {
    fetchSchema()
      .then((ts) => {
        setTables(ts)
        if (ts.length > 0) {
          setTable(ts[0].name)
        }
      })
      .catch((e: unknown) => setError(e instanceof Error ? e.message : 'невідома помилка'))
  }, [])

  const def = useMemo(() => tables.find((t) => t.name === table) ?? null, [tables, table])

  const reload = useCallback(() => {
    if (!table) {
      return
    }
    setLoading(true)
    fetchRows(table)
      .then(setRows)
      .catch((e: unknown) => setError(e instanceof Error ? e.message : 'невідома помилка'))
      .finally(() => setLoading(false))
  }, [table])

  useEffect(() => {
    reload()
  }, [reload])

  const handleDelete = async (id: unknown) => {
    if (!table || typeof id !== 'number') {
      return
    }
    if (!window.confirm(`Видалити запис #${id}?`)) {
      return
    }
    try {
      await deleteRow(table, id)
      reload()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'помилка видалення')
    }
  }

  const switchTable = (name: string) => {
    setTable(name)
    setEditing(null)
    setCreating(false)
  }

  const handleSubmit = async (values: Row) => {
    if (!def) {
      return
    }
    if (editing && typeof editing.id === 'number') {
      await updateRow(def.name, editing.id, values)
    } else {
      await insertRow(def.name, values)
    }
    setCreating(false)
    setEditing(null)
    reload()
  }

  return (
    <section className="data-browser">
      <div className="data-browser__toolbar">
        <h2 className="data-browser__title">Дані</h2>
        <select
          className="data-browser__select"
          value={table ?? ''}
          onChange={(e) => switchTable(e.target.value)}
        >
          {tables.map((t) => (
            <option key={t.name} value={t.name}>
              {t.name}
            </option>
          ))}
        </select>
        <button
          type="button"
          className="data-browser__add"
          onClick={() => setCreating(true)}
          disabled={!def}
        >
          Додати запис
        </button>
      </div>

      {error && <p className="data-browser__hint data-browser__hint--error">Помилка: {error}</p>}
      {loading && <p className="data-browser__hint">Завантаження…</p>}

      <div className="data-browser__table-wrap">
        <table className="data-table">
          <thead>
            <tr className="data-table__row data-table__row--head">
              {def?.columns.map((c) => (
                <th key={c.name} className="data-table__cell data-table__cell--head">
                  {c.name}
                </th>
              ))}
              <th className="data-table__cell data-table__cell--head" />
            </tr>
          </thead>
          <tbody>
            {rows.map((r, i) => (
              <tr key={String(r.id ?? i)} className="data-table__row">
                {def?.columns.map((c) => (
                  <td key={c.name} className="data-table__cell">
                    {formatCell(r[c.name])}
                  </td>
                ))}
                <td className="data-table__cell data-table__cell--actions">
                  <button type="button" className="data-table__btn" onClick={() => setEditing(r)}>
                    ✏
                  </button>
                  <button
                    type="button"
                    className="data-table__btn data-table__btn--danger"
                    onClick={() => handleDelete(r.id)}
                  >
                    ✕
                  </button>
                </td>
              </tr>
            ))}
            {!loading && rows.length === 0 && (
              <tr className="data-table__row">
                <td className="data-table__cell data-table__cell--empty" colSpan={(def?.columns.length ?? 0) + 1}>
                  Немає записів
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {(creating || editing) && def && (
        <div className="modal">
          <div className="modal__body">
            <h3 className="modal__title">
              {editing ? `Редагування: ${def.name}` : `Новий запис: ${def.name}`}
            </h3>
            <RecordForm
              def={def}
              initial={editing}
              onSubmit={handleSubmit}
              onCancel={() => {
                setCreating(false)
                setEditing(null)
              }}
            />
          </div>
        </div>
      )}
    </section>
  )
}

function formatCell(v: unknown): string {
  if (v === null || v === undefined) {
    return '—'
  }
  if (typeof v === 'boolean') {
    return v ? 'так' : 'ні'
  }
  if (typeof v === 'object') {
    return JSON.stringify(v)
  }
  return String(v)
}
