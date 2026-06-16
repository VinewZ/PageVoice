# PageVoice — Agent Guide

## Commands

### Dev & build
| Command | What it does |
|---|---|
| `task dev` | Vite HMR + Go hot-reload |
| `task build` | Production build (frontend → Go binary → `bin/`) |
| `task run` | Run built binary |
| `task build:server` | Server-only (`-tags server`, no GUI) |
| `task run:server` | Run server binary |

### Frontend (in `frontend/`)
| Command | Tool |
|---|---|
| `bun install` | Install deps |
| `bun run dev` | Vite dev on :3000 |
| `bun run build` | Prod build |
| `bun run test` | Vitest |
| `bun run lint` | Biome lint |
| `bun run check` | Biome lint + format check |
| `bun run format` | Biome format (--write) |

### Go
- `go build .` — desktop app (requires `frontend/dist/` to exist)
- `go build -tags server` — server mode
- `go vet ./...` — passes

## Key structure

| Path | Role |
|---|---|
| `main.go` | Entry point. **Must stay at root** — `//go:embed all:frontend/dist` forbids `..` in paths |
| `internal/app/` | Wails app bootstrap (`app.Run(assets)`) |
| `internal/data/` | Embedded language JSON files |
| `frontend/` | React 19 + TanStack Router + Tailwind v4 + Vite 8 + Biome |
| `build/tasks/` | Split Taskfiles: `frontend.yml`, `tooling.yml`, `server.yml`, `docker.yml` |
| `build/{linux,darwin,windows}/` | Platform build configs (icons, manifests, .desktop, packaging) |

## Taskfile namespace map

Root → `common: ./build/Taskfile.yml` → delegates to:

- `common:frontend:*` — `build`, `dev`, `install:deps`
- `common:tooling:*` — `go:mod:tidy`, `generate:bindings`, `generate:icons`, `update:build-assets`
- `common:server:*` — `build`, `run`
- `common:docker:*` — `build`, `run`, `setup`

Platform Taskfiles (`build/{linux,darwin,windows}/Taskfile.yml`) include `common: ../Taskfile.yml` and reference the same `common:*` namespace.

## Generated files (do not edit)

- `frontend/src/routeTree.gen.ts` — TanStack Router codegen (`bun run generate-routes`)
- `frontend/bindings/` — Wails Go→TS bindings (`common:tooling:generate:bindings`)
- `frontend/dist/` — Vite output

## Frontend quirks

- Package manager is **bun**, not npm/pnpm
- Tailwind v4 uses Vite plugin directly — no `tailwind.config.js`
- Biome replaces both ESLint and Prettier
- Imports use `#/*` alias → `./src/*`
- TanStack Router devtools panel injected in `__root.tsx`
- No test files exist yet

## Wails-specific

- Dev mode config: `build/config.yml` (debounce, exec order, file watches)
- After changing `build/config.yml` `info`, run `common:tooling:update:build-assets` (overwrites manual changes)
- `build/config.yml` `info` fields still have placeholder values

## Not applicable

- **iOS**: removed, not a mobile app
- **CI/CD**: none configured
- **README**: still the default Wails template
