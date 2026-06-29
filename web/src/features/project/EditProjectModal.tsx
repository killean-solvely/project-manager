import { useState } from 'react'
import type { Project } from '../../api/types'
import { Modal } from '../../components/Modal'
import { Button } from '../../components/Button'
import { Field, controlClass } from '../../components/Field'
import { useUpdateProject } from '../../api/hooks'

export function EditProjectModal({ project, onClose }: { project: Project; onClose: () => void }) {
  const [name, setName] = useState(project.name)
  const [summary, setSummary] = useState(project.summary)
  const [description, setDescription] = useState(project.description)
  const [tags, setTags] = useState(project.tags.join(', '))
  const update = useUpdateProject()

  const submit = () => {
    if (!name.trim()) return
    update.mutate(
      {
        id: project.id,
        name: name.trim(),
        summary: summary.trim(),
        description: description.trim(),
        tags: tags.split(',').map((t) => t.trim()).filter(Boolean),
      },
      { onSuccess: onClose },
    )
  }

  return (
    <Modal
      title="Edit project"
      onClose={onClose}
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="primary" onClick={submit} disabled={!name.trim() || update.isPending}>
            {update.isPending ? 'Saving…' : 'Save'}
          </Button>
        </>
      }
    >
      <div className="space-y-3">
        <Field label="Name">
          <input className={controlClass} value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <Field label="Summary">
          <input className={controlClass} value={summary} onChange={(e) => setSummary(e.target.value)} />
        </Field>
        <Field label="Tags (comma separated)">
          <input className={controlClass} value={tags} onChange={(e) => setTags(e.target.value)} />
        </Field>
        <Field label="Notes">
          <textarea
            className={`${controlClass} min-h-40 resize-y font-mono`}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </Field>
        {update.isError && <p className="text-sm text-error">{(update.error as Error).message}</p>}
      </div>
    </Modal>
  )
}
