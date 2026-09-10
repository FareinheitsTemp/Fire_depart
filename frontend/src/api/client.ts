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

export async function fetchSchema(): Promise<TableDef[]> {
  const res = await fetch('/api/schema')
  if (!res.ok) {
    throw new Error(`не вдалося завантажити схему (HTTP ${res.status})`)
  }
  return (await res.json()) as TableDef[]
}
