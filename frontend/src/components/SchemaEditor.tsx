import { useEffect, useState, type FormEvent } from 'react'
import {
  addColumn,
  createTable,
  dropColumn,
  dropTable,
  fetchSchema,
  type ColumnInput,
  type TableDef,
} from '../api/client'

const COLUMN_TYPES = [
  'integer',
  'bigint',
  'varchar(120)',
  'varchar(200)',
  'text',
  'boolean',
  'date',
  'timestamp',
  'numeric(14,2)',
  'jsonb',
]

interface SchemaEditorProps {
  onChanged: () => void
}

export function SchemaEditor({ onChanged }: SchemaEditorProps) {
  const [tables, setTables] = useState<TableDef[]>([])
  const [error, setError] = useState<string | null>(null)
  const [creating, setCreating] = useState(false)
  const [target, setTarget] = useState<TableDef | null>(null)

  const reload = () => {
    fetchSchema()
      .then(setTables)
      .catch((e: unknown) => setError(e instanceof Error ? e.message : 'невідома помилка'))
  }

  useEffect(() => {
    reload()
  }, [])

  const act = async (fn: () => Promise<void>) => {
    setError(null)
    try {
      await fn()
      reload()
      onChanged()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'помилка')
    }
  }

  const handleDropTable = (t: TableDef) => {
    if (window.confirm(`Видалити таблицю «${t.name}» разом з усіма даними?`)) {
      void act(() => dropTable(t.name))
    }
  }

  const handleDropColumn = (t: TableDef, column: string) => {
    if (window.confirm(`Видалити колонку «${column}» з таблиці «${t.name}»?`)) {
      void act(() => dropColumn(t.name, column))
    }
  }

  return (
    <section className="schema-editor">
      <div className="schema-editor__header">
        <h3 className="schema-editor__title">Редактор схеми</h3>
        <button type="button" className="schema-editor__create" onClick={() => setCreating(true)}>
          Нова таблиця
        </button>
      </div>
      {error && <p className="schema-editor__hint schema-editor__hint--error">{error}</p>}
      <ul className="schema-editor__list">
        {tables.map((t) => (
          <li key={t.name} className="schema-editor__item">
            <span className="schema-editor__name">{t.name}</span>
            <span className="schema-editor__meta">{t.columns.length} колонок</span>
            <div className="schema-editor__actions">
              <button type="button" className="schema-editor__btn" onClick={() => setTarget(t)}>
                + колонка
              </button>
              <button
                type="button"
                className="schema-editor__btn schema-editor__btn--danger"
                onClick={() => handleDropTable(t)}
              >
                Видалити таблицю
              </button>
            </div>
            <div className="schema-editor__cols">
              {t.columns
                .filter((c) => !c.primary)
                .map((c) => (
                  <button
                    key={c.name}
                    type="button"
                    className="schema-editor__col"
                    title={`Видалити колонку ${c.name}`}
                    onClick={() => handleDropColumn(t, c.name)}
                  >
                    {c.name} ✕
                  </button>
                ))}
            </div>
          </li>
        ))}
      </ul>

      {creating && (
        <div className="modal">
          <div className="modal__body">
            <h3 className="modal__title">Нова таблиця</h3>
            <NewTableForm
              onSubmit={async (name, cols) => {
                await act(() => createTable(name, cols))
                setCreating(false)
              }}
              onCancel={() => setCreating(false)}
            />
          </div>
        </div>
      )}

      {target && (
        <div className="modal">
          <div className="modal__body">
            <h3 className="modal__title">Нова колонка для «{target.name}»</h3>
            <ColumnForm
              onSubmit={async (col) => {
                await act(() => addColumn(target.name, col))
                setTarget(null)
              }}
              onCancel={() => setTarget(null)}
            />
          </div>
        </div>
      )}
    </section>
  )
}

function ColumnForm({
  onSubmit,
  onCancel,
}: {
  onSubmit: (col: ColumnInput) => Promise<void>
  onCancel: () => void
}) {
  const [name, setName] = useState('')
  const [type, setType] = useState('varchar(120)')
  const [nullable, setNullable] = useState(true)
  const [busy, setBusy] = useState(false)

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setBusy(true)
    try {
      await onSubmit({ name: name.trim(), type, nullable })
    } finally {
      setBusy(false)
    }
  }

  return (
    <form className="record-form" onSubmit={handleSubmit}>
      <label className="record-form__field">
        <span className="record-form__label">Назва колонки (напр. notes)</span>
        <input
          className="record-form__input"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
      </label>
      <label className="record-form__field">
        <span className="record-form__label">Тип</span>
        <select className="record-form__select" value={type} onChange={(e) => setType(e.target.value)}>
          {COLUMN_TYPES.map((t) => (
            <option key={t} value={t}>
              {t}
            </option>
          ))}
        </select>
      </label>
      <label className="record-form__field">
        <span className="record-form__label">Дозволено NULL</span>
        <input
          type="checkbox"
          className="record-form__checkbox"
          checked={nullable}
          onChange={(e) => setNullable(e.target.checked)}
        />
      </label>
      <div className="record-form__actions">
        <button type="submit" className="record-form__submit" disabled={busy || name.trim() === ''}>
          Додати
        </button>
        <button type="button" className="record-form__cancel" onClick={onCancel}>
          Скасувати
        </button>
      </div>
    </form>
  )
}

interface ColumnRow {
  name: string
  type: string
  nullable: boolean
}

function NewTableForm({
  onSubmit,
  onCancel,
}: {
  onSubmit: (name: string, cols: ColumnInput[]) => Promise<void>
  onCancel: () => void
}) {
  const [name, setName] = useState('')
  const [rows, setRows] = useState<ColumnRow[]>([{ name: '', type: 'varchar(120)', nullable: true }])
  const [busy, setBusy] = useState(false)

  const setRow = (i: number, patch: Partial<ColumnRow>) => {
    setRows((prev) => prev.map((row, idx) => (idx === i ? { ...row, ...patch } : row)))
  }

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault()
    const cols: ColumnInput[] = rows
      .filter((r) => r.name.trim() !== '')
      .map((r) => ({ name: r.name.trim(), type: r.type, nullable: r.nullable }))
    setBusy(true)
    try {
      await onSubmit(name.trim(), cols)
    } finally {
      setBusy(false)
    }
  }

  return (
    <form className="record-form" onSubmit={handleSubmit}>
      <label className="record-form__field">
        <span className="record-form__label">Назва таблиці (напр. equipment_checks)</span>
        <input
          className="record-form__input"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
      </label>
      {rows.map((row, i) => (
        <div key={i} className="schema-editor__row">
          <input
            className="record-form__input"
            placeholder="колонка"
            value={row.name}
            onChange={(e) => setRow(i, { name: e.target.value })}
          />
          <select
            className="record-form__select"
            value={row.type}
            onChange={(e) => setRow(i, { type: e.target.value })}
          >
            {COLUMN_TYPES.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
          <label className="schema-editor__null">
            <input
              type="checkbox"
              checked={row.nullable}
              onChange={(e) => setRow(i, { nullable: e.target.checked })}
            />
            null
          </label>
          <button
            type="button"
            className="schema-editor__row-del"
            onClick={() => setRows((prev) => prev.filter((_, idx) => idx !== i))}
          >
            ✕
          </button>
        </div>
      ))}
      <button
        type="button"
        className="schema-editor__add-row"
        onClick={() => setRows((prev) => [...prev, { name: '', type: 'varchar(120)', nullable: true }])}
      >
        + колонка
      </button>
      <div className="record-form__actions">
        <button type="submit" className="record-form__submit" disabled={busy || name.trim() === ''}>
          Створити
        </button>
        <button type="button" className="record-form__cancel" onClick={onCancel}>
          Скасувати
        </button>
      </div>
    </form>
  )
}
