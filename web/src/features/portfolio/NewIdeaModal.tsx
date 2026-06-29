import { useState } from 'react'
import { Modal } from '../../components/Modal'
import { Button } from '../../components/Button'
import { Field, controlClass } from '../../components/Field'
import { useCreateIdea } from '../../api/hooks'

export function NewIdeaModal({ onClose }: { onClose: () => void }) {
  const [name, setName] = useState('')
  const [summary, setSummary] = useState('')
  const [tags, setTags] = useState('')
  const [description, setDescription] = useState('')
  const create = useCreateIdea()

  const submit = () => {
    if (!name.trim()) return
    create.mutate(
      {
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
      title="New idea"
      onClose={onClose}
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="primary" onClick={submit} disabled={!name.trim() || create.isPending}>
            {create.isPending ? 'Capturing…' : 'Capture idea'}
          </Button>
        </>
      }
    >
      <div className="space-y-3">
        <Field label="Name">
          <input
            className={controlClass}
            value={name}
            onChange={(e) => setName(e.target.value)}
            autoFocus
            placeholder="What's the idea?"
          />
        </Field>
        <Field label="Summary">
          <input
            className={controlClass}
            value={summary}
            onChange={(e) => setSummary(e.target.value)}
            placeholder="One line"
          />
        </Field>
        <Field label="Tags (comma separated)">
          <input
            className={controlClass}
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            placeholder="cli, go, experiment"
          />
        </Field>
        <Field label="Notes">
          <textarea
            className={`${controlClass} min-h-28 resize-y font-mono`}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Braindump anything you're thinking…"
          />
        </Field>
        {create.isError && <p className="text-sm text-error">{(create.error as Error).message}</p>}
      </div>
    </Modal>
  )
}
