import * as React from 'react'
import { cn } from '../../lib/utils'

export const Select = React.forwardRef<HTMLSelectElement, React.SelectHTMLAttributes<HTMLSelectElement>>(function Select({ className, ...props }, ref) {
  return <select ref={ref} className={cn('flex h-9 w-full rounded-md border border-slate-300 bg-white px-3 py-1 text-sm shadow-xs focus:outline-none focus:ring-2 focus:ring-slate-900/20', className)} {...props} />
})
