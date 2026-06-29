import { useState } from 'react'
import { TASK_PRIORITIES, type Task, type TaskPriority } from '../../api/types'
import { Modal } from '../../components/Modal'
import { Button } from '../../components/Button'
import { Field, controlClass } from '../../components/Field'
import { useDeleteTask, useDocuments, useLinkTaskDocument, useUpdateTask } from '../../api/hooks'
import { docTypeLabel, priorityMeta } from '../../lib/ui'

export function TaskEditorModal({
  projectId,
  task,
  onClose,
}: {
  projectId: string
  task: Task
  onClose: () => void
}) {
  const { data: docs } = useDocuments(projectId)
  const [title, setTitle] = useState(task.title)
  const [description, setDescription] = useState(task.description)
  const [priority, setPriority] = useState<TaskPriority>(task.priority)
  const [labels, setLabels] = useState(task.labels.join(', '))
  const [documentId, setDocumentId] = useState(task.document_id ?? '')
  const [err, setErr] = useState('')

  const update = useUpdateTask(projectId)
  const link = useLinkTaskDocument(projectId)
  const del = useDeleteTask(projectId)
  const busy = update.isPending || link.isPending || del.isPending

  const save = async () => {
    if (!title.trim()) return
    setErr('')
    try {
      await update.mutateAsync({
        id: task.id,
        title: title.trim(),
        description,
        priority,
        labels: labels.split(',').map((t) => t.trim()).filter(Boolean),
      })
      if ((documentId || '') !== (task.document_id || '')) {
        await link.mutateAsync({ id: task.id, document_id: documentId || null })
      }
      onClose()
    } catch (e) {
      setErr((e as Error).message)
    }
  }

  const remove = () => {
    if (!window.confirm('Delete this task?')) return
    del.mutate(task.id, { onSuccess: onClose })
  }

  return (
    <Modal
      title="Task"
      onClose={onClose}
      footer={
        <>
          <Button variant="danger" onClick={remove} disabled={busy} className="mr-auto">
            Delete
          </Button>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="primary" onClick={save} disabled={!title.trim() || busy}>
            {busy ? 'Saving…' : 'Save'}
          </Button>
        </>
      }
    >
      <div className="space-y-3">
        <Field label="Title">
          <input className={controlClass} value={title} onChange={(e) => setTitle(e.target.value)} autoFocus />
        </Field>
        <Field label="Description">
          <textarea
            className={`${controlClass} min-h-28 resize-y`}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </Field>
        <div className="grid grid-cols-2 gap-3">
          <Field label="Priority">
            <select
              className={controlClass}
              value={priority}
              onChange={(e) => setPriority(e.target.value as TaskPriority)}
            >
              {TASK_PRIORITIES.map((p) => (
                <option key={p} value={p}>
                  {priorityMeta[p].label}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Linked document">
            <select
              className={controlClass}
              value={documentId}
              onChange={(e) => setDocumentId(e.target.value)}
            >
              <option value="">None</option>
              {docs?.map((d) => (
                <option key={d.id} value={d.id}>
                  {docTypeLabel[d.type]}
                  {d.title && d.title !== docTypeLabel[d.type] ? ` — ${d.title}` : ''}
                </option>
              ))}
            </select>
          </Field>
        </div>
        <Field label="Labels (comma separated)">
          <input className={controlClass} value={labels} onChange={(e) => setLabels(e.target.value)} />
        </Field>
        {err && <p className="text-sm text-error">{err}</p>}
      </div>
    </Modal>
  )
}
