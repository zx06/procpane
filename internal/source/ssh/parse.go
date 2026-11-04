package ssh

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/zx06/procpane/internal/model"
)

var (
	ssTCPRe = regexp.MustCompile(`(?i)^LISTEN\s+\d+\s+\d+\s+(\S+):(\d+)\s+.*users:\(\("([^"]+)",pid=(\d+)`)
	ssUDPRe = regexp.MustCompile(`(?i)^UNCONN\s+\d+\s+\d+\s+(\S+):(\d+)\s+.*users:\(\("([^"]+)",pid=(\d+)`)
	// ss without process info: LISTEN 0 128 *:8080 *:*
	ssPlainTCP = regexp.MustCompile(`(?i)^LISTEN\s+\d+\s+\d+\s+(\S+):(\d+)\s`)
	ssPlainUDP = regexp.MustCompile(`(?i)^UNCONN\s+\d+\s+\d+\s+(\S+):(\d+)\s`)

	lsofTCPRe = regexp.MustCompile(`(?i)^(\S+)\s+\d+\s+\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+TCP\s+(\S+):(\d+)\s+\(LISTEN\)`)
	lsofUDPRe = regexp.MustCompile(`(?i)^(\S+)\s+\d+\s+\S+\s+\S+\s+\S+\s+\S+\s+\S+\s+UDP\s+(\S+):(\d+)\s`)

	// PID/program may be "888/sshd" or "-" when non-root (no -p permission).
	netstatTCPRe = regexp.MustCompile(`(?i)^tcp\s+\d+\s+\d+\s+(\S+):(\d+)\s+\S+\s+LISTEN\s+(-|\d+)(?:/(\S+))?`)
	netstatUDPRe = regexp.MustCompile(`(?i)^udp\s+\d+\s+\d+\s+(\S+):(\d+)\s+\S*\s*(-|\d+)(?:/(\S+))?`)
)

type rawListen struct {
	Port  uint32
	Proto model.Protocol
	PID   int32
	Name  string
}

func parseListenOutput(tool, output string) []rawListen {
	var out []rawListen
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if r, ok := parseLine(tool, line); ok {
			out = append(out, r)
		}
	}
	return dedupeRaw(out)
}

func parseLine(tool, line string) (rawListen, bool) {
	switch tool {
	case "ss":
		return parseSSLine(line)
	case "lsof":
		return parseLsofLine(line)
	case "netstat":
		return parseNetstatLine(line)
	default:
		return rawListen{}, false
	}
}

func parseSSLine(line string) (rawListen, bool) {
	if m := ssTCPRe.FindStringSubmatch(line); len(m) >= 5 {
		port := parsePort(m[2])
		pid, _ := strconv.ParseInt(m[4], 10, 32)
		return rawListen{Port: port, Proto: model.ProtoTCP, PID: int32(pid), Name: m[3]}, port > 0
	}
	if m := ssUDPRe.FindStringSubmatch(line); len(m) >= 5 {
		port := parsePort(m[2])
		pid, _ := strconv.ParseInt(m[4], 10, 32)
		return rawListen{Port: port, Proto: model.ProtoUDP, PID: int32(pid), Name: m[3]}, port > 0
	}
	if m := ssPlainTCP.FindStringSubmatch(line); len(m) >= 3 {
		port := parsePort(m[2])
		return rawListen{Port: port, Proto: model.ProtoTCP}, port > 0
	}
	if m := ssPlainUDP.FindStringSubmatch(line); len(m) >= 3 {
		port := parsePort(m[2])
		return rawListen{Port: port, Proto: model.ProtoUDP}, port > 0
	}
	return rawListen{}, false
}

func parseLsofLine(line string) (rawListen, bool) {
	if m := lsofTCPRe.FindStringSubmatch(line); len(m) >= 4 {
		port := parsePort(m[3])
		return rawListen{Port: port, Proto: model.ProtoTCP, Name: m[1]}, port > 0
	}
	if m := lsofUDPRe.FindStringSubmatch(line); len(m) >= 4 {
		port := parsePort(m[3])
		return rawListen{Port: port, Proto: model.ProtoUDP, Name: m[1]}, port > 0
	}
	return rawListen{}, false
}

func parseNetstatLine(line string) (rawListen, bool) {
	if m := netstatTCPRe.FindStringSubmatch(line); len(m) >= 4 {
		port := parsePort(m[2])
		pid := netstatPID(m[3])
		name := ""
		if len(m) >= 5 {
			name = m[4]
		}
		return rawListen{Port: port, Proto: model.ProtoTCP, PID: pid, Name: name}, port > 0
	}
	if m := netstatUDPRe.FindStringSubmatch(line); len(m) >= 4 {
		port := parsePort(m[2])
		pid := netstatPID(m[3])
		name := ""
		if len(m) >= 5 {
			name = m[4]
		}
		return rawListen{Port: port, Proto: model.ProtoUDP, PID: pid, Name: name}, port > 0
	}
	return rawListen{}, false
}

func netstatPID(field string) int32 {
	if field == "" || field == "-" {
		return 0
	}
	pid, _ := strconv.ParseInt(field, 10, 32)
	return int32(pid)
}

func parsePort(s string) uint32 {
	// handle [ipv6]:port
	if strings.HasPrefix(s, "[") {
		if idx := strings.LastIndex(s, "]:"); idx >= 0 {
			s = s[idx+2:]
		}
	} else if idx := strings.LastIndex(s, ":"); idx >= 0 {
		s = s[idx+1:]
	}
	n, _ := strconv.ParseUint(s, 10, 32)
	return uint32(n)
}

func dedupeRaw(in []rawListen) []rawListen {
	seen := make(map[string]struct{})
	var out []rawListen
	for _, r := range in {
		key := fmtKey(r)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, r)
	}
	return out
}

func fmtKey(r rawListen) string {
	return string(r.Proto) + ":" + strconv.FormatUint(uint64(r.Port), 10) + ":" + strconv.FormatInt(int64(r.PID), 10)
}

func rawToListeners(host string, raw []rawListen) []model.Listener {
	out := make([]model.Listener, 0, len(raw))
	for _, r := range raw {
		name := r.Name
		if name == "" {
			name = model.Unknown
		}
		pid := r.PID
		id := fmt.Sprintf("ssh-%s-%s-%d-%d", host, r.Proto, r.Port, pid)
		out = append(out, model.Listener{
			ID:      id,
			Port:    r.Port,
			Proto:   r.Proto,
			PID:     pid,
			Name:    name,
			Cmdline: model.Unknown,
			Cwd:     model.Unknown,
			User:    model.Unknown,
			Started: model.Unknown,
			Chain:   []model.ChainNode{},
			Source:  "ssh",
			Host:    host,
		})
	}
	return out
}
