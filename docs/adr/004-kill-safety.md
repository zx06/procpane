# ADR 004: Kill safety and confirmation

## Status

Accepted (2026-11-04)

## Context

Ending processes is destructive; v1 includes CLI and GUI kill paths.

## Decision

- Reject **pid ≤ 1** and **self** in `internal/kill.ValidateTargetPID`.
- CLI: print target summary, prompt unless `--yes`.
- GUI: modal double confirmation before calling `KillProcess`.
- Unix: `SIGTERM` / `SIGKILL` via `os.Process.Signal`.
- Windows: `TerminateProcess` (documented; no Unix signals).

## Consequences

- Permission errors surface as failed `KillResult` with message.
- Future remote HostSources must reuse the same safety layer, not reimplement checks in UI.
