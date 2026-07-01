package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/moby/moby/client"
)

// SubmitJob handles POST /jobs
func (m *Manager) SubmitJob(w http.ResponseWriter, r *http.Request) {
	var req JobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("ERROR SubmitJob: invalid request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Intent == "" || req.Task == "" {
		http.Error(w, "missing required fields: intent, task", http.StatusBadRequest)
		return
	}

	if len(req.RepoIDs) == 0 {
		http.Error(w, "at least one repo_id is required", http.StatusBadRequest)
		return
	}

	allRepos, err := m.readRepos()
	if err != nil {
		log.Printf("ERROR SubmitJob: failed to read repos: %v", err)
		http.Error(w, "failed to read repos", http.StatusInternalServerError)
		return
	}

	repoMap := make(map[string]Repo)
	for _, repo := range allRepos {
		repoMap[repo.ID] = repo
	}

	selectedRepos := make([]Repo, 0)
	for _, id := range req.RepoIDs {
		repo, ok := repoMap[id]
		if !ok {
			http.Error(w, fmt.Sprintf("repo not found: %s", id), http.StatusBadRequest)
			return
		}
		selectedRepos = append(selectedRepos, repo)
	}

	m.mu.RLock()
	for _, repo := range selectedRepos {
		if m.repoTokens[repo.ID] == "" {
			m.mu.RUnlock()
			http.Error(w, fmt.Sprintf("no token set for repo: %s", repo.Name), http.StatusBadRequest)
			return
		}
	}
	m.mu.RUnlock()

	jobID := generateJobID()

	image := req.Image
	if image == "" {
		image = defaultImage
	}

	provider := req.Provider
	model := req.Model

	apiKey := getAPIKeyForProvider(provider, m.config)
	if apiKey == "" {
		http.Error(w, fmt.Sprintf("no API key configured for provider: %s", provider), http.StatusBadRequest)
		return
	}

	jobPath := filepath.Join(m.config.JobsBasePath, jobID)
	for _, dir := range []string{"code", "data", "log"} {
		if err := os.MkdirAll(filepath.Join(jobPath, dir), 0755); err != nil {
			log.Printf("ERROR SubmitJob: failed to create %s dir: %v", dir, err)
			http.Error(w, fmt.Sprintf("failed to create %s dir: %v", dir, err), http.StatusInternalServerError)
			return
		}
	}

	codePath := filepath.Join(jobPath, "code")
	m.mu.RLock()
	for _, repo := range selectedRepos {
		token := m.repoTokens[repo.ID]
		repoCodePath := filepath.Join(codePath, repo.Name)
		if err := CloneRepo(repo.URL, repoCodePath, token); err != nil {
			m.mu.RUnlock()
			os.RemoveAll(jobPath)
			log.Printf("ERROR SubmitJob: failed to clone repo %s: %v", repo.Name, err)
			http.Error(w, fmt.Sprintf("failed to clone repo %s: %v", repo.Name, err), http.StatusInternalServerError)
			return
		}
	}
	m.mu.RUnlock()

	type RepoMount struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Path   string `json:"path"`
		Branch string `json:"branch"`
	}
	repoMounts := make([]RepoMount, len(selectedRepos))
	for i, repo := range selectedRepos {
		repoMounts[i] = RepoMount{
			ID:     repo.ID,
			Name:   repo.Name,
			Path:   fmt.Sprintf("/code/%s", repo.Name),
			Branch: repo.Branch,
		}
	}
	repoMountsJSON, err := json.Marshal(repoMounts)
	if err != nil {
		os.RemoveAll(jobPath)
		log.Printf("ERROR SubmitJob: failed to serialize repo mounts: %v", err)
		http.Error(w, "failed to serialize repo mounts", http.StatusInternalServerError)
		return
	}

	dataPath := filepath.Join(jobPath, "data")
	if err := writeIntentFile(dataPath, req.Intent, m.config.PromptsPath); err != nil {
		os.RemoveAll(jobPath)
		log.Printf("ERROR SubmitJob: failed to write INTENT.md: %v", err)
		http.Error(w, fmt.Sprintf("failed to write INTENT.md: %v", err), http.StatusInternalServerError)
		return
	}

	if err := writeTaskFile(dataPath, req.Task); err != nil {
		os.RemoveAll(jobPath)
		log.Printf("ERROR SubmitJob: failed to write TASK.md: %v", err)
		http.Error(w, fmt.Sprintf("failed to write TASK.md: %v", err), http.StatusInternalServerError)
		return
	}

	if err := writeContextFile(dataPath, m.config.ContextPath); err != nil {
		os.RemoveAll(jobPath)
		log.Printf("ERROR SubmitJob: failed to write CONTEXT.md: %v", err)
		http.Error(w, fmt.Sprintf("failed to write CONTEXT.md: %v", err), http.StatusInternalServerError)
		return
	}

	if err := writeSaveMemoryFile(dataPath, m.config.PromptsPath); err != nil {
		os.RemoveAll(jobPath)
		log.Printf("ERROR SubmitJob: failed to write UPDATE_MEMORY.md: %v", err)
		http.Error(w, fmt.Sprintf("failed to write UPDATE_MEMORY.md: %v", err), http.StatusInternalServerError)
		return
	}

	gitEmail := selectedRepos[0].GitEmail
	gitName := selectedRepos[0].GitName

	containerID, err := m.launchContainer(jobID, string(repoMountsJSON), req.Intent, image, model, provider, apiKey, gitEmail, gitName, jobPath)
	if err != nil {
		os.RemoveAll(jobPath)
		log.Printf("ERROR SubmitJob: failed to launch container: %v", err)
		http.Error(w, fmt.Sprintf("failed to launch container: %v", err), http.StatusInternalServerError)
		return
	}

	status := &JobStatus{
		JobID:       jobID,
		ContainerID: containerID,
		Status:      "running",
		Intent:      req.Intent,
		Image:       image,
		Task:        req.Task,
		RepoIDs:     req.RepoIDs,
		CreatedAt:   time.Now(),
		StartedAt:   time.Now(),
	}
	if err := writeStatusJSON(jobPath, status); err != nil {
		log.Printf("Warning: failed to write status.json for job %s: %v", jobID, err)
	}

	m.mu.Lock()
	m.jobs[jobID] = status
	m.mu.Unlock()

	go m.watchContainer(jobID, jobPath)

	log.Printf("Job %s started: container=%s image=%s repos=%v", jobID, containerID, image, req.RepoIDs)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(JobResponse{
		JobID:       jobID,
		Status:      "running",
		CreatedAt:   status.CreatedAt,
		ContainerID: containerID,
	})
}

// ListJobs handles GET /jobs — reads from disk
func (m *Manager) ListJobs(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(m.config.JobsBasePath)
	if err != nil {
		log.Printf("ERROR ListJobs: failed to read jobs directory: %v", err)
		http.Error(w, "failed to read jobs directory", http.StatusInternalServerError)
		return
	}

	jobs := make([]*JobStatus, 0)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		statusPath := filepath.Join(m.config.JobsBasePath, e.Name(), "status.json")
		data, err := os.ReadFile(statusPath)
		if err != nil {
			continue
		}
		var s JobStatus
		if err := json.Unmarshal(data, &s); err != nil {
			continue
		}
		jobs = append(jobs, &s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"jobs": jobs,
	})
}

// GetJob handles GET /jobs/{job_id} — reads from disk
func (m *Manager) GetJob(w http.ResponseWriter, r *http.Request) {
	jobID := mux.Vars(r)["job_id"]

	statusPath := filepath.Join(m.config.JobsBasePath, jobID, "status.json")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		log.Printf("ERROR GetJob: failed to read job status for %s: %v", jobID, err)
		http.Error(w, "failed to read job status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// PatchJob handles PATCH /jobs/{job_id}
func (m *Manager) PatchJob(w http.ResponseWriter, r *http.Request) {
	jobID := mux.Vars(r)["job_id"]

	var patchReq map[string]string
	if err := json.NewDecoder(r.Body).Decode(&patchReq); err != nil {
		log.Printf("ERROR PatchJob: invalid request body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	action, ok := patchReq["action"]
	if !ok {
		http.Error(w, "missing action field", http.StatusBadRequest)
		return
	}

	m.mu.RLock()
	job, exists := m.jobs[jobID]
	m.mu.RUnlock()

	if !exists {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	if action == "cancel" {
		if job.Status != "running" {
			http.Error(w, "can only cancel running jobs", http.StatusConflict)
			return
		}

		if err := m.stopContainer(job.ContainerID); err != nil {
			log.Printf("ERROR PatchJob: failed to stop container for job %s: %v", jobID, err)
			http.Error(w, fmt.Sprintf("failed to stop container: %v", err), http.StatusInternalServerError)
			return
		}

		m.mu.Lock()
		job.Status = "cancelled"
		job.CompletedAt = time.Now()
		m.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(job)
		return
	}

	http.Error(w, fmt.Sprintf("unknown action: %s", action), http.StatusBadRequest)
}

// DeleteJob handles DELETE /jobs/{job_id}
func (m *Manager) DeleteJob(w http.ResponseWriter, r *http.Request) {
	jobID := mux.Vars(r)["job_id"]

	m.mu.RLock()
	job, exists := m.jobs[jobID]
	m.mu.RUnlock()

	if exists && job.Status == "running" {
		http.Error(w, "cannot delete a running job", http.StatusConflict)
		return
	}

	m.mu.Lock()
	if cancel, ok := m.timers[jobID]; ok {
		cancel()
		delete(m.timers, jobID)
	}
	delete(m.jobs, jobID)
	m.mu.Unlock()

	if exists {
		ctx := context.Background()
		m.docker.ContainerRemove(ctx, job.ContainerID, client.ContainerRemoveOptions{
			Force: true,
		})
	}

	jobPath := filepath.Join(m.config.JobsBasePath, jobID)
	if err := os.RemoveAll(jobPath); err != nil {
		log.Printf("ERROR DeleteJob: failed to delete job directory %s: %v", jobID, err)
		http.Error(w, fmt.Sprintf("failed to delete job directory: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Job %s deleted", jobID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"job_id": jobID,
		"status": "deleted",
	})
}
