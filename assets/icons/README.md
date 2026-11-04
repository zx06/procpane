# procpane icons

| File | Use |
|------|-----|
| `procpane-mark.svg` | Source mark — outer bars `currentColor`, middle bar accent `#1898f0` |
| `procpane-tray-template.png` | macOS menu-bar **template** icon (18×18, anti-aliased; embedded) |
| `procpane-tray-template@2x.png` | Retina reference (36×36 AA; not embedded) |
| `procpane-tray-dark.png` / `procpane-tray-light.png` | Colored tray (64×64 AA; Linux default uses dark) |
| `procpane-app-256.png` | Application / about icon (AA, accent mid bar) |
| `procpane.ico` | Windows system tray (multi-size 16–256) |

Design (2026-11-04): based on option C Pane, without `:P` chip and without outer rounded frame — three horizontal bars, middle shorter + accent.

Go embeds the production assets via `embed.go` — edit PNG/SVG here, then rebuild `cmd/procpane-app`.

**Regeneration (2026-11-04):** Replaced hard-edge 18/32 px rasters with anti-aliased exports. Linux HiDPI benefits from 64×64 `TrayDark` / `TrayLight` without embed API changes.
