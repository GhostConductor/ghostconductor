package main

import (
	"fmt"
	"math/rand"
	"time"
)

// generateJobID generates a job ID in format: job-{timestamp}-{random}
func generateJobID() string {
	ts := time.Now().Format("20060102-150405")
	randomSuffix := fmt.Sprintf("%06x", rand.Int63()%0xffffff)
	return fmt.Sprintf("job-%s-%s", ts, randomSuffix)
}

// getAPIKeyForProvider returns the appropriate API key for a given provider
func getAPIKeyForProvider(provider string, cfg Config) string {
	switch provider {
	case "anthropic":
		return cfg.AnthropicAPIKey
	case "openai":
		return cfg.OpenAIAPIKey
	case "google":
		return cfg.GoogleAPIKey
	default:
		return ""
	}
}
