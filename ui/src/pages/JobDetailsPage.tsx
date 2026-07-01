import { useEffect, useRef, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api } from '../api'
import { JobResult, JobStatus } from '../types'

type Tab = 'logs' | 'events' | 'result'

function truncate(s: string, n: number) {
  return s && s.length > n ? s.slice(0, n) + '…' : s
}

export default function JobDetailPage() {
  const { jobId } = useParams<{ jobId: string }>()
  const navigate = useNavigate()
  const [status, setStatus] = useState<JobStatus | null>(null)
  const [result, setResult] = useState<JobResult | null>(null)
  const [logs, setLogs] = useState('')
  const [events, setEvents] = useState('')
  const [tab, setTab] = useState<Tab>('logs')
  const [cancelling, setCancelling] = useState(false)
  const logsRef = useRef<HTMLPreElement>(null)

  const isTerminal = (s: string) => ['completed', 'failed', 'timed_out', 'crashed', 'cancelled'].includes(s)

  const loadStatus = async () => {
    if (!jobId) return
    try {
      const s = await api.getJobStatus(jobId)
      setStatus(s)
      return s
    } catch { return null }
  }

  const loadLogs = async () => {
    if (!jobId) return
    try {
      const l = await api.getJobLogs(jobId)
      setLogs(l)
      if (logsRef.current) logsRef.current.scrollTop = logsRef.current.scrollHeight
    } catch { /* not ready yet */ }
  }

  const loadEvents = async () => {
    if (!jobId) return
    try { setEvents(await api.getJobEvents(jobId)) } catch { /* not ready */ }
  }

  const loadResult = async () => {
    if (!jobId) return
    try { setResult(await api.getJobResult(jobId)) } catch { /* not ready */ }
  }

  useEffect(() => {
    if (!jobId) return
    loadStatus()
    loadLogs()
    loadEvents()

    const t = setInterval(async () => {
      const s = await loadStatus()
      await loadLogs()
      await loadEvents()
      if (s && isTerminal(s.status)) {
        clearInterval(t)
        await loadResult()
      }
    }, 3000)

    return () => clearInterval(t)
  }, [jobId])

  useEffect(() => {
    if (status && isTerminal(status.status)) loadResult()
  }, [status?.status])

  const cancel = async () => {
    if (!jobId) return
    setCancelling(true)
    try { await api.cancelJob(jobId); await loadStatus() }
    finally { setCancelling(false) }
  }

  const deleteJob = async () => {
    if (!jobId) return
    await api.deleteJob(jobId)
    navigate('/')
  }

  if (!status) return <div className="p-8 text-gray-400">Loading...</div>

  return (
    <div className="max-w-5xl mx-auto p-6 space-y-6">

      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <h1 className="font-mono text-lg text-white">{status.job_id}</h1>
          <div className="flex gap-3 text-xs text-gray-400 flex-wrap">
            {status.intent && <span className="bg-gray-800 px-2 py-0.5 rounded">{status.intent}</span>}
            {status.image && <span className="bg-gray-800 px-2 py-0.5 rounded font-mono">{truncate(status.image, 40)}</span>}
          </div>
          {status.task && (
            <p className="text-sm text-gray-400 max-w-xl">{truncate(status.task, 120)}</p>
          )}
          <p className="text-xs text-gray-500">
            Started: {status.started_at ? new Date(status.started_at).toLocaleString() : '—'}
            {status.completed_at && status.completed_at !== '0001-01-01T00:00:00Z'
              ? ` · Completed: ${new Date(status.completed_at).toLocaleString()}`
              : ''}
          </p>
        </div>
        <div className="flex gap-2 ml-4">
          <Link
            to={`/repo/${status.job_id}`}
            className="text-sm px-3 py-1.5 rounded border border-gray-700 text-gray-400 hover:text-white"
          >
            View Files
          </Link>
          {status.status === 'running' && (
            <button
              onClick={cancel}
              disabled={cancelling}
              className="text-sm px-3 py-1.5 rounded border border-yellow-700 text-yellow-400 hover:bg-yellow-900/30 disabled:opacity-40"
            >
              {cancelling ? 'Cancelling...' : 'Cancel'}
            </button>
          )}
          {isTerminal(status.status) && (
            <button
              onClick={deleteJob}
              className="text-sm px-3 py-1.5 rounded border border-red-800 text-red-400 hover:bg-red-900/30"
            >
              Delete
            </button>
          )}
        </div>
      </div>

      {/* Result summary */}
      {result && (
        <div className="p-4 rounded bg-gray-900 border border-gray-800 text-sm space-y-1">
          <div className="flex gap-4">
            <span className="text-gray-400">Status: <span className="text-white">{result.status}</span></span>
            <span className="text-gray-400">Exit code: <span className="text-white">{result.exit_code}</span></span>
          </div>
          {result.reason && <p className="text-red-400">{result.reason}</p>}
        </div>
      )}

      {/* Tabs */}
      <div>
        <div className="flex gap-1 border-b border-gray-800 mb-3">
          {(['logs', 'events', 'result'] as Tab[]).map(t => (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={`px-4 py-2 text-sm border-b-2 -mb-px ${tab === t ? 'border-indigo-500 text-white' : 'border-transparent text-gray-500 hover:text-gray-300'}`}
            >
              {t}
            </button>
          ))}
        </div>

        {tab === 'logs' && (
          <pre
            ref={logsRef}
            className="bg-gray-900 rounded p-4 text-xs text-gray-300 font-mono overflow-auto max-h-[60vh] whitespace-pre-wrap"
          >
            {logs || 'No logs yet.'}
          </pre>
        )}

        {tab === 'events' && (
          <pre className="bg-gray-900 rounded p-4 text-xs text-gray-300 font-mono overflow-auto max-h-[60vh] whitespace-pre-wrap">
            {events || 'No events yet.'}
          </pre>
        )}

        {tab === 'result' && (
          <pre className="bg-gray-900 rounded p-4 text-xs text-gray-300 font-mono overflow-auto max-h-[60vh] whitespace-pre-wrap">
            {result ? JSON.stringify(result, null, 2) : 'No result yet.'}
          </pre>
        )}
      </div>
    </div>
  )
}
