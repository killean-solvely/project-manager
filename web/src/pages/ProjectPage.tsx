import { useState, type ReactNode } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useProject } from '../api/hooks'
import { ProjectHeader } from '../features/project/ProjectHeader'
import { DocsPanel } from '../features/project/DocsPanel'
import { BoardPanel } from '../features/project/BoardPanel'

export function ProjectPage() {
  const { id = '' } = useParams()
  const { data: project, isLoading, isError, error } = useProject(id)
  const [tab, setTab] = useState<'board' | 'docs'>('board')

  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-10 border-b border-line bg-page/80 backdrop-blur">
        <div className="mx-auto max-w-6xl px-6 py-3">
          <Link to="/" className="text-sm text-ink-secondary hover:text-ink">
            ← Portfolio
          </Link>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-6 py-6">
        {isLoading && <p className="text-ink-secondary">Loading…</p>}
        {isError && (
          <div className="rounded-lg border border-error/30 bg-error/10 px-4 py-3 text-sm text-error">
            {(error as Error).message}
          </div>
        )}
        {project && (
          <>
            <ProjectHeader project={project} />

            <div className="mb-4 mt-6 flex gap-1 border-b border-line">
              <Tab active={tab === 'board'} onClick={() => setTab('board')}>
                Board
              </Tab>
              <Tab active={tab === 'docs'} onClick={() => setTab('docs')}>
                Docs
              </Tab>
            </div>

            {tab === 'board' ? <BoardPanel projectId={id} /> : <DocsPanel projectId={id} />}
          </>
        )}
      </main>
    </div>
  )
}

function Tab({ active, onClick, children }: { active: boolean; onClick: () => void; children: ReactNode }) {
  return (
    <button
      onClick={onClick}
      className={`-mb-px border-b-2 px-4 py-2 text-sm font-medium transition-colors ${
        active ? 'border-brand text-brand' : 'border-transparent text-ink-secondary hover:text-ink'
      }`}
    >
      {children}
    </button>
  )
}
