# Self-test checklist (grokbot-tasks#11)

Use this on **your machine** for items that need macOS/Windows GUI or elevated permissions. Automated Linux CLI evidence is in [`evidence/linux-smoke-2026-11-04.txt`](../evidence/linux-smoke-2026-11-04.txt).

## Quick CLI (all platforms)

```bash
./scripts/test-cli.sh          # or see commands below
CGO_ENABLED=0 go test ./internal/... ./cmd/procpane/...
CGO_ENABLED=0 go build -o bin/procpane ./cmd/procpane
./bin/procpane list
./bin/procpane list --json | jq '.listeners | length'
./bin/procpane kill 1          # must refuse pid 1
```

| Check | Linux (VM) | macOS | Windows |
|-------|------------|-------|---------|
| TCP LISTEN rows with port/pid/name | ✅ auto | Manual | Manual |
| Missing enrich → `?` (not crash) | ✅ auto | Manual | Manual |
| `list --json` jq-parseable | ✅ auto | Manual | Manual |
| `find :port` hits known listener | ✅ auto (python http.server) | Manual | Manual |
| `show <pid>` JSON | ✅ auto | Manual | Manual |
| `kill` confirm / `--yes` | Manual (throwaway process) | Manual | Manual |
| Refuse pid ≤ 1 / self | ✅ pid 1 | Manual | Manual |
| GUI tray/menu bar, same data as CLI | N/A headless | **Manual** | **Manual** |
| GUI kill modal → refresh | N/A headless | **Manual** | **Manual** |

## Linux (headless / CI)

- **Full** `go test ./...` with `CGO_ENABLED=0` **does not** compile `cmd/procpane-app` (Wails needs cgo). Use `./scripts/test-cli.sh` or test CLI packages only.
- GUI (`procpane-app`) needs GTK4 + WebKitGTK 6; not required on headless agents. See README GUI section.

## macOS

1. `go build -o bin/procpane ./cmd/procpane` (native; cross-compile from Linux also works for CLI — see evidence file).
2. `go build -o bin/procpane-app ./cmd/procpane-app` (requires Xcode CLT, cgo).
3. Run `bin/procpane-app`: menu bar icon, left-click show/hide, right-click refresh/quit.
4. Compare row count with `bin/procpane list --json`.
5. Kill: select row → 结束进程 → confirm modal → list refreshes.

## Windows

See README **Windows self-test**. Summary:

1. Install Go 1.27.3; build CLI:  
   `set CGO_ENABLED=0` then `go build -o bin\procpane.exe .\cmd\procpane`  
   (Or cross-build on Linux: `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/procpane.exe ./cmd/procpane` — ✅ verified.)
2. `procpane.exe list`, `procpane.exe find :<port>`, `procpane.exe kill 1` (refuse).
3. GUI: install [WebView2](https://developer.microsoft.com/microsoft-edge/webview2/); `go build -o procpane-app.exe ./cmd/procpane-app`; tray + window parity with CLI.

## Cross-compile CLI from Linux (2026-11-04)

| Target | Command | Result on agent VM |
|--------|---------|-------------------|
| Windows amd64 | `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o bin/procpane.exe ./cmd/procpane` | ✅ success |
| Darwin amd64 | `GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o bin/procpane-darwin-amd64 ./cmd/procpane` | ✅ success |
| Darwin arm64 | `GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o bin/procpane-darwin-arm64 ./cmd/procpane` | ✅ success |

GUI binaries must be built on each OS (Wails + native webview toolchains).
