import { useState } from 'react'
import { clearToken } from './api'
import TransactionsPage from './pages/TransactionsPage'
import AccountsPage from './pages/AccountsPage'
import CategoriesPage from './pages/CategoriesPage'
import ImportPage from './pages/ImportPage'
import ReportsPage from './pages/ReportsPage'
import BudgetsPage from './pages/BudgetsPage'

type Page = 'transactions' | 'accounts' | 'categories' | 'import' | 'reports' | 'budgets'

const NAV: { key: Page; label: string }[] = [
  { key: 'transactions', label: 'Transactions' },
  { key: 'accounts', label: 'Accounts' },
  { key: 'categories', label: 'Categories' },
  { key: 'import', label: 'Import' },
  { key: 'reports', label: 'Reports' },
  { key: 'budgets', label: 'Budgets' },
]

export default function Shell() {
  const [page, setPage] = useState<Page>('transactions')

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white border-b border-gray-200 px-4 py-2 flex items-center gap-4">
        <span className="font-bold text-lg mr-4">Pocket Ledger</span>
        {NAV.map(n => (
          <button
            key={n.key}
            onClick={() => setPage(n.key)}
            className={`px-3 py-1 rounded text-sm ${page === n.key ? 'bg-blue-600 text-white' : 'text-gray-600 hover:bg-gray-100'}`}
          >
            {n.label}
          </button>
        ))}
        <button onClick={() => { clearToken(); window.location.reload() }} className="ml-auto text-sm text-red-600 hover:underline">Logout</button>
      </nav>
      <main className="p-4 max-w-7xl mx-auto">
        {page === 'transactions' && <TransactionsPage />}
        {page === 'accounts' && <AccountsPage />}
        {page === 'categories' && <CategoriesPage />}
        {page === 'import' && <ImportPage />}
        {page === 'reports' && <ReportsPage />}
        {page === 'budgets' && <BudgetsPage />}
      </main>
    </div>
  )
}
