package ssh

import (
	"context"
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

func TestBatchRemoteCommRetriesEmptyEnrichOutput(t *testing.T) {
	var n int
	run := func(cmd string) (string, error) {
		return retryRemoteOutput(remoteEmptyOutputAttempts, func() (string, error) {
			n++
			if n < 2 {
				return "", nil
			}
			return strings.Join([]string{
				enrichMarker + "197",
				"sshd.pam",
				"",
			}, "\n"), nil
		})
	}
	names := batchRemoteCommWithRun(context.Background(), []int32{197}, run)
	if names[197] != "sshd.pam" {
		t.Fatalf("names=%v", names)
	}
	if n != 2 {
		t.Fatalf("attempts=%d want 2", n)
	}
}

func TestEnrichRemoteOverwritesTruncatedNetstatName(t *testing.T) {
	listeners := []model.Listener{
		{PID: 888, Name: "ssh_host_rsa_ke"},
		{PID: 2, Name: model.Unknown},
	}
	pids := pidsForCommEnrich(listeners)
	if len(pids) != 2 || pids[0] != 888 || pids[1] != 2 {
		t.Fatalf("pids=%v", pids)
	}
	names := map[int32]string{888: "sshd.pam", 2: "dhclient"}
	applyCommNames(listeners, names)
	if listeners[0].Name != "sshd.pam" || listeners[1].Name != "dhclient" {
		t.Fatalf("listeners=%+v", listeners)
	}
}
