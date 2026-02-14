import * as React from 'react'
import { cn } from '../../lib/utils'

export const Select = React.forwardRef<HTMLSelectElement, React.SelectHTMLAttributes<HTMLSelectElement>>(function Select({ className, ...props }, ref) {
  return <select ref={ref} className={cn('flex h-10 w-full rounded-lg border border-[var(--border)] bg-white px-3 py-2 text-sm shadow-xs focus:outline-none focus:ring-2 focus:ring-[var(--primary)]/25', className)} {...props} />
})
