import { useNavigate } from 'react-router-dom'
import type { Project } from '../../api/types'
import { Badge } from '../../components/Badge'
import { modeMeta, relativeTime } from '../../lib/ui'
import { usePromoteProject, useReviveProject } from '../../api/hooks'

export function ProjectCard({ project }: { project: Project }) {
  const navigate = useNavigate()
  const promote = usePromoteProject()
  const revive = useReviveProject()
  const mode = modeMeta(project.mode)
  const busy = promote.isPending || revive.isPending

  return (
    <div
      onClick={() => navigate(`/projects/${project.id}`)}
      className="cursor-pointer rounded-card border border-line bg-surface p-4 transition-colors hover:border-line-strong"
    >
      <div className="mb-1 flex items-start justify-between gap-2">
        <h3 className="font-display font-medium leading-snug">{project.name}</h3>
        {mode && <Badge className={mode.badge}>{mode.label}</Badge>}
      </div>

      {project.summary && (
        <p className="mb-3 line-clamp-2 text-sm text-ink-secondary">{project.summary}</p>
      )}

      {project.tags.length > 0 && (
        <div className="mb-3 flex flex-wrap gap-1">
          {project.tags.map((t) => (
            <span key={t} className="rounded-full bg-muted px-2 py-0.5 text-xs text-ink-secondary">
              #{t}
            </span>
          ))}
        </div>
      )}

      <div className="flex items-center justify-between">
        <span className="text-xs text-ink-tertiary">Updated {relativeTime(project.updated_at)}</span>
        {project.status === 'idea' && (
          <button
            onClick={(e) => {
              e.stopPropagation()
              promote.mutate({ id: project.id })
            }}
            disabled={busy}
            className="rounded-lg border border-line px-2.5 py-1 text-xs font-medium hover:bg-muted disabled:opacity-50"
          >
            Promote →
          </button>
        )}
        {project.status === 'archived' && (
          <button
            onClick={(e) => {
              e.stopPropagation()
              revive.mutate({ id: project.id })
            }}
            disabled={busy}
            className="rounded-lg border border-line px-2.5 py-1 text-xs font-medium hover:bg-muted disabled:opacity-50"
          >
            Revive
          </button>
        )}
      </div>
    </div>
  )
}
