import { useState, useEffect, useCallback } from 'react'
import { Transaction, Account, Category, listTransactions, createTransaction, updateTransaction, deleteTransaction, listAccounts, listCategories, getSplits, replaceSplits } from '../api'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Dialog } from '../components/ui/dialog'
import { Input } from '../components/ui/input'
import { Select } from '../components/ui/select'
import { Table, TBody, TD, TH, THead, TR } from '../components/ui/table'
import { Badge } from '../components/ui/badge'

export default function TransactionsPage() {
  const [txs, setTxs] = useState<Transaction[]>([]); const [total, setTotal] = useState(0)
  const [accounts, setAccounts] = useState<Account[]>([]); const [categories, setCategories] = useState<Category[]>([])
  const [offset, setOffset] = useState(0); const limit = 50
  const [fAccount, setFAccount] = useState(''); const [fCategory, setFCategory] = useState(''); const [fFrom, setFFrom] = useState(''); const [fTo, setFTo] = useState(''); const [fQuery, setFQuery] = useState('')
  const [editTx, setEditTx] = useState<Transaction | null>(null); const [editDesc, setEditDesc] = useState(''); const [editCat, setEditCat] = useState(''); const [editDate, setEditDate] = useState(''); const [editAmount, setEditAmount] = useState(''); const [editAccount, setEditAccount] = useState('')
  const [splitTx, setSplitTx] = useState<Transaction | null>(null); const [splits, setSplits] = useState<{ category_id: string; amount_minor: number }[]>([])
  const [showCreate, setShowCreate] = useState(false); const [newDesc, setNewDesc] = useState(''); const [newAmount, setNewAmount] = useState(''); const [newDate, setNewDate] = useState(new Date().toISOString().slice(0, 10)); const [newAccount, setNewAccount] = useState(''); const [newCurrency, setNewCurrency] = useState('EUR'); const [newCategory, setNewCategory] = useState('')

  const load = useCallback(async () => { const p: Record<string, string> = { limit: String(limit), offset: String(offset) }; if (fAccount) p.account_id = fAccount; if (fCategory) p.category_id = fCategory; if (fFrom) p.from = fFrom + 'T00:00:00Z'; if (fTo) p.to = fTo + 'T23:59:59Z'; if (fQuery) p.q = fQuery; const d = await listTransactions(p); setTxs(d.items || []); setTotal(d.total) }, [offset, fAccount, fCategory, fFrom, fTo, fQuery])
  const loadMeta = useCallback(async () => { const [a, c] = await Promise.all([listAccounts(), listCategories()]); setAccounts(a || []); setCategories(c || []) }, [])
  useEffect(() => { loadMeta() }, [loadMeta]); useEffect(() => { load() }, [load])

  const accountName = (id: string) => accounts.find(a => a.ID === id)?.Name || id.slice(0, 8)
  const categoryName = (id: string | null) => !id ? '-' : categories.find(c => c.ID === id)?.Name || id.slice(0, 8)
  const fmt = (minor: number) => (minor / 100).toFixed(2)

  const openEdit = (tx: Transaction) => { setEditTx(tx); setEditDesc(tx.Description); setEditCat(tx.CategoryID || ''); setEditDate(tx.OccurredAt.slice(0, 10)); setEditAmount(String(tx.AmountMinor)); setEditAccount(tx.AccountID) }
  const saveEdit = async () => { if (!editTx) return; await updateTransaction(editTx.ID, { description: editDesc, category_id: editCat || undefined, occurred_at: editDate + 'T00:00:00Z', amount_minor: parseInt(editAmount), account_id: editAccount }); setEditTx(null); load() }
  const openSplits = async (tx: Transaction) => { setSplitTx(tx); const s = await getSplits(tx.ID); setSplits((s || []).map(sp => ({ category_id: sp.CategoryID, amount_minor: sp.AmountMinor }))) }
  const saveSplits = async () => { if (!splitTx) return; await replaceSplits(splitTx.ID, splits); setSplitTx(null); load() }
  const addSplit = () => setSplits([...splits, { category_id: '', amount_minor: 0 }]); const removeSplit = (i: number) => setSplits(splits.filter((_, idx) => idx !== i)); const updateSplit = (i: number, field: 'category_id' | 'amount_minor', val: string) => { const n = [...splits]; n[i] = field === 'amount_minor' ? { ...n[i], amount_minor: parseInt(val) || 0 } : { ...n[i], category_id: val }; setSplits(n) }
  const splitSum = splits.reduce((s, sp) => s + sp.amount_minor, 0)
  const doCreate = async () => { await createTransaction({ account_id: newAccount, occurred_at: newDate + 'T00:00:00Z', amount_minor: parseInt(newAmount) || 0, currency: newCurrency, description: newDesc, category_id: newCategory || undefined }); setShowCreate(false); load() }
  const removeTx = async (id: string) => { if (confirm('Delete transaction?')) { await deleteTransaction(id); load() } }

  return (<div className="space-y-4">
    <div className="flex flex-wrap items-center justify-between gap-2"><h2 className="text-2xl font-semibold">Transactions <span className="text-slate-500">{total}</span></h2><Button onClick={() => setShowCreate(true)}>Add transaction</Button></div>
    <Card><CardHeader><CardTitle className="text-base">Filters</CardTitle></CardHeader><CardContent className="grid gap-2 sm:grid-cols-2 lg:grid-cols-5">
      <Select value={fAccount} onChange={e => { setFAccount(e.target.value); setOffset(0) }}><option value="">All accounts</option>{accounts.map(a => <option key={a.ID} value={a.ID}>{a.Name}</option>)}</Select>
      <Select value={fCategory} onChange={e => { setFCategory(e.target.value); setOffset(0) }}><option value="">All categories</option>{categories.map(c => <option key={c.ID} value={c.ID}>{c.Name}</option>)}</Select>
      <Input type="date" value={fFrom} onChange={e => { setFFrom(e.target.value); setOffset(0) }} />
      <Input type="date" value={fTo} onChange={e => { setFTo(e.target.value); setOffset(0) }} />
      <Input value={fQuery} onChange={e => { setFQuery(e.target.value); setOffset(0) }} placeholder="Search description" />
    </CardContent></Card>

    <Card><CardContent className="p-0 overflow-x-auto"><Table><THead><TR><TH>Date</TH><TH>Description</TH><TH>Account</TH><TH>Category</TH><TH className="text-right">Amount</TH><TH>Actions</TH></TR></THead><TBody>
      {txs.map(tx => <TR key={tx.ID}><TD>{tx.OccurredAt.slice(0,10)}</TD><TD>{tx.Description}</TD><TD>{accountName(tx.AccountID)}</TD><TD><Badge>{categoryName(tx.CategoryID)}</Badge></TD><TD className={`text-right font-medium ${tx.AmountMinor < 0 ? 'text-rose-600' : 'text-emerald-600'}`}>{fmt(tx.AmountMinor)} {tx.Currency}</TD><TD><div className="flex gap-1"><Button size="sm" variant="ghost" onClick={() => openEdit(tx)}>Edit</Button><Button size="sm" variant="ghost" onClick={() => openSplits(tx)}>Splits</Button><Button size="sm" variant="destructive" onClick={() => removeTx(tx.ID)}>Del</Button></div></TD></TR>)}
      {txs.length === 0 && <TR><TD colSpan={6} className="py-6 text-center text-slate-400">No transactions</TD></TR>}
    </TBody></Table></CardContent></Card>

    {total > limit && <div className="flex items-center justify-center gap-3"><Button variant="outline" onClick={() => setOffset(Math.max(0, offset-limit))} disabled={offset===0}>Previous</Button><span className="text-sm text-slate-500">{offset+1}-{Math.min(offset+limit,total)} of {total}</span><Button variant="outline" onClick={() => setOffset(offset+limit)} disabled={offset+limit>=total}>Next</Button></div>}

    <Dialog open={showCreate} onClose={() => setShowCreate(false)} title="New Transaction" footer={<><Button onClick={doCreate}>Create</Button><Button variant="outline" onClick={() => setShowCreate(false)}>Cancel</Button></>}>
      <div className="grid grid-cols-1 gap-2 md:grid-cols-2"><Select value={newAccount} onChange={e => setNewAccount(e.target.value)}><option value="">Account...</option>{accounts.map(a => <option key={a.ID} value={a.ID}>{a.Name} ({a.Currency})</option>)}</Select><Input type="date" value={newDate} onChange={e => setNewDate(e.target.value)} /><Input value={newAmount} onChange={e => setNewAmount(e.target.value)} placeholder="Amount (minor units)" /><Input value={newCurrency} onChange={e => setNewCurrency(e.target.value)} placeholder="Currency" maxLength={3} /><Input className="md:col-span-2" value={newDesc} onChange={e => setNewDesc(e.target.value)} placeholder="Description" /><Select className="md:col-span-2" value={newCategory} onChange={e => setNewCategory(e.target.value)}><option value="">No category</option>{categories.map(c => <option key={c.ID} value={c.ID}>{c.Name}</option>)}</Select></div>
    </Dialog>

    <Dialog open={!!editTx} onClose={() => setEditTx(null)} title="Edit Transaction" footer={<><Button onClick={saveEdit}>Save</Button><Button variant="outline" onClick={() => setEditTx(null)}>Cancel</Button></>}>
      <div className="grid grid-cols-1 gap-2 md:grid-cols-2"><Select value={editAccount} onChange={e => setEditAccount(e.target.value)}>{accounts.map(a => <option key={a.ID} value={a.ID}>{a.Name}</option>)}</Select><Input type="date" value={editDate} onChange={e => setEditDate(e.target.value)} /><Input value={editAmount} onChange={e => setEditAmount(e.target.value)} /><Input value={editDesc} onChange={e => setEditDesc(e.target.value)} /><Select className="md:col-span-2" value={editCat} onChange={e => setEditCat(e.target.value)}><option value="">No category</option>{categories.map(c => <option key={c.ID} value={c.ID}>{c.Name}</option>)}</Select></div>
    </Dialog>

    <Dialog open={!!splitTx} onClose={() => setSplitTx(null)} title={`Splits${splitTx ? `: ${splitTx.Description}` : ''}`} footer={<><Button onClick={saveSplits} disabled={splits.length > 0 && splitTx ? splitSum !== splitTx.AmountMinor : false}>Save splits</Button><Button variant="outline" onClick={() => setSplitTx(null)}>Cancel</Button></>} width="max-w-3xl">
      {splitTx && <p className="mb-3 text-sm text-slate-500">Amount: {fmt(splitTx.AmountMinor)} {splitTx.Currency}</p>}
      {splits.map((sp, i) => <div key={i} className="mb-2 flex gap-2"><Select value={sp.category_id} onChange={e => updateSplit(i, 'category_id', e.target.value)} className="flex-1"><option value="">Category...</option>{categories.map(c => <option key={c.ID} value={c.ID}>{c.Name}</option>)}</Select><Input value={sp.amount_minor} onChange={e => updateSplit(i, 'amount_minor', e.target.value)} className="w-40" /><Button variant="ghost" onClick={() => removeSplit(i)}>Remove</Button></div>)}
      <div className="flex items-center justify-between"><Button variant="outline" onClick={addSplit}>Add split</Button>{splitTx && <span className={`text-sm ${splitSum === splitTx.AmountMinor ? 'text-emerald-600' : 'text-rose-600'}`}>Sum {fmt(splitSum)} / {fmt(splitTx.AmountMinor)}</span>}</div>
    </Dialog>
  </div>)
}
