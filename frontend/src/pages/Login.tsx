import { useState } from 'react'
import { authRequestCode, authConfirmCode, setToken } from '../api'

export default function Login({ onLogin }: { onLogin: () => void }) {
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [step, setStep] = useState<'email' | 'code'>('email')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const requestCode = async () => {
    setLoading(true); setError('')
    try {
      await authRequestCode(email)
      setStep('code')
    } catch (e: unknown) { setError((e as Error).message) }
    finally { setLoading(false) }
  }

  const confirmCode = async () => {
    setLoading(true); setError('')
    try {
      const res = await authConfirmCode(email, code)
      setToken(res.token)
      onLogin()
    } catch (e: unknown) { setError((e as Error).message) }
    finally { setLoading(false) }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="bg-white p-8 rounded-lg shadow-md w-96">
        <h1 className="text-2xl font-bold mb-6 text-center">Pocket Ledger</h1>
        {error && <p className="text-red-600 text-sm mb-4">{error}</p>}
        {step === 'email' ? (
          <>
            <input type="email" value={email} onChange={e => setEmail(e.target.value)} placeholder="Email" className="w-full border rounded px-3 py-2 mb-4" />
            <button onClick={requestCode} disabled={loading || !email} className="w-full bg-blue-600 text-white py-2 rounded disabled:opacity-50">
              {loading ? 'Sending...' : 'Send Code'}
            </button>
          </>
        ) : (
          <>
            <p className="text-sm text-gray-600 mb-4">Code sent to {email}</p>
            <input type="text" value={code} onChange={e => setCode(e.target.value)} placeholder="6-digit code" className="w-full border rounded px-3 py-2 mb-4" maxLength={6} />
            <button onClick={confirmCode} disabled={loading || code.length < 6} className="w-full bg-blue-600 text-white py-2 rounded disabled:opacity-50">
              {loading ? 'Verifying...' : 'Login'}
            </button>
          </>
        )}
      </div>
    </div>
  )
}
