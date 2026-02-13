import { useState, useEffect, useCallback } from 'react'
import {
  ImportProfile, ImportSession, ImportRow,
  listImportProfiles, createImportProfile,
  uploadImport, previewImport, commitImport,
  listAccounts, Account,
} from '../api'

export default function ImportPage() {
  const [profiles, setProfiles] = useState<ImportProfile[]>([])
  const [accounts, setAccounts] = useState<Account[]>([])
  const [showProfileForm, setShowProfileForm] = useState(false)
  const [pName, setPName] = useState('')
  const [pAccount, setPAccount] = useState('')
  const [pSep, setPSep] = useState(',')
  const [pDateFmt, setPDateFmt] = useState('2006-01-02')
  const [pMapping, setPMapping] = useState('{"date":0,"amount":1,"currency":2,"description":3}')
  const [pFlip, setPFlip] = useState(false)
  const [pSkip, setPSkip] = useState(1)

  const [selectedProfile, setSelectedProfile] = useState('')
  const [file, setFile] = useState<File | null>(null)
  const [session, setSession] = useState<ImportSession | null>(null)
  const [rows, setRows] = useState<ImportRow[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const loadProfiles = useCallback(async () => {
    try {
      const [p, a] = await Promise.all([listImportProfiles(), listAccounts()])
      setProfiles(p || [])
      setAccounts(a || [])
    } catch { /* profiles endpoint may not exist yet */ }
  }, [])

  useEffect(() => { loadProfiles() }, [loadProfiles])

  const saveProfile = async () => {
    try {
      await createImportProfile({
        name: pName, account_id: pAccount, separator: pSep,
        date_format: pDateFmt, column_mapping: JSON.parse(pMapping),
        amount_sign_flip: pFlip, skip_header_rows: pSkip,
      })
      setShowProfileForm(false)
      loadProfiles()
    } catch (e: unknown) { setError((e as Error).message) }
  }

  const doUpload = async () => {
    if (!file || !selectedProfile) return
    setLoading(true); setError('')
    try {
      const sess = await uploadImport(selectedProfile, file)
      setSession(sess)
      const preview = await previewImport(sess.ID)
      setSession(preview.import)
      setRows(preview.rows || [])
    } catch (e: unknown) { setError((e as Error).message) }
    finally { setLoading(false) }
  }

  const doCommit = async () => {
    if (!session) return
    setLoading(true); setError('')
    try {
      const result = await commitImport(session.ID)
      setSession(result)
      const preview = await previewImport(result.ID)
      setRows(preview.rows || [])
    } catch (e: unknown) { setError((e as Error).message) }
    finally { setLoading(false) }
  }

  const profileAccount = (id: string) => accounts.find(a => a.ID === id)?.Name || id.slice(0, 8)

  return (
    <div>
      <h2 className="text-xl font-bold mb-4">CSV Import</h2>

      {error && <p className="text-red-600 text-sm mb-4">{error}</p>}

      {/* Profiles */}
      <div className="bg-white p-4 rounded shadow mb-4">
        <div className="flex justify-between items-center mb-2">
          <h3 className="font-semibold text-sm">Import Profiles</h3>
          <button onClick={() => setShowProfileForm(true)} className="text-blue-600 text-xs">+ New Profile</button>
        </div>
        {showProfileForm && (
          <div className="border p-3 rounded mb-3">
            <div className="grid grid-cols-2 gap-2 mb-2 text-sm">
              <input value={pName} onChange={e => setPName(e.target.value)} placeholder="Profile name" className="border rounded px-2 py-1" />
              <select value={pAccount} onChange={e => setPAccount(e.target.value)} className="border rounded px-2 py-1">
                <option value="">Account...</option>
                {accounts.map(a => <option key={a.ID} value={a.ID}>{a.Name}</option>)}
              </select>
              <input value={pSep} onChange={e => setPSep(e.target.value)} placeholder="Separator" className="border rounded px-2 py-1" />
              <input value={pDateFmt} onChange={e => setPDateFmt(e.target.value)} placeholder="Date format (Go)" className="border rounded px-2 py-1" />
              <textarea value={pMapping} onChange={e => setPMapping(e.target.value)} placeholder="Column mapping JSON" className="border rounded px-2 py-1 col-span-2" rows={2} />
              <label className="flex items-center gap-1 text-xs">
                <input type="checkbox" checked={pFlip} onChange={e => setPFlip(e.target.checked)} /> Flip amount sign
              </label>
              <input type="number" value={pSkip} onChange={e => setPSkip(parseInt(e.target.value) || 0)} placeholder="Skip header rows" className="border rounded px-2 py-1" />
            </div>
            <div className="flex gap-2">
              <button onClick={saveProfile} className="bg-green-600 text-white px-3 py-1 rounded text-sm">Save</button>
              <button onClick={() => setShowProfileForm(false)} className="text-gray-600 text-sm">Cancel</button>
            </div>
          </div>
        )}
        <table className="w-full text-sm">
          <tbody>
            {profiles.map(p => (
              <tr key={p.ID} className="border-b">
                <td className="px-2 py-1">{p.Name}</td>
                <td className="px-2 py-1">{profileAccount(p.AccountID)}</td>
                <td className="px-2 py-1">sep: {p.Separator}</td>
              </tr>
            ))}
            {profiles.length === 0 && <tr><td className="px-2 py-1 text-gray-400">No profiles yet</td></tr>}
          </tbody>
        </table>
      </div>

      {/* Upload */}
      <div className="bg-white p-4 rounded shadow mb-4">
        <h3 className="font-semibold text-sm mb-2">Upload CSV</h3>
        <div className="flex gap-2 items-center">
          <select value={selectedProfile} onChange={e => setSelectedProfile(e.target.value)} className="border rounded px-2 py-1 text-sm">
            <option value="">Select profile...</option>
            {profiles.map(p => <option key={p.ID} value={p.ID}>{p.Name}</option>)}
          </select>
          <input type="file" accept=".csv" onChange={e => setFile(e.target.files?.[0] || null)} className="text-sm" />
          <button onClick={doUpload} disabled={!file || !selectedProfile || loading} className="bg-blue-600 text-white px-4 py-1 rounded text-sm disabled:opacity-50">
            {loading ? 'Processing...' : 'Upload & Preview'}
          </button>
        </div>
      </div>

      {/* Preview */}
      {session && (
        <div className="bg-white p-4 rounded shadow">
          <div className="flex justify-between items-center mb-2">
            <h3 className="font-semibold text-sm">Import: {session.Filename} ({session.Status})</h3>
            {session.Status === 'previewed' && (
              <button onClick={doCommit} disabled={loading} className="bg-green-600 text-white px-4 py-1 rounded text-sm disabled:opacity-50">
                Commit Import
              </button>
            )}
          </div>
          <div className="text-sm text-gray-500 mb-2">
            Total: {session.TotalRows} | Imported: {session.ImportedRows} | Skipped: {session.SkippedRows} | Errors: {session.ErrorRows}
          </div>
          <table className="w-full text-xs">
            <thead><tr className="border-b text-left text-gray-500">
              <th className="px-2 py-1">#</th><th className="px-2 py-1">Status</th><th className="px-2 py-1">Data</th><th className="px-2 py-1">Error</th>
            </tr></thead>
            <tbody>
              {rows.map(r => (
                <tr key={r.ID} className={`border-b ${r.Status === 'error' ? 'bg-red-50' : r.Status === 'skipped' ? 'bg-yellow-50' : ''}`}>
                  <td className="px-2 py-1">{r.RowNumber}</td>
                  <td className="px-2 py-1">{r.Status}</td>
                  <td className="px-2 py-1 truncate max-w-xs">{r.RawData}</td>
                  <td className="px-2 py-1 text-red-600">{r.ErrorMessage}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
