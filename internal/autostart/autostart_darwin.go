//go:build darwin

package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func setEnabledPlatform(on bool, exe string) error {
	if err := smAppService(on); err == nil {
		return nil
	}
	return setEnabledLaunchAgent(on, exe)
}

func smAppService(register bool) error {
	scriptPath, err := writeSwiftHelper()
	if err != nil {
		return err
	}
	args := []string{scriptPath}
	if register {
		args = append(args, "register")
	} else {
		args = append(args, "unregister")
	}
	out, err := exec.Command("swift", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("SMAppService: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func writeSwiftHelper() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir = filepath.Join(dir, "procpane")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "autostart-mainapp.swift")
	const src = `import Foundation
import ServiceManagement

let action = CommandLine.arguments.dropFirst().first ?? ""
let svc = SMAppService.mainApp
do {
    switch action {
    case "register":
        try svc.register()
    case "unregister":
        try svc.unregister()
    default:
        fputs("usage: register|unregister\n", stderr)
        exit(2)
    }
} catch {
    fputs("error: \(error)\n", stderr)
    exit(1)
}
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func launchAgentPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Library", "LaunchAgents", macBundleID+".plist")
}

func setEnabledLaunchAgent(on bool, exe string) error {
	path := launchAgentPath()
	if path == "" {
		return os.ErrInvalid
	}
	if !on {
		_ = exec.Command("launchctl", "bootout", fmt.Sprintf("gui/%d", os.Getuid()), path).Run()
		err := os.Remove(path)
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>%s</string>
	<key>ProgramArguments</key>
	<array>
		<string>%s</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
</dict>
</plist>
`, macBundleID, exe)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(plist), 0o644); err != nil {
		return err
	}
	return exec.Command("launchctl", "bootstrap", fmt.Sprintf("gui/%d", os.Getuid()), path).Run()
}

// IsRegistered best-effort: SMAppService status when swift works, else LaunchAgent plist.
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
