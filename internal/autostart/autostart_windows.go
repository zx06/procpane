//go:build windows

package autostart

import (
	"errors"
	"os"

	"golang.org/x/sys/windows/registry"
)

const defaultWinRunKey = `Software\Microsoft\Windows\CurrentVersion\Run`

func winRunKeyPath() string {
	if override := os.Getenv("PROCPANE_AUTOSTART_RUN_KEY"); override != "" {
		return override
	}
	return defaultWinRunKey
}

func setEnabledPlatform(on bool, exe string) error {
	k, _, err := registry.CreateKey(registry.CURRENT_USER, winRunKeyPath(), registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if on {
		return k.SetStringValue(winRunValue, exe)
	}
	err = k.DeleteValue(winRunValue)
	if err == registry.ErrNotExist {
		return nil
	}
	return err
}

// IsRegistered reports whether the Run key entry exists.
func IsRegistered() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, winRunKeyPath(), registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	defer k.Close()
	_, _, err = k.GetStringValue(winRunValue)
	if err == registry.ErrNotExist {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
