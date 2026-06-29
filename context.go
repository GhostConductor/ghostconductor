package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// GetContext handles GET /context
func (m *Manager) GetContext(w http.ResponseWriter, r *http.Request) {
	content, err := os.ReadFile(m.config.ContextPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Write([]byte(""))
			return
		}
		http.Error(w, "failed to read context", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}

// PutContext handles PUT /context
func (m *Manager) PutContext(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(filepath.Dir(m.config.ContextPath), 0755); err != nil {
		http.Error(w, "failed to create context directory", http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(m.config.ContextPath, body, 0644); err != nil {
		http.Error(w, "failed to write context", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetContextTemplates handles GET /context/templates
func (m *Manager) GetContextTemplates(w http.ResponseWriter, r *http.Request) {
	templatesPath := filepath.Join(m.config.BasePath, "etc", "context")
	entries, err := os.ReadDir(templatesPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string][]string{"templates": {}})
			return
		}
		http.Error(w, "failed to read context templates", http.StatusInternalServerError)
		return
	}

	templates := []string{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			templates = append(templates, e.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"templates": templates})
}

// LoadContextTemplate handles POST /context/load
func (m *Manager) LoadContextTemplate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Template string `json:"template"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Template == "" {
		http.Error(w, "template is required", http.StatusBadRequest)
		return
	}

	templatePath := filepath.Join(m.config.BasePath, "etc", "context", req.Template)
	content, err := os.ReadFile(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "template not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to read template", http.StatusInternalServerError)
		return
	}

	contextPath := m.config.ContextPath
	if err := os.WriteFile(contextPath, content, 0644); err != nil {
		http.Error(w, "failed to write context", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "template": req.Template})
}
