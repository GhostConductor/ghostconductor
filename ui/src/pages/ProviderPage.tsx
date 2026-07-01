import { useEffect, useState } from 'react'
import { api } from '../api'
import { ConfigStatus, UsageEntry } from '../types'

const PROVIDERS = ['anthropic', 'openai', 'google'] as const
type ProviderID = typeof PROVIDERS[number]

const PROVIDER_LABELS: Record<ProviderID, string> = {
  anthropic: 'Anthropic',
  openai: 'OpenAI',
  google: 'Google',
}

const PROVIDER_ENV_KEYS: Record<ProviderID, string> = {
  anthropic: 'anthropic_api_key',
  openai: 'openai_api_key',
  google: 'google_api_key',
}

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(2)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`
  return n.toString()
}

function formatCost(n: number): string {
  return `$${n.toFixed(2)}`
}

export default function ProviderPage() {
  const [configStatus, setConfigStatus] = useState<ConfigStatus | null>(null)
  const [usage, setUsage] = useState<UsageEntry[]>([])
  const [keyInputs, setKeyInputs] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState<string | null>(null)
  const [saved, setSaved] = useState<string | null>(null)
  const [clearing, setClearing] = useState<string | null>(null)
  const [showAdd, setShowAdd] = useState(false)
  const [addProvider, setAddProvider] = useState<ProviderID>('anthropic')
  const [addKey, setAddKey] = useState('')
  const [error, setError] = useState('')
  const [loaded, setLoaded] = useState(false)

  const load = () => Promise.all([
    api.getUsage(),
    api.getConfig(),
  ]).then(([u, c]) => {
    setUsage(u)
    setConfigStatus(c)
    setLoaded(true)
  }).catch(() => setError('Failed to load provider config'))

  useEffect(() => { load() }, [])

  const keySet: Record<ProviderID, boolean> = {
    anthropic: configStatus?.anthropic_api_key ?? false,
    openai: configStatus?.openai_api_key ?? false,
    google: configStatus?.google_api_key ?? false,
  }

  const configuredProviders = PROVIDERS.filter(id => keySet[id])
  const availableProviders = PROVIDERS.filter(id => !keySet[id])

  useEffect(() => {
    if (availableProviders.length > 0 && !availableProviders.includes(addProvider)) {
      setAddProvider(availableProviders[0])
    }
  }, [configStatus])

  const addProvider_ = async () => {
    if (!addKey.trim()) return
    setSaving('add')
    setError('')
    try {
      await api.updateConfig({ [PROVIDER_ENV_KEYS[addProvider]]: addKey.trim() })
      setConfigStatus(await api.getConfig())
      setAddKey('')
      setShowAdd(false)
      setSaved('add')
      setTimeout(() => setSaved(null), 2000)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to save key')
    } finally {
      setSaving(null)
    }
  }

  const updateKey = async (id: ProviderID) => {
    const val = keyInputs[id]
    if (!val?.trim()) return
    setSaving(id)
    setError('')
    try {
      await api.updateConfig({ [PROVIDER_ENV_KEYS[id]]: val.trim() })
      setConfigStatus(await api.getConfig())
      setKeyInputs(prev => ({ ...prev, [id]: '' }))
      setSaved(id)
      setTimeout(() => setSaved(null), 2000)
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to save key')
    } finally {
      setSaving(null)
    }
  }

  const clearKey = async (id: ProviderID) => {
    setClearing(id)
    setError('')
    try {
      await api.deleteConfig(PROVIDER_ENV_KEYS[id])
      setConfigStatus(await api.getConfig())
      setUsage(await api.getUsage())
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : 'Failed to clear key')
    } finally {
      setClearing(null)
    }
  }

  const usageByProvider = (id: ProviderID) => {
    const entries = usage.filter(u => u.provider === id)
    const input = entries.reduce((s, e) => s + e.input_tokens, 0)
    const output = entries.reduce((s, e) => s + e.output_tokens, 0)
    const cost = entries.reduce((s, e) => s + (e.cost_usd ?? 0), 0)
    return { input, output, cost, jobs: entries.length }
  }

  if (!loaded && !error) return <div className="p-8 text-gray-400">Loading...</div>

  const selectClass = 'bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white'
  const inputClass = 'flex-1 bg-gray-800 border border-gray-700 rounded px-3 py-1.5 text-sm text-white'

  return (
    <div className="max-w-5xl mx-auto p-6 space-y-6">
      <h1 className="font-semibold text-white">Providers</h1>
      {error && <p className="text-sm text-red-400">{error}</p>}

      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-xs text-gray-500 uppercase tracking-wider">API Keys</h2>
          {availableProviders.length > 0 && (
            <button
              onClick={() => setShowAdd(s => !s)}
              className="text-sm px-3 py-1.5 rounded bg-indigo-600 hover:bg-indigo-500 text-white"
            >
              + Add Provider
            </button>
          )}
        </div>

        {showAdd && (
          <div className="p-4 rounded bg-gray-900 border border-gray-700 space-y-3">
            <h3 className="text-sm font-medium text-white">Add Provider</h3>
            <div className="flex gap-2">
              <select
                value={addProvider}
                onChange={e => setAddProvider(e.target.value as ProviderID)}
                className={selectClass}
              >
                {availableProviders.map(id => (
                  <option key={id} value={id}>{PROVIDER_LABELS[id]}</option>
                ))}
              </select>
              <input
                type="password"
                value={addKey}
                onChange={e => setAddKey(e.target.value)}
                placeholder="API key..."
                autoComplete="new-password"
                className={inputClass}
              />
            </div>
            <div className="flex gap-2 justify-end">
              <button
                onClick={() => { setShowAdd(false); setAddKey('') }}
                className="text-sm px-3 py-1.5 rounded border border-gray-700 text-gray-400 hover:text-white"
              >
                Cancel
              </button>
              <button
                onClick={addProvider_}
                disabled={saving === 'add' || !addKey.trim()}
                className="text-sm px-3 py-1.5 rounded bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white"
              >
                {saving === 'add' ? 'Adding...' : 'Add'}
              </button>
            </div>
          </div>
        )}

        {configuredProviders.length === 0 && !showAdd && (
          <p className="text-sm text-gray-500">No providers configured.</p>
        )}

        {configuredProviders.map(id => {
          const stats = usageByProvider(id)
          return (
            <div key={id} className="p-4 rounded bg-gray-900 border border-gray-800 space-y-3">
              <div className="flex items-center gap-3">
                <span className="text-sm font-medium text-gray-300">{PROVIDER_LABELS[id]}</span>
                <span className="text-xs text-green-400">● key set</span>
              </div>
              <div className="flex gap-2">
                <input
                  type="password"
                  value={keyInputs[id] ?? ''}
                  onChange={e => setKeyInputs(prev => ({ ...prev, [id]: e.target.value }))}
                  placeholder="Enter new API key..."
                  autoComplete="new-password"
                  className={inputClass}
                />
                <button
                  onClick={() => updateKey(id)}
                  disabled={saving === id || !keyInputs[id]?.trim()}
                  className="px-3 py-1.5 text-sm rounded bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white"
                >
                  {saved === id ? 'Saved' : saving === id ? '...' : 'Set'}
                </button>
                <button
                  onClick={() => clearKey(id)}
                  disabled={clearing === id}
                  className="px-3 py-1.5 text-sm rounded border border-red-700 text-red-400 hover:bg-red-900/30 disabled:opacity-40"
                >
                  {clearing === id ? '...' : 'Clear'}
                </button>
              </div>
              {stats.jobs > 0 && (
                <div className="flex gap-4 text-xs text-gray-500 pt-1">
                  <span>Jobs: <span className="text-gray-300">{stats.jobs}</span></span>
                  <span>In: <span className="text-gray-300">{formatTokens(stats.input)}</span></span>
                  <span>Out: <span className="text-gray-300">{formatTokens(stats.output)}</span></span>
                  {stats.cost > 0 && <span>Cost: <span className="text-green-400">{formatCost(stats.cost)}</span></span>}
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
