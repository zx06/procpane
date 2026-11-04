package cmdroot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/zx06/procpane/internal/config"
	"github.com/zx06/procpane/internal/format"
	sshsrc "github.com/zx06/procpane/internal/source/ssh"
)

func newSSHCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh",
		Short: "Remote LISTEN via SSH (enabled hosts only for ls)",
	}
	cmd.AddCommand(newSSHLsCmd(), newSSHHostsCmd())
	return cmd
}

func newSSHLsCmd() *cobra.Command {
	var asJSON bool
	var watch time.Duration

	cmd := &cobra.Command{
		Use:   "ls <host>",
		Short: "List LISTEN ports on an enabled SSH host alias",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			alias := args[0]
			store, err := config.DefaultStore()
			if err != nil {
				exitErr(err)
			}
			on, err := store.IsSSHEnabled(alias)
			if err != nil {
				exitErr(err)
			}
			if !on {
				exitErr(fmt.Errorf("%w: %s", config.ErrHostNotEnabled, alias))
			}
			src := sshsrc.New(alias)
			run := func() error {
				ctx := context.Background()
				listeners, err := src.ListListeners(ctx)
				if err != nil {
					return err
				}
				if asJSON {
					return format.WriteListenersJSON(os.Stdout, listeners)
				}
				return format.WriteListenersTable(os.Stdout, listeners)
			}
			if watch <= 0 {
				if err := run(); err != nil {
					exitErr(err)
				}
				return
			}
			for {
				if err := run(); err != nil {
					exitErr(err)
				}
				time.Sleep(watch)
			}
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit JSON")
	cmd.Flags().DurationVarP(&watch, "watch", "w", 0, "Refresh interval")
	return cmd
}

func newSSHHostsCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "hosts",
		Short: "List OpenSSH config host aliases and enable state",
		Run: func(cmd *cobra.Command, args []string) {
			store, err := config.DefaultStore()
			if err != nil {
				exitErr(err)
			}
			st, err := store.Load()
			if err != nil {
				exitErr(err)
			}
			hosts, err := sshsrc.ListHosts()
			if err != nil {
				exitErr(err)
			}
			type row struct {
				Alias    string `json:"alias"`
				Hostname string `json:"hostname"`
				Port     string `json:"port,omitempty"`
				Enabled  bool   `json:"enabled"`
			}
			var rows []row
			for _, h := range hosts {
				rows = append(rows, row{
					Alias:    h.Alias,
					Hostname: h.Hostname,
					Port:     h.Port,
					Enabled:  st.SSH.Enabled[h.Alias],
				})
			}
			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(rows); err != nil {
					exitErr(err)
				}
				return
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "ALIAS\tHOSTNAME\tPORT\tENABLED")
			for _, r := range rows {
				port := r.Port
				if port == "" {
					port = "22"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%v\n", r.Alias, r.Hostname, port, r.Enabled)
			}
			_ = tw.Flush()
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit JSON")
	return cmd
}
