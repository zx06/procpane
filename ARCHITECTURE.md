# Architecture

## Layers

```text
┌─────────────────────────────────────────────────────────┐
│  CLI (cobra)          │  GUI (Wails + Svelte 5)         │
└───────────┬───────────┴──────────────┬──────────────────┘
            │                          │
            ▼                          ▼
     internal/app.Service        internal/gui.ProcpaneService
            │                          │
            └──────────┬───────────────┘
                       ▼
              source.HostSource
                       │
         ┌─────────────┴─────────────┐
         ▼                           ▼
  source/local (v1)          source/docker|ssh (future)
         │
         ├── scan.ListRawListeners
         └── enrich.EnrichListeners
```

## Data flow (v1 local)

1. **scan** reads TCP/UDP connections via gopsutil; keeps rows in LISTEN (UDP may show `NONE` on Linux).
2. **enrich** attaches process fields; missing values → `"?"` in model.
3. **format** renders table or JSON for CLI.
4. **kill** validates PID, then platform-specific terminate.

## Extension: HostSource

Implement `ListListeners(ctx) ([]model.Listener, error)` for new hosts. CLI/GUI swap the injected `HostSource` — no frontend changes.

## v1 boundary

| In scope | Out of scope |
|----------|----------------|
| LISTEN tcp/udp | ESTABLISHED view |
| Local host | Docker/SSH |
| Kill with confirm | Autostart, alerts |

## Decisions

See `docs/adr/` (Wails3, Svelte5, gopsutil, kill safety, HostSource).
