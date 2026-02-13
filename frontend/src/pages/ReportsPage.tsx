import { useState } from 'react'
import { reportSpending, reportBalances, SpendingRow, AccountBalance, syncFXRates } from '../api'

export default function ReportsPage() {
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [spending, setSpending] = useState<SpendingRow[]>([])
  const [balances, setBalances] = useState<AccountBalance[]>([])
  const [tab, setTab] = useState<'spending' | 'balances'>('spending')
  const [error, setError] = useState('')
  const [syncing, setSyncing] = useState(false)

  const loadSpending = async () => {
    setError('')
    try {
      const params: Record<string, string> = {}
      if (from) params.from = from + 'T00:00:00Z'
      if (to) params.to = to + 'T23:59:59Z'
      const data = await reportSpending(params)
      setSpending(data || [])
    } catch (e: unknown) { setError((e as Error).message) }
  }

  const loadBalances = async () => {
    setError('')
    try {
      const data = await reportBalances()
      setBalances(data || [])
    } catch (e: unknown) { setError((e as Error).message) }
  }

  const doSync = async () => {
    setSyncing(true); setError('')
    try {
      const r = await syncFXRates()
      alert(`Synced ${r.synced} rates`)
    } catch (e: unknown) { setError((e as Error).message) }
    finally { setSyncing(false) }
  }

  const fmt = (minor: number) => (minor / 100).toFixed(2)

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">Reports</h2>
        <button onClick={doSync} disabled={syncing} className="bg-purple-600 text-white px-4 py-2 rounded text-sm disabled:opacity-50">
          {syncing ? 'Syncing...' : 'Sync FX Rates'}
        </button>
      </div>

      {error && <p className="text-red-600 text-sm mb-4">{error}</p>}

      <div className="flex gap-2 mb-4">
        <button onClick={() => setTab('spending')} className={`px-3 py-1 rounded text-sm ${tab === 'spending' ? 'bg-blue-600 text-white' : 'bg-gray-200'}`}>Spending by Category</button>
        <button onClick={() => setTab('balances')} className={`px-3 py-1 rounded text-sm ${tab === 'balances' ? 'bg-blue-600 text-white' : 'bg-gray-200'}`}>Account Balances</button>
      </div>

      {tab === 'spending' && (
        <div className="bg-white p-4 rounded shadow">
          <div className="flex gap-2 items-end mb-4">
            <div>
              <label className="text-xs text-gray-500">From</label>
              <input type="date" value={from} onChange={e => setFrom(e.target.value)} className="block border rounded px-2 py-1 text-sm" />
            </div>
            <div>
              <label className="text-xs text-gray-500">To</label>
              <input type="date" value={to} onChange={e => setTo(e.target.value)} className="block border rounded px-2 py-1 text-sm" />
            </div>
            <button onClick={loadSpending} className="bg-blue-600 text-white px-4 py-1 rounded text-sm">Load</button>
          </div>
          <table className="w-full text-sm">
            <thead><tr className="border-b text-left text-gray-500">
              <th className="px-3 py-2">Category</th><th className="px-3 py-2 text-right">Total (base)</th>
            </tr></thead>
            <tbody>
              {spending.map(s => (
                <tr key={s.CategoryID} className="border-b">
                  <td className="px-3 py-2">{s.CategoryName}</td>
                  <td className="px-3 py-2 text-right">{fmt(s.TotalMinor)}</td>
                </tr>
              ))}
              {spending.length === 0 && <tr><td colSpan={2} className="px-3 py-4 text-center text-gray-400">No data</td></tr>}
            </tbody>
          </table>
        </div>
      )}

      {tab === 'balances' && (
        <div className="bg-white p-4 rounded shadow">
          <button onClick={loadBalances} className="bg-blue-600 text-white px-4 py-1 rounded text-sm mb-4">Load Balances</button>
          <table className="w-full text-sm">
            <thead><tr className="border-b text-left text-gray-500">
              <th className="px-3 py-2">Account</th><th className="px-3 py-2">Currency</th><th className="px-3 py-2 text-right">Balance</th><th className="px-3 py-2 text-right">Base Equiv.</th>
            </tr></thead>
            <tbody>
              {balances.map(b => (
                <tr key={b.AccountID} className="border-b">
                  <td className="px-3 py-2">{b.AccountName}</td>
                  <td className="px-3 py-2">{b.Currency}</td>
                  <td className="px-3 py-2 text-right">{fmt(b.BalanceMinor)}</td>
                  <td className="px-3 py-2 text-right">{fmt(b.BalanceBaseMinor)}</td>
                </tr>
              ))}
              {balances.length === 0 && <tr><td colSpan={4} className="px-3 py-4 text-center text-gray-400">No data</td></tr>}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
