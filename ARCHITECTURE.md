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
  source/local (v1)          source/docker|ssh (v0.2)
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
| Local host | Remote kill on docker/ssh |
| Docker A1 + SSH B-go HostSources | hosts.yaml, LAN discovery |
| Kill with confirm (local) | Alerts |
| Login autostart (Settings toggle → `internal/autostart`) | |

## Autostart (GUI)

Preference: `gui.autostart` in `~/.config/procpane/state.yaml` (default off). Toggling in Settings calls `internal/autostart.SetEnabled` immediately; app launch re-applies registration when enabled.

| OS | Mechanism |
|----|-----------|
| Linux | XDG `~/.config/autostart/procpane-app.desktop` |
| Windows | `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` → `procpane` = path to `procpane-app.exe` |
| macOS | User LaunchAgent `~/Library/LaunchAgents/com.zx06.procpane.plist` (`ProgramArguments` = path to `procpane.app` binary); `launchctl bootstrap` / `bootout` on toggle |

## Decisions

See `docs/adr/` (Wails3, Svelte5, gopsutil, kill safety, HostSource).
