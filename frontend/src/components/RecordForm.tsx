import { useEffect, useState, type FormEvent } from 'react'
import { fetchRows, type ForeignKeyDef, type Row, type TableDef } from '../api/client'

interface RecordFormProps {
  def: TableDef
  initial?: Row | null
  onSubmit: (values: Row) => Promise<void>
  onCancel: () => void
}

type FormValue = string | boolean
type InputKind = 'text' | 'number' | 'decimal' | 'boolean' | 'datetime' | 'date' | 'json'

function inputKind(type: string): InputKind {
  if (type.startsWith('integer')) {
    return 'number'
  }
  if (type.startsWith('numeric')) {
    return 'decimal'
  }
  if (type === 'boolean') {
    return 'boolean'
  }
  if (type.includes('timestamp')) {
    return 'datetime'
  }
  if (type === 'date') {
    return 'date'
  }
  if (type === 'jsonb') {
    return 'json'
  }
  return 'text'
}

function fkOf(def: TableDef, column: string): ForeignKeyDef | null {
  return def.foreignKeys.find((f) => f.column === column) ?? null
}

function rowLabel(r: Row): string {
  for (const [key, value] of Object.entries(r)) {
    if (key !== 'id' && typeof value === 'string' && value !== '') {
      return value
    }
  }
  return `#${String(r.id)}`
}

function initValues(def: TableDef, initial: Row | null): Record<string, FormValue> {
  const vals: Record<string, FormValue> = {}
  for (const c of def.columns) {
    if (c.primary) {
      continue
    }
    const raw = initial?.[c.name]
    const kind = inputKind(c.type)
    if (raw === null || raw === undefined) {
      vals[c.name] = kind === 'boolean' ? true : ''
      continue
    }
    if (kind === 'boolean') {
      vals[c.name] = Boolean(raw)
    } else if (kind === 'datetime') {
      vals[c.name] = String(raw).slice(0, 16)
    } else if (kind === 'date') {
      vals[c.name] = String(raw).slice(0, 10)
    } else {
      vals[c.name] = String(raw)
    }
  }
  return vals
}

export function RecordForm({ def, initial, onSubmit, onCancel }: RecordFormProps) {
  const [values, setValues] = useState<Record<string, FormValue>>(initValues(def, initial ?? null))
  const [fkOptions, setFkOptions] = useState<Record<string, { value: number; label: string }[]>>({})
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    setError(null)
    Promise.all(
      def.foreignKeys.map(async (f) => ({ f, rows: await fetchRows(f.refTable) })),
    )
      .then((entries) => {
        const opts: Record<string, { value: number; label: string }[]> = {}
        for (const { f, rows } of entries) {
          opts[f.column] = rows.map((r) => ({ value: Number(r.id), label: rowLabel(r) }))
        }
        setFkOptions(opts)
      })
      .catch((e: unknown) =>
        setError(e instanceof Error ? e.message : 'не вдалося завантажити довідники'),
      )
  }, [def])

  const editable = def.columns.filter((c) => !c.primary)

  const set = (name: string, v: FormValue) => {
    setValues((prev) => ({ ...prev, [name]: v }))
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    const payload: Row = {}
    for (const c of editable) {
      const v = values[c.name]
      const kind = inputKind(c.type)
      if (v === '' || v === undefined) {
        continue // порожні поля пропускаємо — БД застосує DEFAULT
      }
      switch (kind) {
        case 'number':
        case 'decimal': {
          const n = Number(v)
          if (Number.isNaN(n)) {
            setError(`Некоректне число у полі «${c.name}»`)
            return
          }
          payload[c.name] = n
          break
        }
        case 'boolean':
          payload[c.name] = Boolean(v)
          break
        case 'datetime':
          payload[c.name] = new Date(String(v)).toISOString()
          break
        case 'date':
          payload[c.name] = new Date(`${String(v)}T00:00:00`).toISOString()
          break
        case 'json': {
          try {
            payload[c.name] = JSON.parse(String(v))
          } catch {
            setError(`Некоректний JSON у полі «${c.name}»`)
            return
          }
          break
        }
        default:
          payload[c.name] = String(v)
      }
    }

    setBusy(true)
    try {
      await onSubmit(payload)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'помилка збереження')
    } finally {
      setBusy(false)
    }
  }

  return (
    <form className="record-form" onSubmit={handleSubmit}>
      {editable.map((c) => {
        const kind = inputKind(c.type)
        const fk = fkOf(def, c.name)
        return (
          <label key={c.name} className="record-form__field">
            <span className="record-form__label">
              {c.name}
              {fk ? ` → ${fk.refTable}.${fk.refColumn}` : ''}
            </span>
            {fk ? (
              <select
                className="record-form__select"
                value={String(values[c.name] ?? '')}
                onChange={(e) => set(c.name, e.target.value)}
              >
                <option value="">— не обрано —</option>
                {(fkOptions[c.name] ?? []).map((o) => (
                  <option key={o.value} value={String(o.value)}>
                    {o.label}
                  </option>
                ))}
              </select>
            ) : kind === 'boolean' ? (
              <input
                type="checkbox"
                className="record-form__checkbox"
                checked={Boolean(values[c.name])}
                onChange={(e) => set(c.name, e.target.checked)}
              />
            ) : kind === 'json' ? (
              <textarea
                className="record-form__textarea"
                rows={4}
                value={String(values[c.name] ?? '')}
                onChange={(e) => set(c.name, e.target.value)}
              />
            ) : (
              <input
                className="record-form__input"
                type={
                  kind === 'number' || kind === 'decimal'
                    ? 'number'
                    : kind === 'datetime'
                      ? 'datetime-local'
                      : kind === 'date'
                        ? 'date'
                        : 'text'
                }
                step={kind === 'decimal' ? '0.01' : undefined}
                value={String(values[c.name] ?? '')}
                onChange={(e) => set(c.name, e.target.value)}
              />
            )}
          </label>
        )
      })}
      {error && <p className="record-form__error">{error}</p>}
      <div className="record-form__actions">
        <button type="submit" className="record-form__submit" disabled={busy}>
          {busy ? 'Збереження…' : 'Зберегти'}
        </button>
        <button type="button" className="record-form__cancel" onClick={onCancel}>
          Скасувати
        </button>
      </div>
    </form>
  )
}
