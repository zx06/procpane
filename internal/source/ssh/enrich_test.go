package ssh

import (
	"strings"
	"testing"

	"github.com/zx06/procpane/internal/model"
)

func TestParseBatchCommOutput(t *testing.T) {
	raw := strings.Join([]string{
		enrichMarker + "100",
		"node",
		"",
		enrichMarker + "200",
		"sshd",
	}, "\n")
	out := parseBatchCommOutput(raw)
	if out[100] != "node" || out[200] != "sshd" {
		t.Fatalf("out=%v", out)
	}
}

func TestEnrichRemoteSkipsNamedListeners(t *testing.T) {
	listeners := []model.Listener{
		{PID: 1, Name: "nginx"},
		{PID: 2, Name: model.Unknown},
	}
	var need []int32
	for i := range listeners {
		if listeners[i].PID <= 0 {
			continue
		}
		if listeners[i].Name != "" && listeners[i].Name != model.Unknown {
			continue
		}
		need = append(need, listeners[i].PID)
	}
	if len(need) != 1 || need[0] != 2 {
		t.Fatalf("need=%v", need)
	}
}
