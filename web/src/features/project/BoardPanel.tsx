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
import type { Column as Col, Task } from '../../api/types'
import { useBoard, useCreateTask, useMoveTask, useTasks } from '../../api/hooks'
import { DraggableTask, TaskCardView } from './TaskCard'
import { TaskEditorModal } from './TaskEditorModal'

export function BoardPanel({ projectId }: { projectId: string }) {
  const { data: board, isLoading } = useBoard(projectId)
  const { data: tasks } = useTasks(projectId)
  const move = useMoveTask(projectId)
  const [editing, setEditing] = useState<Task | null>(null)
  const [activeId, setActiveId] = useState<string | null>(null)
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }))

  const byColumn = useMemo(() => {
    const m: Record<string, Task[]> = {}
    for (const c of board?.columns ?? []) m[c.id] = []
    for (const t of tasks ?? []) (m[t.column_id] ??= []).push(t)
    return m
  }, [board, tasks])

  const activeTask = tasks?.find((t) => t.id === activeId) ?? null

  const onDragEnd = (e: DragEndEvent) => {
    setActiveId(null)
    const { active, over } = e
    if (!over) return
    const target = String(over.id)
    const task = tasks?.find((t) => t.id === active.id)
    if (!task || task.column_id === target) return
    const count = (byColumn[target] ?? []).length
    move.mutate({ id: task.id, column_id: target, position: count })
  }

  if (isLoading || !board) return <p className="text-ink-secondary">Loading board…</p>

  return (
    <>
      <DndContext
        sensors={sensors}
        collisionDetection={closestCorners}
        onDragStart={(e: DragStartEvent) => setActiveId(String(e.active.id))}
        onDragEnd={onDragEnd}
      >
        <div className="flex gap-4 overflow-x-auto pb-2">
          {board.columns.map((col) => (
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
        {tasks.map((t) => (
          <DraggableTask key={t.id} task={t} onOpen={onOpen} />
        ))}
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
