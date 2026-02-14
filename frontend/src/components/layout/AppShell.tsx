import type { ReactNode } from 'react'
import { clearToken } from '../../api'
import { Button } from '../ui/button'
import { cn } from '../../lib/utils'

export type Page = 'transactions' | 'accounts' | 'categories' | 'import' | 'reports' | 'budgets'

const NAV: { key: Page; label: string; desc: string }[] = [
  { key: 'transactions', label: 'Transactions', desc: 'Track day-to-day money movement' },
  { key: 'accounts', label: 'Accounts', desc: 'Manage wallets and balances' },
  { key: 'categories', label: 'Categories', desc: 'Organize spending and income labels' },
  { key: 'import', label: 'Import', desc: 'Bring CSV data into your ledger' },
  { key: 'reports', label: 'Reports', desc: 'Understand trends and balances' },
  { key: 'budgets', label: 'Budgets', desc: 'Plan monthly limits and monitor progress' },
]

export function AppShell({ page, setPage, children }: { page: Page; setPage: (p: Page) => void; children: ReactNode }) {
  const current = NAV.find(n => n.key === page)

  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-40 border-b border-[var(--border)] bg-[#fffaf4]/95 backdrop-blur">
        <div className="mx-auto flex max-w-7xl flex-wrap items-center gap-2 px-3 py-3 md:px-6">
          <div className="mr-3">
            <div className="font-semibold text-[var(--foreground)]">Pocket Ledger</div>
            <div className="text-xs text-[var(--muted-foreground)]">Personal money companion</div>
          </div>
          <nav className="flex flex-wrap gap-1">
            {NAV.map(n => (
              <button key={n.key} onClick={() => setPage(n.key)} className={cn('rounded-lg px-3 py-1.5 text-sm transition', page === n.key ? 'bg-[var(--primary)] text-[var(--primary-foreground)]' : 'text-[var(--muted-foreground)] hover:bg-[var(--surface-muted)] hover:text-[var(--foreground)]')}>
                {n.label}
              </button>
            ))}
          </nav>
          <div className="ml-auto"><Button variant="outline" size="sm" onClick={() => { clearToken(); window.location.reload() }}>Logout</Button></div>
        </div>
      </header>
      <main className="mx-auto max-w-7xl space-y-4 p-3 md:p-6">
        {current && (
          <div className="rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3 md:px-6 md:py-4">
            <h1 className="text-xl font-semibold text-[var(--foreground)] md:text-2xl">{current.label}</h1>
            <p className="text-sm text-[var(--muted-foreground)]">{current.desc}</p>
          </div>
        )}
        {children}
      </main>
    </div>
  )
}
