# procpane icons (aa-v4 clarity pass, 2026-11-04)

| File | Use |
|------|-----|
| `procpane-mark.svg` | Source mark — outer bars `currentColor`, middle `#1898f0` |
| `procpane-tray-template.png` | macOS menu-bar **template** (88×88; Wails `setSize`s to status-bar thickness ≈22pt — extra pixels keep Retina sharp) |
| `procpane-tray-template@2x.png` | 44×44 reference (not embedded) |
| `procpane-tray-dark.png` / `procpane-tray-light.png` | Colored tray 64×64 (Linux default: dark) |
| `procpane-app-256.png` / `procpane-app-512.png` | Application icons |
| `procpane.ico` | Windows tray multi-size 16–256 |

## Design
C′ identity kept: three horizontal bars, middle shorter + accent, no `:P`, no outer frame.
Bars thickened (2.5 on 22pt canvas) so they match neighboring menubar icon weight.

## Why 88×88 for Mac template
Wails v3 `systemTraySetIcon` always `[image setSize:NSMakeSize(thickness, thickness)]` (~22×22 pt).
Embedding only 18×18 forced Retina upscale → blur vs WeChat/battery/etc.
88×88 provides 4× samples; downsample stays crisp.

Replace files under `assets/icons/` (same names). No `embed.go` API change required.
