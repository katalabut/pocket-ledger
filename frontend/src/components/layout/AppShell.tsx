import type { ReactNode } from 'react'
import { clearToken } from '../../api'
import { Button } from '../ui/button'
import { cn } from '../../lib/utils'

export type Page = 'transactions' | 'accounts' | 'categories' | 'import' | 'reports' | 'budgets'

const NAV: { key: Page; label: string }[] = [
  { key: 'transactions', label: 'Transactions' },
  { key: 'accounts', label: 'Accounts' },
  { key: 'categories', label: 'Categories' },
  { key: 'import', label: 'Import' },
  { key: 'reports', label: 'Reports' },
  { key: 'budgets', label: 'Budgets' },
]

export function AppShell({ page, setPage, children }: { page: Page; setPage: (p: Page) => void; children: ReactNode }) {
  return (
    <div className="min-h-screen bg-slate-100/60">
      <header className="sticky top-0 z-40 border-b border-slate-200 bg-white/95 backdrop-blur">
        <div className="mx-auto flex max-w-7xl flex-wrap items-center gap-2 px-3 py-3 md:px-6">
          <div className="mr-3 font-semibold text-slate-900">Pocket Ledger</div>
          <nav className="flex flex-wrap gap-1">
            {NAV.map(n => (
              <button key={n.key} onClick={() => setPage(n.key)} className={cn('rounded-md px-3 py-1.5 text-sm', page === n.key ? 'bg-slate-900 text-white' : 'text-slate-600 hover:bg-slate-100')}>
                {n.label}
              </button>
            ))}
          </nav>
          <div className="ml-auto"><Button variant="outline" size="sm" onClick={() => { clearToken(); window.location.reload() }}>Logout</Button></div>
        </div>
      </header>
      <main className="mx-auto max-w-7xl p-3 md:p-6">{children}</main>
    </div>
  )
}
