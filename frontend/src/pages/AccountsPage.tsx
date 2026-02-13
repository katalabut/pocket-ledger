import { useState, useEffect, useCallback } from 'react'
import { Account, listAccounts, createAccount, updateAccount, deleteAccount } from '../api'

export default function AccountsPage() {
  const [accounts, setAccounts] = useState<Account[]>([])
  const [showForm, setShowForm] = useState(false)
  const [name, setName] = useState('')
  const [currency, setCurrency] = useState('EUR')
  const [type, setType] = useState('card')
  const [balance, setBalance] = useState('0')
  const [editId, setEditId] = useState<string | null>(null)

  const load = useCallback(async () => {
    const data = await listAccounts()
    setAccounts(data || [])
  }, [])

  useEffect(() => { load() }, [load])

  const save = async () => {
    if (editId) {
      await updateAccount(editId, { name, currency, type, initial_balance_minor: parseInt(balance) || 0 })
    } else {
      await createAccount({ name, currency, type, initial_balance_minor: parseInt(balance) || 0 })
    }
    setShowForm(false); setEditId(null); setName(''); setCurrency('EUR'); setType('card'); setBalance('0')
    load()
  }

  const edit = (a: Account) => {
    setEditId(a.ID); setName(a.Name); setCurrency(a.Currency); setType(a.Type); setBalance(String(a.InitialBalanceMinor))
    setShowForm(true)
  }

  const remove = async (id: string) => {
    if (confirm('Delete this account?')) { await deleteAccount(id); load() }
  }

  const fmt = (minor: number) => (minor / 100).toFixed(2)

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">Accounts</h2>
        <button onClick={() => { setShowForm(true); setEditId(null); setName(''); setCurrency('EUR'); setType('card'); setBalance('0') }}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm">Add Account</button>
      </div>

      {showForm && (
        <div className="bg-white p-4 rounded shadow mb-4">
          <div className="grid grid-cols-4 gap-2 mb-2">
            <input value={name} onChange={e => setName(e.target.value)} placeholder="Name" className="border rounded px-2 py-1" />
            <input value={currency} onChange={e => setCurrency(e.target.value)} placeholder="Currency" className="border rounded px-2 py-1" maxLength={3} />
            <select value={type} onChange={e => setType(e.target.value)} className="border rounded px-2 py-1">
              <option value="card">Card</option>
              <option value="cash">Cash</option>
              <option value="savings">Savings</option>
              <option value="other">Other</option>
            </select>
            <input value={balance} onChange={e => setBalance(e.target.value)} placeholder="Initial balance (minor)" className="border rounded px-2 py-1" />
          </div>
          <div className="flex gap-2">
            <button onClick={save} className="bg-green-600 text-white px-4 py-1 rounded text-sm">Save</button>
            <button onClick={() => setShowForm(false)} className="text-gray-600 px-4 py-1 text-sm">Cancel</button>
          </div>
        </div>
      )}

      <table className="w-full bg-white rounded shadow">
        <thead><tr className="border-b text-left text-sm text-gray-500">
          <th className="px-4 py-2">Name</th><th className="px-4 py-2">Currency</th><th className="px-4 py-2">Type</th><th className="px-4 py-2">Initial Balance</th><th className="px-4 py-2">Actions</th>
        </tr></thead>
        <tbody>
          {accounts.map(a => (
            <tr key={a.ID} className="border-b text-sm">
              <td className="px-4 py-2">{a.Name}</td>
              <td className="px-4 py-2">{a.Currency}</td>
              <td className="px-4 py-2">{a.Type}</td>
              <td className="px-4 py-2">{fmt(a.InitialBalanceMinor)}</td>
              <td className="px-4 py-2 flex gap-2">
                <button onClick={() => edit(a)} className="text-blue-600 text-xs">Edit</button>
                <button onClick={() => remove(a.ID)} className="text-red-600 text-xs">Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
