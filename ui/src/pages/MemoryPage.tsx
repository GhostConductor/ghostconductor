import { useEffect, useState } from 'react'
import { api } from '../api'

export default function MemoryPage() {
  const [content, setContent] = useState('')
  const [original, setOriginal] = useState('')
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)

  useEffect(() => {
    api.getMemory().then(m => { setContent(m); setOriginal(m) })
  }, [])

  const save = async () => {
    setSaving(true)
    try {
      await api.putMemory(content)
      setOriginal(content)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } finally {
      setSaving(false)
    }
  }

  const dirty = content !== original

  return (
    <div className="max-w-5xl mx-auto p-6 space-y-4">
      <h1 className="font-semibold text-white">Project Memory</h1>
      <textarea
        value={content}
        onChange={e => setContent(e.target.value)}
        className="w-full bg-gray-900 border border-gray-800 rounded p-4 text-sm font-mono text-gray-200 resize-none focus:outline-none focus:border-gray-600"
        rows={30}
        spellCheck={false}
      />
      <div className="flex justify-end">
        <button
          onClick={save}
          disabled={saving || !dirty}
          className="text-sm px-4 py-1.5 rounded bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white"
        >
          {saved ? 'Saved' : saving ? 'Saving...' : 'Save'}
        </button>
      </div>
    </div>
  )
}
