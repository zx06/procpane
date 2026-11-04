# ADR 001: Wails v3 for GUI

## Status

Accepted (2026-11-04)

## Context

procpane needs a cross-platform tray/menu-bar window with a web UI and Go backend sharing scan/enrich/kill logic with the CLI.

## Decision

Use **Wails v3.1.0** (stable) with embedded Vite assets and `application.SystemTray` for tray behavior.

## Consequences

- GUI build requires platform webview deps (GTK4/WebKit on Linux, WebView2 on Windows, WebKit on macOS).
- CLI remains **CGO_ENABLED=0**; only `cmd/procpane-app` links Wails.
- Tray APIs differ from Wails v2; follow v3 `SystemTray.AttachWindow` + menu patterns.
