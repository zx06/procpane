# ADR 005: HostSource extension point

## Status

Accepted (2026-11-04)

## Decision

All listener collection goes through `source.HostSource`:

```go
ListListeners(ctx context.Context) ([]model.Listener, error)
```

v1 implements `internal/source/local`. Docker/SSH add packages without changing CLI/GUI contracts.

## Consequences

- Stable JSON field names live in `internal/model`.
- Scan/enrich stay host-agnostic helpers used by local (and future) sources.
