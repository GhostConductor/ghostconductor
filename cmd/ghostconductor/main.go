package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/moby/moby/client"
)

const defaultPort = "7777"
const configDir = ".ghostconductor"

var defaultImage = "ghcr.io/ghostconductor/ghost:dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "uninstall" {
		runUninstall()
		return
	}

	portFlag := flag.String("port", "", "Port to listen on (default 7777)")
	flag.Parse()

	cfg := loadConfig(portFlag)

	if err := precheck(cfg); err != nil {
		os.Exit(1)
	}

	docker, err := connectDocker()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer docker.Close()

	pullGhostImage(docker, defaultImage)

	mgr := &Manager{
		docker:     docker,
		config:     cfg,
		jobs:       make(map[string]*JobStatus),
		timers:     make(map[string]context.CancelFunc),
		repoTokens: make(map[string]string),
	}

	router := mux.NewRouter()
	router.Use(Recovery)
	router.Use(RequestLogger)
	RegisterRoutes(router, mgr)

	url := fmt.Sprintf("http://localhost:%s", cfg.Port)
	log.Printf("ghostconductor starting on %s", url)
	log.Printf("  base path:    %s", cfg.BasePath)
	log.Printf("  jobs path:    %s", cfg.JobsBasePath)
	log.Printf("  prompts path: %s", cfg.PromptsPath)
	log.Printf("  repos path:   %s", cfg.ReposPath)
	log.Printf("  usage path:   %s", cfg.UsagePath)

	openBrowser(url)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func precheck(cfg Config) error {
	dirs := []string{
		cfg.JobsBasePath,
		cfg.ReposBasePath,
		cfg.PromptsPath,
		filepath.Dir(cfg.ContextPath),
	}

	var missing []string
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			missing = append(missing, dir)
		}
	}

	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "GhostConductor: required directories are missing:\n")
		for _, dir := range missing {
			fmt.Fprintf(os.Stderr, "  missing: %s\n", dir)
		}
		return fmt.Errorf("preflight check failed")
	}

	return nil
}

func defaultBasePath() string {
	if v := os.Getenv("GC_BASE_PATH"); v != "" {
		return v
	}
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "ghostconductor")
	default:
		return "/opt/ghostconductor"
	}
}

func connectDocker() (*client.Client, error) {
	home, _ := os.UserHomeDir()
	sockets := []string{
		os.Getenv("GC_DOCKER_SOCKET"),
		filepath.Join(home, ".rd", "docker.sock"),
		"/var/run/docker.sock",
	}

	for _, sock := range sockets {
		if sock == "" {
			continue
		}
		if _, err := os.Stat(sock); err != nil {
			continue
		}
		docker, err := client.NewClientWithOpts(
			client.WithHost("unix://"+sock),
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			continue
		}
		if _, err := docker.Ping(context.Background(), client.PingOptions{}); err != nil {
			continue
		}
		log.Printf("Connected to Docker at %s", sock)
		return docker, nil
	}

	return nil, fmt.Errorf(`No Docker daemon found.

	GhostConductor requires Rancher Desktop.
	Download it at https://rancherdesktop.io

	If Rancher Desktop is already installed, make sure it is running and try again.

	Using Docker Desktop instead? Set GC_DOCKER_SOCKET=/var/run/docker.sock and try again.`)
}

func loadConfig(portFlag *string) Config {
	base := defaultBasePath()

	port := defaultPort
	if portFlag != nil && *portFlag != "" {
		port = *portFlag
	} else if v := os.Getenv("GC_PORT"); v != "" {
		port = v
	}

	return Config{
		Port:            port,
		DockerSocket:    envOrDefault("GC_DOCKER_SOCKET", "/var/run/docker.sock"),
		BasePath:        base,
		JobsBasePath:    filepath.Join(base, "jobs"),
		ReposBasePath:   filepath.Join(base, "repos"),
		PromptsPath:     filepath.Join(base, "etc", "prompts"),
		ContextPath:     filepath.Join(base, "etc", "CONTEXT.md"),
		ReposPath:       filepath.Join(base, "etc", "repos.json"),
		UsagePath:       filepath.Join(base, "shared", "usage.json"),
		GracePeriodSecs: envOrDefaultInt("GC_GRACE_PERIOD_SECS", 60),
		AnthropicAPIKey: os.Getenv("GC_ANTHROPIC_API_KEY"),
		OpenAIAPIKey:    os.Getenv("GC_OPENAI_API_KEY"),
		GoogleAPIKey:    os.Getenv("GC_GOOGLE_API_KEY"),
	}
}

func openBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	go func() {
		if err := exec.Command(cmd, args...).Start(); err != nil {
			log.Printf("Failed to open browser: %v", err)
		}
	}()
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func pullGhostImage(docker *client.Client, image string) {
	ctx := context.Background()
	log.Printf("Pulling agent image: %s", image)
	out, err := docker.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		log.Printf("Warning: failed to pull agent image: %v", err)
		return
	}
	defer out.Close()
	io.Copy(io.Discard, out)
	log.Printf("Agent image ready: %s", image)
}

func runUninstall() {
	home, _ := os.UserHomeDir()
	basePath := defaultBasePath()

	fmt.Println("Ghost Conductor Uninstall")
	fmt.Println("─────────────────────────")
	fmt.Printf("This will remove:\n")
	fmt.Printf("  %s\n", basePath)
	fmt.Printf("  Docker image: ghcr.io/ghostconductor/ghost:latest\n")
	fmt.Printf("  Docker image: ghcr.io/ghostconductor/ghost:dev\n")
	fmt.Printf("  All ghost containers\n\n")
	fmt.Print("Are you sure? (yes/no): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "yes" {
		fmt.Println("Uninstall cancelled.")
		return
	}

	if err := os.RemoveAll(basePath); err != nil {
		fmt.Printf("Warning: failed to remove data dir: %v\n", err)
	} else {
		fmt.Printf("Removed: %s\n", basePath)
	}

	docker, err := connectDocker()
	if err != nil {
		fmt.Printf("Warning: could not connect to Docker — skipping container and image removal\n")
	} else {
		defer docker.Close()
		ctx := context.Background()

		containers, err := docker.ContainerList(ctx, client.ContainerListOptions{All: true})
		if err == nil {
			for _, c := range containers.Items {
				for _, name := range c.Names {
					if strings.HasPrefix(name, "/ghost-") {
						docker.ContainerRemove(ctx, c.ID, client.ContainerRemoveOptions{Force: true})
						fmt.Printf("Removed container: %s\n", name)
					}
				}
			}
		}

		for _, image := range []string{
			"ghcr.io/ghostconductor/ghost:latest",
			"ghcr.io/ghostconductor/ghost:dev",
		} {
			if _, err := docker.ImageRemove(ctx, image, client.ImageRemoveOptions{Force: true}); err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", image, err)
			} else {
				fmt.Printf("Removed: %s\n", image)
			}
		}
	}

	fmt.Println("\nDone. If installed via Homebrew, also run:")
	fmt.Println("  brew uninstall ghostconductor")
	_ = home
}
