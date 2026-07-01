package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// GetConfig handles GET /config
func (m *Manager) GetConfig(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"anthropic_api_key": m.config.AnthropicAPIKey != "",
		"openai_api_key":    m.config.OpenAIAPIKey != "",
		"google_api_key":    m.config.GoogleAPIKey != "",
	})
}

// GetIntents handles GET /intents — scans prompts/intent dir for *.md files
func (m *Manager) GetIntents(w http.ResponseWriter, r *http.Request) {
	intentPath := filepath.Join(m.config.PromptsPath, "intent")
	entries, err := os.ReadDir(intentPath)
	if err != nil {
		http.Error(w, "failed to read intents directory", http.StatusInternalServerError)
		return
	}

	intents := []string{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			intents = append(intents, strings.TrimSuffix(e.Name(), ".md"))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"intents": intents})
}

// GetJobLogs handles GET /jobs/{job_id}/logs
func (m *Manager) GetJobLogs(w http.ResponseWriter, r *http.Request) {
	jobID := mux.Vars(r)["job_id"]
	logPath := filepath.Join(m.config.JobsBasePath, jobID, "log", "gc-ghost.log")

	content, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "log not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to read log", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}

// GetJobEvents handles GET /jobs/{job_id}/events
func (m *Manager) GetJobEvents(w http.ResponseWriter, r *http.Request) {
	jobID := mux.Vars(r)["job_id"]
	eventsPath := filepath.Join(m.config.JobsBasePath, jobID, "data", "events.json")

	content, err := os.ReadFile(eventsPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "events not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to read events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
}

// GetJobStatus handles GET /jobs/{job_id}/status — reads from disk
func (m *Manager) GetJobStatus(w http.ResponseWriter, r *http.Request) {
	jobID := mux.Vars(r)["job_id"]
	statusPath := filepath.Join(m.config.JobsBasePath, jobID, "status.json")

	content, err := os.ReadFile(statusPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "status not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to read status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
}

// GetJobResult handles GET /jobs/{job_id}/result
func (m *Manager) GetJobResult(w http.ResponseWriter, r *http.Request) {
	jobID := mux.Vars(r)["job_id"]
	resultPath := filepath.Join(m.config.JobsBasePath, jobID, "data", "result.json")

	content, err := os.ReadFile(resultPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "result not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to read result", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(content)
}

// GetMemory handles GET /memory
func (m *Manager) GetMemory(w http.ResponseWriter, r *http.Request) {
	memPath := filepath.Join(m.config.BasePath, "shared", "memory.md")
	content, err := os.ReadFile(memPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte(""))
			return
		}
		http.Error(w, "failed to read memory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}

// PutMemory handles PUT /memory
func (m *Manager) PutMemory(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	memPath := filepath.Join(m.config.BasePath, "shared", "memory.md")
	if err := os.MkdirAll(filepath.Dir(memPath), 0755); err != nil {
		http.Error(w, "failed to create memory directory", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(memPath, body, 0644); err != nil {
		http.Error(w, "failed to write memory", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ClearMemory handles DELETE /memory
func (m *Manager) ClearMemory(w http.ResponseWriter, r *http.Request) {
	memPath := filepath.Join(m.config.BasePath, "shared", "memory.md")
	if err := os.Remove(memPath); err != nil && !os.IsNotExist(err) {
		http.Error(w, "failed to clear memory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}
