package scan

import (
	"context"
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/zx06/procpane/internal/model"
)

// RawListener is a LISTEN socket before enrich.
type RawListener struct {
	Port  uint32
	Proto model.Protocol
	PID   int32
}

// ListRawListeners returns LISTEN sockets for tcp/udp from the OS.
func ListRawListeners(ctx context.Context, protoFilter string) ([]RawListener, error) {
	_ = ctx
	kind := []string{"tcp", "udp"}
	switch strings.ToLower(protoFilter) {
	case "tcp":
		kind = []string{"tcp"}
	case "udp":
		kind = []string{"udp"}
	case "all", "":
	default:
		return nil, fmt.Errorf("unknown proto filter %q (use tcp, udp, or all)", protoFilter)
	}

	var out []RawListener
	seen := make(map[string]struct{})

	for _, k := range kind {
		conns, err := net.Connections(k)
		if err != nil {
			return nil, fmt.Errorf("connections(%s): %w", k, err)
		}
		for _, c := range conns {
			if !isListen(c.Status) {
				continue
			}
			if c.Laddr.Port == 0 {
				continue
			}
			p := model.Protocol(strings.ToLower(k))
			key := fmt.Sprintf("%s:%d:%d", p, c.Laddr.Port, c.Pid)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, RawListener{
				Port:  c.Laddr.Port,
				Proto: p,
				PID:   c.Pid,
			})
		}
	}
	return out, nil
}

func isListen(status string) bool {
	s := strings.ToUpper(strings.TrimSpace(status))
	return s == "LISTEN" || s == "NONE" // UDP often reports NONE on Linux
}
