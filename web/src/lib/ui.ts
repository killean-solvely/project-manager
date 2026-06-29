import type {
  DocumentStatus,
  DocumentType,
  ProjectMode,
  ProjectStatus,
  TaskPriority,
} from '../api/types'

export function relativeTime(iso?: string): string {
  if (!iso) return ''
  const then = new Date(iso).getTime()
  if (Number.isNaN(then)) return ''
  const diff = Date.now() - then
  const min = Math.round(diff / 60000)
  if (min < 1) return 'just now'
  if (min < 60) return `${min}m ago`
  const hr = Math.round(min / 60)
  if (hr < 24) return `${hr}h ago`
  const day = Math.round(hr / 24)
  if (day < 30) return `${day}d ago`
  return new Date(iso).toLocaleDateString()
}

export const statusMeta: Record<ProjectStatus, { label: string; badge: string }> = {
  idea: { label: 'Idea', badge: 'bg-warning/15 text-warning' },
  active: { label: 'Active', badge: 'bg-success/15 text-success' },
  archived: { label: 'Archived', badge: 'bg-muted text-ink-tertiary' },
}

export function modeMeta(mode?: ProjectMode): { label: string; badge: string } | null {
  if (mode === 'developing') return { label: 'Developing', badge: 'bg-brand/15 text-brand' }
  if (mode === 'maintaining') return { label: 'Maintaining', badge: 'bg-note/15 text-note' }
  return null
}

export const docStatusMeta: Record<DocumentStatus, { label: string; badge: string }> = {
  draft: { label: 'Draft', badge: 'bg-warning/15 text-warning' },
  in_review: { label: 'In review', badge: 'bg-help/15 text-help' },
  complete: { label: 'Complete', badge: 'bg-success/15 text-success' },
}

export const docTypeLabel: Record<DocumentType, string> = {
  overview: 'Overview',
  technical: 'Technical design',
  spec: 'Spec',
  api: 'API reference',
  runbook: 'Runbook',
  other: 'Other',
}

export const priorityMeta: Record<TaskPriority, { label: string; dot: string; text: string }> = {
  none: { label: 'None', dot: 'bg-line', text: 'text-ink-tertiary' },
  low: { label: 'Low', dot: 'bg-ink-tertiary', text: 'text-ink-secondary' },
  medium: { label: 'Medium', dot: 'bg-warning', text: 'text-warning' },
  high: { label: 'High', dot: 'bg-error', text: 'text-error' },
}

export function missingReasonBadge(reason: string): string {
  return reason === 'missing' ? 'bg-error/12 text-error' : 'bg-warning/15 text-warning'
}
