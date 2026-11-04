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

func TestRetryRemoteOutputEmptyThenData(t *testing.T) {
	var n int
	out, err := retryRemoteOutput(3, func() (string, error) {
		n++
		if n < 3 {
			return "", nil
		}
		return "tcp listen row\n", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if out != "tcp listen row\n" {
		t.Fatalf("out=%q", out)
	}
	if n != 3 {
		t.Fatalf("attempts=%d want 3", n)
	}
}

func TestRetryRemoteOutputStopsOnError(t *testing.T) {
	var n int
	_, err := retryRemoteOutput(3, func() (string, error) {
		n++
		return "", fmt.Errorf("127")
	})
	if err == nil {
		t.Fatal("want error")
	}
	if n != 1 {
		t.Fatalf("attempts=%d want 1", n)
	}
}

func TestCollectRemoteWithRunRemoteStyleRetries(t *testing.T) {
	attempts := map[string]int{}
	run := func(cmd string) (string, error) {
		return retryRemoteOutput(3, func() (string, error) {
			attempts[cmd]++
			switch cmd {
			case "ss -H -ltnp 2>/dev/null", "ss -H -lunp 2>/dev/null",
				"lsof -nP -iTCP -sTCP:LISTEN 2>/dev/null", "lsof -nP -iUDP 2>/dev/null":
				return "", fmt.Errorf("missing")
			case "netstat -ltnp 2>/dev/null":
				if attempts[cmd] < 2 {
					return "", nil
				}
				return "tcp        0      0 0.0.0.0:8080            0.0.0.0:*               LISTEN      -\n", nil
			case "netstat -lunp 2>/dev/null":
				return "", nil
			default:
				return "", fmt.Errorf("unexpected %q", cmd)
			}
		})
	}
	raw, tool, err := collectRemoteWithRun(context.Background(), run)
	if err != nil {
		t.Fatal(err)
	}
	if tool != "netstat" || len(raw) != 1 {
		t.Fatalf("tool=%q raw=%+v", tool, raw)
	}
	if attempts["netstat -ltnp 2>/dev/null"] != 2 {
		t.Fatalf("tcp attempts=%d want 2", attempts["netstat -ltnp 2>/dev/null"])
	}
}

func TestCollectRemoteSSSupplementsNetstat(t *testing.T) {
	run := func(cmd string) (string, error) {
		switch cmd {
		case "ss -H -ltnp 2>/dev/null":
			return "LISTEN 0 128 *:22 *:* users:((\"sshd\",pid=1,fd=3))\n", nil
		case "ss -H -lunp 2>/dev/null":
			return "UNCONN 0 0 0.0.0.0:68 0.0.0.0:*\nUNCONN 0 0 0.0.0.0:111 0.0.0.0:*\n", nil
		case "netstat -ltnp 2>/dev/null":
			return "", nil
		case "netstat -lunp 2>/dev/null":
			return "udp        0      0 0.0.0.0:68              0.0.0.0:*                           2/dhclient\n", nil
		default:
			return "", fmt.Errorf("skip")
		}
	}
	raw, tool, err := collectRemoteWithRun(context.Background(), run)
	if err != nil {
		t.Fatal(err)
	}
	if tool != "ss" {
		t.Fatalf("tool=%q", tool)
	}
	var port68 *rawListen
	for i := range raw {
		if raw[i].Port == 68 {
			port68 = &raw[i]
			break
		}
	}
	if port68 == nil || port68.PID != 2 || port68.Name != "dhclient" {
		t.Fatalf("port 68: %+v all=%+v", port68, raw)
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
