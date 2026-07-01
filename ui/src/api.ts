import { ConfigStatus, FileNode, JobResult, JobStatus, Repo, RepoStatus, UsageEntry } from './types'

const BASE = '/api/v1'

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`)
  if (!res.ok) throw new Error(`GET ${path} failed: ${res.status}`)
  return res.json()
}

async function getText(path: string): Promise<string> {
  const res = await fetch(`${BASE}${path}`)
  if (res.status === 404) return ''
  if (!res.ok) throw new Error(`GET ${path} failed: ${res.status}`)
  return res.text()
}

async function put(path: string, body: string): Promise<void> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'text/plain' },
    body,
  })
  if (!res.ok) throw new Error(`PUT ${path} failed: ${res.status}`)
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`POST ${path} failed: ${res.status}`)
  return res.json()
}

async function del<T>(path: string, body?: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'DELETE',
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) throw new Error(`DELETE ${path} failed: ${res.status}`)
  return res.json()
}

async function putJson<T>(path: string, body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  if (!res.ok) throw new Error(`PUT ${path} failed: ${res.status}`)
  return res.json()
}

export const api = {
  getConfig: () => get<ConfigStatus>('/config'),
  updateConfig: (data: Partial<Record<string, string>>) => post<{ status: string }>('/config', data),
  deleteConfig: (key: string) => del<{ status: string }>('/config', { key }),

  getIntents: () => get<{ intents: string[] }>('/intents'),

  listJobs: () => get<{ jobs: JobStatus[] }>('/jobs'),
  submitJob: (intent: string, task: string, model: string, provider: string, repoIds: string[]) =>
    post<JobStatus>('/jobs', { intent, task, model, provider, repo_ids: repoIds }),
  getJob: (id: string) => get<JobStatus>(`/jobs/${id}`),
  getJobStatus: (id: string) => get<JobStatus>(`/jobs/${id}/status`),
  getJobResult: (id: string) => get<JobResult>(`/jobs/${id}/result`),
  getJobLogs: (id: string) => getText(`/jobs/${id}/logs`),
  getJobEvents: (id: string) => getText(`/jobs/${id}/events`),
  getJobTree: (id: string) => get<FileNode>(`/jobs/${id}/tree`),
  cancelJob: (id: string) =>
    fetch(`${BASE}/jobs/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ action: 'cancel' }),
    }),
  deleteJob: (id: string) => del<{ job_id: string; status: string }>(`/jobs/${id}`),
  clearJobs: () => post<{ status: string }>('/jobs/clear', {}),

  getMemory: () => getText('/memory'),
  putMemory: (content: string) => put('/memory', content),
  clearMemory: () => del<{ status: string }>('/memory'),

  getContext: () => getText('/context'),
  putContext: (content: string) => put('/context', content),
  getContextTemplates: () => get<{ templates: string[] }>('/context/templates'),
  loadContextTemplate: (template: string) => post<{ status: string; template: string }>('/context/load', { template }),

  factoryReset: () => post<{ status: string }>('/factory-reset', {}),

  getRepos: () => get<RepoStatus[]>('/repos'),
  createRepo: (data: { name: string; url: string; branch: string }) =>
    post<Repo>('/repos', data),
  updateRepo: (id: string, data: { name?: string; url?: string; branch?: string; git_email?: string; git_name?: string }) =>
    putJson<{ status: string }>(`/repos/${id}`, data),
  deleteRepo: (id: string) => del<{ status: string }>(`/repos/${id}`),
  setRepoToken: (id: string, token: string) =>
    post<{ status: string }>(`/repos/${id}/token`, { token }),
  deleteRepoToken: (id: string) => del<{ status: string }>(`/repos/${id}/token`),

  getRepoTree: (repoId?: string, jobId?: string) => {
    const params = new URLSearchParams()
    if (repoId) params.set('repo_id', repoId)
    if (jobId) params.set('job_id', jobId)
    const qs = params.toString()
    return get<FileNode>(`/repo/tree${qs ? `?${qs}` : ''}`)
  },
  getRepoFile: (path: string, repoId?: string, jobId?: string) => {
    const params = new URLSearchParams({ path })
    if (repoId) params.set('repo_id', repoId)
    if (jobId) params.set('job_id', jobId)
    return getText(`/repo/file?${params.toString()}`)
  },
  pullRepo: (repoId: string) => post<{ status: string }>(`/repos/${repoId}/pull`, {}),

  getUsage: () => get<UsageEntry[]>('/usage'),
  clearUsage: () => del<{ status: string }>('/usage'),
  clearRepos: () => del<{ status: string }>('/repos'),
  clearProviders: () => del<{ status: string }>('/providers'),
}
