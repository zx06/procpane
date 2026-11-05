//go:build linux

package autostart

import (
	"os"
	"path/filepath"
)

var linuxAutostartDir = func() string {
	if override := os.Getenv("PROCPANE_AUTOSTART_DIR"); override != "" {
		return override
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "autostart")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "autostart")
}

func setEnabledPlatform(on bool, exe string) error {
	dir := linuxAutostartDir()
	if dir == "" {
		return os.ErrInvalid
	}
	path := filepath.Join(dir, desktopName)
	if !on {
		err := os.Remove(path)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(desktopFileContent(exe)), 0o644)
}

// IsRegistered reports whether the XDG autostart desktop file exists.
func IsRegistered() (bool, error) {
	dir := linuxAutostartDir()
	if dir == "" {
		return false, os.ErrInvalid
	}
	path := filepath.Join(dir, desktopName)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
