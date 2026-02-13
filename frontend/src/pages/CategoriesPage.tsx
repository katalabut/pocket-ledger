import { useState, useEffect, useCallback } from 'react'
import { Category, listCategories, createCategory, updateCategory, deleteCategory } from '../api'

export default function CategoriesPage() {
  const [categories, setCategories] = useState<Category[]>([])
  const [showForm, setShowForm] = useState(false)
  const [name, setName] = useState('')
  const [parentId, setParentId] = useState('')
  const [editId, setEditId] = useState<string | null>(null)

  const load = useCallback(async () => {
    const data = await listCategories()
    setCategories(data || [])
  }, [])

  useEffect(() => { load() }, [load])

  const save = async () => {
    const data: { name: string; parent_id?: string } = { name }
    if (parentId) data.parent_id = parentId
    if (editId) {
      await updateCategory(editId, data)
    } else {
      await createCategory(data)
    }
    setShowForm(false); setEditId(null); setName(''); setParentId('')
    load()
  }

  const edit = (c: Category) => {
    setEditId(c.ID); setName(c.Name); setParentId(c.ParentID || '')
    setShowForm(true)
  }

  const remove = async (id: string) => {
    if (confirm('Delete this category?')) { await deleteCategory(id); load() }
  }

  const parentName = (id: string | null) => {
    if (!id) return '-'
    return categories.find(c => c.ID === id)?.Name || id
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-bold">Categories</h2>
        <button onClick={() => { setShowForm(true); setEditId(null); setName(''); setParentId('') }}
          className="bg-blue-600 text-white px-4 py-2 rounded text-sm">Add Category</button>
      </div>

      {showForm && (
        <div className="bg-white p-4 rounded shadow mb-4">
          <div className="grid grid-cols-2 gap-2 mb-2">
            <input value={name} onChange={e => setName(e.target.value)} placeholder="Name" className="border rounded px-2 py-1" />
            <select value={parentId} onChange={e => setParentId(e.target.value)} className="border rounded px-2 py-1">
              <option value="">No parent</option>
              {categories.filter(c => c.ID !== editId).map(c => (
                <option key={c.ID} value={c.ID}>{c.Name}</option>
              ))}
            </select>
          </div>
          <div className="flex gap-2">
            <button onClick={save} className="bg-green-600 text-white px-4 py-1 rounded text-sm">Save</button>
            <button onClick={() => setShowForm(false)} className="text-gray-600 px-4 py-1 text-sm">Cancel</button>
          </div>
        </div>
      )}

      <table className="w-full bg-white rounded shadow">
        <thead><tr className="border-b text-left text-sm text-gray-500">
          <th className="px-4 py-2">Name</th><th className="px-4 py-2">Parent</th><th className="px-4 py-2">Actions</th>
        </tr></thead>
        <tbody>
          {categories.map(c => (
            <tr key={c.ID} className="border-b text-sm">
              <td className="px-4 py-2">{c.Name}</td>
              <td className="px-4 py-2">{parentName(c.ParentID)}</td>
              <td className="px-4 py-2 flex gap-2">
                <button onClick={() => edit(c)} className="text-blue-600 text-xs">Edit</button>
                <button onClick={() => remove(c.ID)} className="text-red-600 text-xs">Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
