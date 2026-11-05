package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRefreshIntervalDefault(t *testing.T) {
	dir := t.TempDir()
	s := &Store{path: filepath.Join(dir, "state.yaml")}
	sec, err := s.RefreshIntervalSec()
	if err != nil {
		t.Fatal(err)
	}
	if sec != DefaultRefreshIntervalSec {
		t.Fatalf("default=%d want %d", sec, DefaultRefreshIntervalSec)
	}
}

func TestRefreshIntervalPersistAndClamp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")
	s := &Store{path: path}

	if err := s.SetRefreshIntervalSec(60); err != nil {
		t.Fatal(err)
	}
	sec, err := s.RefreshIntervalSec()
	if err != nil || sec != 60 {
		t.Fatalf("sec=%d err=%v", sec, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "refreshIntervalSec") {
		t.Fatalf("yaml missing refreshIntervalSec: %s", data)
	}

	if err := s.SetRefreshIntervalSec(5); err != nil {
		t.Fatal(err)
	}
	sec, _ = s.RefreshIntervalSec()
	if sec != minRefreshIntervalSec {
		t.Fatalf("clamped low=%d want %d", sec, minRefreshIntervalSec)
	}

	if err := s.SetRefreshIntervalSec(9999); err != nil {
		t.Fatal(err)
	}
	sec, _ = s.RefreshIntervalSec()
	if sec != maxRefreshIntervalSec {
		t.Fatalf("clamped high=%d want %d", sec, maxRefreshIntervalSec)
	}
}

func TestAutostartDefaultOff(t *testing.T) {
	dir := t.TempDir()
	s := &Store{path: filepath.Join(dir, "state.yaml")}
	on, err := s.Autostart()
	if err != nil {
		t.Fatal(err)
	}
	if on {
		t.Fatal("expected default autostart off")
	}
}

func TestAutostartPersist(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.yaml")
	s := &Store{path: path}

	if err := s.SetAutostart(true); err != nil {
		t.Fatal(err)
	}
	on, err := s.Autostart()
	if err != nil || !on {
		t.Fatalf("on=%v err=%v", on, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "autostart: true") {
		t.Fatalf("yaml missing autostart: %s", data)
	}

	if err := s.SetAutostart(false); err != nil {
		t.Fatal(err)
	}
	on, _ = s.Autostart()
	if on {
		t.Fatal("expected autostart off after clear")
	}
}
