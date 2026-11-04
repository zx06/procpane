package enrich

import (
	"fmt"
	"os/user"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/zx06/procpane/internal/model"
	"github.com/zx06/procpane/internal/scan"
)

// ListenerID builds a stable row id from port/proto/pid.
func ListenerID(port uint32, proto model.Protocol, pid int32) string {
	return fmt.Sprintf("p%d-%s-%d", port, proto, pid)
}

// EnrichListeners fills process metadata; missing fields become "?".
func EnrichListeners(raw []scan.RawListener) []model.Listener {
	out := make([]model.Listener, 0, len(raw))
	for _, r := range raw {
		out = append(out, enrichOne(r))
	}
	return out
}

func enrichOne(r scan.RawListener) model.Listener {
	l := model.Listener{
		ID:    ListenerID(r.Port, r.Proto, r.PID),
		Port:  r.Port,
		Proto: r.Proto,
		PID:   r.PID,
		Name:  model.Unknown,
		Cmdline: model.Unknown,
		Cwd:     model.Unknown,
		User:    model.Unknown,
		Started: model.Unknown,
	}
	if r.PID <= 0 {
		return l
	}
	p, err := process.NewProcess(r.PID)
	if err != nil {
		return l
	}
	if name, err := p.Name(); err == nil && name != "" {
		l.Name = name
	}
	if cmd, err := p.Cmdline(); err == nil && cmd != "" {
		l.Cmdline = cmd
	} else if parts, err := p.CmdlineSlice(); err == nil && len(parts) > 0 {
		l.Cmdline = strings.Join(parts, " ")
	}
	if cwd, err := p.Cwd(); err == nil && cwd != "" {
		l.Cwd = cwd
	}
	if u, err := p.Username(); err == nil && u != "" {
		l.User = u
	} else if u := lookupUserFromUID(p); u != "" {
		l.User = u
	}
	if ct, err := p.CreateTime(); err == nil && ct > 0 {
		l.Started = time.UnixMilli(ct).Format("2006-01-02 15:04:05")
	}
	if cpu, err := p.CPUPercent(); err == nil {
		l.CPU = round1(cpu)
	}
	if mem, err := p.MemoryInfo(); err == nil && mem != nil {
		l.MemBytes = mem.RSS
		l.MemMB = float64(mem.RSS) / (1024 * 1024)
	}
	l.Chain = buildChain(r.PID)
	if l.Chain == nil {
		l.Chain = []model.ChainNode{}
	}
	l.Leftover = detectLeftover(l.Chain)
	return l
}

// EnrichProcess returns detail for a single PID.
func EnrichProcess(pid int32) model.ProcessInfo {
	raw := scan.RawListener{Port: 0, Proto: model.ProtoTCP, PID: pid}
	l := enrichOne(raw)
	return model.ProcessInfo{
		PID:      l.PID,
		Name:     l.Name,
		Cmdline:  l.Cmdline,
		Cwd:      l.Cwd,
		User:     l.User,
		Started:  l.Started,
		CPU:      l.CPU,
		MemBytes: l.MemBytes,
		MemMB:    l.MemMB,
		Chain:    l.Chain,
	}
}

func buildChain(pid int32) []model.ChainNode {
	var chain []model.ChainNode
	seen := make(map[int32]struct{})
	cur := pid
	for depth := 0; depth < 32 && cur > 0; depth++ {
		if _, ok := seen[cur]; ok {
			break
		}
		seen[cur] = struct{}{}
		p, err := process.NewProcess(cur)
		if err != nil {
			chain = append([]model.ChainNode{{
				PID:  cur,
				Name: model.Unknown,
				Note: "unreadable",
			}}, chain...)
			break
		}
		name, _ := p.Name()
		if name == "" {
			name = model.Unknown
		}
		note := ""
		if cur == pid {
			note = "listener"
		}
		ppid, err := p.Ppid()
		if err != nil {
			note = "orphan? parent unreadable"
		} else if ppid == 0 && cur != 1 {
			note = "orphan? parent exited"
		}
		chain = append([]model.ChainNode{{PID: cur, Name: name, Note: note}}, chain...)
		if ppid <= 0 || ppid == cur {
			break
		}
		cur = ppid
	}
	if len(chain) == 0 {
		return chain
	}
	// annotate root
	if chain[0].Note == "" {
		chain[0].Note = "root"
	}
	return chain
}

func detectLeftover(chain []model.ChainNode) bool {
	if len(chain) < 2 {
		return false
	}
	last := chain[len(chain)-1]
	return strings.Contains(last.Note, "orphan")
}

func lookupUserFromUID(p *process.Process) string {
	uids, err := p.Uids()
	if err != nil || len(uids) == 0 {
		return ""
	}
	u, err := user.LookupId(fmt.Sprintf("%d", uids[0]))
	if err != nil {
		return ""
	}
	return u.Username
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
