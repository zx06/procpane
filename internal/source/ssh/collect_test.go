package ssh

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestRemoteListenAttemptsSeparateCommands(t *testing.T) {
	for _, a := range remoteListenAttempts() {
		if len(a.cmds) < 2 {
			t.Fatalf("%s: want tcp+udp commands, got %v", a.tool, a.cmds)
		}
		for _, cmd := range a.cmds {
			if strings.Contains(cmd, ";") {
				t.Fatalf("%s command must not chain with ';' (BusyBox netstat quirk): %q", a.tool, cmd)
			}
		}
	}
}

func TestCollectRemoteAggregatesSeparateRuns(t *testing.T) {
	var calls []string
	run := func(cmd string) (string, error) {
		calls = append(calls, cmd)
		switch {
		case strings.HasPrefix(cmd, "ss "):
			return "", fmt.Errorf("ss: not found")
		case strings.HasPrefix(cmd, "lsof"):
			return "", fmt.Errorf("lsof: not found")
		case cmd == "netstat -ltnp 2>/dev/null":
			return "tcp        0      0 0.0.0.0:8080            0.0.0.0:*               LISTEN      -\n", nil
		case cmd == "netstat -lunp 2>/dev/null":
			return "", nil
		default:
			return "", fmt.Errorf("unexpected cmd: %q", cmd)
		}
	}
	raw, tool, err := collectRemoteWithRun(context.Background(), run)
	if err != nil {
		t.Fatal(err)
	}
	if tool != "netstat" {
		t.Fatalf("tool=%q", tool)
	}
	if len(raw) != 1 || raw[0].Port != 8080 {
		t.Fatalf("raw=%+v", raw)
	}
	wantCalls := []string{
		"ss -H -ltnp 2>/dev/null",
		"ss -H -lunp 2>/dev/null",
		"lsof -nP -iTCP -sTCP:LISTEN 2>/dev/null",
		"lsof -nP -iUDP 2>/dev/null",
		"netstat -ltnp 2>/dev/null",
		"netstat -lunp 2>/dev/null",
	}
	if len(calls) != len(wantCalls) {
		t.Fatalf("calls=%v", calls)
	}
	for i := range wantCalls {
		if calls[i] != wantCalls[i] {
			t.Fatalf("call[%d]=%q want %q", i, calls[i], wantCalls[i])
		}
	}
}

func TestCollectRemoteMergesTCPAndUDP(t *testing.T) {
	run := func(cmd string) (string, error) {
		switch cmd {
		case "netstat -ltnp 2>/dev/null":
			return "tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN      1/-\n", nil
		case "netstat -lunp 2>/dev/null":
			return "udp        0      0 0.0.0.0:68              0.0.0.0:*                           2/dhclient\n", nil
		default:
			return "", fmt.Errorf("skip")
		}
	}
	// Force netstat path only by failing ss/lsof quickly.
	wrapped := func(cmd string) (string, error) {
		if strings.HasPrefix(cmd, "ss ") || strings.HasPrefix(cmd, "lsof") {
			return "", fmt.Errorf("missing")
		}
		return run(cmd)
	}
	raw, tool, err := collectRemoteWithRun(context.Background(), wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if tool != "netstat" || len(raw) != 2 {
		t.Fatalf("tool=%q raw=%+v", tool, raw)
	}
}
