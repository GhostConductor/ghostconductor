import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { api } from '../api'
import { FileNode, JobStatus, RepoStatus } from '../types'

export default function RepoViewPage() {
  const { repoId, jobId } = useParams<{ repoId: string; jobId?: string }>()
  const navigate = useNavigate()
  const [repo, setRepo] = useState<RepoStatus | null>(null)
  const [jobs, setJobs] = useState<JobStatus[]>([])
  const [tree, setTree] = useState<FileNode | null>(null)
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [fileContent, setFileContent] = useState<string>('')
  const [pulling, setPulling] = useState(false)
  const [error, setError] = useState('')

  // Token management
  const [tokenInput, setTokenInput] = useState('')
  const [savingToken, setSavingToken] = useState(false)
  const [savedToken, setSavedToken] = useState(false)
  const [clearingToken, setClearingToken] = useState(false)

  // Repo edit
  const [editing, setEditing] = useState(false)
  const [editName, setEditName] = useState('')
  const [editUrl, setEditUrl] = useState('')
  const [editBranch, setEditBranch] = useState('')
  const [saving, setSaving] = useState(false)

  // Git User
  const [editGitEmail, setEditGitEmail] = useState('')
  const [editGitName, setEditGitName] = useState('')

  const loadRepo = async () => {
    if (!repoId) return
    const repos = await api.getRepos()
    const found = repos.find(r => r.id === repoId) ?? null
    setRepo(found)
    if (found) {
      setEditName(found.name)
      setEditUrl(found.url)
      setEditBranch(found.branch)
      setEditGitEmail(found.git_email ?? '')
      setEditGitName(found.git_name ?? '')
    }
  }

  const loadTree = async () => {
    if (!repoId) return
    try {
      const t = await api.getRepoTree(repoId, jobId)
      setTree(t)
    } catch {
      setTree(null)
    }
  }

  useEffect(() => {
    loadRepo()
    api.listJobs().then(r => setJobs(r.jobs ?? []))
    loadTree()
    setSelectedFile(null)
    setFileContent('')
  }, [repoId, jobId])

  const pull = async () => {
    if (!repoId) return
    setPulling(true)
    setError('')
    try {
      await api.pullRepo(repoId)
      await loadTree()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Pull failed')
    } finally {
      setPulling(false)
    }
  }

  const selectFile = async (path: string) => {
    setSelectedFile(path)
    try {
      const content = await api.getRepoFile(path, repoId, jobId)
      setFileContent(content)
    } catch {
      setFileContent('Failed to load file.')
    }
  }

  const saveToken = async () => {
    if (!repoId || !tokenInput.trim()) return
    setSavingToken(true)
    try {
      await api.setRepoToken(repoId, tokenInput.trim())
      setTokenInput('')
      setSavedToken(true)
      setTimeout(() => setSavedToken(false), 2000)
      await loadRepo()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to save token')
    } finally {
      setSavingToken(false)
    }
  }

  const clearToken = async () => {
    if (!repoId) return
    setClearingToken(true)
    try {
      await api.deleteRepoToken(repoId)
      await loadRepo()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to clear token')
    } finally {
      setClearingToken(false)
    }
  }

  const saveRepo = async () => {
    if (!repoId) return
    setSaving(true)
    try {
      await api.updateRepo(repoId, { name: editName, url: editUrl, branch: editBranch, git_email: editGitEmail, git_name: editGitName })
      setEditing(false)
      await loadRepo()
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  const handleJobSelect = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const val = e.target.value
    if (val === '') navigate(`/repos/${repoId}`)
    else navigate(`/repos/${repoId}/jobs/${val}`)
  }

  if (!repo) return <div className="p-8 text-gray-400">Loading...</div>

  return (
    <div className="max-w-5xl mx-auto p-6 space-y-6">
      {/* Repo config header */}
      <div className="p-4 rounded bg-gray-900 border border-gray-800 space-y-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-sm font-medium text-white">{repo.name}</span>
            {repo.token_set
              ? <span className="text-xs text-green-400">● token set</span>
              : <span className="text-xs text-red-400">● no token</span>
            }
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setEditing(e => !e)}
              className="text-xs px-2 py-1 rounded border border-gray-700 text-gray-400 hover:text-white"
            >
              {editing ? 'Cancel' : 'Edit'}
            </button>
            <button
              onClick={() => navigate('/repos')}
              className="text-xs px-2 py-1 rounded border border-gray-700 text-gray-400 hover:text-white"
            >
              ← Repos
            </button>
          </div>
        </div>

        {editing ? (
          <div className="space-y-2">
            <input
              value={editName}
              onChange={e => setEditName(e.target.value)}
              placeholder="Name"
              className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white"
            />
            <input
              value={editUrl}
              onChange={e => setEditUrl(e.target.value)}
              placeholder="URL"
              className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white"
            />
            <input
              value={editBranch}
              onChange={e => setEditBranch(e.target.value)}
              placeholder="Branch"
              className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white"
            />
            <input
              value={editGitEmail}
              onChange={e => setEditGitEmail(e.target.value)}
              placeholder="Git email (e.g. you@example.com)"
              className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white"
            />
            <input
              value={editGitName}
              onChange={e => setEditGitName(e.target.value)}
              placeholder="Git name (e.g. Kenny Chen)"
              className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white"
            />
            <div className="flex justify-end">
              <button
                onClick={saveRepo}
                disabled={saving}
                className="text-sm px-3 py-1.5 rounded bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white"
              >
                {saving ? 'Saving...' : 'Save'}
              </button>
            </div>
          </div>
        ) : (
          <div className="text-xs text-gray-500 space-y-0.5">
            <div className="font-mono">{repo.url}</div>
            <div>Branch: {repo.branch}</div>
          </div>
        )}

        {/* Token management */}
        <div className="flex gap-2 pt-1">
          <input
            type="password"
            value={tokenInput}
            onChange={e => setTokenInput(e.target.value)}
            placeholder="Enter GitHub token..."
            autoComplete="new-password"
            className="flex-1 bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white"
          />
          <button
            onClick={saveToken}
            disabled={savingToken || !tokenInput.trim()}
            className="px-3 py-1.5 text-sm rounded bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white"
          >
            {savedToken ? 'Saved' : savingToken ? '...' : 'Set Token'}
          </button>
          {repo.token_set && (
            <button
              onClick={clearToken}
              disabled={clearingToken}
              className="px-3 py-1.5 text-sm rounded border border-red-700 text-red-400 hover:bg-red-900/30 disabled:opacity-40"
            >
              {clearingToken ? '...' : 'Clear'}
            </button>
          )}
        </div>
      </div>

      {error && <p className="text-sm text-red-400">{error}</p>}

      {/* File viewer */}
      <div className="flex items-center justify-between">
        <h2 className="text-sm text-gray-400 font-mono">{repo.url.replace('https://github.com/', '')} @ {repo.branch}</h2>
        <div className="flex items-center gap-2">
          <select
            value={jobId ?? ''}
            onChange={handleJobSelect}
            className="bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white"
          >
            <option value="">Source</option>
            {[...jobs].reverse().map(j => (
              <option key={j.job_id} value={j.job_id}>{j.job_id}</option>
            ))}
          </select>
          {!jobId && (
            <button
              onClick={pull}
              disabled={pulling}
              className="text-sm px-3 py-1.5 rounded bg-gray-800 border border-gray-700 text-gray-300 hover:text-white disabled:opacity-40"
            >
              {pulling ? 'Pulling...' : '↓ Pull'}
            </button>
          )}
        </div>
      </div>

      <div className="flex gap-4 h-[60vh]">
        <div className="w-64 flex-none bg-gray-900 border border-gray-800 rounded overflow-auto p-2">
          {tree
            ? <TreeNode node={tree} onSelect={selectFile} selected={selectedFile} />
            : <p className="text-xs text-gray-500 p-2">No files found. Pull to load.</p>
          }
        </div>
        <div className="flex-1 bg-gray-900 border border-gray-800 rounded overflow-auto">
          {selectedFile
            ? (
              <div className="h-full flex flex-col">
                <div className="px-4 py-2 border-b border-gray-800 text-xs text-gray-400 font-mono">{selectedFile}</div>
                <pre className="flex-1 p-4 text-xs text-gray-300 font-mono whitespace-pre overflow-auto">{fileContent}</pre>
              </div>
            )
            : <p className="text-sm text-gray-500 p-4">Select a file to view.</p>
          }
        </div>
      </div>
    </div>
  )
}

function TreeNode({ node, onSelect, selected, depth = 0 }: {
  node: FileNode
  onSelect: (path: string) => void
  selected: string | null
  depth?: number
}) {
  const [open, setOpen] = useState(depth === 0)

  if (node.is_dir) {
    return (
      <div>
        <div
          onClick={() => setOpen(o => !o)}
          className="flex items-center gap-1 px-2 py-0.5 rounded cursor-pointer hover:bg-gray-800 text-sm text-gray-400"
          style={{ paddingLeft: `${depth * 12 + 8}px` }}
        >
          <span className="text-xs">{open ? '▾' : '▸'}</span>
          <span>{node.name}</span>
        </div>
        {open && node.children?.map(child => (
          <TreeNode key={child.path} node={child} onSelect={onSelect} selected={selected} depth={depth + 1} />
        ))}
      </div>
    )
  }

  return (
    <div
      onClick={() => onSelect(node.path)}
      className={`px-2 py-0.5 rounded cursor-pointer text-sm hover:bg-gray-800 ${selected === node.path ? 'bg-gray-800 text-white' : 'text-gray-400'}`}
      style={{ paddingLeft: `${depth * 12 + 8}px` }}
    >
      {node.name}
    </div>
  )
}
