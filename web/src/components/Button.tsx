import type { ButtonHTMLAttributes } from 'react'

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger'
type Size = 'sm' | 'md'

const variants: Record<Variant, string> = {
  primary: 'bg-brand text-white hover:bg-brand-hover active:bg-brand-active',
  secondary: 'bg-surface text-ink border border-line hover:bg-muted',
  ghost: 'bg-transparent text-ink-secondary hover:bg-muted',
  danger: 'bg-surface text-error border border-error/40 hover:bg-error/10',
}

const sizes: Record<Size, string> = {
  sm: 'text-sm px-2.5 py-1 gap-1',
  md: 'text-sm px-3.5 py-2 gap-1.5',
}

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  size?: Size
}

export function Button({ variant = 'secondary', size = 'md', className = '', ...rest }: Props) {
  return (
    <button
      className={`inline-flex items-center justify-center rounded-lg font-medium transition-colors disabled:opacity-50 disabled:pointer-events-none ${variants[variant]} ${sizes[size]} ${className}`}
      {...rest}
    />
  )
}
