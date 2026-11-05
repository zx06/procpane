package gui

import (
	"sync"
	"testing"
)

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
