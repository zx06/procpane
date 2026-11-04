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
- Run `./scripts/test-cli.sh` (or `CGO_ENABLED=0 go test ./internal/... ./cmd/procpane/...`) before push.
- Note: `go test ./...` with `CGO_ENABLED=0` fails on `cmd/procpane-app` (Wails needs cgo). CI uses `scripts/test-cli.sh` only.

## Cross-platform checklist

- [x] Linux CLI: list / json / find / show / refuse kill 1 (`./scripts/test-cli.sh` / CI)
- [ ] macOS: menu bar tray + GUI parity with CLI
- [ ] Windows: CLI steps in README “Windows self-test”; GUI + WebView2
- [ ] Kill confirmation on a throwaway process (CLI + GUI modal)

Do **not** commit local smoke logs under `evidence/` or a `docs/self-test.md` dump — keep checklist notes out of the repo.

## AI contributors

Start with `AGENTS.md` and `ARCHITECTURE.md`. Do not expand scope into Docker/SSH without a new ADR.
