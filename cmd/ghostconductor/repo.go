package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// GetRepoTree handles GET /repo/tree?repo_id={id}
func (m *Manager) GetRepoTree(w http.ResponseWriter, r *http.Request) {
	repoID := r.URL.Query().Get("repo_id")
	jobID := r.URL.Query().Get("job_id")

	var root string
	if jobID != "" {
		root = filepath.Join(m.config.JobsBasePath, jobID, "code")
		if repoID != "" {
			repos, _ := m.readRepos()
			for _, repo := range repos {
				if repo.ID == repoID {
					root = filepath.Join(m.config.JobsBasePath, jobID, "code", repo.Name)
					break
				}
			}
		}
	} else if repoID != "" {
		repos, err := m.readRepos()
		if err != nil {
			http.Error(w, "failed to read repos", http.StatusInternalServerError)
			return
		}
		found := false
		for _, repo := range repos {
			if repo.ID == repoID {
				root = filepath.Join(m.config.ReposBasePath, repo.Name)
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "repo not found", http.StatusNotFound)
			return
		}
	} else {
		http.Error(w, "repo_id or job_id required", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(root); os.IsNotExist(err) {
		http.Error(w, "repo not found on disk", http.StatusNotFound)
		return
	}

	tree, err := buildFileTree(root, root)
	if err != nil {
		http.Error(w, "failed to build file tree", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
}

// GetRepoFile handles GET /repo/file?repo_id={id}&path={path}
func (m *Manager) GetRepoFile(w http.ResponseWriter, r *http.Request) {
	repoID := r.URL.Query().Get("repo_id")
	jobID := r.URL.Query().Get("job_id")
	filePath := r.URL.Query().Get("path")

	if filePath == "" {
		http.Error(w, "missing path param", http.StatusBadRequest)
		return
	}

	var root string
	if jobID != "" {
		root = filepath.Join(m.config.JobsBasePath, jobID, "code")
		if repoID != "" {
			repos, _ := m.readRepos()
			for _, repo := range repos {
				if repo.ID == repoID {
					root = filepath.Join(m.config.JobsBasePath, jobID, "code", repo.Name)
					break
				}
			}
		}
	} else if repoID != "" {
		repos, err := m.readRepos()
		if err != nil {
			http.Error(w, "failed to read repos", http.StatusInternalServerError)
			return
		}
		found := false
		for _, repo := range repos {
			if repo.ID == repoID {
				root = filepath.Join(m.config.ReposBasePath, repo.Name)
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "repo not found", http.StatusNotFound)
			return
		}
	} else {
		http.Error(w, "repo_id or job_id required", http.StatusBadRequest)
		return
	}

	abs := filepath.Join(root, filePath)
	if !strings.HasPrefix(abs, root) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	content, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write(content)
}

// GetJobTree handles GET /jobs/{job_id}/tree
func (m *Manager) GetJobTree(w http.ResponseWriter, r *http.Request) {
	jobID := mux.Vars(r)["job_id"]
	root := filepath.Join(m.config.JobsBasePath, jobID, "code")

	if _, err := os.Stat(root); os.IsNotExist(err) {
		http.Error(w, "job code not found", http.StatusNotFound)
		return
	}

	tree, err := buildFileTree(root, root)
	if err != nil {
		http.Error(w, "failed to build file tree", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
}

// buildFileTree recursively builds a FileNode tree, skipping .git
func buildFileTree(root, current string) (*FileNode, error) {
	info, err := os.Stat(current)
	if err != nil {
		return nil, err
	}

	rel, _ := filepath.Rel(root, current)
	node := &FileNode{
		Name:  info.Name(),
		Path:  rel,
		IsDir: info.IsDir(),
	}

	if !info.IsDir() {
		return node, nil
	}

	entries, err := os.ReadDir(current)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		child, err := buildFileTree(root, filepath.Join(current, e.Name()))
		if err != nil {
			continue
		}
		node.Children = append(node.Children, child)
	}

	return node, nil
}
