# procpane

Cross-platform **local** process/port panel: see who is **LISTEN**ing on tcp/udp, inspect cmdline/cwd/parent chain, and end processes — via **CLI** or **menu-bar/tray GUI**.

Spec: [grokbot-tasks#11](https://github.com/zx06/grokbot-tasks/issues/11) · Interactive UI reference: [`docs/design/prototype.html`](docs/design/prototype.html)

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
