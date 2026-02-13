import { useState, useEffect, useCallback } from 'react'
import { Budget, listBudgets, upsertBudget, listCategories, Category } from '../api'

function currentMonth() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

export default function BudgetsPage() {
  const [month, setMonth] = useState(currentMonth())
  const [budgets, setBudgets] = useState<Budget[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [showForm, setShowForm] = useState(false)
  const [formCat, setFormCat] = useState('')
  const [formPlanned, setFormPlanned] = useState('')
  const [error, setError] = useState('')

  const loadCats = useCallback(async () => {
    const cats = await listCategories()
    setCategories(cats || [])
  }, [])

  const load = useCallback(async () => {
    setError('')
    try {
      const data = await listBudgets(month)
      setBudgets(data || [])
    } catch (e: unknown) { setError((e as Error).message) }
  }, [month])

  useEffect(() => { loadCats() }, [loadCats])
  useEffect(() => { load() }, [load])

  const save = async () => {
    setError('')
    try {
      await upsertBudget({ month, category_id: formCat, planned_minor: parseInt(formPlanned) || 0 })
      setShowForm(false)
      load()
    } catch (e: unknown) { setError((e as Error).message) }
  }

  const fmt = (minor: number) => (minor / 100).toFixed(2)
  const catName = (id: string) => categories.find(c => c.ID === id)?.Name || id.slice(0, 8)

  const totalPlanned = budgets.reduce((s, b) => s + b.PlannedMinor, 0)
  const totalSpent = budgets.reduce((s, b) => s + (b.SpentMinor || 0), 0)
  const totalRemaining = budgets.reduce((s, b) => s + (b.RemainingMinor || 0), 0)

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">Budgets</h2>
        <div className="flex gap-2 items-center">
          <input type="month" value={month} onChange={e => setMonth(e.target.value)} className="border rounded px-2 py-1 text-sm" />
          <button onClick={() => setShowForm(true)} className="bg-blue-600 text-white px-4 py-2 rounded text-sm">Set Budget</button>
        </div>
      </div>

      {error && <p className="text-red-600 text-sm mb-4">{error}</p>}

      {showForm && (
        <div className="bg-white p-4 rounded shadow mb-4">
          <div className="grid grid-cols-2 gap-2 mb-2">
            <select value={formCat} onChange={e => setFormCat(e.target.value)} className="border rounded px-2 py-1 text-sm">
              <option value="">Category...</option>
              {categories.map(c => <option key={c.ID} value={c.ID}>{c.Name}</option>)}
            </select>
            <input value={formPlanned} onChange={e => setFormPlanned(e.target.value)} placeholder="Planned (minor units, base currency)" className="border rounded px-2 py-1 text-sm" />
          </div>
          <div className="flex gap-2">
            <button onClick={save} className="bg-green-600 text-white px-4 py-1 rounded text-sm">Save</button>
            <button onClick={() => setShowForm(false)} className="text-gray-600 px-4 py-1 text-sm">Cancel</button>
          </div>
        </div>
      )}

      <table className="w-full bg-white rounded shadow text-sm">
        <thead><tr className="border-b text-left text-gray-500">
          <th className="px-3 py-2">Category</th>
          <th className="px-3 py-2 text-right">Planned</th>
          <th className="px-3 py-2 text-right">Spent</th>
          <th className="px-3 py-2 text-right">Remaining</th>
        </tr></thead>
        <tbody>
          {budgets.map(b => (
            <tr key={b.ID} className="border-b">
              <td className="px-3 py-2">{b.CategoryName || catName(b.CategoryID)}</td>
              <td className="px-3 py-2 text-right">{fmt(b.PlannedMinor)}</td>
              <td className="px-3 py-2 text-right">{fmt(b.SpentMinor || 0)}</td>
              <td className={`px-3 py-2 text-right ${(b.RemainingMinor || 0) < 0 ? 'text-red-600' : 'text-green-600'}`}>
                {fmt(b.RemainingMinor || 0)}
              </td>
            </tr>
          ))}
          {budgets.length === 0 && <tr><td colSpan={4} className="px-3 py-4 text-center text-gray-400">No budgets set for {month}</td></tr>}
          {budgets.length > 0 && (
            <tr className="border-t-2 font-bold">
              <td className="px-3 py-2">Total</td>
              <td className="px-3 py-2 text-right">{fmt(totalPlanned)}</td>
              <td className="px-3 py-2 text-right">{fmt(totalSpent)}</td>
              <td className={`px-3 py-2 text-right ${totalRemaining < 0 ? 'text-red-600' : 'text-green-600'}`}>{fmt(totalRemaining)}</td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )
}
