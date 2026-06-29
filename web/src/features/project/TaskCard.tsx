import { useDraggable } from '@dnd-kit/core'
import { CSS } from '@dnd-kit/utilities'
import type { Task } from '../../api/types'
import { priorityMeta } from '../../lib/ui'

export function TaskCardView({ task, dragging }: { task: Task; dragging?: boolean }) {
  const pr = priorityMeta[task.priority]
  return (
    <div className={`rounded-lg border border-line bg-surface p-3 ${dragging ? 'shadow-lg' : ''}`}>
      <div className="flex items-start gap-2">
        {task.priority !== 'none' && (
          <span
            className={`mt-1.5 h-2 w-2 shrink-0 rounded-full ${pr.dot}`}
            title={`${pr.label} priority`}
          />
        )}
        <p className={`text-sm ${task.completed_at ? 'text-ink-tertiary line-through' : 'text-ink'}`}>
          {task.title}
        </p>
      </div>
      {(task.labels.length > 0 || task.document_id) && (
        <div className="mt-2 flex flex-wrap items-center gap-1">
          {task.document_id && (
            <span className="rounded bg-brand-50 px-1.5 py-0.5 text-[11px] font-medium text-brand">doc</span>
          )}
          {task.labels.map((l) => (
            <span key={l} className="rounded bg-muted px-1.5 py-0.5 text-[11px] text-ink-secondary">
              {l}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}

export function DraggableTask({ task, onOpen }: { task: Task; onOpen: (t: Task) => void }) {
  const { attributes, listeners, setNodeRef, transform, isDragging } = useDraggable({ id: task.id })
  return (
    <div
      ref={setNodeRef}
      style={{ transform: CSS.Translate.toString(transform), opacity: isDragging ? 0.4 : 1 }}
      {...attributes}
      {...listeners}
      onClick={() => onOpen(task)}
      className="cursor-grab active:cursor-grabbing"
    >
      <TaskCardView task={task} />
    </div>
  )
}
