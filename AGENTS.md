# AGENTS.md — procpane

## Goal

Cross-platform **local** LISTEN (tcp/udp) panel: **CLI** + **tray GUI**. v1 = list, enrich, kill. Future: Docker/SSH via `HostSource`.

## Map

| Path | Role |
|------|------|
| `cmd/procpane` | Cobra CLI (CGO_ENABLED=0) |
| `cmd/procpane-app` | Wails v3 tray app |
| `internal/model` | Stable JSON structs — change carefully |
| `internal/source` | `HostSource` interface |
| `internal/source/local` | v1 local implementation |
| `internal/scan` | OS connections → raw LISTEN rows |
| `internal/enrich` | PID → name/cmdline/cwd/cpu/mem/chain |
| `internal/kill` | Safety + platform send |
| `internal/format` | table / json |
| `internal/gui` | Wails service for Svelte |
| `frontend/` | Svelte 5 UI (`src/App.svelte`, `src/app.css`) |
| `assets/icons/` | Tray + app icons (embedded via `assets/icons/embed.go`) |

## Commands

```bash
# Pins check
./scripts/check-pins.sh

# CLI (no cgo)
CGO_ENABLED=0 go build -o bin/procpane ./cmd/procpane
./scripts/test-cli.sh
# or: CGO_ENABLED=0 go test ./internal/... ./cmd/procpane/...
# Full go test ./... needs cgo for cmd/procpane-app

# Frontend
cd frontend && npm ci && npm run build   # or ./scripts/build-frontend.sh

# GUI (needs OS webview + cgo on Linux)
go build -o bin/procpane-app ./cmd/procpane-app
# macOS: conn-graph frosted glass needs transparent WKWebView — use private_mac_apis:
# go build -tags private_mac_apis -o bin/procpane-app ./cmd/procpane-app
# Windows GUI only: -ldflags="-H windowsgui" (PE GUI subsystem, no CMD flash)
```

## Taboos

- Do **not** kill pid ≤ 1 or self (see `internal/kill`).
- Do **not** bypass CLI/GUI confirmation for kill in user-facing flows.
- Do **not** break `internal/model` JSON field names without ADR.
- Do **not** call gopsutil from Svelte — only Go services.
- Do **not** commit standalone HTML/UI prototype mocks — iterate in `frontend/` only.

## GUI window (tray popover)

Default size in `cmd/procpane-app/main.go`: **440×520** — compact menu-bar panel, not a full installer window. Design tokens and layout: `frontend/src/app.css`, `frontend/src/App.svelte`.

## Testing

- Linux VM: `procpane list`, `procpane list --json | jq .listeners`.
- Kill: use a throwaway sleep process; always test refusal for pid 1 and self.
- **Browser UI (Playwright + Wails server build):** exercises the Svelte UI and real Go services over HTTP — not native WebView, OS menus, **tray icons**, or native file dialogs. See [Wails v3 E2E guide](https://v3.wails.io/guides/e2e-testing/).

  ```bash
  export PATH="$(go env GOPATH)/bin:$PATH"
  wails3 generate bindings -f '-tags=server' ./cmd/procpane-app ./internal/gui
  cd frontend && npm ci --ignore-engines && npm run build
  npx playwright install chromium
  npx playwright test
  ```

  Host: `go run -tags server,production ./cmd/procpane-app` (Playwright `webServer` starts this in CI). Requires `application.Options.Server` in `cmd/procpane-app/main.go` (127.0.0.1:8080).

- **GUI tray:** manual on macOS/Windows/Linux — out of scope for Playwright; Linux VM may lack full tray/webview in CI.
- **CI:** `.github/workflows/e2e-ui.yml` uploads `playwright-report` and `test-results` artifacts (screenshots, traces) on every run — download from the Actions run → Artifacts.

## Read order

1. `AGENTS.md` (this file)
2. `ARCHITECTURE.md`
3. Relevant `docs/adr/*`
4. Package you are editing
