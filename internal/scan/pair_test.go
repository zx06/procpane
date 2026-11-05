package scan

import (
	"testing"

	"github.com/zx06/procpane/internal/model"
)

func TestPairLocalConnections_localInterconnect(t *testing.T) {
	raw := []RawConnection{
		{Proto: model.ProtoTCP, Status: "ESTABLISHED", LocalHost: "127.0.0.1", LocalPort: 53102, RemoteHost: "127.0.0.1", RemotePort: 80, PID: 4012},
		{Proto: model.ProtoTCP, Status: "ESTABLISHED", LocalHost: "127.0.0.1", LocalPort: 80, RemoteHost: "127.0.0.1", RemotePort: 53102, PID: 412},
	}
	paired := PairLocalConnections(raw)
	if len(paired) != 2 {
		t.Fatalf("got %d paired rows, want 2", len(paired))
	}
	for _, p := range paired {
		if p.Direction != model.ConnLocal {
			t.Errorf("pid %d: direction %q, want local", p.Raw.PID, p.Direction)
		}
	}
	byPID := map[int32]int32{}
	for _, p := range paired {
		byPID[p.Raw.PID] = p.PeerPID
	}
	if byPID[4012] != 412 || byPID[412] != 4012 {
		t.Errorf("peer pairing missing: %+v", byPID)
	}
}

func TestPairLocalConnections_inboundOutboundHeuristic(t *testing.T) {
	raw := []RawConnection{
		{LocalHost: "10.0.0.12", LocalPort: 80, RemoteHost: "203.0.113.9", RemotePort: 54321, PID: 412},
		{LocalHost: "10.0.0.12", LocalPort: 55110, RemoteHost: "140.82.112.3", RemotePort: 443, PID: 2201},
		{LocalHost: "10.0.0.12", LocalPort: 53188, RemoteHost: "192.168.1.8", RemotePort: 22, PID: 771},
	}
	localAddrs := map[string]struct{}{"10.0.0.12": {}}
	for _, r := range raw {
		if IsLocalHost(r.RemoteHost, localAddrs) {
			t.Fatalf("remote should be external for this fixture")
		}
	}
	paired := PairLocalConnections(raw)
	dirByPID := map[int32]model.ConnDirection{}
	for _, p := range paired {
		dirByPID[p.Raw.PID] = p.Direction
	}
	if dirByPID[412] != model.ConnInbound {
		t.Errorf("nginx: got %q want inbound", dirByPID[412])
	}
	if dirByPID[2201] != model.ConnOutbound {
		t.Errorf("node: got %q want outbound", dirByPID[2201])
	}
	if dirByPID[771] != model.ConnOutbound {
		t.Errorf("ssh: got %q want outbound", dirByPID[771])
	}
}

func TestIsLocalHost_loopback(t *testing.T) {
	localAddrs, _ := localInterfaceAddrs()
	for _, ip := range []string{"127.0.0.1", "127.0.0.2", "::1", "::ffff:127.0.0.1"} {
		if !IsLocalHost(ip, localAddrs) {
			t.Errorf("%q should be local", ip)
		}
	}
	if IsLocalHost("203.0.113.9", localAddrs) {
		t.Error("203.0.113.9 should not be local")
	}
}

func TestIsEstablished_filter(t *testing.T) {
	for _, s := range []string{"ESTABLISHED", "established", " Established "} {
		if !isEstablished(s) {
			t.Errorf("%q should be established", s)
		}
	}
	for _, s := range []string{"LISTEN", "TIME_WAIT", "NONE"} {
		if isEstablished(s) {
			t.Errorf("%q should not be established", s)
		}
	}
}
