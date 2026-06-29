# ghostconductor

Ghost Conductor — run autonomous AI software engineering agents on your local machine.

Set your context, choose an intent, and describe the task. Support for Anthropic, OpenAI, and Google models — bring your own API keys. Mix and match depending on the task, or run them simultaneously on the same project. Track token usage and costs per job, per provider, and per model. Memory persists across runs and is fully editable, so your ghosts get smarter over time.

Each agent runs in an isolated Docker container on a dedicated branch, torn down after each job leaving only the code it wrote.

## Requirements

- macOS (Apple Silicon or Intel)
- [Rancher Desktop](https://rancherdesktop.io) — required for Docker support
  - Container engine must be set to **dockerd (moby)**

## Install

**Via Homebrew:**
```bash
# install 
brew tap ghostconductor/ghostconductor
brew install ghostconductor
# run
ghostconductor
```

**Or download the binary directly:**

[Download latest release](https://github.com/GhostConductor/ghostconductor/releases/latest)

Then run:
```bash
chmod +x ghostconductor
xattr -rd com.apple.quarantine ghostconductor
./ghostconductor
```

On first launch, ghostconductor will:
1. Ask where to store your data (default: `~/ghostconductor/`)
2. Check that Rancher Desktop is running
3. Pull the gc-ghost agent image from ghcr.io
4. Open the UI at `http://localhost:7777`

## Uninstall

```bash
brew uninstall ghostconductor
```

To also remove all data and Docker images:
```bash
rm -rf ~/.ghostconductor
rm -rf ~/ghostconductor
docker container prune -f
docker rmi ghcr.io/ghostconductor/gc-ghost:latest
docker rmi ghcr.io/ghostconductor/gc-ghost:dev
```

## Port

ghostconductor runs on port `7777` by default. To use a different port:

```bash
ghostconductor --port 8888
```

## Using Docker Desktop instead of Rancher Desktop

```bash
GC_DOCKER_SOCKET=/var/run/docker.sock ghostconductor
```

## Contributing

ghostconductor is part of the Ghost Conductor platform. At this stage we are not accepting outside PRs. Issues and feedback welcome.

## License

MIT
