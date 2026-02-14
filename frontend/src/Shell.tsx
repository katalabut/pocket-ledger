import { useState } from 'react'
import TransactionsPage from './pages/TransactionsPage'
import AccountsPage from './pages/AccountsPage'
import CategoriesPage from './pages/CategoriesPage'
import ImportPage from './pages/ImportPage'
import ReportsPage from './pages/ReportsPage'
import BudgetsPage from './pages/BudgetsPage'
import { AppShell, Page } from './components/layout/AppShell'

export default function Shell() {
  const [page, setPage] = useState<Page>('transactions')

  return (
    <AppShell page={page} setPage={setPage}>
      {page === 'transactions' && <TransactionsPage />}
      {page === 'accounts' && <AccountsPage />}
      {page === 'categories' && <CategoriesPage />}
      {page === 'import' && <ImportPage />}
      {page === 'reports' && <ReportsPage />}
      {page === 'budgets' && <BudgetsPage />}
    </AppShell>
  )
}
