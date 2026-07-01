package main

import (
	"context"
	"sync"
	"time"

	"github.com/moby/moby/client"
)

type Config struct {
	Port            string
	DockerSocket    string
	BasePath        string
	JobsBasePath    string
	PromptsPath     string
	ContextPath     string
	ReposPath       string
	ReposBasePath   string
	UsagePath       string
	GracePeriodSecs int
	AnthropicAPIKey string
	OpenAIAPIKey    string
	GoogleAPIKey    string
}

type Repo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	Branch   string `json:"branch"`
	GitEmail string `json:"git_email"`
	GitName  string `json:"git_name"`
}

type RepoStatus struct {
	Repo
	TokenSet bool `json:"token_set"`
}

type JobRequest struct {
	Intent  string   `json:"intent"`
	Task    string   `json:"task"`
	Image   string   `json:"image"`
	Model   string   `json:"model"`
	Provider string  `json:"provider"`
	RepoIDs []string `json:"repo_ids"`
}

type JobResponse struct {
	JobID       string    `json:"job_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ContainerID string    `json:"container_id"`
}

type JobStatus struct {
	JobID       string    `json:"job_id"`
	ContainerID string    `json:"container_id"`
	Status      string    `json:"status"`
	Intent      string    `json:"intent"`
	Image       string    `json:"image"`
	Task        string    `json:"task"`
	RepoIDs     []string  `json:"repo_ids,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	ExitCode    int       `json:"exit_code,omitempty"`
}

type FileNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"is_dir"`
	Children []*FileNode `json:"children,omitempty"`
}

type UsageEntry struct {
	JobID        string    `json:"job_id"`
	Model        string    `json:"model"`
	Provider     string    `json:"provider"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	CostUSD      *float64  `json:"cost_usd"`
	Timestamp    time.Time `json:"timestamp"`
}

type Manager struct {
	docker     *client.Client
	config     Config
	mu         sync.RWMutex
	jobs       map[string]*JobStatus
	timers     map[string]context.CancelFunc
	repoTokens map[string]string
}
