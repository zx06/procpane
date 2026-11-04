# Contributing

## Toolchain pins (verified 2026-11-04)

| Tool | Version |
|------|---------|
| Go | 1.27.3 (`go.mod` + `toolchain`) |
| Node.js | 26.13.0 LTS |
| Wails | v3.1.0 |
| Svelte | 5.57.3 |
| gopsutil | v4.26.10 |
| cobra | v1.10.2 |

Re-run `./scripts/check-pins.sh` after bumping versions; update README in the same commit.

## Commits

- Focused messages; behavior changes that affect contracts should mention ADR/model.
- Run `CGO_ENABLED=0 go test ./internal/... ./cmd/procpane/...` before push.

## Cross-platform checklist

- [ ] Linux: `procpane list --json`
- [ ] macOS: tray show/hide (manual)
- [ ] Windows: `go build ./cmd/procpane-app` on dev machine with WebView2
- [ ] Kill confirmation CLI + GUI modal

## AI contributors

Start with `AGENTS.md` and `ARCHITECTURE.md`. Do not expand scope into Docker/SSH without a new ADR.
