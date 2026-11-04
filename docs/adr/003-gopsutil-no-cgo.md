# ADR 003: gopsutil v4 without cgo

## Status

Accepted (2026-11-04)

## Context

LISTEN discovery and process enrichment must work on Linux, macOS, and Windows with minimal native build pain.

## Decision

Use **github.com/shirou/gopsutil/v4 v4.26.10** with **CGO_ENABLED=0** for the CLI and shared `internal/*` libraries.

## Consequences

- Some fields (cwd, user on certain OS configs) may return `?`; UI and JSON must tolerate gaps.
- If a platform regresses without cgo, document the exception in README before enabling cgo.
