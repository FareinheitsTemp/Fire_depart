import { useCallback, useEffect, useState } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  type Edge,
  type Node,
  type NodeProps,
  type OnNodeDrag,
  useEdgesState,
  useNodesState,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import {
  fetchLayout,
  fetchSchema,
  saveLayout,
  type LayoutPositions,
  type TableDef,
} from '../api/client'

const LAYOUT_VIEW = 'schema'

type TableNodeData = { title: string; columns: string[] }
type TableNode = Node<TableNodeData, 'table'>

function TableNodeCard({ data }: NodeProps<TableNode>) {
  return (
    <div className="erd-table">
      <div className="erd-table__header">{data.title}</div>
      <ul className="erd-table__list">
        {data.columns.map((c) => (
          <li key={c} className="erd-table__item">
            {c}
          </li>
        ))}
      </ul>
    </div>
  )
}

const nodeTypes = { table: TableNodeCard }

export function SchemaMap() {
  const [nodes, setNodes, onNodesChange] = useNodesState<TableNode>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [saveState, setSaveState] = useState<'idle' | 'saving' | 'saved' | 'error'>('idle')

  useEffect(() => {
    Promise.all([fetchSchema(), fetchLayout(LAYOUT_VIEW)])
      .then(([tables, layout]) => {
        setNodes(toNodes(tables, layout))
        setEdges(toEdges(tables))
        setLoading(false)
      })
      .catch((e: unknown) => {
        setError(e instanceof Error ? e.message : 'невідома помилка')
        setLoading(false)
      })
  }, [setNodes, setEdges])

  const persistLayout = useCallback((current: TableNode[]) => {
    const positions: LayoutPositions = {}
    for (const n of current) {
      positions[n.id] = n.position
    }
    setSaveState('saving')
    saveLayout(LAYOUT_VIEW, positions)
      .then(() => setSaveState('saved'))
      .catch(() => setSaveState('error'))
  }, [])

  const handleNodeDragStop: OnNodeDrag = useCallback(
    () => {
      persistLayout(nodes)
    },
    [nodes, persistLayout],
  )

  return (
    <section className="schema-map">
      <h2 className="schema-map__title">Структура бази даних</h2>
      <div className="schema-map__toolbar">
        <button type="button" className="schema-map__save" onClick={() => persistLayout(nodes)}>
          Зберегти розташування
        </button>
        {saveState === 'saving' && <span className="schema-map__status">Збереження…</span>}
        {saveState === 'saved' && (
          <span className="schema-map__status schema-map__status--saved">Збережено</span>
        )}
        {saveState === 'error' && (
          <span className="schema-map__status schema-map__status--error">Помилка збереження</span>
        )}
      </div>
      {loading && <p className="schema-map__hint">Завантаження схеми…</p>}
      {error && <p className="schema-map__hint schema-map__hint--error">Помилка: {error}</p>}
      <div className="schema-map__canvas">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onNodeDragStop={handleNodeDragStop}
          fitView
        >
          <Background />
          <Controls />
        </ReactFlow>
      </div>
    </section>
  )
}

function toNodes(tables: TableDef[], layout: LayoutPositions | null): TableNode[] {
  return tables.map((t, i) => {
    const fallback = { x: (i % 4) * 300 + 40, y: Math.floor(i / 4) * 340 + 40 }
    return {
      id: t.name,
      type: 'table' as const,
      position: layout?.[t.name] ?? fallback,
      data: {
        title: t.name,
        columns: t.columns.map((c) => `${c.primary ? 'PK ' : ''}${c.name}: ${c.type}`),
      },
    }
  })
}

function toEdges(tables: TableDef[]): Edge[] {
  return tables.flatMap((t) =>
    t.foreignKeys.map((fk) => ({
      id: `${t.name}.${fk.column}->${fk.refTable}`,
      source: t.name,
      target: fk.refTable,
      label: fk.column,
      animated: true,
    })),
  )
}
