// Package autostart registers the procpane tray app to launch at OS login.
//
// Platform behavior:
//   - Linux: XDG autostart desktop file in ~/.config/autostart/procpane-app.desktop
//   - Windows: HKCU\Software\Microsoft\Windows\CurrentVersion\Run value "procpane" (standard per-user Run key for tray apps)
//   - macOS: SMAppService.mainApp via a short swift helper (Login Items, no Dock icon; requires distributed .app with com.zx06.procpane)
package autostart

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var errUnsupported = errors.New("autostart: unsupported platform")

const (
	appName    = "procpane"
	desktopName = "procpane-app.desktop"
	winRunValue = "procpane"
	macBundleID = "com.zx06.procpane"
)

// SetEnabled registers or removes OS-level autostart for the GUI binary.
func SetEnabled(on bool) error {
	exe, err := resolveExecutable()
	if err != nil {
		return err
	}
	return setEnabledPlatform(on, exe)
}

func resolveExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	// Go test binaries on Windows CI often lack a resolvable final path; Abs is enough for Run-key registration.
	if resolved, symErr := filepath.EvalSymlinks(exe); symErr == nil {
		exe = resolved
	}
	return filepath.Abs(exe)
}

func desktopFileContent(exe string) string {
	return fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=procpane
Comment=Local LISTEN port panel
Exec=%s
Terminal=false
Categories=Utility;
StartupNotify=false
X-GNOME-Autostart-enabled=true
`, exe)
}
