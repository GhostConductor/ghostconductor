import { useState } from 'react'
import { api } from '../api'

export default function SettingsPage() {
  const [clearingJobs, setClearingJobs] = useState(false)
  const [clearingMemory, setClearingMemory] = useState(false)
  const [clearingProviders, setClearingProviders] = useState(false)
  const [clearingUsage, setClearingUsage] = useState(false)
  const [clearingRepos, setClearingRepos] = useState(false)
  const [resetting, setResetting] = useState(false)
  const [error, setError] = useState('')

  const clearJobs = async () => {
    if (!confirm('Delete all jobs? This cannot be undone.')) return
    setClearingJobs(true)
    try { await api.clearJobs() }
    catch (e: unknown) { setError(e instanceof Error ? e.message : 'Failed to clear jobs') }
    finally { setClearingJobs(false) }
  }

  const clearMemory = async () => {
    if (!confirm('Clear project memory? This cannot be undone.')) return
    setClearingMemory(true)
    try { await api.clearMemory() }
    catch (e: unknown) { setError(e instanceof Error ? e.message : 'Failed to clear memory') }
    finally { setClearingMemory(false) }
  }

  const clearProviders = async () => {
    if (!confirm('Clear all provider keys? This cannot be undone.')) return
    setClearingProviders(true)
    try { await api.clearProviders() }
    catch (e: unknown) { setError(e instanceof Error ? e.message : 'Failed to clear providers') }
    finally { setClearingProviders(false) }
  }

  const clearUsage = async () => {
    if (!confirm('Clear all usage history? This cannot be undone.')) return
    setClearingUsage(true)
    try { await api.clearUsage() }
    catch (e: unknown) { setError(e instanceof Error ? e.message : 'Failed to clear usage') }
    finally { setClearingUsage(false) }
  }

  const clearRepos = async () => {
    if (!confirm('Delete all repos and tokens? This cannot be undone.')) return
    setClearingRepos(true)
    try { await api.clearRepos() }
    catch (e: unknown) { setError(e instanceof Error ? e.message : 'Failed to clear repos') }
    finally { setClearingRepos(false) }
  }

  const factoryReset = async () => {
    if (!confirm('Factory reset? This will clear all jobs, memory, repos, usage, and reset all config values. This cannot be undone.')) return
    setResetting(true)
    try { await api.factoryReset() }
    catch (e: unknown) { setError(e instanceof Error ? e.message : 'Factory reset failed') }
    finally { setResetting(false) }
  }

  const btnClass = 'text-sm px-3 py-1.5 w-36 rounded border border-red-700 text-red-400 hover:bg-red-900/30 disabled:opacity-40 whitespace-nowrap'

  return (
    <div className="max-w-5xl mx-auto p-6 space-y-6">
      <h1 className="font-semibold text-white">Settings</h1>

      {error && <p className="text-sm text-red-400">{error}</p>}

      <div className="space-y-3">
        <h2 className="text-xs text-gray-500 uppercase tracking-wider">Danger Zone</h2>
        <div className="p-4 rounded bg-gray-900 border border-red-900 space-y-3">

          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-300">Clear Jobs</p>
              <p className="text-xs text-gray-500">Stops all containers and deletes all job directories.</p>
            </div>
            <button onClick={clearJobs} disabled={clearingJobs} className={btnClass}>
              {clearingJobs ? 'Clearing...' : 'Clear Jobs'}
            </button>
          </div>
          <div className="border-t border-gray-800" />

          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-300">Clear Memory</p>
              <p className="text-xs text-gray-500">Wipes the project memory file.</p>
            </div>
            <button onClick={clearMemory} disabled={clearingMemory} className={btnClass}>
              {clearingMemory ? 'Clearing...' : 'Clear Memory'}
            </button>
          </div>
          <div className="border-t border-gray-800" />

          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-300">Clear Providers</p>
              <p className="text-xs text-gray-500">Clears all ghost and medium provider API keys from memory.</p>
            </div>
            <button onClick={clearProviders} disabled={clearingProviders} className={btnClass}>
              {clearingProviders ? 'Clearing...' : 'Clear Providers'}
            </button>
          </div>
          <div className="border-t border-gray-800" />

          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-300">Clear Usage</p>
              <p className="text-xs text-gray-500">Wipes all token usage and cost history.</p>
            </div>
            <button onClick={clearUsage} disabled={clearingUsage} className={btnClass}>
              {clearingUsage ? 'Clearing...' : 'Clear Usage'}
            </button>
          </div>
          <div className="border-t border-gray-800" />

          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-300">Clear Repos</p>
              <p className="text-xs text-gray-500">Deletes all registered repos and clears all tokens.</p>
            </div>
            <button onClick={clearRepos} disabled={clearingRepos} className={btnClass}>
              {clearingRepos ? 'Clearing...' : 'Clear Repos'}
            </button>
          </div>
          <div className="border-t border-gray-800" />

          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-300">Factory Reset</p>
              <p className="text-xs text-gray-500">Clears all jobs, memory, repos, usage, and resets all config values.</p>
            </div>
            <button onClick={factoryReset} disabled={resetting} className={btnClass}>
              {resetting ? 'Resetting...' : 'Factory Reset'}
            </button>
          </div>

        </div>
      </div>
    </div>
  )
}
