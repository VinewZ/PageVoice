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

**Go services** at `internal/services/`:
- `textupload/` — 3 Wails-bound methods:
  - `ProcessFile(fileName, fileData, language) → UploadResult` — accepts file bytes, extracts text (PDF/EPUB/TXT), persists to XDG
  - `GetLibrary() → LibraryEntry[]` — reads `library.json` from `books/` dir
  - `GetBook(dirName) → BookDetail` — reads `metadata.json` + `state.json`
- `tts/` *(planned)* — TTS with Piper CLI

**Piper TTS** (`internal/tts/piper/`):
- Embeds `piper_linux_x86_64.tar.gz` (rhasspy/piper v2023.11.14-2, MIT) via `//go:embed`
- Extracts to `$XDG_DATA_HOME/page-voice/piper/` on first run (version marker file prevents re-extraction)
- Linux-only: `//go:build linux && amd64` (add `arm64` support by embedding `piper_linux_aarch64.tar.gz`)
- `EnsureExtracted()` — called at app startup in `internal/app/app.go`, creates the piper directory with binary, shared libs, and espeak-ng-data
- `Piper.Synthesize(text, modelPath, configPath) → WAV bytes` — spawns per-call subprocess, pipes text via stdin, reads `--output-raw` PCM, encodes to WAV
- Uses `LD_LIBRARY_PATH` to resolve `libespeak-ng.so`, `libonnxruntime.so`, `libpiper_phonemize.so`
- Voice config JSON provides sample rate (defaults to 22050)

**XDG data layout** (`$XDG_DATA_HOME/page-voice/`):
```
piper/
├── piper                        ← extracted binary
├── libespeak-ng.so.1.52.0.1
├── libonnxruntime.so.1.14.1
├── libpiper_phonemize.so.1.2.0
├── espeak-ng-data/              ← phoneme dictionaries
└── voices/                      ← downloaded ONNX voice models
    └── en_US-lessac-medium/
        ├── en_US-lessac-medium.onnx
        └── en_US-lessac-medium.onnx.json
books/
├── library.json                 ← [{id, title, dirName}]
└── <sanitized-title-8c8f9c>/
    ├── original.txt             ← raw text
    ├── metadata.json            ← {title, author, language, sourceFile, importedAt}
    ├── state.json               ← {status, chunkLength:2500, currentChunk, totalChunks, createdAt, updatedAt}
    ├── sentences.json           ← split sentences with chunk mapping
    └── audio/                   ← generated WAV files per chunk
        ├── chunk_001.wav
        └── ...
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
