//go:build windows

package autostart

import (
	"golang.org/x/sys/windows/registry"
)

const winRunKey = `Software\Microsoft\Windows\CurrentVersion\Run`

func setEnabledPlatform(on bool, exe string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, winRunKey, registry.SET_VALUE)
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
	k, err := registry.OpenKey(registry.CURRENT_USER, winRunKey, registry.QUERY_VALUE)
	if err != nil {
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
