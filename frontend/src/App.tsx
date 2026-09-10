import { useState } from 'react'
import { Sidebar } from './components/Sidebar'
import { SchemaMap } from './components/SchemaMap'

type Page = 'dashboard' | 'incidents' | 'schema'

const PAGES: { id: Page; label: string }[] = [
  { id: 'dashboard', label: 'Дашборд' },
  { id: 'incidents', label: 'Виклики' },
  { id: 'schema', label: 'Структура БД' },
]

export default function App() {
  const [page, setPage] = useState<Page>('dashboard')

  return (
    <div className="app">
      <Sidebar pages={PAGES} active={page} onSelect={setPage} />
      <main className="app__content">
        {page === 'dashboard' && (
          <section className="page">
            <h2 className="page__title">Дашборд</h2>
            <p className="page__hint">Активні виклики, техніка та зміни — наступний етап.</p>
          </section>
        )}
        {page === 'incidents' && (
          <section className="page">
            <h2 className="page__title">Виклики</h2>
            <p className="page__hint">CRUD викликів — наступний етап.</p>
          </section>
        )}
        {page === 'schema' && <SchemaMap />}
      </main>
    </div>
  )
}
