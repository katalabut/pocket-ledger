import { useState, useEffect, useCallback } from 'react'
import { Budget, listBudgets, upsertBudget, listCategories, Category } from '../api'
import { Alert } from '../components/ui/alert'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Dialog } from '../components/ui/dialog'
import { Input } from '../components/ui/input'
import { Select } from '../components/ui/select'
import { Table, TBody, TD, TH, THead, TR } from '../components/ui/table'

function currentMonth() { const d = new Date(); return `${d.getFullYear()}-${String(d.getMonth()+1).padStart(2,'0')}` }

export default function BudgetsPage() {
  const [month, setMonth] = useState(currentMonth()); const [budgets, setBudgets] = useState<Budget[]>([]); const [categories, setCategories] = useState<Category[]>([])
  const [showForm, setShowForm] = useState(false); const [formCat, setFormCat] = useState(''); const [formPlanned, setFormPlanned] = useState(''); const [error, setError] = useState('')
  const loadCats = useCallback(async () => setCategories((await listCategories()) || []), [])
  const load = useCallback(async () => { setError(''); try { setBudgets((await listBudgets(month)) || []) } catch (e: unknown) { setError((e as Error).message) } }, [month])
  useEffect(() => { loadCats() }, [loadCats]); useEffect(() => { load() }, [load])
  const save = async () => { setError(''); try { await upsertBudget({ month, category_id: formCat, planned_minor: parseInt(formPlanned) || 0 }); setShowForm(false); load() } catch (e: unknown) { setError((e as Error).message) } }
  const fmt = (minor: number) => (minor / 100).toFixed(2)
  const catName = (id: string) => categories.find(c => c.ID === id)?.Name || id.slice(0, 8)
  const totalPlanned = budgets.reduce((s,b)=>s+b.PlannedMinor,0); const totalSpent = budgets.reduce((s,b)=>s+(b.SpentMinor||0),0); const totalRemaining = budgets.reduce((s,b)=>s+(b.RemainingMinor||0),0)

  return <div className="space-y-4"><div className="flex flex-wrap items-center justify-between gap-2"><h2 className="text-2xl font-semibold">Budgets</h2><div className="flex items-center gap-2"><Input type="month" value={month} onChange={e => setMonth(e.target.value)} className="w-40" /><Button onClick={() => setShowForm(true)}>Set budget</Button></div></div>{error && <Alert>{error}</Alert>}
    <div className="grid gap-3 md:grid-cols-3"><Card><CardContent className="pt-4"><p className="text-xs text-slate-500">Planned</p><p className="text-xl font-semibold">{fmt(totalPlanned)}</p></CardContent></Card><Card><CardContent className="pt-4"><p className="text-xs text-slate-500">Spent</p><p className="text-xl font-semibold">{fmt(totalSpent)}</p></CardContent></Card><Card><CardContent className="pt-4"><p className="text-xs text-slate-500">Remaining</p><p className={`text-xl font-semibold ${totalRemaining < 0 ? 'text-rose-600' : 'text-emerald-600'}`}>{fmt(totalRemaining)}</p></CardContent></Card></div>
    <Card><CardHeader><CardTitle className="text-base">Monthly categories</CardTitle></CardHeader><CardContent className="p-0 overflow-x-auto"><Table><THead><TR><TH>Category</TH><TH className="text-right">Planned</TH><TH className="text-right">Spent</TH><TH className="text-right">Remaining</TH></TR></THead><TBody>{budgets.map(b => <TR key={b.ID}><TD>{b.CategoryName || catName(b.CategoryID)}</TD><TD className="text-right">{fmt(b.PlannedMinor)}</TD><TD className="text-right">{fmt(b.SpentMinor || 0)}</TD><TD className={`text-right ${(b.RemainingMinor || 0) < 0 ? 'text-rose-600' : 'text-emerald-600'}`}>{fmt(b.RemainingMinor || 0)}</TD></TR>)}{budgets.length === 0 && <TR><TD colSpan={4} className="text-center text-slate-400">No budgets set for {month}</TD></TR>}</TBody></Table></CardContent></Card>
    <Dialog open={showForm} onClose={() => setShowForm(false)} title="Set budget" footer={<><Button onClick={save}>Save</Button><Button variant="outline" onClick={() => setShowForm(false)}>Cancel</Button></>}><div className="grid grid-cols-1 gap-2 md:grid-cols-2"><Select value={formCat} onChange={e => setFormCat(e.target.value)}><option value="">Category...</option>{categories.map(c => <option key={c.ID} value={c.ID}>{c.Name}</option>)}</Select><Input value={formPlanned} onChange={e => setFormPlanned(e.target.value)} placeholder="Planned (minor units)" /></div></Dialog>
  </div>
}
