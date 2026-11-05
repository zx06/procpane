//go:build windows

package autostart

import (
	"testing"
)

func TestWindowsAutostartEnableDisable(t *testing.T) {
	t.Setenv("PROCPANE_AUTOSTART_RUN_KEY", `Software\procpane-test\Autostart`)
	t.Cleanup(func() { _ = setEnabledPlatform(false, "") })

	exe := `C:\Program Files\procpane\procpane-app.exe`
	if err := setEnabledPlatform(true, exe); err != nil {
		t.Fatal(err)
	}
	ok, err := IsRegistered()
	if err != nil || !ok {
		t.Fatalf("registered=%v err=%v", ok, err)
	}

	if err := setEnabledPlatform(false, exe); err != nil {
		t.Fatal(err)
	}
	ok, err = IsRegistered()
	if err != nil || ok {
		t.Fatalf("registered=%v err=%v", ok, err)
	}
}

func TestSetEnabledUsesExecutable(t *testing.T) {
	t.Setenv("PROCPANE_AUTOSTART_RUN_KEY", `Software\procpane-test\SetEnabled`)
	t.Cleanup(func() { _ = SetEnabled(false) })

	if err := SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	ok, err := IsRegistered()
	if err != nil || !ok {
		t.Fatalf("registered=%v err=%v", ok, err)
	}
}
