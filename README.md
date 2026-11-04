# procpane

Cross-platform **local** process/port panel: see who is **LISTEN**ing on tcp/udp, inspect cmdline/cwd/parent chain, and end processes — via **CLI** or **menu-bar/tray GUI**.

Spec: [grokbot-tasks#11](https://github.com/zx06/grokbot-tasks/issues/11) · GUI: [`frontend/src/`](frontend/src/) (Svelte 5)

## Pinned stack (2026-11-04)

| Component | Version |
|-----------|---------|
| Go | **1.27.3** |
| Node.js | **26.13.0** LTS |
| Wails | **v3.1.0** |
| Svelte | **5.57.3** |
| gopsutil | **v4.26.10** |
| cobra | **v1.10.2** |

## Install / build CLI

Requires Go 1.27.3+.

```bash
git clone https://github.com/zx06/procpane.git
cd procpane
CGO_ENABLED=0 go build -o bin/procpane ./cmd/procpane
./bin/procpane version
```

No cgo required for the CLI.

### Cross-compile CLI (from Linux/macOS/Windows)

```bash
# Windows (verified on Linux CI agent)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/procpane.exe ./cmd/procpane

# macOS CLI binaries (pure Go — cross-build from Linux works; run on Mac to verify)
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o bin/procpane-darwin-arm64 ./cmd/procpane
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o bin/procpane-darwin-amd64 ./cmd/procpane
```

**GUI** (`cmd/procpane-app`) must be built on the target OS (Wails + webview SDK). macOS menu bar / Windows tray builds are not cross-compiled from Linux.

Automated CLI smoke: `./scripts/test-cli.sh`.

## CLI examples

```bash
procpane list
procpane list --proto tcp --json | jq '.listeners[] | {port, pid, name, cmdline}'
procpane find :8080
procpane show 12345
procpane kill 12345          # prompts for confirmation
procpane kill 12345 --yes --signal KILL
```

### Flags

- `list --proto tcp|udp|all` `--json` `-w 2s` (watch)
- `kill --signal TERM|KILL` `--yes`

## GUI (Wails + Svelte)

1. Install [Node.js 26.13.0 LTS](https://nodejs.org/) and Go 1.27.3.
2. Build frontend (committed `frontend/dist` is also kept for embed):

   ```bash
   ./scripts/build-frontend.sh
   ```

3. Install OS deps:
   - **macOS**: Xcode CLT; app uses menu bar (`ActivationPolicyAccessory`).
   - **Windows**: [WebView2](https://developer.microsoft.com/microsoft-edge/webview2/) runtime.
   - **Linux**: GTK4 + WebKitGTK 6 (Wails v3). Tray + popup tested manually; headless CI often cannot run the full app.

4. Build:

   ```bash
   go build -o bin/procpane-app ./cmd/procpane-app
   ./bin/procpane-app
   ```

Tray: **left-click** toggles window; **right-click** menu → refresh / quit. UI auto-refreshes ~2s (slower when unfocused).

The popover defaults to **440×520** px (see `cmd/procpane-app/main.go`). CSS follows **system light/dark** via `prefers-color-scheme`. Branded icons live under [`assets/icons/`](assets/icons/) (macOS template PNG + app/tray assets).

### Automated browser UI tests (Playwright)

CI runs Playwright against a **Wails server build** (`-tags server,production`), which hits the embedded Svelte UI and real Go bindings over HTTP. This does **not** cover the native tray, WebView window, or OS menus — test those manually per platform. Guide: [Wails v3 E2E testing](https://v3.wails.io/guides/e2e-testing/).

```bash
wails3 generate bindings -f '-tags=server' ./cmd/procpane-app ./internal/gui
cd frontend && npm ci --ignore-engines && npm run build
npx playwright install chromium && npm run test:e2e
```

On GitHub Actions, download **Artifacts** → `playwright-report` (HTML) or `playwright-test-results` (screenshots/traces).

Headless Linux CI **does not** run the native GUI (GTK4/WebKitGTK 6 not installed by design). Validate tray/WebView on a desktop Mac or Windows machine (see Windows self-test below).

## Windows self-test

On a Windows 10/11 machine with Go 1.27.3:

1. **CLI** (no cgo):  
   `set CGO_ENABLED=0`  
   `go build -o procpane.exe .\cmd\procpane`
2. **LISTEN**: `procpane.exe list` and `procpane.exe list --json` — expect tcp rows with port/pid/name (or `?` if enrich fails).
3. **find**: start any local listener (e.g. `python -m http.server 8765`), then `procpane.exe find :8765`.
4. **kill safety**: `procpane.exe kill 1` must print a refusal (non-zero exit).
5. **kill confirm**: `procpane.exe kill <pid>` on a test process must prompt unless `--yes`.
6. **GUI** (optional): install [WebView2 Evergreen Runtime](https://developer.microsoft.com/microsoft-edge/webview2/); build `go build -o procpane-app.exe .\cmd\procpane-app`; confirm system tray icon, window list matches CLI, kill uses confirmation modal.

You can also copy `bin/procpane.exe` from a Linux cross-compile (`GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build …`) and run steps 2–5 without Go on Windows.

## Permissions & kill behavior

- Refuses **pid ≤ 1** and **own pid**.
- **Unix/macOS**: `SIGTERM` (default) or `SIGKILL`.
- **Windows**: `TerminateProcess` (see `docs/adr/004-kill-safety.md`). CLI/GUI labels still say TERM/KILL for consistency.
- Killing other users' processes may fail with permission errors — surfaced in CLI JSON / GUI toast.

## Project docs

- [`AGENTS.md`](AGENTS.md) — map & commands for AI/humans
- [`ARCHITECTURE.md`](ARCHITECTURE.md) — layers & HostSource
- [`CONTRIBUTING.md`](CONTRIBUTING.md)
- [`docs/adr/`](docs/adr/) — decisions

## License

Private repository — all rights reserved unless otherwise noted by owner.
