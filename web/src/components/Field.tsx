import type { ReactNode } from 'react'

export const controlClass =
  'w-full rounded-lg border border-line bg-surface px-3 py-2 text-sm text-ink outline-none transition-colors focus:border-brand focus:ring-2 focus:ring-brand/20'

export function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="block">
      <span className="mb-1 block font-display text-xs font-medium text-ink-secondary">{label}</span>
      {children}
    </label>
  )
}
