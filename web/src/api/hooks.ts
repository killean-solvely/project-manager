import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from './client'
import type {
  Board,
  Doc,
  DocumentStatus,
  DocumentType,
  MissingDoc,
  Project,
  ProjectMode,
  Task,
  TaskPriority,
} from './types'

// --- queries ---

export function useProjects() {
  return useQuery({ queryKey: ['projects'], queryFn: () => api.get<Project[]>('/projects') })
}

export function useProject(id: string) {
  return useQuery({
    queryKey: ['project', id],
    queryFn: () => api.get<Project>(`/projects/${id}`),
    enabled: !!id,
  })
}

export function useDocuments(projectId: string) {
  return useQuery({
    queryKey: ['documents', projectId],
    queryFn: () => api.get<Doc[]>(`/projects/${projectId}/documents`),
    enabled: !!projectId,
  })
}

export function useMissingDocs(projectId: string) {
  return useQuery({
    queryKey: ['missingDocs', projectId],
    queryFn: () => api.get<MissingDoc[]>(`/projects/${projectId}/documents/missing`),
    enabled: !!projectId,
  })
}

export function useBoard(projectId: string) {
  return useQuery({
    queryKey: ['board', projectId],
    queryFn: () => api.get<Board>(`/projects/${projectId}/board`),
    enabled: !!projectId,
  })
}

export function useTasks(projectId: string) {
  return useQuery({
    queryKey: ['tasks', projectId],
    queryFn: () => api.get<Task[]>(`/projects/${projectId}/tasks`),
    enabled: !!projectId,
  })
}

// --- project mutations ---

export function useCreateIdea() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: { name: string; summary?: string; description?: string; tags?: string[] }) =>
      api.post<Project>('/projects', vars),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
  })
}

function useProjectWrite<V extends { id: string }>(
  fn: (vars: V) => Promise<Project>,
) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: fn,
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['projects'] })
      qc.invalidateQueries({ queryKey: ['project', vars.id] })
    },
  })
}

export function useUpdateProject() {
  return useProjectWrite(
    (vars: { id: string; name?: string; summary?: string; description?: string; tags?: string[] }) => {
      const { id, ...patch } = vars
      return api.patch<Project>(`/projects/${id}`, patch)
    },
  )
}

export function usePromoteProject() {
  return useProjectWrite((vars: { id: string; mode?: ProjectMode }) =>
    api.post<Project>(`/projects/${vars.id}/promote`, { mode: vars.mode ?? '' }),
  )
}

export function useSetProjectMode() {
  return useProjectWrite((vars: { id: string; mode: ProjectMode }) =>
    api.post<Project>(`/projects/${vars.id}/mode`, { mode: vars.mode }),
  )
}

export function useArchiveProject() {
  return useProjectWrite((vars: { id: string; reason?: string }) =>
    api.post<Project>(`/projects/${vars.id}/archive`, { reason: vars.reason ?? '' }),
  )
}

export function useReviveProject() {
  return useProjectWrite((vars: { id: string; mode?: ProjectMode }) =>
    api.post<Project>(`/projects/${vars.id}/revive`, { mode: vars.mode ?? '' }),
  )
}

// --- document mutations ---

export function useUpsertDocument(projectId: string) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: { type: DocumentType; title: string; content: string; status: DocumentStatus }) =>
      api.put<Doc>(`/projects/${projectId}/documents/${vars.type}`, {
        title: vars.title,
        content: vars.content,
        status: vars.status,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['documents', projectId] })
      qc.invalidateQueries({ queryKey: ['missingDocs', projectId] })
    },
  })
}

// --- task mutations (scoped to a project for cache invalidation) ---

function useTaskWrite<V, R>(projectId: string, fn: (vars: V) => Promise<R>) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: fn,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['tasks', projectId] }),
  })
}

export function useCreateTask(projectId: string) {
  return useTaskWrite(projectId, (vars: {
    column_id: string
    title: string
    description?: string
    priority?: TaskPriority
    labels?: string[]
    document_id?: string | null
  }) => api.post<Task>(`/projects/${projectId}/tasks`, vars))
}

export function useUpdateTask(projectId: string) {
  return useTaskWrite(projectId, (vars: {
    id: string
    title?: string
    description?: string
    priority?: TaskPriority
    labels?: string[]
  }) => {
    const { id, ...patch } = vars
    return api.patch<Task>(`/tasks/${id}`, patch)
  })
}

// reorderTasks mirrors the server's insert-and-renumber so the optimistic cache
// update matches what the backend will return.
function reorderTasks(tasks: Task[], id: string, toCol: string, position: number): Task[] {
  const moving = tasks.find((t) => t.id === id)
  if (!moving) return tasks
  const dest = tasks
    .filter((t) => t.column_id === toCol && t.id !== id)
    .sort((a, b) => a.position - b.position)
  const pos = Math.max(0, Math.min(position, dest.length))
  dest.splice(pos, 0, moving)
  const newPos = new Map(dest.map((t, i) => [t.id, i]))
  return tasks.map((t) => {
    if (t.id === id) return { ...t, column_id: toCol, position: pos }
    return newPos.has(t.id) ? { ...t, position: newPos.get(t.id)! } : t
  })
}

export function useMoveTask(projectId: string) {
  const qc = useQueryClient()
  const key = ['tasks', projectId]
  return useMutation({
    mutationFn: (vars: { id: string; column_id: string; position: number }) =>
      api.post<Task>(`/tasks/${vars.id}/move`, { column_id: vars.column_id, position: vars.position }),
    // Optimistically reorder so the card lands instantly; reconcile on settle.
    onMutate: async (vars) => {
      await qc.cancelQueries({ queryKey: key })
      const prev = qc.getQueryData<Task[]>(key)
      if (prev) qc.setQueryData<Task[]>(key, reorderTasks(prev, vars.id, vars.column_id, vars.position))
      return { prev }
    },
    onError: (_e, _vars, ctx) => {
      if (ctx?.prev) qc.setQueryData(key, ctx.prev)
    },
    onSettled: () => qc.invalidateQueries({ queryKey: key }),
  })
}

export function useLinkTaskDocument(projectId: string) {
  return useTaskWrite(projectId, (vars: { id: string; document_id: string | null }) =>
    api.post<Task>(`/tasks/${vars.id}/link`, { document_id: vars.document_id }),
  )
}

export function useDeleteTask(projectId: string) {
  return useTaskWrite(projectId, (id: string) => api.del<void>(`/tasks/${id}`))
}
