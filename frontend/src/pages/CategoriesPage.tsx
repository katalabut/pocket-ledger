import { useState, useEffect, useCallback } from 'react'
import { Category, listCategories, createCategory, updateCategory, deleteCategory } from '../api'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Dialog } from '../components/ui/dialog'
import { Input } from '../components/ui/input'
import { Select } from '../components/ui/select'
import { Table, TBody, TD, TH, THead, TR } from '../components/ui/table'

export default function CategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([])
  const [showForm, setShowForm] = useState(false)
  const [name, setName] = useState(''); const [parentId, setParentId] = useState(''); const [editId, setEditId] = useState<string | null>(null)
  const load = useCallback(async () => setCategories((await listCategories()) || []), [])
  useEffect(() => { load() }, [load])
  const reset = () => { setShowForm(false); setEditId(null); setName(''); setParentId('') }
  const save = async () => { const data: { name: string; parent_id?: string } = { name }; if (parentId) data.parent_id = parentId; if (editId) await updateCategory(editId, data); else await createCategory(data); reset(); load() }
  const edit = (c: Category) => { setEditId(c.ID); setName(c.Name); setParentId(c.ParentID || ''); setShowForm(true) }
  const remove = async (id: string) => { if (confirm('Delete this category?')) { await deleteCategory(id); load() } }
  const parentName = (id: string | null) => !id ? '-' : categories.find(c => c.ID === id)?.Name || id

  return <div className="space-y-4"><div className="flex items-center justify-between"><h2 className="text-2xl font-semibold">Categories</h2><Button onClick={() => setShowForm(true)}>Add category</Button></div>
    <Card><CardHeader><CardTitle className="text-base">Category tree</CardTitle></CardHeader><CardContent className="p-0 overflow-x-auto"><Table><THead><TR><TH>Name</TH><TH>Parent</TH><TH>Actions</TH></TR></THead><TBody>{categories.map(c => <TR key={c.ID}><TD>{c.Name}</TD><TD>{parentName(c.ParentID)}</TD><TD><div className="flex gap-1"><Button size="sm" variant="ghost" onClick={() => edit(c)}>Edit</Button><Button size="sm" variant="destructive" onClick={() => remove(c.ID)}>Delete</Button></div></TD></TR>)}</TBody></Table></CardContent></Card>
    <Dialog open={showForm} onClose={reset} title={editId ? 'Edit category' : 'New category'} footer={<><Button onClick={save}>Save</Button><Button variant="outline" onClick={reset}>Cancel</Button></>}><div className="grid grid-cols-1 gap-2 md:grid-cols-2"><Input value={name} onChange={e => setName(e.target.value)} placeholder="Name" /><Select value={parentId} onChange={e => setParentId(e.target.value)}><option value="">No parent</option>{categories.filter(c => c.ID !== editId).map(c => <option key={c.ID} value={c.ID}>{c.Name}</option>)}</Select></div></Dialog>
  </div>
}
