import type { ReactNode } from 'react'
import { cn } from '../../lib/utils'

export function Alert({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={cn('rounded-md border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700', className)}>{children}</div>
}
