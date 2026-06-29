// Mirrors the Go server DTOs.

export type ProjectStatus = 'idea' | 'active' | 'archived'
export type ProjectMode = 'developing' | 'maintaining' | ''

export interface Project {
  id: string
  name: string
  summary: string
  description: string
  status: ProjectStatus
  mode?: ProjectMode
  tags: string[]
  archived_reason?: string
  promoted_at?: string
  archived_at?: string
  created_at: string
  updated_at: string
}

export type DocumentType =
  | 'overview'
  | 'technical'
  | 'spec'
  | 'api'
  | 'runbook'
  | 'other'

export type DocumentStatus = 'draft' | 'in_review' | 'complete'

export interface Doc {
  id: string
  project_id: string
  type: DocumentType
  title: string
  content: string
  status: DocumentStatus
  created_at: string
  updated_at: string
}

export interface MissingDoc {
  type: DocumentType
  reason: string // "missing" | "incomplete"
}

export interface Column {
  id: string
  board_id: string
  name: string
  position: number
}

export interface Board {
  id: string
  project_id: string
  name: string
  columns: Column[]
}

export type TaskPriority = 'none' | 'low' | 'medium' | 'high'

export interface Task {
  id: string
  board_id: string
  column_id: string
  title: string
  description: string
  priority: TaskPriority
  labels: string[]
  document_id?: string
  position: number
  completed_at?: string
  created_at: string
  updated_at: string
}

export const DOCUMENT_TYPES: DocumentType[] = [
  'overview',
  'technical',
  'spec',
  'api',
  'runbook',
  'other',
]

export const DOCUMENT_STATUSES: DocumentStatus[] = ['draft', 'in_review', 'complete']

export const TASK_PRIORITIES: TaskPriority[] = ['none', 'low', 'medium', 'high']
