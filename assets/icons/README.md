# procpane icons

| File | Use |
|------|-----|
| `procpane-mark.svg` | Source mark (panel + port rows) |
| `procpane-tray-template.png` | macOS menu-bar **template** icon (embedded in app) |
| `procpane-tray-template@2x.png` | Retina reference (not embedded; regenerate template from SVG if needed) |
| `procpane-tray-dark.png` / `procpane-tray-light.png` | Colored tray references |
| `procpane-app-256.png` | Application / about icon |
| `procpane.ico` | Windows system tray |

Go embeds the production assets via `embed.go` — edit PNG/SVG here, then rebuild `cmd/procpane-app`.
