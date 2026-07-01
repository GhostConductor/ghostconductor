import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { api } from '../api'
import { ConfigStatus, JobStatus, RepoStatus, UsageEntry } from '../types'

const STATUS_COLORS: Record<string, string> = {
  running:   'text-blue-400',
  completed: 'text-green-400',
  failed:    'text-red-400',
  timed_out: 'text-yellow-400',
  crashed:   'text-red-500',
  cancelled: 'text-gray-500',
}

interface GhostModel {
  label:    string
  model:    string
  provider: string
}

const ALL_MODELS: GhostModel[] = [
  { label: 'Claude Sonnet', model: 'claude-sonnet-4-6', provider: 'anthropic' },
  { label: 'Claude Haiku',  model: 'claude-haiku-4-5',  provider: 'anthropic' },
  { label: 'Claude Opus',   model: 'claude-opus-4-6',   provider: 'anthropic' },
  { label: 'GPT-4.1',       model: 'gpt-4.1',           provider: 'openai' },
  { label: 'GPT-4.1 Mini',  model: 'gpt-4.1-mini',      provider: 'openai' },
  { label: 'GPT-4.1 Nano',  model: 'gpt-4.1-nano',      provider: 'openai' },
  { label: 'Gemini Pro',    model: 'gemini-2.0-pro',     provider: 'google' },
  { label: 'Gemini Flash',  model: 'gemini-2.5-flash',   provider: 'google' },
  { label: 'Gemini Lite',   model: 'gemini-2.5-flash-lite', provider: 'google' },
]

function truncate(s: string, n: number) {
  return s && s.length > n ? s.slice(0, n) + '…' : s
}

function getAvailableModels(cfg: ConfigStatus | null): GhostModel[] {
  if (!cfg) return []
  return ALL_MODELS.filter(m => {
    if (m.provider === 'anthropic') return cfg.anthropic_api_key
    if (m.provider === 'openai')    return cfg.openai_api_key
    if (m.provider === 'google')    return cfg.google_api_key
    return false
  })
}

function getModelLabel(model: string): string {
  return ALL_MODELS.find(m => m.model === model)?.label ?? model
}

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(2)}M`
  if (n >= 1_000)     return `${(n / 1_000).toFixed(1)}K`
  return n.toString()
}

function formatCost(n: number): string {
  return `$${n.toFixed(2)}`
}

export default function JobsPage() {
  const [jobs, setJobs]                   = useState<JobStatus[]>([])
  const [usage, setUsage]                 = useState<UsageEntry[]>([])
  const [intents, setIntents]             = useState<string[]>([])
  const [config, setConfig]               = useState<ConfigStatus | null>(null)
  const [repos, setRepos]                 = useState<RepoStatus[]>([])
  const [showModal, setShowModal]         = useState(false)
  const [intent, setIntent]               = useState('')
  const [selectedModel, setSelectedModel] = useState<GhostModel | null>(null)
  const [selectedRepoIds, setSelectedRepoIds] = useState<string[]>([])
  const [task, setTask]                   = useState('')
  const [submitting, setSubmitting]       = useState(false)
  const [deleting, setDeleting]           = useState<string | null>(null)
  const [error, setError]                 = useState('')
  const navigate = useNavigate()

  const load = () => Promise.all([
    api.listJobs(),
    api.getUsage(),
  ]).then(([r, u]) => {
    setJobs(r.jobs ?? [])
    setUsage(u)
  })

  useEffect(() => {
    load()
    api.getConfig().then(c => {
      setConfig(c)
      const models = getAvailableModels(c)
      if (models.length > 0) setSelectedModel(models[0])
    })
    api.getIntents().then(r => {
      setIntents(r.intents)
      if (r.intents.length > 0) setIntent(r.intents[0])
    })
    api.getRepos().then(r => {
      setRepos(r.filter(repo => repo.token_set))
    })
    const t = setInterval(load, 5000)
    return () => clearInterval(t)
  }, [])

  const availableModels = getAvailableModels(config)
  const hasApiKey = !!(config?.anthropic_api_key || config?.openai_api_key || config?.google_api_key)

  const getJobUsage = (jobId: string): UsageEntry | undefined =>
    usage.find(u => u.job_id === jobId)

  const toggleRepo = (id: string) => {
    setSelectedRepoIds(prev =>
      prev.includes(id) ? prev.filter(r => r !== id) : [...prev, id]
    )
  }

  const submit = async () => {
    if (!intent || !task.trim() || selectedRepoIds.length === 0 || !selectedModel) return
    setSubmitting(true)
    setError('')
    try {
      const job = await api.submitJob(intent, task, selectedModel.model, selectedModel.provider, selectedRepoIds)
      setTask('')
      setSelectedRepoIds([])
      setShowModal(false)
      navigate(`/jobs/${job.job_id}`)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to summon spirit')
    } finally {
      setSubmitting(false)
    }
  }

  const deleteJob = async (e: React.MouseEvent, jobId: string) => {
    e.stopPropagation()
    setDeleting(jobId)
    try {
      await api.deleteJob(jobId)
      await load()
    } finally {
      setDeleting(null)
    }
  }

  return (
    <div className="max-w-5xl mx-auto p-6 space-y-6">

      {!hasApiKey && (
        <div className="p-4 rounded bg-yellow-900/40 border border-yellow-700 text-yellow-300 text-sm flex items-center justify-between">
          <span>An API key is required to summon spirits. Configure Anthropic, OpenAI, or Google.</span>
          <Link to="/providers" className="underline hover:text-yellow-100">Go to Providers →</Link>
        </div>
      )}

      <div className="flex items-center justify-between">
        <h2 className="font-semibold text-white">Summonings</h2>
        {hasApiKey && (
          <button
            onClick={() => setShowModal(true)}
            className="text-sm px-3 py-1.5 rounded bg-indigo-600 hover:bg-indigo-500 text-white"
          >
            Conjure Spirit
          </button>
        )}
      </div>

      {jobs.length === 0
        ? <p className="text-sm text-gray-500">No summonings yet.</p>
        : (
          <div className="space-y-2">
            {[...jobs].reverse().map(job => {
              const jobUsage = getJobUsage(job.job_id)
              const isTerminal = ['completed', 'failed', 'timed_out', 'crashed', 'cancelled'].includes(job.status)
              return (
                <div
                  key={job.job_id}
                  onClick={() => navigate(`/jobs/${job.job_id}`)}
                  className="flex items-center justify-between p-3 rounded bg-gray-900 border border-gray-800 cursor-pointer hover:border-gray-600"
                >
                  <div className="flex-1 min-w-0 space-y-0.5">
                    <div className="font-mono text-sm text-gray-300">{job.job_id}</div>
                    <div className="text-xs text-gray-500">
                      {job.intent && <span>{job.intent}</span>}
                      {job.task && <span className="text-gray-600"> — {truncate(job.task, 60)}</span>}
                    </div>
                    {isTerminal && jobUsage && (
                      <div className="text-xs text-gray-500 flex gap-3">
                        <span className="text-gray-400">{getModelLabel(jobUsage.model)}</span>
                        <span>In: <span className="text-gray-300">{formatTokens(jobUsage.input_tokens)}</span></span>
                        <span>Out: <span className="text-gray-300">{formatTokens(jobUsage.output_tokens)}</span></span>
                        {jobUsage.cost_usd != null && (
                          <span>Cost: <span className="text-green-400">{formatCost(jobUsage.cost_usd)}</span></span>
                        )}
                      </div>
                    )}
                  </div>
                  <div className="flex items-center gap-3 ml-4">
                    <span className={`text-xs font-medium ${STATUS_COLORS[job.status] ?? 'text-gray-400'}`}>
                      {job.status}
                    </span>
                    <button
                      onClick={e => deleteJob(e, job.job_id)}
                      disabled={job.status === 'running' || deleting === job.job_id}
                      className="text-xs px-2 py-1 rounded border border-red-900 text-red-500 hover:bg-red-900/30 disabled:opacity-30"
                    >
                      {deleting === job.job_id ? '…' : 'Delete'}
                    </button>
                  </div>
                </div>
              )
            })}
          </div>
        )
      }

      {showModal && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-lg space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="font-semibold text-white">Conjure Spirit</h2>
              <button onClick={() => setShowModal(false)} className="text-gray-500 hover:text-white text-lg">✕</button>
            </div>

            <div className="flex gap-3">
              <div className="flex-none">
                <label className="text-xs text-gray-400 block mb-1">Intent</label>
                <select
                  value={intent}
                  onChange={e => setIntent(e.target.value)}
                  className="bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-white"
                >
                  {intents.map(i => <option key={i} value={i}>{i}</option>)}
                </select>
              </div>
              <div className="flex-1">
                <label className="text-xs text-gray-400 block mb-1">Model</label>
                <select
                  value={selectedModel?.model ?? ''}
                  onChange={e => setSelectedModel(availableModels.find(m => m.model === e.target.value) ?? null)}
                  className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-white"
                >
                  {availableModels.map(m => (
                    <option key={m.model} value={m.model}>{m.label}</option>
                  ))}
                </select>
              </div>
            </div>

            <div>
              <label className="text-xs text-gray-400 block mb-1">Repos</label>
              {repos.length === 0
                ? <p className="text-xs text-yellow-400">No repos with tokens configured. Add repos in the Repos page.</p>
                : (
                  <div className="space-y-1">
                    {repos.map(repo => (
                      <label key={repo.id} className="flex items-center gap-2 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={selectedRepoIds.includes(repo.id)}
                          onChange={() => toggleRepo(repo.id)}
                          className="w-4 h-4 accent-indigo-500"
                        />
                        <span className="text-sm text-gray-300">{repo.name}</span>
                        <span className="text-xs text-gray-600 font-mono">{repo.branch}</span>
                      </label>
                    ))}
                  </div>
                )
              }
            </div>

            <div>
              <label className="text-xs text-gray-400 block mb-1">Task</label>
              <textarea
                value={task}
                onChange={e => setTask(e.target.value)}
                rows={5}
                className="w-full bg-gray-800 border border-gray-700 rounded px-3 py-2 text-sm text-white resize-none"
                placeholder="Describe what the spirit should do..."
                autoFocus
              />
            </div>

            {error && <p className="text-sm text-red-400">{error}</p>}

            <div className="flex justify-end gap-2">
              <button
                onClick={() => setShowModal(false)}
                className="text-sm px-4 py-2 rounded border border-gray-700 text-gray-400 hover:text-white"
              >
                Cancel
              </button>
              <button
                onClick={submit}
                disabled={submitting || !task.trim() || !selectedModel || selectedRepoIds.length === 0}
                className="text-sm px-4 py-2 rounded bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white"
              >
                {submitting ? 'Conjuring...' : 'Conjure Spirit'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
