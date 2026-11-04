package ssh

import (
	"testing"

	"github.com/zx06/procpane/internal/model"
)

func TestParseSS(t *testing.T) {
	sample := `LISTEN 0 128 *:8080 *:* users:(("nginx",pid=1234,fd=3))
LISTEN 0 128 127.0.0.1:5432 *:* users:(("postgres",pid=99,fd=5))
UNCONN 0 0 0.0.0.0:5353 0.0.0.0:* users:(("systemd-resolve",pid=500,fd=12))
`
	raw := parseListenOutput("ss", sample)
	if len(raw) < 3 {
		t.Fatalf("expected 3 rows, got %d", len(raw))
	}
	if raw[0].Port != 8080 || raw[0].Proto != model.ProtoTCP || raw[0].PID != 1234 {
		t.Fatalf("tcp row: %+v", raw[0])
	}
}

func TestParseLsof(t *testing.T) {
	sample := `nginx   1234 user  3u  IPv4 12345      0t0  TCP *:8080 (LISTEN)
`
	raw := parseListenOutput("lsof", sample)
	if len(raw) != 1 || raw[0].Port != 8080 || raw[0].Name != "nginx" {
		t.Fatalf("lsof: %+v", raw)
	}
}

func TestParseNetstat(t *testing.T) {
	sample := `tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN      888/sshd
udp        0      0 0.0.0.0:68              0.0.0.0:*                           777/dhclient
`
	raw := parseListenOutput("netstat", sample)
	if len(raw) != 2 {
		t.Fatalf("expected 2, got %d: %+v", len(raw), raw)
	}
	if raw[0].PID != 888 || raw[0].Name != "sshd" {
		t.Fatalf("tcp: %+v", raw[0])
	}
}

func TestParseNetstatNoRootPID(t *testing.T) {
	// Non-root netstat -ltnp: LISTEN rows show "-" instead of pid/program.
	sample := `tcp        0      0 0.0.0.0:8080            0.0.0.0:*               LISTEN      -
tcp        0      0 127.0.0.1:445           0.0.0.0:*               LISTEN      -
`
	raw := parseListenOutput("netstat", sample)
	if len(raw) != 2 {
		t.Fatalf("expected 2, got %d: %+v", len(raw), raw)
	}
	if raw[0].Port != 8080 || raw[0].Proto != model.ProtoTCP || raw[0].PID != 0 {
		t.Fatalf("first row: %+v", raw[0])
	}
	if raw[1].Port != 445 || raw[1].PID != 0 {
		t.Fatalf("second row: %+v", raw[1])
	}
}
