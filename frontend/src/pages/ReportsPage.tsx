import { useState } from 'react'
import { reportSpending, reportBalances, SpendingRow, AccountBalance, syncFXRates } from '../api'
import { Alert } from '../components/ui/alert'
import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Input } from '../components/ui/input'
import { Table, TBody, TD, TH, THead, TR } from '../components/ui/table'

export default function ReportsPage() {
  const [from, setFrom] = useState(''); const [to, setTo] = useState('')
  const [spending, setSpending] = useState<SpendingRow[]>([]); const [balances, setBalances] = useState<AccountBalance[]>([])
  const [tab, setTab] = useState<'spending' | 'balances'>('spending'); const [error, setError] = useState(''); const [syncing, setSyncing] = useState(false)
  const loadSpending = async () => { setError(''); try { const p: Record<string, string> = {}; if (from) p.from = from + 'T00:00:00Z'; if (to) p.to = to + 'T23:59:59Z'; setSpending((await reportSpending(p)) || []) } catch (e: unknown) { setError((e as Error).message) } }
  const loadBalances = async () => { setError(''); try { setBalances((await reportBalances()) || []) } catch (e: unknown) { setError((e as Error).message) } }
  const doSync = async () => { setSyncing(true); setError(''); try { const r = await syncFXRates(); alert(`Synced ${r.synced} rates`) } catch (e: unknown) { setError((e as Error).message) } finally { setSyncing(false) } }
  const fmt = (m: number) => (m / 100).toFixed(2)

  return <div className="space-y-4"><div className="flex items-center justify-between"><h2 className="text-2xl font-semibold">Reports</h2><Button onClick={doSync} disabled={syncing}>{syncing ? 'Syncing...' : 'Sync FX Rates'}</Button></div>{error && <Alert>{error}</Alert>}
    <div className="flex gap-2"><Button variant={tab === 'spending' ? 'default' : 'outline'} onClick={() => setTab('spending')}>Spending</Button><Button variant={tab === 'balances' ? 'default' : 'outline'} onClick={() => setTab('balances')}>Balances</Button></div>
    {tab === 'spending' && <Card><CardHeader><CardTitle className="text-base">Spending by Category</CardTitle></CardHeader><CardContent><div className="mb-3 flex flex-wrap items-end gap-2"><Input type="date" value={from} onChange={e => setFrom(e.target.value)} className="w-44" /><Input type="date" value={to} onChange={e => setTo(e.target.value)} className="w-44" /><Button onClick={loadSpending}>Load</Button></div><Table><THead><TR><TH>Category</TH><TH className="text-right">Total (base)</TH></TR></THead><TBody>{spending.map(s => <TR key={s.CategoryID}><TD>{s.CategoryName}</TD><TD className="text-right">{fmt(s.TotalMinor)}</TD></TR>)}{spending.length === 0 && <TR><TD colSpan={2} className="text-center text-slate-400">No data</TD></TR>}</TBody></Table></CardContent></Card>}
    {tab === 'balances' && <Card><CardHeader className="flex flex-row items-center justify-between"><CardTitle className="text-base">Account Balances</CardTitle><Button onClick={loadBalances}>Load balances</Button></CardHeader><CardContent><Table><THead><TR><TH>Account</TH><TH>Currency</TH><TH className="text-right">Balance</TH><TH className="text-right">Base</TH></TR></THead><TBody>{balances.map(b => <TR key={b.AccountID}><TD>{b.AccountName}</TD><TD><Badge>{b.Currency}</Badge></TD><TD className="text-right">{fmt(b.BalanceMinor)}</TD><TD className="text-right">{fmt(b.BalanceBaseMinor)}</TD></TR>)}{balances.length === 0 && <TR><TD colSpan={4} className="text-center text-slate-400">No data</TD></TR>}</TBody></Table></CardContent></Card>}
  </div>
}
