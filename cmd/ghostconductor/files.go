package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeIntentFile(dataPath, intent, promptsPath string) error {
	promptFile := filepath.Join(promptsPath, "intent", fmt.Sprintf("%s.md", strings.ToUpper(intent)))
	content, err := os.ReadFile(promptFile)
	if err != nil {
		return fmt.Errorf("failed to read intent prompt from disk (%s): %w", promptFile, err)
	}
	return os.WriteFile(filepath.Join(dataPath, "INTENT.md"), content, 0644)
}

func writeTaskFile(dataPath, task string) error {
	return os.WriteFile(filepath.Join(dataPath, "TASK.md"), []byte(task), 0644)
}

func writeContextFile(dataPath, contextPath string) error {
	content, err := os.ReadFile(contextPath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(filepath.Join(dataPath, "CONTEXT.md"), []byte(""), 0644)
		}
		return fmt.Errorf("failed to read context file from disk (%s): %w", contextPath, err)
	}
	return os.WriteFile(filepath.Join(dataPath, "CONTEXT.md"), content, 0644)
}

func writeStatusJSON(jobPath string, status *JobStatus) error {
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(jobPath, "status.json"), data, 0644)
}

func writeSaveMemoryFile(dataPath, promptsPath string) error {
	promptFile := filepath.Join(promptsPath, "UPDATE_MEMORY.md")
	content, err := os.ReadFile(promptFile)
	if err != nil {
		return fmt.Errorf("failed to read UPDATE_MEMORY prompt (%s): %w", promptFile, err)
	}
	return os.WriteFile(filepath.Join(dataPath, "UPDATE_MEMORY.md"), content, 0644)
}
