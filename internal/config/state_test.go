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
