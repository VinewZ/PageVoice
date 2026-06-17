# PageVoice

Give voice to your texts. A desktop application for importing PDF/EPUB/TXT documents, splitting them into sentences, and (future) generating TTS audio.

Built with [Wails v3](https://v3.wails.io/) (Go + React).

## Prerequisites

- Go 1.25+
- [Bun](https://bun.sh/)
- Wails v3 CLI: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`

## Development

```bash
task dev
```

This starts Vite HMR on port 9245 and a hot-reload Go backend.

## Build

```bash
task build
```

Produces a binary in `bin/`.

### Server mode (no GUI)

```bash
task build:server
task run:server
```

## Architecture

- **Go backend**: single `textupload` service (`internal/services/textupload/`) — file extraction, sentence splitting (17 languages), XDG persistence
- **Frontend**: React 19 + TanStack Router + Tailwind v4 + Biome + shadcn (base-ui variant)
- **Data**: stored under `$XDG_DATA_HOME/page-voice/` — see `AGENTS.md` for layout

## Commands

| Command | Purpose |
|---|---|
| `task dev` | Dev server with HMR |
| `task build` | Production build |
| `bun run check` | Biome lint + format check |
| `bun run build:dev` | `tsc` + Vite dev build (run before PRs) |
| `go vet ./...` | Go static analysis |
