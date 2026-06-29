package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

//go:embed prompts
var promptFiles embed.FS

//go:embed context
var contextFiles embed.FS

func copyPromptsToBase(basePath string) error {
	// Copy intent prompts
	if err := fs.WalkDir(promptFiles, "prompts", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel("prompts", path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(basePath, "etc", "prompts", relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		if _, err := os.Stat(destPath); err == nil {
			return nil
		}

		data, err := promptFiles.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return err
		}

		log.Printf("Installed prompt: %s", destPath)
		return nil
	}); err != nil {
		return err
	}

	// Copy context templates
	if err := fs.WalkDir(contextFiles, "context", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel("context", path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(basePath, "etc", "context", relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		if _, err := os.Stat(destPath); err == nil {
			return nil
		}

		data, err := contextFiles.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.WriteFile(destPath, data, 0644); err != nil {
			return err
		}

		log.Printf("Installed context template: %s", destPath)
		return nil
	}); err != nil {
		return err
	}

	// Copy TEMPLATE.md to etc/CONTEXT.md if it doesn't exist
	contextPath := filepath.Join(basePath, "etc", "CONTEXT.md")
	if _, err := os.Stat(contextPath); os.IsNotExist(err) {
		data, err := contextFiles.ReadFile("context/TEMPLATE.md")
		if err != nil {
			return fmt.Errorf("failed to read context template: %w", err)
		}
		if err := os.WriteFile(contextPath, data, 0644);
 err != nil {
			return fmt.Errorf("failed to write CONTEXT.md: %w", err)
		}
		log.Printf("Installed default context: %s", contextPath)
	}

	return nil
}
