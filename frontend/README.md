# PageVoice Frontend

React 19 + TanStack Router + Tailwind v4 + Vite 8 + Biome + shadcn (base-ui variant).

## Commands

| Command | Purpose |
|---|---|
| `bun install` | Install dependencies |
| `bun run dev` | Vite dev on `:3000` |
| `bun run build` | Production build |
| `bun run build:dev` | `tsc` + dev build |
| `bun run lint` | Biome lint |
| `bun run check` | Biome lint + format |
| `bun run test` | Vitest (no tests yet) |
| `bun run generate-routes` | Regenerate TanStack Router types |

## Structure

- `src/routes/` — file-based routing via TanStack Router
- `src/components/ui/` — shadcn components (base-ui variant, no `asChild`)
- `src/components/` — app components (theme provider, mode toggle)
- `bindings/` — auto-generated Wails Go→TS bindings (`.gitignore`d)
- `src/routeTree.gen.ts` — auto-generated router tree (`.gitignore`d)

## Conventions

- **Imports**: `#/*` → `./src/*`, `@bindings/*` → `./bindings/*`
- **Package manager**: Bun only
- **Styling**: Tailwind v4 via Vite plugin (no `tailwind.config.js`)
- **Linting**: Biome (ESLint + Prettier replacement)
- **Theming**: dark/light/system via `ThemeProvider` + `ModeToggle`
