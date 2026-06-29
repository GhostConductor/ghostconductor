package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// GetUsage handles GET /usage
func (m *Manager) GetUsage(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(m.config.UsagePath)
	if os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]UsageEntry{})
		return
	}
	if err != nil {
		http.Error(w, "failed to read usage file", http.StatusInternalServerError)
		return
	}

	var entries []UsageEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		http.Error(w, "failed to parse usage file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// ClearUsage handles DELETE /usage — wipes usage.json entirely
func (m *Manager) ClearUsage(w http.ResponseWriter, r *http.Request) {
	if err := os.WriteFile(m.config.UsagePath, []byte("[]"), 0666); err != nil {
		http.Error(w, "failed to clear usage", http.StatusInternalServerError)
		return
	}
	log.Printf("Usage cleared")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}

// filterUsageByProvider removes entries for a given provider from usage.json
func (m *Manager) filterUsageByProvider(provider string) {
	data, err := os.ReadFile(m.config.UsagePath)
	if err != nil {
		return
	}
	var entries []UsageEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return
	}
	filtered := make([]UsageEntry, 0)
	for _, e := range entries {
		if e.Provider != provider {
			filtered = append(filtered, e)
		}
	}
	out, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(m.config.UsagePath, out, 0666)
}

// helpers
func splitLines(s string) []string {
	var lines []string
	for _, line := range splitNewlines(s) {
		if line != "" && line[0] != '#' {
			lines = append(lines, line)
		}
	}
	return lines
}

func splitNewlines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func getValue(s string) string {
	for i, c := range s {
		if c == '=' {
			return s[i+1:]
		}
	}
	return ""
}

// ClearProviders handles DELETE /providers — clears all provider API keys from memory
func (m *Manager) ClearProviders(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.config.AnthropicAPIKey = ""
	m.config.OpenAIAPIKey = ""
	m.config.GoogleAPIKey = ""
	m.mu.Unlock()

	log.Printf("All provider keys cleared")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}
