package gui

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestSetAutostartPersistsAndEmits(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("PROCPANE_AUTOSTART_DIR", filepath.Join(dir, "autostart"))

	svc := NewService()
	var got map[string]any
	svc.BindEventEmitter(func(name string, data ...any) {
		if name != "procpane:settings-changed" {
			return
		}
		if len(data) == 1 {
			if m, ok := data[0].(map[string]any); ok {
				got = m
			}
		}
	})

	if err := svc.SetAutostart(true); err != nil {
		t.Fatal(err)
	}
	view, err := svc.GetSettings()
	if err != nil || !view.Autostart {
		t.Fatalf("autostart=%v err=%v", view.Autostart, err)
	}
	if got["autostart"] != true {
		t.Fatalf("emit=%v", got)
	}

	if err := svc.SetAutostart(false); err != nil {
		t.Fatal(err)
	}
	view, _ = svc.GetSettings()
	if view.Autostart {
		t.Fatal("expected autostart off")
	}

	desktop := filepath.Join(dir, "autostart", "procpane-app.desktop")
	if _, err := os.Stat(desktop); !os.IsNotExist(err) {
		t.Fatalf("desktop should be removed: %v", err)
	}
}

func TestRefreshCurrentSourceEmitsActiveTab(t *testing.T) {
	svc := NewService()
	var mu sync.Mutex
	var last map[string]any
	svc.BindEventEmitter(func(name string, data ...any) {
		mu.Lock()
		defer mu.Unlock()
		if name != "procpane:refresh" {
			t.Fatalf("event=%q", name)
		}
		if len(data) != 1 {
			t.Fatalf("data len=%d", len(data))
		}
		m, ok := data[0].(map[string]string)
		if !ok {
			if ma, ok2 := data[0].(map[string]any); ok2 {
				last = map[string]any{"source": ma["source"]}
				return
			}
			t.Fatalf("type %T", data[0])
		}
		last = map[string]any{"source": m["source"]}
	})
	svc.SetActiveSource("docker")
	svc.RefreshCurrentSource()
	mu.Lock()
	defer mu.Unlock()
	if last["source"] != "docker" {
		t.Fatalf("source=%v", last["source"])
	}
}

func TestSetDockerContextEmitsSettings(t *testing.T) {
	svc := NewService()
	var got bool
	svc.BindEventEmitter(func(name string, data ...any) {
		if name == "procpane:settings-changed" {
			got = true
		}
	})
	_ = svc.SetDockerContext("")
	if !got {
		t.Fatal("expected settings-changed emit")
	}
}
