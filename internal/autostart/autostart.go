// Package autostart registers the procpane tray app to launch at OS login.
//
// Platform behavior:
//   - Linux: XDG autostart desktop file in ~/.config/autostart/procpane-app.desktop
//   - Windows: HKCU\Software\Microsoft\Windows\CurrentVersion\Run value "procpane" (standard per-user Run key for tray apps)
//   - macOS: user LaunchAgent plist ~/Library/LaunchAgents/com.zx06.procpane.plist → installed .app binary (Login Items shows procpane)
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
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
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
