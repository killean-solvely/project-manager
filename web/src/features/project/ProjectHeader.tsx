import { useState } from 'react'
import type { Project } from '../../api/types'
import { Badge } from '../../components/Badge'
import { Button } from '../../components/Button'
import { EditProjectModal } from './EditProjectModal'
import {
  useArchiveProject,
  usePromoteProject,
  useReviveProject,
  useSetProjectMode,
} from '../../api/hooks'
import { modeMeta, statusMeta } from '../../lib/ui'
import { renderMarkdown } from '../../lib/markdown'

export function ProjectHeader({ project }: { project: Project }) {
  const [editing, setEditing] = useState(false)
  const promote = usePromoteProject()
  const archive = useArchiveProject()
  const revive = useReviveProject()
  const setMode = useSetProjectMode()

  const status = statusMeta[project.status]
  const mode = modeMeta(project.mode)
  const busy = promote.isPending || archive.isPending || revive.isPending || setMode.isPending

  const onArchive = () => {
    const reason = window.prompt('Reason for archiving (optional):', '')
    if (reason === null) return
    archive.mutate({ id: project.id, reason })
  }

  return (
    <div className="rounded-card border border-line bg-surface p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h1 className="font-display text-2xl font-semibold">{project.name}</h1>
            <Badge className={status.badge}>{status.label}</Badge>
            {mode && <Badge className={mode.badge}>{mode.label}</Badge>}
          </div>
          {project.summary && <p className="mt-1 text-ink-secondary">{project.summary}</p>}
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {project.status === 'idea' && (
            <Button variant="primary" disabled={busy} onClick={() => promote.mutate({ id: project.id })}>
              Promote to active
            </Button>
          )}
          {project.status === 'active' && (
            <div className="inline-flex rounded-lg border border-line p-0.5">
              {(['developing', 'maintaining'] as const).map((m) => (
                <button
                  key={m}
                  disabled={busy}
                  onClick={() => setMode.mutate({ id: project.id, mode: m })}
                  className={`rounded-md px-3 py-1 text-sm font-medium capitalize transition-colors ${
                    project.mode === m ? 'bg-brand text-white' : 'text-ink-secondary hover:bg-muted'
                  }`}
                >
                  {m}
                </button>
              ))}
            </div>
          )}
          {project.status === 'archived' ? (
            <Button variant="secondary" disabled={busy} onClick={() => revive.mutate({ id: project.id })}>
              Revive
            </Button>
          ) : (
            <Button variant="danger" disabled={busy} onClick={onArchive}>
              Archive
            </Button>
          )}
          <Button variant="ghost" onClick={() => setEditing(true)}>
            Edit
          </Button>
        </div>
      </div>

      {project.tags.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-1">
          {project.tags.map((t) => (
            <span key={t} className="rounded-full bg-muted px-2 py-0.5 text-xs text-ink-secondary">
              #{t}
            </span>
          ))}
        </div>
      )}

      {project.description && (
        <div
          className="prose-doc mt-4 border-t border-line pt-4 text-sm"
          dangerouslySetInnerHTML={{ __html: renderMarkdown(project.description) }}
        />
      )}

      {project.status === 'archived' && project.archived_reason && (
        <p className="mt-3 text-sm text-ink-tertiary">Archived — {project.archived_reason}</p>
      )}

      {editing && <EditProjectModal project={project} onClose={() => setEditing(false)} />}
    </div>
  )
}
