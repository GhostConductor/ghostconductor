package main

import (
	"fmt"
	"net/url"
	"os/exec"
)

// CloneRepo clones the target repo into destPath
func CloneRepo(repoURL, destPath, githubToken string) error {
	cloneURL := repoURL
	if githubToken != "" {
		u, err := url.Parse(repoURL)
		if err != nil {
			return fmt.Errorf("failed to parse repo URL: %w", err)
		}
		u.User = url.UserPassword(githubToken, "")
		cloneURL = u.String()
	}

	cmd := exec.Command("git", "clone", cloneURL, destPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s: %w", string(output), err)
	}
	return nil
}

// gitCmd returns an exec.Cmd for a git command run in the given directory
func gitCmd(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	return cmd
}

// DestroyRepo is a no-op — os.RemoveAll in DeleteJob handles cleanup
func DestroyRepo(jobID string) error {
	return nil
}
