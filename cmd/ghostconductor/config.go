package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var secretKeys = map[string]bool{
	"GC_ANTHROPIC_API_KEY": true,
	"GC_OPENAI_API_KEY":    true,
	"GC_GOOGLE_API_KEY":    true,
}

type ConfigUpdateRequest struct {
	AnthropicAPIKey string `json:"anthropic_api_key,omitempty"`
	OpenAIAPIKey    string `json:"openai_api_key,omitempty"`
	GoogleAPIKey    string `json:"google_api_key,omitempty"`
}

func (m *Manager) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req ConfigUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	updates := map[string]string{}
	if req.AnthropicAPIKey != "" {
		updates["GC_ANTHROPIC_API_KEY"] = req.AnthropicAPIKey
	}
	if req.OpenAIAPIKey != "" {
		updates["GC_OPENAI_API_KEY"] = req.OpenAIAPIKey
	}
	if req.GoogleAPIKey != "" {
		updates["GC_GOOGLE_API_KEY"] = req.GoogleAPIKey
	}

	if len(updates) == 0 {
		http.Error(w, "no fields to update", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	for k, v := range updates {
		switch k {
		case "GC_ANTHROPIC_API_KEY":
			m.config.AnthropicAPIKey = v
		case "GC_OPENAI_API_KEY":
			m.config.OpenAIAPIKey = v
		case "GC_GOOGLE_API_KEY":
			m.config.GoogleAPIKey = v
		}
	}
	m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (m *Manager) DeleteConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	switch req.Key {
	case "anthropic_api_key":
		m.config.AnthropicAPIKey = ""
		m.mu.Unlock()
		m.filterUsageByProvider("anthropic")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	case "openai_api_key":
		m.config.OpenAIAPIKey = ""
		m.mu.Unlock()
		m.filterUsageByProvider("openai")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	case "google_api_key":
		m.config.GoogleAPIKey = ""
		m.mu.Unlock()
		m.filterUsageByProvider("google")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		return
	default:
		m.mu.Unlock()
		http.Error(w, "unknown key: "+req.Key, http.StatusBadRequest)
		return
	}
}

func updateEnvFile(path string, updates map[string]string) error {
	existing := map[string]string{}
	order := []string{}

	if data, err := os.ReadFile(path); err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || strings.HasPrefix(line, "#") {
				order = append(order, line)
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				existing[parts[0]] = parts[1]
				order = append(order, parts[0])
			}
		}
	}

	for k, v := range updates {
		if _, found := existing[k]; !found {
			order = append(order, k)
		}
		existing[k] = v
	}

	var b strings.Builder
	seen := map[string]bool{}
	for _, key := range order {
		if key == "" || strings.HasPrefix(key, "#") {
			b.WriteString(key + "\n")
			continue
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		b.WriteString(fmt.Sprintf("%s=%s\n", key, existing[key]))
	}

	return os.WriteFile(path, []byte(b.String()), 0600)
}

func (m *Manager) ClearJobs(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	for jobID, job := range m.jobs {
		if job.Status == "running" {
			if err := m.stopContainer(job.ContainerID); err != nil {
				log.Printf("Warning: failed to stop container for job %s: %v", jobID, err)
			}
		}
		if cancel, ok := m.timers[jobID]; ok {
			cancel()
		}
	}
	m.jobs = make(map[string]*JobStatus)
	m.timers = make(map[string]context.CancelFunc)
	m.mu.Unlock()

	entries, err := os.ReadDir(m.config.JobsBasePath)
	if err != nil && !os.IsNotExist(err) {
		http.Error(w, "failed to read jobs directory", http.StatusInternalServerError)
		return
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(m.config.JobsBasePath, e.Name())); err != nil {
			log.Printf("Warning: failed to delete job dir %s: %v", e.Name(), err)
		}
	}

	log.Printf("All jobs cleared")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}

func (m *Manager) FactoryReset(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	for jobID, job := range m.jobs {
		if job.Status == "running" {
			if err := m.stopContainer(job.ContainerID); err != nil {
				log.Printf("Warning: failed to stop container for job %s: %v", jobID, err)
			}
		}
		if cancel, ok := m.timers[jobID]; ok {
			cancel()
		}
	}
	m.jobs = make(map[string]*JobStatus)
	m.timers = make(map[string]context.CancelFunc)
	m.repoTokens = make(map[string]string)
	m.mu.Unlock()

	entries, err := os.ReadDir(m.config.JobsBasePath)
	if err != nil && !os.IsNotExist(err) {
		http.Error(w, "failed to read jobs directory", http.StatusInternalServerError)
		return
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(m.config.JobsBasePath, e.Name())); err != nil {
			log.Printf("Warning: failed to delete job dir %s: %v", e.Name(), err)
		}
	}

	m.mu.Lock()
	m.config.AnthropicAPIKey = ""
	m.config.OpenAIAPIKey = ""
	m.config.GoogleAPIKey = ""
	m.mu.Unlock()

	if err := os.WriteFile(m.config.UsagePath, []byte("[]"), 0666); err != nil {
		log.Printf("Warning: failed to clear usage: %v", err)
	}

	memPath := filepath.Join(m.config.BasePath, "shared", "memory.md")
	if err := os.Remove(memPath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: failed to clear memory: %v", err)
	}

	if err := os.WriteFile(m.config.ReposPath, []byte("[]"), 0644); err != nil {
		log.Printf("Warning: failed to clear repos: %v", err)
	}

	log.Printf("Factory reset complete")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "reset"})
}
