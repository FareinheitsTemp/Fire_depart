interface SidebarProps<T extends string> {
  pages: { id: T; label: string }[]
  active: T
  onSelect: (id: T) => void
}

export function Sidebar<T extends string>({ pages, active, onSelect }: SidebarProps<T>) {
  return (
    <nav className="sidebar">
      <h1 className="sidebar__title">Fire Depart</h1>
      <ul className="sidebar__list">
        {pages.map((p) => (
          <li key={p.id}>
            <button
              type="button"
              className={`sidebar__link${p.id === active ? ' sidebar__link--active' : ''}`}
              onClick={() => onSelect(p.id)}
            >
              {p.label}
            </button>
          </li>
        ))}
      </ul>
    </nav>
  )
}
