export interface ColumnDef {
  name: string
  type: string
  nullable: boolean
  primary: boolean
}

export interface ForeignKeyDef {
  column: string
  refTable: string
  refColumn: string
}

export interface TableDef {
  name: string
  columns: ColumnDef[]
  foreignKeys: ForeignKeyDef[]
}

export interface LayoutPositions {
  [nodeId: string]: { x: number; y: number }
}

export type Row = Record<string, unknown>

export interface IncidentCard {
  id: number
  type: string
  address: string
  station: string
  status: string
  receivedAt: string
}

export interface StatusCount {
  status: string
  count: number
}

export interface DashboardStats {
  activeIncidents: number
  incidentsToday: number
  vehiclesAvailable: number
  vehiclesTotal: number
  activeEmployees: number
  recentIncidents: IncidentCard[]
  vehiclesByStatus: StatusCount[]
}

export interface ColumnInput {
  name: string
  type: string
  nullable: boolean
}

export async function fetchSchema(): Promise<TableDef[]> {
  const res = await fetch('/api/schema')
  if (!res.ok) {
    throw new Error(`не вдалося завантажити схему (HTTP ${res.status})`)
  }
  return (await res.json()) as TableDef[]
}

export async function fetchDashboard(): Promise<DashboardStats> {
  const res = await fetch('/api/dashboard')
  if (!res.ok) {
    throw new Error(`не вдалося завантажити дані дашборда (HTTP ${res.status})`)
  }
  return (await res.json()) as DashboardStats
}

export async function fetchLayout(view: string): Promise<LayoutPositions | null> {
  const res = await fetch(`/api/layout/${encodeURIComponent(view)}`)
  if (!res.ok) {
    return null
  }
  const data = (await res.json()) as { positions: LayoutPositions | null }
  return data.positions ?? null
}

export async function saveLayout(view: string, positions: LayoutPositions): Promise<void> {
  const res = await fetch(`/api/layout/${encodeURIComponent(view)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(positions),
  })
  if (!res.ok) {
    throw new Error(`не вдалося зберегти розташування (HTTP ${res.status})`)
  }
}

export async function fetchRows(table: string): Promise<Row[]> {
  const res = await fetch(`/api/tables/${encodeURIComponent(table)}`)
  if (!res.ok) {
    throw new Error(`не вдалося завантажити рядки "${table}" (HTTP ${res.status})`)
  }
  return (await res.json()) as Row[]
}

export async function insertRow(table: string, values: Row): Promise<Row> {
  const res = await fetch(`/api/tables/${encodeURIComponent(table)}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(values),
  })
  return expectOk(res, `не вдалося створити запис у "${table}"`)
}

export async function updateRow(table: string, id: number, values: Row): Promise<Row> {
  const res = await fetch(`/api/tables/${encodeURIComponent(table)}/${id}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(values),
  })
  return expectOk(res, `не вдалося оновити запис ${id} у "${table}"`)
}

export async function deleteRow(table: string, id: number): Promise<void> {
  const res = await fetch(`/api/tables/${encodeURIComponent(table)}/${id}`, {
    method: 'DELETE',
  })
  if (!res.ok) {
    throw new Error(`не вдалося видалити запис ${id} з "${table}" (HTTP ${res.status})`)
  }
}

export async function createTable(name: string, columns: ColumnInput[]): Promise<void> {
  const res = await fetch('/api/admin/tables', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, columns }),
  })
  await expectOkNoContent(res, `не вдалося створити таблицю "${name}"`)
}

export async function dropTable(name: string): Promise<void> {
  const res = await fetch(`/api/admin/tables/${encodeURIComponent(name)}`, {
    method: 'DELETE',
  })
  await expectOkNoContent(res, `не вдалося видалити таблицю "${name}"`)
}

export async function addColumn(table: string, col: ColumnInput): Promise<void> {
  const res = await fetch(`/api/admin/tables/${encodeURIComponent(table)}/columns`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(col),
  })
  await expectOkNoContent(res, `не вдалося додати колонку до "${table}"`)
}

export async function dropColumn(table: string, column: string): Promise<void> {
  const res = await fetch(
    `/api/admin/tables/${encodeURIComponent(table)}/columns/${encodeURIComponent(column)}`,
    { method: 'DELETE' },
  )
  await expectOkNoContent(res, `не вдалося видалити колонку "${column}"`)
}

async function expectOk(res: Response, message: string): Promise<Row> {
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { error?: string } | null
    throw new Error(body?.error ?? `${message} (HTTP ${res.status})`)
  }
  return (await res.json()) as Row
}

async function expectOkNoContent(res: Response, message: string): Promise<void> {
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { error?: string } | null
    throw new Error(body?.error ?? `${message} (HTTP ${res.status})`)
  }
}
