package enrich

import (
	"fmt"
	"time"

	"github.com/zx06/procpane/internal/model"
	"github.com/zx06/procpane/internal/scan"
)

// ConnectionID builds a stable row id from endpoints and pid.
func ConnectionID(r scan.RawConnection) string {
	return fmt.Sprintf("c-%s-%s:%d-%s:%d-%d",
		r.Proto, r.LocalHost, r.LocalPort, r.RemoteHost, r.RemotePort, r.PID)
}

// EnrichConnections fills process names for connection rows.
func EnrichConnections(paired []scan.PairedConnection) []model.Connection {
	names := map[int32]string{}
	out := make([]model.Connection, 0, len(paired))
	for _, p := range paired {
		r := p.Raw
		name := lookupName(r.PID, names)
		peerName := ""
		if p.PeerPID > 0 {
			peerName = lookupName(p.PeerPID, names)
		}
		out = append(out, model.Connection{
			ID:         ConnectionID(r),
			Proto:      r.Proto,
			Status:     r.Status,
			LocalHost:  r.LocalHost,
			LocalPort:  r.LocalPort,
			RemoteHost: r.RemoteHost,
			RemotePort: r.RemotePort,
			PID:        r.PID,
			Name:       name,
			Direction:  p.Direction,
			PeerPID:    p.PeerPID,
			PeerName:   peerName,
		})
	}
	return out
}

func lookupName(pid int32, cache map[int32]string) string {
	if pid <= 0 {
		return model.Unknown
	}
	if n, ok := cache[pid]; ok {
		return n
	}
	raw := scan.RawListener{Port: 0, Proto: model.ProtoTCP, PID: pid}
	l := enrichOne(raw)
	cache[pid] = l.Name
	return l.Name
}

// BuildConnectionGraphResponse pairs, enriches, and timestamps a graph payload.
func BuildConnectionGraphResponse(raw []scan.RawConnection) model.ConnectionGraphResponse {
	paired := scan.PairLocalConnections(raw)
	return model.ConnectionGraphResponse{
		GeneratedAt:   time.Now().UTC(),
		Connections: EnrichConnections(paired),
	}
}
