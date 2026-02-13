import { useState } from 'react'
import { hasToken } from './api'
import Login from './pages/Login'
import Shell from './Shell'

export default function App() {
  const [authed, setAuthed] = useState(hasToken())
  if (!authed) return <Login onLogin={() => setAuthed(true)} />
  return <Shell />
}
