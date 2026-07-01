package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

func (m *Manager) launchContainer(jobID, gcRepos, intent, image, model, provider, apiKey, gitEmail, gitName, jobPath string) (string, error) {
	ctx := context.Background()

	stopTimeout := m.config.GracePeriodSecs

	// Build capability drop list
	var capDrop []string
	for _, cap := range m.containerPolicy.DropCapabilities {
		capDrop = append(capDrop, cap)
	}

	// Memory limit in bytes
	memoryLimit := int64(m.containerPolicy.MemoryLimitMB) * 1024 * 1024

	// CPU limit (NanoCPUs = CPULimit * 1e9)
	nanoCPUs := int64(m.containerPolicy.CPULimit * 1e9)

	resp, err := m.docker.ContainerCreate(ctx, client.ContainerCreateOptions{
		Name: fmt.Sprintf("ghost-%s", jobID),
		Config: &container.Config{
			Image: image,
			Env: []string{
				"GC_JOB_ID=" + jobID,
				"GC_REPOS=" + gcRepos,
				"GC_INTENT=" + intent,
				"GC_TIME_LIMIT=1800",
				"GC_PROVIDER=" + provider,
				"GC_MODEL=" + model,
				"GC_USAGE_PATH=/shared/usage.json",
				"AI_API_KEY=" + apiKey,
				"ANTHROPIC_API_KEY=" + apiKey,
				"GC_GIT_EMAIL=" + gitEmail,
				"GC_GIT_NAME=" + gitName,
			},
			StopTimeout: &stopTimeout,
		},
		HostConfig: &container.HostConfig{
			Binds: []string{
				filepath.Join(jobPath, "code") + ":/code",
				filepath.Join(jobPath, "data") + ":/data",
				filepath.Join(jobPath, "log") + ":/log",
				filepath.Join(m.config.BasePath, "shared") + ":/shared",
			},
			Resources: container.Resources{
				Memory:    memoryLimit,
				NanoCPUs:  nanoCPUs,
				PidsLimit: &m.containerPolicy.PidsLimit,
			},
			CapDrop:        capDrop,
			ReadonlyRootfs: m.containerPolicy.ReadonlyRootfs,
			SecurityOpt:    securityOpts(m.containerPolicy.NoNewPrivileges),
			NetworkMode:    "ghostconductor",
		},
		NetworkingConfig: &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				"ghostconductor": {
					NetworkID: m.networkID,
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	if _, err := m.docker.ContainerStart(ctx, resp.ID, client.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return resp.ID, nil
}

func securityOpts(noNewPrivileges bool) []string {
	if noNewPrivileges {
		return []string{"no-new-privileges:true"}
	}
	return nil
}

func (m *Manager) stopContainer(containerID string) error {
	ctx := context.Background()
	stopTimeout := m.config.GracePeriodSecs

	if _, err := m.docker.ContainerStop(ctx, containerID, client.ContainerStopOptions{
		Timeout: &stopTimeout,
	}); err != nil {
		log.Printf("Failed to stop container %s gracefully: %v, force killing", containerID, err)
		if _, err := m.docker.ContainerKill(ctx, containerID, client.ContainerKillOptions{
			Signal: "SIGKILL",
		}); err != nil {
			return fmt.Errorf("failed to kill container: %w", err)
		}
	}
	return nil
}

func (m *Manager) watchContainer(jobID, jobPath string) {
	ctx, cancel := context.WithCancel(context.Background())

	m.mu.Lock()
	m.timers[jobID] = cancel
	containerID := m.jobs[jobID].ContainerID
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.timers, jobID)
		m.mu.Unlock()
	}()

	f := client.Filters{}
	f.Add("container", containerID)
	f.Add("type", "container")

	eventsResult := m.docker.Events(ctx, client.EventsListOptions{
		Filters: f,
	})

	timeLimitTimer := time.NewTimer(30 * time.Minute)
	defer timeLimitTimer.Stop()

	timedOut := false

	for {
		select {
		case event := <-eventsResult.Messages:
			if event.Action == "die" {
				status := "completed"
				exitCode := 0

				if code, ok := event.Actor.Attributes["exitCode"]; ok {
					if c, err := strconv.Atoi(code); err == nil {
						exitCode = c
					}
				}

				if timedOut {
					status = "timed_out"
				} else if exitCode != 0 {
					status = "failed"
				}

				m.mu.Lock()
				m.jobs[jobID].Status = status
				m.jobs[jobID].CompletedAt = time.Now()
				m.jobs[jobID].ExitCode = exitCode
				m.mu.Unlock()

				m.mu.RLock()
				jobStatus := m.jobs[jobID]
				m.mu.RUnlock()
				if err := writeStatusJSON(jobPath, jobStatus); err != nil {
					log.Printf("Warning: failed to write status.json for job %s: %v", jobID, err)
				}

				cancel()
				return
			}

			if event.Action == "oom" {
				log.Printf("Container OOM for job %s", jobID)

				m.mu.Lock()
				m.jobs[jobID].Status = "crashed"
				m.jobs[jobID].CompletedAt = time.Now()
				m.mu.Unlock()

				cancel()
				return
			}

		case err := <-eventsResult.Err:
			if err != nil {
				log.Printf("Docker event stream error for job %s: %v", jobID, err)
			}
			cancel()
			return

		case <-timeLimitTimer.C:
			log.Printf("Time limit expired for job %s, stopping container", jobID)
			timedOut = true
			if err := m.stopContainer(containerID); err != nil {
				log.Printf("Failed to stop container for job %s: %v", jobID, err)
			}

		case <-ctx.Done():
			return
		}
	}
}
