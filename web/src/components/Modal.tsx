import { useEffect, type ReactNode } from 'react'

interface Props {
  title: string
  onClose: () => void
  children: ReactNode
  footer?: ReactNode
  wide?: boolean
}

export function Modal({ title, onClose, children, footer, wide }: Props) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose])

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4 py-10"
      onClick={onClose}
    >
      <div
        className={`w-full ${wide ? 'max-w-3xl' : 'max-w-lg'} rounded-card bg-surface`}
        style={{ boxShadow: '0 20px 40px 0 rgba(0,0,0,.18), 0 8px 16px 0 rgba(0,0,0,.10)' }}
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b border-line px-5 py-3.5">
          <h3 className="text-lg font-medium">{title}</h3>
          <button
            onClick={onClose}
            aria-label="Close"
            className="text-2xl leading-none text-ink-tertiary hover:text-ink"
          >
            &times;
          </button>
        </div>
        <div className="px-5 py-4">{children}</div>
        {footer && (
          <div className="flex justify-end gap-2 border-t border-line px-5 py-3.5">{footer}</div>
        )}
      </div>
    </div>
  )
}
