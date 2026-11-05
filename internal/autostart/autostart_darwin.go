//go:build darwin

package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func setEnabledPlatform(on bool, exe string) error {
	removeLegacySwiftHelper()
	return setEnabledLaunchAgent(on, exe)
}

func launchAgentPath() string {
	if p := os.Getenv("PROCPANE_LAUNCH_AGENT_PLIST"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Library", "LaunchAgents", macBundleID+".plist")
}

func launchAgentDomain() string {
	return fmt.Sprintf("gui/%d", os.Getuid())
}

func launchAgentBootout(path string) {
	if path == "" {
		return
	}
	_ = exec.Command("launchctl", "bootout", launchAgentDomain(), path).Run()
}

func setEnabledLaunchAgent(on bool, exe string) error {
	path := launchAgentPath()
	if path == "" {
		return os.ErrInvalid
	}
	if !on {
		launchAgentBootout(path)
		err := os.Remove(path)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	launchAgentBootout(path)
	plist := launchAgentPlistXML(macBundleID, exe)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
		return err
	}
	if err := exec.Command("launchctl", "bootstrap", launchAgentDomain(), path).Run(); err != nil {
		return fmt.Errorf("launchctl bootstrap: %w", err)
	}
	return nil
}

// IsRegistered reports whether the user LaunchAgent plist is installed.
func IsRegistered() (bool, error) {
	path := launchAgentPath()
	if path == "" {
		return false, os.ErrInvalid
	}
	if _, err := os.Stat(path); err == nil {
		return true, nil
	}
	return false, nil
}

func removeLegacySwiftHelper() {
	dir, err := os.UserCacheDir()
	if err != nil {
		return
	}
	_ = os.Remove(filepath.Join(dir, "procpane", "autostart-mainapp.swift"))
}
