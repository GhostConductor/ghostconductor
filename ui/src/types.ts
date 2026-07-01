export interface JobStatus {
  job_id: string
  container_id: string
  status: 'running' | 'completed' | 'failed' | 'timed_out' | 'crashed' | 'cancelled'
  intent?: string
  image?: string
  task?: string
  repo_ids?: string[]
  created_at: string
  started_at?: string
  completed_at?: string
  exit_code?: number
}

export interface JobResult {
  status: string
  exit_code: number
  completed_at: string
  reason?: string
  usage?: {
    input_tokens: number
    output_tokens: number
    cache_creation_input_tokens?: number
    cache_read_input_tokens?: number
    total_cost_usd?: number
  }
}

export interface ConfigStatus {
  anthropic_api_key: boolean
  openai_api_key: boolean
  google_api_key: boolean
}

export interface FileNode {
  name: string
  path: string
  is_dir: boolean
  children?: FileNode[]
}


export interface UsageEntry {
  job_id: string
  model: string
  provider: string
  input_tokens: number
  output_tokens: number
  cost_usd: number | null
  timestamp: string
}

export interface Repo {
  id: string
  name: string
  url: string
  branch: string
  git_email?: string
  git_name?: string
}

export interface RepoStatus extends Repo {
  token_set: boolean
}
