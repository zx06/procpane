//go:build linux

package autostart

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinuxAutostartEnableDisable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PROCPANE_AUTOSTART_DIR", dir)

	exe := filepath.Join(t.TempDir(), "procpane-app")
	if err := os.WriteFile(exe, []byte{0}, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := setEnabledPlatform(true, exe); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, desktopName)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), exe) {
		t.Fatalf("desktop missing exe: %s", data)
	}
	ok, err := IsRegistered()
	if err != nil || !ok {
		t.Fatalf("registered=%v err=%v", ok, err)
	}

	if err := setEnabledPlatform(false, exe); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("desktop still present: %v", err)
	}
}

func TestSetEnabledUsesExecutable(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PROCPANE_AUTOSTART_DIR", dir)

	if err := SetEnabled(true); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, desktopName)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	self, err := resolveExecutable()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), self) {
		t.Fatalf("want %q in %s", self, data)
	}
	t.Cleanup(func() { _ = SetEnabled(false) })
}
