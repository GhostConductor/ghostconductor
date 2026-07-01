package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

func generateRepoID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *Manager) readRepos() ([]Repo, error) {
	data, err := os.ReadFile(m.config.ReposPath)
	if os.IsNotExist(err) {
		return []Repo{}, nil
	}
	if err != nil {
		return nil, err
	}
	var repos []Repo
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func (m *Manager) writeRepos(repos []Repo) error {
	data, err := json.MarshalIndent(repos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.config.ReposPath, data, 0644)
}

// GetRepos handles GET /repos
func (m *Manager) GetRepos(w http.ResponseWriter, r *http.Request) {
	repos, err := m.readRepos()
	if err != nil {
		http.Error(w, "failed to read repos", http.StatusInternalServerError)
		return
	}

	m.mu.RLock()
	result := make([]RepoStatus, len(repos))
	for i, repo := range repos {
		result[i] = RepoStatus{
			Repo:     repo,
			TokenSet: m.repoTokens[repo.ID] != "",
		}
	}
	m.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CreateRepo handles POST /repos
func (m *Manager) CreateRepo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Branch   string `json:"branch"`
		GitEmail string `json:"git_email"`
		GitName  string `json:"git_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.URL == "" {
		http.Error(w, "name and url are required", http.StatusBadRequest)
		return
	}
	if req.Branch == "" {
		req.Branch = "main"
	}

	repos, err := m.readRepos()
	if err != nil {
		http.Error(w, "failed to read repos", http.StatusInternalServerError)
		return
	}

	repo := Repo{
		ID:       generateRepoID(),
		Name:     req.Name,
		URL:      req.URL,
		Branch:   req.Branch,
		GitEmail: req.GitEmail,
		GitName:  req.GitName,
	}
	repos = append(repos, repo)

	if err := m.writeRepos(repos); err != nil {
		http.Error(w, "failed to write repos", http.StatusInternalServerError)
		return
	}

	log.Printf("Repo created: %s (%s)", repo.Name, repo.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(repo)
}

// UpdateRepo handles PUT /repos/{id}
func (m *Manager) UpdateRepo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Branch   string `json:"branch"`
		GitEmail string `json:"git_email"`
		GitName  string `json:"git_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	repos, err := m.readRepos()
	if err != nil {
		http.Error(w, "failed to read repos", http.StatusInternalServerError)
		return
	}

	found := false
	for i, repo := range repos {
		if repo.ID == id {
			if req.Name != "" {
				repos[i].Name = req.Name
			}
			if req.URL != "" {
				repos[i].URL = req.URL
			}
			if req.Branch != "" {
				repos[i].Branch = req.Branch
			}
			if req.GitEmail != "" {
				repos[i].GitEmail = req.GitEmail
			}
			if req.GitName != "" {
				repos[i].GitName = req.GitName
			}
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "repo not found", http.StatusNotFound)
		return
	}

	if err := m.writeRepos(repos); err != nil {
		http.Error(w, "failed to write repos", http.StatusInternalServerError)
		return
	}

	log.Printf("Repo updated: %s", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DeleteRepo handles DELETE /repos/{id}
func (m *Manager) DeleteRepo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	repos, err := m.readRepos()
	if err != nil {
		http.Error(w, "failed to read repos", http.StatusInternalServerError)
		return
	}

	filtered := make([]Repo, 0)
	found := false
	for _, repo := range repos {
		if repo.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, repo)
	}

	if !found {
		http.Error(w, "repo not found", http.StatusNotFound)
		return
	}

	if err := m.writeRepos(filtered); err != nil {
		http.Error(w, "failed to write repos", http.StatusInternalServerError)
		return
	}

	m.mu.Lock()
	delete(m.repoTokens, id)
	m.mu.Unlock()

	log.Printf("Repo deleted: %s", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

// SetRepoToken handles POST /repos/{id}/token
func (m *Manager) SetRepoToken(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	m.mu.Lock()
	m.repoTokens[id] = req.Token
	m.mu.Unlock()

	log.Printf("Repo token set: %s", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// DeleteRepoToken handles DELETE /repos/{id}/token
func (m *Manager) DeleteRepoToken(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	m.mu.Lock()
	delete(m.repoTokens, id)
	m.mu.Unlock()

	log.Printf("Repo token cleared: %s", id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// PullRepo handles POST /repos/{id}/pull
func (m *Manager) PullRepo(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	repos, err := m.readRepos()
	if err != nil {
		http.Error(w, "failed to read repos", http.StatusInternalServerError)
		return
	}

	var repo *Repo
	for i := range repos {
		if repos[i].ID == id {
			repo = &repos[i]
			break
		}
	}
	if repo == nil {
		http.Error(w, "repo not found", http.StatusNotFound)
		return
	}

	m.mu.RLock()
	token := m.repoTokens[id]
	m.mu.RUnlock()

	if token == "" {
		http.Error(w, "no token set for repo", http.StatusBadRequest)
		return
	}

	repoPath := filepath.Join(m.config.ReposBasePath, repo.Name)
	if err := ensureRepo(repoPath, repo.URL, token, repo.Branch); err != nil {
		http.Error(w, fmt.Sprintf("failed to pull repo: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Repo pulled: %s", repo.Name)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ensureRepo clones the repo if it doesn't exist, or pulls if it does
func ensureRepo(repoPath, repoURL, githubToken, baseBranch string) error {
	if _, err := os.Stat(filepath.Join(repoPath, ".git")); err == nil {
		return pullRepo(repoPath, baseBranch)
	}
	return CloneRepo(repoURL, repoPath, githubToken)
}

// pullRepo checks out base branch and pulls latest
func pullRepo(repoPath, baseBranch string) error {
	if out, err := gitCmd(repoPath, "git", "checkout", baseBranch).CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout %s failed: %s: %w", baseBranch, string(out), err)
	}
	if out, err := gitCmd(repoPath, "git", "pull").CombinedOutput(); err != nil {
		return fmt.Errorf("git pull failed: %s: %w", string(out), err)
	}
	return nil
}

// ClearRepos handles DELETE /repos — deletes all repos and clears all tokens
func (m *Manager) ClearRepos(w http.ResponseWriter, r *http.Request) {
	if err := os.WriteFile(m.config.ReposPath, []byte("[]"), 0644); err != nil {
		http.Error(w, "failed to clear repos", http.StatusInternalServerError)
		return
	}

	m.mu.Lock()
	m.repoTokens = make(map[string]string)
	m.mu.Unlock()

	log.Printf("All repos cleared")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}
