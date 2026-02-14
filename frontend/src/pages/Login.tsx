import { useState } from 'react'
import { authRequestCode, authConfirmCode, setToken } from '../api'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { Input } from '../components/ui/input'
import { Button } from '../components/ui/button'
import { Alert } from '../components/ui/alert'

export default function Login({ onLogin }: { onLogin: () => void }) {
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [step, setStep] = useState<'email' | 'code'>('email')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const requestCode = async () => {
    setLoading(true); setError('')
    try { await authRequestCode(email); setStep('code') } catch (e: unknown) { setError((e as Error).message) } finally { setLoading(false) }
  }

  const confirmCode = async () => {
    setLoading(true); setError('')
    try { const res = await authConfirmCode(email, code); setToken(res.token); onLogin() } catch (e: unknown) { setError((e as Error).message) } finally { setLoading(false) }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-100 p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-2xl">Pocket Ledger</CardTitle>
          <CardDescription>Passwordless sign in via email code</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && <Alert>{error}</Alert>}
          {step === 'email' ? (
            <>
              <Input type="email" value={email} onChange={e => setEmail(e.target.value)} placeholder="you@example.com" />
              <Button onClick={requestCode} disabled={loading || !email} className="w-full">{loading ? 'Sending…' : 'Send code'}</Button>
            </>
          ) : (
            <>
              <p className="text-sm text-slate-500">Code sent to {email}</p>
              <Input type="text" value={code} onChange={e => setCode(e.target.value)} placeholder="6-digit code" maxLength={6} />
              <Button onClick={confirmCode} disabled={loading || code.length < 6} className="w-full">{loading ? 'Verifying…' : 'Login'}</Button>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
