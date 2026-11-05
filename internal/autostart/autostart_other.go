//go:build !linux && !windows && !darwin

package autostart

func setEnabledPlatform(on bool, exe string) error {
	if on {
		return errUnsupported
	}
	return nil
}

// IsRegistered is always false on unsupported platforms.
func IsRegistered() (bool, error) {
	return false, nil
}
