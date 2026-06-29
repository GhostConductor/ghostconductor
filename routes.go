package main

import (
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, m *Manager) {
	api := r.PathPrefix("/api/v1").Subrouter()

	// Health
	api.HandleFunc("/health", m.Health).Methods("GET")

	// Config
	api.HandleFunc("/config", m.GetConfig).Methods("GET")
	api.HandleFunc("/config", m.UpdateConfig).Methods("POST")
	api.HandleFunc("/config", m.DeleteConfig).Methods("DELETE")

	// Intents
	api.HandleFunc("/intents", m.GetIntents).Methods("GET")

	// Factory Reset
	api.HandleFunc("/factory-reset", m.FactoryReset).Methods("POST")

	// Memory
	api.HandleFunc("/memory", m.GetMemory).Methods("GET")
	api.HandleFunc("/memory", m.PutMemory).Methods("PUT")
	api.HandleFunc("/memory", m.ClearMemory).Methods("DELETE")

	// Context
	api.HandleFunc("/context", m.GetContext).Methods("GET")
	api.HandleFunc("/context", m.PutContext).Methods("PUT")
	api.HandleFunc("/context/templates", m.GetContextTemplates).Methods("GET")
	api.HandleFunc("/context/load", m.LoadContextTemplate).Methods("POST")

	// Repos
	api.HandleFunc("/repos", m.GetRepos).Methods("GET")
	api.HandleFunc("/repos", m.CreateRepo).Methods("POST")
	api.HandleFunc("/repos", m.ClearRepos).Methods("DELETE")
	api.HandleFunc("/repos/{id}", m.UpdateRepo).Methods("PUT")
	api.HandleFunc("/repos/{id}", m.DeleteRepo).Methods("DELETE")
	api.HandleFunc("/repos/{id}/token", m.SetRepoToken).Methods("POST")
	api.HandleFunc("/repos/{id}/token", m.DeleteRepoToken).Methods("DELETE")
	api.HandleFunc("/repos/{id}/pull", m.PullRepo).Methods("POST")

	// Repo file viewer
	api.HandleFunc("/repo/tree", m.GetRepoTree).Methods("GET")
	api.HandleFunc("/repo/file", m.GetRepoFile).Methods("GET")

	// Usage
	api.HandleFunc("/usage", m.GetUsage).Methods("GET")
	api.HandleFunc("/usage", m.ClearUsage).Methods("DELETE")
	api.HandleFunc("/providers", m.ClearProviders).Methods("DELETE")

	// Job sub-resources
	api.HandleFunc("/jobs/clear", m.ClearJobs).Methods("POST")
	api.HandleFunc("/jobs/{job_id}/logs", m.GetJobLogs).Methods("GET")
	api.HandleFunc("/jobs/{job_id}/events", m.GetJobEvents).Methods("GET")
	api.HandleFunc("/jobs/{job_id}/status", m.GetJobStatus).Methods("GET")
	api.HandleFunc("/jobs/{job_id}/result", m.GetJobResult).Methods("GET")
	api.HandleFunc("/jobs/{job_id}/tree", m.GetJobTree).Methods("GET")

	// Jobs
	api.HandleFunc("/jobs", m.ListJobs).Methods("GET")
	api.HandleFunc("/jobs", m.SubmitJob).Methods("POST")
	api.HandleFunc("/jobs/{job_id}", m.GetJob).Methods("GET")
	api.HandleFunc("/jobs/{job_id}", m.PatchJob).Methods("PATCH")
	api.HandleFunc("/jobs/{job_id}", m.DeleteJob).Methods("DELETE")

	// UI — serve React app (catch-all)
	r.PathPrefix("/").Handler(uiHandler())
}
