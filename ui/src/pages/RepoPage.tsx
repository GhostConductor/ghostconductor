import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'
import { RepoStatus } from '../types'

export default function RepoPage() {
  const [repos, setRepos] = useState<RepoStatus[]>([])
  const [showAdd, setShowAdd] = useState(false)
  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const [branch, setBranch] = useState('main')
  const [adding, setAdding] = useState(false)
  const [deleting, setDeleting] = useState<string | null>(null)
  const [error, setError] = useState('')
  const navigate = useNavigate()

  const load = () => api.getRepos().then(setRepos)

  useEffect(() => { load() }, [])

  const addRepo = async () => {
    if (!name.trim() || !url.trim()) return
    setAdding(true)
    setError('')
    try {
      await api.createRepo({ name: name.trim(), url: url.trim(), branch: branch.trim() || 'main' })
      setName('')
      setUrl('')
      setBranch('main')
      setShowAdd(false)
      await load()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to add repo')
    } finally {
      setAdding(false)
    }
  }

  const deleteRepo = async (id: string) => {
    if (!confirm('Delete this repo? This cannot be undone.')) return
    setDeleting(id)
    try {
      await api.deleteRepo(id)
      await load()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to delete repo')
    } finally {
      setDeleting(null)
    }
  }

  return (
    <div className="max-w-5xl mx-auto p-6 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="font-semibold text-white">Repos</h1>
        <button
          onClick={() => setShowAdd(s => !s)}
          className="text-sm px-3 py-1.5 rounded bg-indigo-600 hover:bg-indigo-500 text-white"
        >
          + Add Repo
        </button>
      </div>

      {error && <p className="text-sm text-red-400">{error}</p>}

      {showAdd && (
        <div className="p-4 rounded bg-gray-900 border border-gray-700 space-y-3">
          <h2 className="text-sm font-medium text-white">Add Repo</h2>
          <div className="space-y-2">
            <input
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="Name (e.g. gc-cman)"
              className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white"
            />
            <input
              value={url}
              onChange={e => setUrl(e.target.value)}
              placeholder="URL (e.g. https://github.com/org/repo)"
              className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white"
            />
            <input
              value={branch}
              onChange={e => setBranch(e.target.value)}
              placeholder="Branch (default: main)"
              className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white"
            />
          </div>
          <div className="flex gap-2 justify-end">
            <button
              onClick={() => setShowAdd(false)}
              className="text-sm px-3 py-1.5 rounded border border-gray-700 text-gray-400 hover:text-white"
            >
              Cancel
            </button>
            <button
              onClick={addRepo}
              disabled={adding || !name.trim() || !url.trim()}
              className="text-sm px-3 py-1.5 rounded bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white"
            >
              {adding ? 'Adding...' : 'Add'}
            </button>
          </div>
        </div>
      )}

      {repos.length === 0
        ? <p className="text-sm text-gray-500">No repos configured.</p>
        : (
          <div className="space-y-3">
            {repos.map(repo => (
              <div
                key={repo.id}
                onClick={() => navigate(`/repos/${repo.id}`)}
                className="flex items-center justify-between p-4 rounded bg-gray-900 border border-gray-800 cursor-pointer hover:border-gray-600"
              >
                <div className="space-y-0.5">
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-medium text-white">{repo.name}</span>
                    {repo.token_set
                      ? <span className="text-xs text-green-400">● token set</span>
                      : <span className="text-xs text-red-400">● no token</span>
                    }
                  </div>
                  <div className="text-xs text-gray-500 font-mono">{repo.url}</div>
                  <div className="text-xs text-gray-600">{repo.branch}</div>
                </div>
                <button
                  onClick={e => { e.stopPropagation(); deleteRepo(repo.id) }}
                  disabled={deleting === repo.id}
                  className="text-xs px-2 py-1 rounded border border-red-900 text-red-500 hover:bg-red-900/30 disabled:opacity-30"
                >
                  {deleting === repo.id ? '…' : 'Delete'}
                </button>
              </div>
            ))}
          </div>
        )
      }
    </div>
  )
}
