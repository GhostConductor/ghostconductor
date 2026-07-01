# Project Context
This file is injected into every agent job as CONTEXT.md.

---

## Language & Runtime
- Language: Go (gc-desktop, gc-server), Python (gc-ghost), TypeScript (gc-web-ui)
- Go Version: 1.22
- Python Version: 3.12
- Node.js Version: 20

## Frameworks & Libraries
- Go: gorilla/mux, moby/moby/client (Docker), embed.FS
- Python: anthropic, openai, google-adk, anyio, httpx
- TypeScript: React 18, Vite, Tailwind CSS, React Router

## Project Structure
- gc-desktop/ — Go binary, embeds gc-web-ui, manages Docker containers and job lifecycle
- gc-web-ui/src/pages/ — React pages (JobsPage, RepoPage, ProviderPage, ContextPage, etc.)
- gc-web-ui/src/api.ts — all API calls to gc-desktop backend at /api/v1
- gc-web-ui/src/types.ts — shared TypeScript interfaces
- gc-ghost/src/ — Python agent runtime (main.py, config.py, runner.py, git.py, memory.py)
- prompts/intent/ — markdown intent prompts embedded in gc-desktop binary
- prompts/context/ — context templates embedded in gc-desktop binary

## Code Conventions
- Go: multi-file flat package main, handlers in handlers.go, routes in routes.go
- Go: all errors returned and logged before http.Error()
- Go: snake_case for JSON fields, CamelCase for Go structs
- TypeScript: functional components only, no class components
- TypeScript: all API calls go through src/api.ts
- Python: async/await with anyio, all file paths as pathlib.Path

## Success Criteria
- go build ./... passes with no errors
- npm run build passes with no TypeScript errors
- No new dependencies without discussion

## Constraints
- Do not modify existing API contracts at /api/v1
- Do not add AWS dependencies to gc-desktop or gc-ghost
- Do not hardcode secrets, tokens, or API keys
- Do not modify the Docker socket detection logic in main.go
- Prompts are read-only at runtime — do not write to prompts/
