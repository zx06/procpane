package format

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/zx06/procpane/internal/model"
)

// WriteListenersTable prints a human-readable table.
func WriteListenersTable(w io.Writer, listeners []model.Listener) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "PORT\tPROTO\tPID\tNAME\tCPU\tMEM")
	for _, l := range listeners {
		mem := formatMem(l.MemMB)
		cpu := fmt.Sprintf("%.1f%%", l.CPU)
		fmt.Fprintf(tw, ":%d\t%s\t%d\t%s\t%s\t%s\n",
			l.Port, strings.ToUpper(string(l.Proto)), l.PID, l.Name, cpu, mem)
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
