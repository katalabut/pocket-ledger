import * as React from 'react'
import { cn } from '../../lib/utils'

export function Table({ className, ...props }: React.TableHTMLAttributes<HTMLTableElement>) { return <table className={cn('w-full caption-bottom text-sm', className)} {...props} /> }
export function THead(props: React.HTMLAttributes<HTMLTableSectionElement>) { return <thead className='bg-[var(--surface-muted)]/70' {...props} /> }
export function TBody(props: React.HTMLAttributes<HTMLTableSectionElement>) { return <tbody {...props} /> }
export function TR({ className, ...props }: React.HTMLAttributes<HTMLTableRowElement>) { return <tr className={cn('border-b border-[var(--border)] hover:bg-[var(--surface-muted)]/60', className)} {...props} /> }
export function TH({ className, ...props }: React.ThHTMLAttributes<HTMLTableCellElement>) { return <th className={cn('px-3 py-2 text-left text-xs font-medium uppercase tracking-wide text-[var(--muted-foreground)]', className)} {...props} /> }
export function TD({ className, ...props }: React.TdHTMLAttributes<HTMLTableCellElement>) { return <td className={cn('px-3 py-2.5 align-middle', className)} {...props} /> }
