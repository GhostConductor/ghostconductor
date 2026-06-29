package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/moby/moby/client"
)

// HealthResponse is the payload returned by GET /health
type HealthResponse struct {
	Status string `json:"status"`
	Docker string `json:"docker"`
	Uptime string `json:"uptime"`
}

var startTime = time.Now()

// Health handles GET /health
func (m *Manager) Health(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status: "ok",
		Uptime: time.Since(startTime).Truncate(time.Second).String(),
	}

	// Check Docker daemon
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	_, err := m.docker.Ping(ctx, client.PingOptions{})
	if err != nil {
		resp.Status = "degraded"
		resp.Docker = "unreachable"
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		resp.Docker = "ok"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
