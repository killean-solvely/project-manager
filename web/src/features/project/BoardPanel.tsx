import { useMemo, useState } from 'react'
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  closestCorners,
  useDroppable,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import type { Column as Col, Task } from '../../api/types'
import { useBoard, useCreateTask, useMoveTask, useTasks } from '../../api/hooks'
import { SortableTask, TaskCardView } from './TaskCard'
import { TaskEditorModal } from './TaskEditorModal'

export function BoardPanel({ projectId }: { projectId: string }) {
  const { data: board, isLoading } = useBoard(projectId)
  const { data: tasks } = useTasks(projectId)
  const move = useMoveTask(projectId)
  const [editing, setEditing] = useState<Task | null>(null)
  const [activeId, setActiveId] = useState<string | null>(null)
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }))

  const columns = board?.columns ?? []
  const columnIds = useMemo(() => new Set(columns.map((c) => c.id)), [columns])

  const byColumn = useMemo(() => {
    const m: Record<string, Task[]> = {}
    for (const c of columns) m[c.id] = []
    for (const t of tasks ?? []) (m[t.column_id] ??= []).push(t)
    return m
  }, [columns, tasks])

  const activeTask = tasks?.find((t) => t.id === activeId) ?? null

  const onDragEnd = (e: DragEndEvent) => {
    setActiveId(null)
    const { active, over } = e
    if (!over) return
    const dragId = String(active.id)
    const overId = String(over.id)

    const moving = tasks?.find((t) => t.id === dragId)
    if (!moving) return

    // The drop target is either a column (empty space) or another card.
    const toCol = columnIds.has(overId)
      ? overId
      : tasks?.find((t) => t.id === overId)?.column_id
    if (!toCol) return

    const destIds = (byColumn[toCol] ?? []).map((t) => t.id).filter((id) => id !== dragId)
    const index = columnIds.has(overId) ? destIds.length : Math.max(0, destIds.indexOf(overId))

    // No-op if it would land exactly where it already is.
    const fromIds = (byColumn[moving.column_id] ?? []).map((t) => t.id)
    if (toCol === moving.column_id && fromIds.indexOf(dragId) === index) return

    move.mutate({ id: dragId, column_id: toCol, position: index })
  }

  if (isLoading || !board) return <p className="text-ink-secondary">Loading board…</p>

  return (
    <>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCorners}
        onDragStart={(e: DragStartEvent) => setActiveId(String(e.active.id))}
        onDragCancel={() => setActiveId(null)}
        onDragEnd={onDragEnd}
      >
        <div className="flex gap-4 overflow-x-auto pb-2">
          {columns.map((col) => (
            <BoardColumn
              key={col.id}
              column={col}
              tasks={byColumn[col.id] ?? []}
              projectId={projectId}
              onOpen={setEditing}
            />
          ))}
        </div>
        <DragOverlay>
          {activeTask && (
            <div className="w-72">
              <TaskCardView task={activeTask} dragging />
            </div>
          )}
        </DragOverlay>
      </DndContext>

      {editing && (
        <TaskEditorModal projectId={projectId} task={editing} onClose={() => setEditing(null)} />
      )}
    </>
  )
}

function BoardColumn({
  column,
  tasks,
  projectId,
  onOpen,
}: {
  column: Col
  tasks: Task[]
  projectId: string
  onOpen: (t: Task) => void
}) {
  const { setNodeRef, isOver } = useDroppable({ id: column.id })
  return (
    <div className="flex w-72 shrink-0 flex-col">
      <div className="mb-2 flex items-center justify-between px-1">
        <h3 className="font-display text-sm font-medium">{column.name}</h3>
        <span className="text-xs text-ink-tertiary">{tasks.length}</span>
      </div>
      <div
        ref={setNodeRef}
        className={`min-h-24 flex-1 space-y-2 rounded-card border p-2 transition-colors ${
          isOver ? 'border-brand bg-brand-50' : 'border-line bg-muted/40'
        }`}
      >
        <SortableContext items={tasks.map((t) => t.id)} strategy={verticalListSortingStrategy}>
          {tasks.map((t) => (
            <SortableTask key={t.id} task={t} onOpen={onOpen} />
          ))}
        </SortableContext>
        <AddCard columnId={column.id} projectId={projectId} />
      </div>
    </div>
  )
}

function AddCard({ columnId, projectId }: { columnId: string; projectId: string }) {
  const [title, setTitle] = useState('')
  const create = useCreateTask(projectId)
  const submit = () => {
    if (!title.trim()) return
    create.mutate({ column_id: columnId, title: title.trim() }, { onSuccess: () => setTitle('') })
  }
  return (
    <input
      value={title}
      onChange={(e) => setTitle(e.target.value)}
      onKeyDown={(e) => {
        if (e.key === 'Enter') submit()
      }}
      placeholder="+ Add a card"
      className="w-full rounded-lg border border-transparent bg-transparent px-2 py-1.5 text-sm outline-none placeholder:text-ink-tertiary focus:border-line focus:bg-surface"
    />
  )
}
