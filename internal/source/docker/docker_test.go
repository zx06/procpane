package docker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zx06/procpane/internal/model"
)

func TestListListenersMockAPI(t *testing.T) {
	body := []map[string]any{
		{
			"Id":    "abc123def4567890123456789012345678901234",
			"Names": []string{"/web"},
			"Image": "nginx:latest",
			"Ports": []map[string]any{
				{"IP": "0.0.0.0", "PrivatePort": 80, "PublicPort": 8080, "Type": "tcp"},
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_ping" || strings.HasSuffix(r.URL.Path, "/_ping") {
			w.WriteHeader(http.StatusOK)
			return
		}
		if strings.Contains(r.URL.Path, "containers/json") {
			_ = json.NewEncoder(w).Encode(body)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	src := &Source{Host: srv.URL}
	list, err := src.ListListeners(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 listener, got %d", len(list))
	}
	l := list[0]
	if l.Source != "docker" || l.Port != 8080 || l.ContainerName != "web" {
		t.Fatalf("listener: %+v", l)
	}
	if l.Proto != model.ProtoTCP {
		t.Fatalf("proto %s", l.Proto)
	}
}
