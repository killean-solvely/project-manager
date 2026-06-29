import { useMemo, useState } from 'react'
import { useProjects } from '../api/hooks'
import type { Project, ProjectStatus } from '../api/types'
import { ProjectCard } from '../features/portfolio/ProjectCard'
import { NewIdeaModal } from '../features/portfolio/NewIdeaModal'
import { Button } from '../components/Button'

const LANES: { status: ProjectStatus; title: string; hint: string }[] = [
  { status: 'idea', title: 'Ideas', hint: 'Spitball and curate' },
  { status: 'active', title: 'Active', hint: 'Developing or maintaining' },
  { status: 'archived', title: 'Archived', hint: 'Deprecated or shelved' },
]

export function PortfolioPage() {
  const { data: projects, isLoading, isError, error } = useProjects()
  const [showNew, setShowNew] = useState(false)

  const byStatus = useMemo(() => {
    const map: Record<ProjectStatus, Project[]> = { idea: [], active: [], archived: [] }
    for (const p of projects ?? []) map[p.status]?.push(p)
    return map
  }, [projects])

  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-10 border-b border-line bg-page/80 backdrop-blur">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
          <div>
            <h1 className="font-display text-xl font-semibold">Project Manager</h1>
            <p className="text-sm text-ink-secondary">Your ideas, from spark to shipped</p>
          </div>
          <Button variant="primary" onClick={() => setShowNew(true)}>
            + New idea
          </Button>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-6">
        {isLoading && <p className="text-ink-secondary">Loading…</p>}
        {isError && (
          <div className="rounded-lg border border-error/30 bg-error/10 px-4 py-3 text-sm text-error">
            {(error as Error).message}. Is the API running on :4523?
          </div>
        )}
        {projects && (
          <div className="grid grid-cols-1 gap-5 md:grid-cols-3">
            {LANES.map((lane) => (
              <Lane
                key={lane.status}
                title={lane.title}
                hint={lane.hint}
                projects={byStatus[lane.status]}
              />
            ))}
          </div>
        )}
      </main>

      {showNew && <NewIdeaModal onClose={() => setShowNew(false)} />}
    </div>
  )
}

function Lane({ title, hint, projects }: { title: string; hint: string; projects: Project[] }) {
  return (
    <section>
      <h2 className="mb-3 font-display text-base font-medium">
        {title} <span className="text-ink-tertiary">{projects.length}</span>
      </h2>
      <div className="space-y-3">
        {projects.length === 0 && (
          <p className="rounded-card border border-dashed border-line px-4 py-6 text-center text-sm text-ink-tertiary">
            {hint}
          </p>
        )}
        {projects.map((p) => (
          <ProjectCard key={p.id} project={p} />
        ))}
      </div>
    </section>
  )
}
