import { useState, useEffect, useCallback } from 'react'
import {
  Transaction, Account, Category,
  listTransactions, createTransaction, updateTransaction, deleteTransaction,
  listAccounts, listCategories, getSplits, replaceSplits, Split,
} from '../api'

export default function TransactionsPage() {
  const [txs, setTxs] = useState<Transaction[]>([])
  const [total, setTotal] = useState(0)
  const [accounts, setAccounts] = useState<Account[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [offset, setOffset] = useState(0)
  const limit = 50

  // Filters
  const [fAccount, setFAccount] = useState('')
  const [fCategory, setFCategory] = useState('')
  const [fFrom, setFFrom] = useState('')
  const [fTo, setFTo] = useState('')
  const [fQuery, setFQuery] = useState('')

  // Edit modal
  const [editTx, setEditTx] = useState<Transaction | null>(null)
  const [editDesc, setEditDesc] = useState('')
  const [editCat, setEditCat] = useState('')
  const [editDate, setEditDate] = useState('')
  const [editAmount, setEditAmount] = useState('')
  const [editAccount, setEditAccount] = useState('')

  // Splits modal
  const [splitTx, setSplitTx] = useState<Transaction | null>(null)
  const [splits, setSplits] = useState<{ category_id: string; amount_minor: number }[]>([])
  const [existingSplits, setExistingSplits] = useState<Split[]>([])

  // Create modal
  const [showCreate, setShowCreate] = useState(false)
  const [newDesc, setNewDesc] = useState('')
  const [newAmount, setNewAmount] = useState('')
  const [newDate, setNewDate] = useState(new Date().toISOString().slice(0, 10))
  const [newAccount, setNewAccount] = useState('')
  const [newCurrency, setNewCurrency] = useState('EUR')
  const [newCategory, setNewCategory] = useState('')

  const load = useCallback(async () => {
    const params: Record<string, string> = { limit: String(limit), offset: String(offset) }
    if (fAccount) params.account_id = fAccount
    if (fCategory) params.category_id = fCategory
    if (fFrom) params.from = fFrom + 'T00:00:00Z'
    if (fTo) params.to = fTo + 'T23:59:59Z'
    if (fQuery) params.q = fQuery
    const data = await listTransactions(params)
    setTxs(data.items || [])
    setTotal(data.total)
  }, [offset, fAccount, fCategory, fFrom, fTo, fQuery])

  const loadMeta = useCallback(async () => {
    const [a, c] = await Promise.all([listAccounts(), listCategories()])
    setAccounts(a || [])
    setCategories(c || [])
  }, [])

  useEffect(() => { loadMeta() }, [loadMeta])
  useEffect(() => { load() }, [load])

  const accountName = (id: string) => accounts.find(a => a.ID === id)?.Name || id.slice(0, 8)
  const categoryName = (id: string | null) => {
    if (!id) return '-'
    return categories.find(c => c.ID === id)?.Name || id.slice(0, 8)
  }
  const fmt = (minor: number) => (minor / 100).toFixed(2)

  const openEdit = (tx: Transaction) => {
    setEditTx(tx)
    setEditDesc(tx.Description)
    setEditCat(tx.CategoryID || '')
    setEditDate(tx.OccurredAt.slice(0, 10))
    setEditAmount(String(tx.AmountMinor))
    setEditAccount(tx.AccountID)
  }

  const saveEdit = async () => {
    if (!editTx) return
    await updateTransaction(editTx.ID, {
      description: editDesc,
      category_id: editCat || undefined,
      occurred_at: editDate + 'T00:00:00Z',
      amount_minor: parseInt(editAmount),
      account_id: editAccount,
    })
    setEditTx(null)
    load()
  }

  const openSplits = async (tx: Transaction) => {
    setSplitTx(tx)
    const s = await getSplits(tx.ID)
    setExistingSplits(s || [])
    setSplits((s || []).map(sp => ({ category_id: sp.CategoryID, amount_minor: sp.AmountMinor })))
  }

  const saveSplits = async () => {
    if (!splitTx) return
    await replaceSplits(splitTx.ID, splits)
    setSplitTx(null)
    load()
  }

  const addSplit = () => setSplits([...splits, { category_id: '', amount_minor: 0 }])
  const removeSplit = (i: number) => setSplits(splits.filter((_, idx) => idx !== i))
  const updateSplit = (i: number, field: 'category_id' | 'amount_minor', val: string) => {
    const next = [...splits]
    if (field === 'amount_minor') next[i] = { ...next[i], amount_minor: parseInt(val) || 0 }
    else next[i] = { ...next[i], category_id: val }
    setSplits(next)
  }
  const splitSum = splits.reduce((s, sp) => s + sp.amount_minor, 0)

  const doCreate = async () => {
    await createTransaction({
      account_id: newAccount,
      occurred_at: newDate + 'T00:00:00Z',
      amount_minor: parseInt(newAmount) || 0,
      currency: newCurrency,
      description: newDesc,
      category_id: newCategory || undefined,
    })
    setShowCreate(false)
    load()
  }

  const removeTx = async (id: string) => {
    if (confirm('Delete transaction?')) { await deleteTransaction(id); load() }
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">Transactions ({total})</h2>
        <button onClick={() => setShowCreate(true)} className="bg-blue-600 text-white px-4 py-2 rounded text-sm">Add Transaction</button>
      </div>

      {/* Filters */}
      <div className="bg-white p-3 rounded shadow mb-4 flex flex-wrap gap-2 items-end">
        <div>
          <label className="text-xs text-gray-500">Account</label>
          <select value={fAccount} onChange={e => { setFAccount(e.target.value); setOffset(0) }} className="block border rounded px-2 py-1 text-sm">
            <option value="">All</option>
            {accounts.map(a => <option key={a.ID} value={a.ID}>{a.Name}</option>)}
          </select>
        </div>
        <div>
          <label className="text-xs text-gray-500">Category</label>
          <select value={fCategory} onChange={e => { setFCategory(e.target.value); setOffset(0) }} className="block border rounded px-2 py-1 text-sm">
            <option value="">All</option>
            {categories.map(c => <option key={c.ID} value={c.ID}>{c.Name}</option>)}
          </select>
        </div>
        <div>
          <label className="text-xs text-gray-500">From</label>
          <input type="date" value={fFrom} onChange={e => { setFFrom(e.target.value); setOffset(0) }} className="block border rounded px-2 py-1 text-sm" />
        </div>
        <div>
          <label className="text-xs text-gray-500">To</label>
          <input type="date" value={fTo} onChange={e => { setFTo(e.target.value); setOffset(0) }} className="block border rounded px-2 py-1 text-sm" />
        </div>
        <div>
          <label className="text-xs text-gray-500">Search</label>
          <input type="text" value={fQuery} onChange={e => { setFQuery(e.target.value); setOffset(0) }} placeholder="description..." className="block border rounded px-2 py-1 text-sm" />
        </div>
      </div>

      {/* Create modal */}
      {showCreate && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-xl w-[500px]">
            <h3 className="font-bold mb-4">New Transaction</h3>
            <div className="grid grid-cols-2 gap-2 mb-3">
              <select value={newAccount} onChange={e => setNewAccount(e.target.value)} className="border rounded px-2 py-1 text-sm">
                <option value="">Account...</option>
                {accounts.map(a => <option key={a.ID} value={a.ID}>{a.Name} ({a.Currency})</option>)}
              </select>
              <input type="date" value={newDate} onChange={e => setNewDate(e.target.value)} className="border rounded px-2 py-1 text-sm" />
              <input value={newAmount} onChange={e => setNewAmount(e.target.value)} placeholder="Amount (minor units)" className="border rounded px-2 py-1 text-sm" />
              <input value={newCurrency} onChange={e => setNewCurrency(e.target.value)} placeholder="Currency" className="border rounded px-2 py-1 text-sm" maxLength={3} />
              <input value={newDesc} onChange={e => setNewDesc(e.target.value)} placeholder="Description" className="border rounded px-2 py-1 text-sm col-span-2" />
              <select value={newCategory} onChange={e => setNewCategory(e.target.value)} className="border rounded px-2 py-1 text-sm col-span-2">
                <option value="">No category</option>
                {categories.map(c => <option key={c.ID} value={c.ID}>{c.Name}</option>)}
              </select>
            </div>
            <div className="flex gap-2">
              <button onClick={doCreate} className="bg-green-600 text-white px-4 py-1 rounded text-sm">Create</button>
              <button onClick={() => setShowCreate(false)} className="text-gray-600 px-4 py-1 text-sm">Cancel</button>
            </div>
          </div>
        </div>
      )}

      {/* Edit modal */}
      {editTx && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-xl w-[500px]">
            <h3 className="font-bold mb-4">Edit Transaction</h3>
            <div className="grid grid-cols-2 gap-2 mb-3">
              <select value={editAccount} onChange={e => setEditAccount(e.target.value)} className="border rounded px-2 py-1 text-sm">
                {accounts.map(a => <option key={a.ID} value={a.ID}>{a.Name}</option>)}
              </select>
              <input type="date" value={editDate} onChange={e => setEditDate(e.target.value)} className="border rounded px-2 py-1 text-sm" />
              <input value={editAmount} onChange={e => setEditAmount(e.target.value)} className="border rounded px-2 py-1 text-sm" />
              <input value={editDesc} onChange={e => setEditDesc(e.target.value)} className="border rounded px-2 py-1 text-sm" />
              <select value={editCat} onChange={e => setEditCat(e.target.value)} className="border rounded px-2 py-1 text-sm col-span-2">
                <option value="">No category</option>
                {categories.map(c => <option key={c.ID} value={c.ID}>{c.Name}</option>)}
              </select>
            </div>
            <div className="flex gap-2">
              <button onClick={saveEdit} className="bg-green-600 text-white px-4 py-1 rounded text-sm">Save</button>
              <button onClick={() => setEditTx(null)} className="text-gray-600 px-4 py-1 text-sm">Cancel</button>
            </div>
          </div>
        </div>
      )}

      {/* Splits modal */}
      {splitTx && (
        <div className="fixed inset-0 bg-black/30 flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-xl w-[600px]">
            <h3 className="font-bold mb-2">Splits for: {splitTx.Description}</h3>
            <p className="text-sm text-gray-500 mb-4">Transaction amount: {fmt(splitTx.AmountMinor)} {splitTx.Currency}</p>
            {existingSplits.length > 0 && splits.length === 0 && (
              <p className="text-sm text-gray-400 mb-2">Existing splits cleared. Add new ones below.</p>
            )}
            {splits.map((sp, i) => (
              <div key={i} className="flex gap-2 mb-2">
                <select value={sp.category_id} onChange={e => updateSplit(i, 'category_id', e.target.value)} className="border rounded px-2 py-1 text-sm flex-1">
                  <option value="">Category...</option>
                  {categories.map(c => <option key={c.ID} value={c.ID}>{c.Name}</option>)}
                </select>
                <input value={sp.amount_minor} onChange={e => updateSplit(i, 'amount_minor', e.target.value)} className="border rounded px-2 py-1 text-sm w-32" />
                <button onClick={() => removeSplit(i)} className="text-red-600 text-xs">x</button>
              </div>
            ))}
            <div className="flex justify-between items-center mb-4">
              <button onClick={addSplit} className="text-blue-600 text-sm">+ Add split</button>
              <span className={`text-sm ${splitSum === splitTx.AmountMinor ? 'text-green-600' : 'text-red-600'}`}>
                Sum: {fmt(splitSum)} / {fmt(splitTx.AmountMinor)}
              </span>
            </div>
            <div className="flex gap-2">
              <button onClick={saveSplits} disabled={splits.length > 0 && splitSum !== splitTx.AmountMinor} className="bg-green-600 text-white px-4 py-1 rounded text-sm disabled:opacity-50">Save Splits</button>
              <button onClick={() => setSplitTx(null)} className="text-gray-600 px-4 py-1 text-sm">Cancel</button>
            </div>
          </div>
        </div>
      )}

      {/* Table */}
      <table className="w-full bg-white rounded shadow text-sm">
        <thead><tr className="border-b text-left text-gray-500">
          <th className="px-3 py-2">Date</th>
          <th className="px-3 py-2">Description</th>
          <th className="px-3 py-2">Account</th>
          <th className="px-3 py-2">Category</th>
          <th className="px-3 py-2 text-right">Amount</th>
          <th className="px-3 py-2">Actions</th>
        </tr></thead>
        <tbody>
          {txs.map(tx => (
            <tr key={tx.ID} className="border-b hover:bg-gray-50">
              <td className="px-3 py-2">{tx.OccurredAt.slice(0, 10)}</td>
              <td className="px-3 py-2">{tx.Description}</td>
              <td className="px-3 py-2">{accountName(tx.AccountID)}</td>
              <td className="px-3 py-2">{categoryName(tx.CategoryID)}</td>
              <td className={`px-3 py-2 text-right ${tx.AmountMinor < 0 ? 'text-red-600' : 'text-green-600'}`}>
                {fmt(tx.AmountMinor)} {tx.Currency}
              </td>
              <td className="px-3 py-2 flex gap-2">
                <button onClick={() => openEdit(tx)} className="text-blue-600 text-xs">Edit</button>
                <button onClick={() => openSplits(tx)} className="text-purple-600 text-xs">Splits</button>
                <button onClick={() => removeTx(tx.ID)} className="text-red-600 text-xs">Del</button>
              </td>
            </tr>
          ))}
          {txs.length === 0 && <tr><td colSpan={6} className="px-3 py-4 text-center text-gray-400">No transactions</td></tr>}
        </tbody>
      </table>

      {/* Pagination */}
      {total > limit && (
        <div className="flex justify-center gap-4 mt-4">
          <button onClick={() => setOffset(Math.max(0, offset - limit))} disabled={offset === 0} className="text-sm text-blue-600 disabled:text-gray-300">Previous</button>
          <span className="text-sm text-gray-500">{offset + 1}-{Math.min(offset + limit, total)} of {total}</span>
          <button onClick={() => setOffset(offset + limit)} disabled={offset + limit >= total} className="text-sm text-blue-600 disabled:text-gray-300">Next</button>
        </div>
      )}
    </div>
  )
}
