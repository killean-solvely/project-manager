import { useState } from 'react'
import type { Doc, DocumentType } from '../../api/types'
import { useDocuments, useMissingDocs } from '../../api/hooks'
import { Badge } from '../../components/Badge'
import { Button } from '../../components/Button'
import { docStatusMeta, docTypeLabel, missingReasonBadge } from '../../lib/ui'
import { DocEditorModal } from './DocEditorModal'

type EditorState =
  | { mode: 'edit'; doc: Doc }
  | { mode: 'create'; presetType?: DocumentType }
  | null

export function DocsPanel({ projectId }: { projectId: string }) {
  const { data: docs, isLoading } = useDocuments(projectId)
  const { data: missing } = useMissingDocs(projectId)
  const [editor, setEditor] = useState<EditorState>(null)

  return (
    <div>
      {/* completeness */}
      <div className="mb-5 rounded-card border border-line bg-surface p-4">
        <div className="mb-2 flex items-center justify-between">
          <h3 className="font-display font-medium">Documentation</h3>
          <Button size="sm" variant="secondary" onClick={() => setEditor({ mode: 'create' })}>
            + Add document
          </Button>
        </div>
        {missing && missing.length === 0 ? (
          <p className="text-sm text-success">All required documents are complete.</p>
        ) : (
          <div className="flex flex-wrap items-center gap-2">
            <span className="text-sm text-ink-secondary">Still needed:</span>
            {missing?.map((m) => (
              <button
                key={m.type}
                onClick={() => setEditor({ mode: 'create', presetType: m.type })}
                title="Write this document"
              >
                <Badge className={`${missingReasonBadge(m.reason)} hover:opacity-80`}>
                  {docTypeLabel[m.type]} · {m.reason}
                </Badge>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* document list */}
      {isLoading && <p className="text-ink-secondary">Loading…</p>}
      {docs && docs.length === 0 && (
        <p className="rounded-card border border-dashed border-line px-4 py-10 text-center text-sm text-ink-tertiary">
          No documents yet. Add an overview, technical design, or spec to get started.
        </p>
      )}
      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
        {docs?.map((doc) => (
          <button
            key={doc.id}
            onClick={() => setEditor({ mode: 'edit', doc })}
            className="rounded-card border border-line bg-surface p-4 text-left transition-colors hover:border-line-strong"
          >
            <div className="mb-1 flex items-center justify-between gap-2">
              <span className="font-display font-medium">{docTypeLabel[doc.type]}</span>
              <Badge className={docStatusMeta[doc.status].badge}>{docStatusMeta[doc.status].label}</Badge>
            </div>
            {doc.title && doc.title !== docTypeLabel[doc.type] && (
              <p className="text-sm text-ink-secondary">{doc.title}</p>
            )}
            <p className="mt-2 line-clamp-2 text-xs text-ink-tertiary">
              {doc.content.replace(/[#*`>_-]/g, '').trim().slice(0, 160) || 'Empty'}
            </p>
          </button>
        ))}
      </div>

      {editor?.mode === 'edit' && (
        <DocEditorModal projectId={projectId} doc={editor.doc} onClose={() => setEditor(null)} />
      )}
      {editor?.mode === 'create' && (
        <DocEditorModal projectId={projectId} presetType={editor.presetType} onClose={() => setEditor(null)} />
      )}
    </div>
  )
}
