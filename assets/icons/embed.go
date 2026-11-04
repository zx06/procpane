// Package icons embeds procpane tray and application icons.
package icons

import _ "embed"

// TrayMacTemplate is a monochrome PNG for macOS menu-bar template rendering.
//
//go:embed procpane-tray-template.png
var TrayMacTemplate []byte

// App256 is the primary application icon (256×256 PNG).
//
//go:embed procpane-app-256.png
var App256 []byte

// TrayDark is a colored tray icon for dark menu bars / Windows.
//
//go:embed procpane-tray-dark.png
var TrayDark []byte

// TrayLight is a colored tray icon for light menu bars.
//
//go:embed procpane-tray-light.png
var TrayLight []byte

// WindowsICO is the Windows .ico bundle.
//
//go:embed procpane.ico
var WindowsICO []byte
