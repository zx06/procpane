package scan

import (
	"fmt"
	"sort"

	"github.com/zx06/procpane/internal/model"
)

// PairedConnection augments a raw row with direction and optional local peer PID.
type PairedConnection struct {
	Raw           RawConnection
	Direction     model.ConnDirection
	PeerPID       int32
	PeerLocalHost string
	PeerLocalPort uint32
}

type connKey struct {
	localHost  string
	localPort  uint32
	remoteHost string
	remotePort uint32
}

func (k connKey) String() string {
	return fmt.Sprintf("%s:%d-%s:%d", k.localHost, k.localPort, k.remoteHost, k.remotePort)
}

// PairLocalConnections classifies direction and matches reverse ESTABLISHED rows for local↔local links.
//
// Local address heuristics: loopback (127.0.0.0/8, ::1), wildcard bind addrs, and primary NIC IPs
// from net.Interfaces. Remote IPs matching those count as on-host for pairing.
//
// Inbound vs outbound (external peers only): compare local and remote ports — server-side sockets
// usually bind the lower/well-known port, so localPort <= remotePort → inbound, else outbound.
// This is a heuristic (documented); it is not PTR/DNS and does not inspect packet flow.
func PairLocalConnections(raw []RawConnection) []PairedConnection {
	localAddrs, _ := localInterfaceAddrs()
	index := make(map[string]RawConnection, len(raw))
	for _, r := range raw {
		k := connKey{r.LocalHost, r.LocalPort, r.RemoteHost, r.RemotePort}
		index[k.String()] = r
	}

	out := make([]PairedConnection, 0, len(raw))
	for _, r := range raw {
		pc := PairedConnection{Raw: r}
		remoteLocal := IsLocalHost(r.RemoteHost, localAddrs)
		if remoteLocal {
			// Always local interconnect class — never inbound/outbound external, even when unpaired.
			pc.Direction = model.ConnLocal
			rev := connKey{r.RemoteHost, r.RemotePort, r.LocalHost, r.LocalPort}
			if peer, ok := index[rev.String()]; ok && peer.PID > 0 && peer.PID != r.PID {
				pc.PeerPID = peer.PID
				pc.PeerLocalHost = peer.LocalHost
				pc.PeerLocalPort = peer.LocalPort
			}
		} else if r.LocalPort <= r.RemotePort {
			pc.Direction = model.ConnInbound
		} else {
			pc.Direction = model.ConnOutbound
		}
		out = append(out, pc)
	}

	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].Raw, out[j].Raw
		if a.LocalPort != b.LocalPort {
			return a.LocalPort < b.LocalPort
		}
		if a.RemotePort != b.RemotePort {
			return a.RemotePort < b.RemotePort
		}
		return a.PID < b.PID
	})
	return out
}
