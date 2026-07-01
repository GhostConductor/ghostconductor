import { useEffect, useState } from 'react'
import { api } from '../api'

export default function ContextPage() {
  const [content, setContent] = useState('')
  const [original, setOriginal] = useState('')
  const [saving, setSaving] = useState(false)
  const [saved, setSaved] = useState(false)
  const [templates, setTemplates] = useState<string[]>([])
  const [selectedTemplate, setSelectedTemplate] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    api.getContext().then(c => { setContent(c); setOriginal(c) })
    api.getContextTemplates().then(r => {
      setTemplates(r.templates)
      if (r.templates.length > 0) setSelectedTemplate(r.templates[0])
    })
  }, [])

  const save = async () => {
    setSaving(true)
    try {
      await api.putContext(content)
      setOriginal(content)
      setSaved(true)
      setTimeout(() => setSaved(false), 2000)
    } finally {
      setSaving(false)
    }
  }

  const loadTemplate = async () => {
    if (!selectedTemplate) return
    setLoading(true)
    try {
      await api.loadContextTemplate(selectedTemplate)
      const c = await api.getContext()
      setContent(c)
      setOriginal(c)
    } finally {
      setLoading(false)
    }
  }

  const dirty = content !== original

  return (
    <div className="max-w-5xl mx-auto p-6 space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="font-semibold text-white">Project Context</h1>
        {templates.length > 0 && (
          <div className="flex items-center gap-2">
            <select
              value={selectedTemplate}
              onChange={e => setSelectedTemplate(e.target.value)}
              className="bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white"
            >
              {templates.map(t => (
                <option key={t} value={t}>{t.replace('.md', '')}</option>
              ))}
            </select>
            <button
              onClick={loadTemplate}
              disabled={loading}
              className="text-sm px-3 py-1.5 rounded border border-gray-700 text-gray-400 hover:text-white disabled:opacity-40"
            >
              {loading ? 'Loading...' : 'Load Template'}
            </button>
          </div>
        )}
      </div>
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
