package scan

import (
	"context"
	"fmt"
	"net"
	"strings"

	gnet "github.com/shirou/gopsutil/v4/net"
	"github.com/zx06/procpane/internal/model"
)

// RawConnection is one TCP ESTABLISHED socket row before enrich/pairing.
type RawConnection struct {
	Proto      model.Protocol
	Status     string
	LocalHost  string
	LocalPort  uint32
	RemoteHost string
	RemotePort uint32
	PID        int32
}

// ListRawConnections returns TCP ESTABLISHED rows from the OS (local host only).
func ListRawConnections(ctx context.Context) ([]RawConnection, error) {
	_ = ctx
	conns, err := gnet.Connections("tcp")
	if err != nil {
		return nil, fmt.Errorf("connections(tcp): %w", err)
	}
	localAddrs, _ := localInterfaceAddrs()

	var out []RawConnection
	seen := make(map[string]struct{})

	for _, c := range conns {
		if !isEstablished(c.Status) {
			continue
		}
		if c.Laddr.Port == 0 {
			continue
		}
		lh := normalizeHost(c.Laddr.IP)
		rh := normalizeHost(c.Raddr.IP)
		if rh == "" || c.Raddr.Port == 0 {
			// Half-open or not fully connected; skip for graph.
			continue
		}
		key := fmt.Sprintf("%s:%d-%s:%d-%d", lh, c.Laddr.Port, rh, c.Raddr.Port, c.Pid)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, RawConnection{
			Proto:      model.ProtoTCP,
			Status:     strings.ToUpper(strings.TrimSpace(c.Status)),
			LocalHost:  lh,
			LocalPort:  c.Laddr.Port,
			RemoteHost: rh,
			RemotePort: c.Raddr.Port,
			PID:        c.Pid,
		})
	}
	_ = localAddrs // used by pairing via IsLocalHost
	return out, nil
}

func isEstablished(status string) bool {
	return strings.EqualFold(strings.TrimSpace(status), "ESTABLISHED")
}

// normalizeHost strips zone ids and maps IPv4-mapped IPv6 to IPv4 text.
func normalizeHost(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}
	if strings.HasPrefix(ip, "::ffff:") {
		ip = strings.TrimPrefix(ip, "::ffff:")
	}
	if strings.Contains(ip, "%") {
		ip, _, _ = strings.Cut(ip, "%")
	}
	return ip
}

// localInterfaceAddrs returns normalized IPs assigned to this host (loopback + NICs).
func localInterfaceAddrs() (map[string]struct{}, error) {
	out := map[string]struct{}{
		"127.0.0.1": {},
		"::1":       {},
		"0.0.0.0":   {},
		"::":        {},
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return out, err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				out[ip4.String()] = struct{}{}
				continue
			}
			out[ip.String()] = struct{}{}
		}
	}
	return out, nil
}

// IsLocalHost reports whether ip belongs to this machine (loopback, NIC, or wildcard).
func IsLocalHost(ip string, localAddrs map[string]struct{}) bool {
	ip = normalizeHost(ip)
	if ip == "" {
		return false
	}
	if strings.HasPrefix(ip, "127.") {
		return true
	}
	if ip == "::1" {
		return true
	}
	if _, ok := localAddrs[ip]; ok {
		return true
	}
	return false
}
