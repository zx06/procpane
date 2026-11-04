# ADR 002: Svelte 5 frontend

## Status

Accepted (2026-11-04)

## Context

The GUI lives in `frontend/` (Svelte) and should stay easy for AI/humans to iterate without standalone HTML mocks in the repo.

## Decision

**Svelte 5.57.3** + Vite, calling Go via `@wailsio/runtime` `Call.ByName` to `internal/gui.ProcpaneService`.

## Consequences

- Frontend never imports gopsutil or OS APIs.
- Run `scripts/build-frontend.sh` before embedding; `frontend/dist` is committed so `go build ./cmd/procpane-app` works after clone.
- Regenerate typed bindings optional; v1 uses stable service method names.
