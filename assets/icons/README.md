# procpane icons

| File | Use |
|------|-----|
| `procpane-mark.svg` | Source mark (three pane rows, no outer frame) |
| `procpane-tray-template.png` | macOS menu-bar **template** icon (embedded in app) |
| `procpane-tray-template@2x.png` | Retina reference (not embedded; regenerate template from SVG if needed) |
| `procpane-tray-dark.png` / `procpane-tray-light.png` | Colored tray references |
| `procpane-app-256.png` | Application / about icon |
| `procpane.ico` | Windows system tray |

Design (2026-11-04): based on option C Pane, without `:P` chip and without outer rounded frame — three horizontal bars, middle shorter + accent.

Go embeds the production assets via `embed.go` — edit PNG/SVG here, then rebuild `cmd/procpane-app`.
