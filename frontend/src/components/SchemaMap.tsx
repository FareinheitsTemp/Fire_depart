import { useEffect, useState } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  type Edge,
  type Node,
  type NodeProps,
  useEdgesState,
  useNodesState,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { fetchSchema, type TableDef } from '../api/client'

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

  useEffect(() => {
    fetchSchema()
      .then((tables) => {
        setNodes(toNodes(tables))
        setEdges(toEdges(tables))
        setLoading(false)
      })
      .catch((e: unknown) => {
        setError(e instanceof Error ? e.message : 'невідома помилка')
        setLoading(false)
      })
  }, [setNodes, setEdges])

  return (
    <section className="schema-map">
      <h2 className="schema-map__title">Структура бази даних</h2>
      {loading && <p className="schema-map__hint">Завантаження схеми…</p>}
      {error && <p className="schema-map__hint schema-map__hint--error">Помилка: {error}</p>}
      <div className="schema-map__canvas">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          fitView
        >
          <Background />
          <Controls />
        </ReactFlow>
      </div>
    </section>
  )
}

function toNodes(tables: TableDef[]): TableNode[] {
  return tables.map((t, i) => ({
    id: t.name,
    type: 'table' as const,
    position: { x: (i % 4) * 300 + 40, y: Math.floor(i / 4) * 340 + 40 },
    data: {
      title: t.name,
      columns: t.columns.map((c) => `${c.primary ? 'PK ' : ''}${c.name}: ${c.type}`),
    },
  }))
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
