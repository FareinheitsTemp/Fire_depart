import { useState } from 'react'
import { Sidebar } from './components/Sidebar'
import { Dashboard } from './components/Dashboard'
import { DataBrowser } from './components/DataBrowser'
import { SchemaMap } from './components/SchemaMap'

type Page = 'dashboard' | 'data' | 'schema'

const PAGES: { id: Page; label: string }[] = [
  { id: 'dashboard', label: 'Дашборд' },
  { id: 'data', label: 'Дані (CRUD)' },
  { id: 'schema', label: 'Структура БД' },
]

export default function App() {
  const [page, setPage] = useState<Page>('dashboard')

  return (
    <div className="app">
      <Sidebar pages={PAGES} active={page} onSelect={setPage} />
      <main className="app__content">
        {page === 'dashboard' && <Dashboard />}
        {page === 'data' && <DataBrowser />}
        {page === 'schema' && <SchemaMap />}
      </main>
    </div>
  )
}
