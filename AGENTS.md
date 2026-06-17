# PageVoice — Agent Guide

## Commands

### Dev & build
| Command | What it does |
|---|---|
| `task dev` | Vite HMR + Go hot-reload |
| `task build` | Prod build → `bin/` |
| `task run` | Run built binary |
| `task build:server` | `-tags server` (no GUI, HTTP-only) |
| `task run:server` | Run server binary |

### Frontend (`frontend/`)
| Command | Tool |
|---|---|
| `bun install` | Install deps |
| `bun run dev` | Vite dev on :3000 |
| `bun run build` | Vite prod build |
| `bun run build:dev` | `tsc` + Vite dev build (run before PR) |
| `bun run lint` | Biome lint |
| `bun run check` | Biome lint + format check |
| `bun run test` | Vitest (no tests exist yet) |

### Go
- `go build .` — desktop app (requires `frontend/dist/`)
- `go build -tags server` — server mode
- `go vet ./...` — passes

## Architecture

**Single Go service** at `internal/services/textupload/` with 3 Wails-bound methods:
- `ProcessFile(fileName, base64Data, language) → UploadResult` — accepts file as base64, extracts text (PDF/EPUB/TXT), splits into sentences, persists to XDG, returns sentences + metadata
- `GetLibrary() → LibraryEntry[]` — reads `library.json` from XDG data dir
- `GetBook(dirName) → BookDetail` — reads `metadata.json` + `state.json` for a given book dir

**XDG data layout** (`$XDG_DATA_HOME/page-voice/`):
```
library.json  ← [{id, title, dirName}]
books/<sanitized-title-8c8f9c>/
  ├── original.txt    ← raw text
  ├── metadata.json   ← {title, author, language, sourceFile, importedAt}
  ├── state.json      ← {status, chunkLength:250, currentChunk, totalChunks, createdAt, updatedAt}
  └── audio/          ← future TTS WAV files
```

State statuses: `pending`, `running`, `paused`, `completed`, `failed`. `library.json` uses full language names (e.g. `"english"`), `sourceFile` stores upload filename only (no path).

## Frontend quirks

- **Bun** only, no npm/pnpm
- **Biome** replaces ESLint + Prettier. Run `bun run check` to validate. Don't fix lint errors in `src/components/ui/` (pre-existing from shadcn)
- **Tailwind v4** via Vite plugin — no `tailwind.config.js`
- **shadcn (base-ui variant)**: NO `asChild` prop on Button/DropdownMenuTrigger. Use `buttonVariants()` className on the native element instead
- **TanStack Router**: routes in `src/routes/`, codegen creates `src/routeTree.gen.ts` (`.gitignore`d, run `bun run generate-routes`)
- **Imports**: `#/*` → `./src/*`, `@bindings/*` → `./bindings/*`
- **Theme**: dark/light/system via `ThemeProvider` + `ModeToggle` in root layout
- **Upload flow**: frontend reads file → `arrayBuffer` → binary string → `btoa(base64)` → calls Go `ProcessFile`

## Sentence splitting

17 languages (czech through turkish, see `language-data/`). Uses `github.com/neurosnap/sentences` with **embedded JSON training data** (`//go:embed language-data/*.json` in service.go). NOT the `english` subpackage. Language JSONs are per-service (duplicated at `internal/services/textupload/language-data/` and `internal/data/language-data/`).

## Wails-specific

- v3 alpha.74 bug: `*BrowserWindow` doesn't implement `Window` (missing `AttachModal`). Affects `go build -tags server` but NOT `wails3 dev` or `wails3 build`
- Regenerate bindings after Go changes: `task common:tooling:generate:bindings` (runs `wails3 generate bindings -clean=true -ts`)
- Dev config: `build/config.yml`. After changing `info` field, run `common:tooling:update:build-assets`

## Generated files (do not edit, in .gitignore)

- `frontend/src/routeTree.gen.ts` — TanStack Router codegen
- `frontend/bindings/` — Wails Go→TS bindings
- `frontend/dist/` — Vite output
