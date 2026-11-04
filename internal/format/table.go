package format

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/zx06/procpane/internal/model"
)

type pidRow struct {
	pid     int32
	name    string
	cpu     float64
	memMB   float64
	ports   []string
}

// WriteListenersTable prints a human-readable table (one row per PID).
func WriteListenersTable(w io.Writer, listeners []model.Listener) error {
	byPID := make(map[int32]*pidRow)
	var order []int32
	for _, l := range listeners {
		portLabel := fmt.Sprintf(":%d/%s", l.Port, strings.ToUpper(string(l.Proto)))
		row, ok := byPID[l.PID]
		if !ok {
			row = &pidRow{
				pid:   l.PID,
				name:  l.Name,
				cpu:   l.CPU,
				memMB: l.MemMB,
				ports: []string{portLabel},
			}
			byPID[l.PID] = row
			order = append(order, l.PID)
		} else {
			row.ports = append(row.ports, portLabel)
		}
	}

	sort.Slice(order, func(i, j int) bool {
		a, b := byPID[order[i]], byPID[order[j]]
		if len(a.ports) == 0 || len(b.ports) == 0 {
			return a.pid < b.pid
		}
		return a.ports[0] < b.ports[0] || (a.ports[0] == b.ports[0] && a.pid < b.pid)
	})

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "PORTS\tPID\tNAME\tCPU\tMEM")
	for _, pid := range order {
		row := byPID[pid]
		sort.Strings(row.ports)
		mem := formatMem(row.memMB)
		cpu := fmt.Sprintf("%.1f%%", row.cpu)
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t%s\n",
			strings.Join(row.ports, " "), row.pid, row.name, cpu, mem)
	}
	return tw.Flush()
}

func formatMem(mb float64) string {
	if mb <= 0 {
		return model.Unknown
	}
	if mb >= 100 {
		return fmt.Sprintf("%.0f MB", mb)
	}
	return fmt.Sprintf("%.1f MB", mb)
}
