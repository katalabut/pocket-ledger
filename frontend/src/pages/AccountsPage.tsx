import { useState, useEffect, useCallback } from 'react'
import { Account, listAccounts, createAccount, updateAccount, deleteAccount } from '../api'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '../components/ui/card'
import { Dialog } from '../components/ui/dialog'
import { Input } from '../components/ui/input'
import { Select } from '../components/ui/select'
import { Table, TBody, TD, TH, THead, TR } from '../components/ui/table'

export default function AccountsPage() {
  const [accounts, setAccounts] = useState<Account[]>([])
  const [showForm, setShowForm] = useState(false)
  const [name, setName] = useState(''); const [currency, setCurrency] = useState('EUR'); const [type, setType] = useState('card'); const [balance, setBalance] = useState('0'); const [editId, setEditId] = useState<string | null>(null)
  const load = useCallback(async () => setAccounts((await listAccounts()) || []), [])
  useEffect(() => { load() }, [load])
  const reset = () => { setShowForm(false); setEditId(null); setName(''); setCurrency('EUR'); setType('card'); setBalance('0') }
  const save = async () => { if (editId) await updateAccount(editId, { name, currency, type, initial_balance_minor: parseInt(balance) || 0 }); else await createAccount({ name, currency, type, initial_balance_minor: parseInt(balance) || 0 }); reset(); load() }
  const edit = (a: Account) => { setEditId(a.ID); setName(a.Name); setCurrency(a.Currency); setType(a.Type); setBalance(String(a.InitialBalanceMinor)); setShowForm(true) }
  const remove = async (id: string) => { if (confirm('Delete this account?')) { await deleteAccount(id); load() } }
  const fmt = (minor: number) => (minor / 100).toFixed(2)

  return <div className="space-y-4"><div className="flex flex-wrap items-center justify-between gap-2"><div><h2 className="text-lg font-semibold">Your accounts</h2><p className="text-sm text-[var(--muted-foreground)]">Cards, cash and savings in one place</p></div><Button onClick={() => setShowForm(true)}>Add account</Button></div>
    <Card><CardHeader><CardTitle className="text-base">Account list</CardTitle><CardDescription>Manage account setup without affecting transactions</CardDescription></CardHeader><CardContent className="overflow-x-auto p-0"><Table><THead><TR><TH>Name</TH><TH>Currency</TH><TH>Type</TH><TH className="text-right">Initial balance</TH><TH>Actions</TH></TR></THead><TBody>{accounts.map(a => <TR key={a.ID}><TD>{a.Name}</TD><TD>{a.Currency}</TD><TD>{a.Type}</TD><TD className="text-right">{fmt(a.InitialBalanceMinor)}</TD><TD><div className="flex gap-1"><Button size="sm" variant="ghost" onClick={() => edit(a)}>Edit</Button><Button size="sm" variant="destructive" onClick={() => remove(a.ID)}>Delete</Button></div></TD></TR>)}{accounts.length===0 && <TR><TD colSpan={5} className="py-8 text-center text-[var(--muted-foreground)]">No accounts created yet</TD></TR>}</TBody></Table></CardContent></Card>
    <Dialog open={showForm} onClose={reset} title={editId ? 'Edit account' : 'New account'} footer={<><Button onClick={save}>Save</Button><Button variant="outline" onClick={reset}>Cancel</Button></>}><div className="grid grid-cols-1 gap-2 md:grid-cols-2"><Input value={name} onChange={e => setName(e.target.value)} placeholder="Name" /><Input value={currency} onChange={e => setCurrency(e.target.value)} placeholder="Currency" maxLength={3} /><Select value={type} onChange={e => setType(e.target.value)}><option value="card">Card</option><option value="cash">Cash</option><option value="savings">Savings</option><option value="other">Other</option></Select><Input value={balance} onChange={e => setBalance(e.target.value)} placeholder="Initial balance (minor)" /></div></Dialog>
  </div>
}
