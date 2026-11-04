package cmdroot

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zx06/procpane/internal/app"
	"github.com/zx06/procpane/internal/kill"
)

func newKillCmd() *cobra.Command {
	var signal string
	var yes bool
	svc := app.NewService()

	cmd := &cobra.Command{
		Use:   "kill <pid>",
		Short: "Send TERM/KILL to a process (with confirmation)",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pid64, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				exitErr(fmt.Errorf("invalid pid: %w", err))
			}
			pid := int32(pid64)
			if err := kill.ValidateTargetPID(pid); err != nil {
				exitErr(err)
			}
			info := svc.Show(context.Background(), pid)
			fmt.Fprintf(os.Stderr, "Target: PID %d  name=%s  cmdline=%s\n", pid, info.Name, info.Cmdline)
			if !yes {
				fmt.Fprint(os.Stderr, "Send signal "+signal+"? [y/N]: ")
				line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				if strings.TrimSpace(strings.ToLower(line)) != "y" {
					fmt.Fprintln(os.Stderr, "aborted")
					return
				}
			}
			res := svc.KillProcess(pid, signal)
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(res); err != nil {
				exitErr(err)
			}
			if !res.Success {
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringVar(&signal, "signal", "TERM", "Signal: TERM or KILL")
	cmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation")
	return cmd
}
