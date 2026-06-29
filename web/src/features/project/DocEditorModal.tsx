import { useState } from 'react'
import {
  DOCUMENT_STATUSES,
  DOCUMENT_TYPES,
  type Doc,
  type DocumentStatus,
  type DocumentType,
} from '../../api/types'
import { Modal } from '../../components/Modal'
import { Button } from '../../components/Button'
import { Field, controlClass } from '../../components/Field'
import { useUpsertDocument } from '../../api/hooks'
import { docStatusMeta, docTypeLabel } from '../../lib/ui'
import { renderMarkdown } from '../../lib/markdown'

interface Props {
  projectId: string
  doc?: Doc
  presetType?: DocumentType
  onClose: () => void
}

export function DocEditorModal({ projectId, doc, presetType, onClose }: Props) {
  const isEdit = !!doc
  const [type, setType] = useState<DocumentType>(doc?.type ?? presetType ?? 'overview')
  const [title, setTitle] = useState(doc?.title ?? '')
  const [content, setContent] = useState(doc?.content ?? '')
  const [status, setStatus] = useState<DocumentStatus>(doc?.status ?? 'draft')
  const upsert = useUpsertDocument(projectId)

  const submit = () => {
    upsert.mutate(
      { type, title: title.trim() || docTypeLabel[type], content, status },
      { onSuccess: onClose },
    )
  }

  return (
    <Modal
      title={isEdit ? `Edit ${docTypeLabel[type]}` : 'New document'}
      onClose={onClose}
      wide
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="primary" onClick={submit} disabled={upsert.isPending}>
            {upsert.isPending ? 'Saving…' : 'Save document'}
          </Button>
        </>
      }
    >
      <div className="space-y-3">
        <div className="grid grid-cols-2 gap-3">
          <Field label="Type">
            <select
              className={controlClass}
              value={type}
              disabled={isEdit}
              onChange={(e) => setType(e.target.value as DocumentType)}
            >
              {DOCUMENT_TYPES.map((t) => (
                <option key={t} value={t}>
                  {docTypeLabel[t]}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Status">
            <select
              className={controlClass}
              value={status}
              onChange={(e) => setStatus(e.target.value as DocumentStatus)}
            >
              {DOCUMENT_STATUSES.map((s) => (
                <option key={s} value={s}>
                  {docStatusMeta[s].label}
                </option>
              ))}
            </select>
          </Field>
        </div>
        <Field label="Title">
          <input
            className={controlClass}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder={docTypeLabel[type]}
          />
        </Field>
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
          <div>
            <span className="mb-1 block font-display text-xs font-medium text-ink-secondary">
              Markdown
            </span>
            <textarea
              className={`${controlClass} min-h-72 resize-y font-mono text-xs`}
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="# Heading&#10;&#10;Write the doc…"
            />
          </div>
          <div>
            <span className="mb-1 block font-display text-xs font-medium text-ink-secondary">
              Preview
            </span>
            <div
              className="prose-doc min-h-72 overflow-auto rounded-lg border border-line bg-page px-3 py-2 text-sm"
              dangerouslySetInnerHTML={{ __html: renderMarkdown(content || '_Nothing yet…_') }}
            />
          </div>
        </div>
        {upsert.isError && <p className="text-sm text-error">{(upsert.error as Error).message}</p>}
      </div>
    </Modal>
  )
}
