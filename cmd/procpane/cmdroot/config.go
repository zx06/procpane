package cmdroot

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/zx06/procpane/internal/config"
	dockersrc "github.com/zx06/procpane/internal/source/docker"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "procpane settings (SSH host enable state)",
	}
	cmd.AddCommand(newConfigShowCmd(), newConfigEnableCmd(), newConfigDisableCmd())
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show config paths, Docker endpoint, SSH enable flags",
		Run: func(cmd *cobra.Command, args []string) {
			store, err := config.DefaultStore()
			if err != nil {
				exitErr(err)
			}
			host := dockersrc.ResolveHost()
			dockerErr := dockersrc.Ping(cmd.Context(), host)
			sum, err := config.ShowSummary(store, host, dockerErr == nil, dockerErr)
			if err != nil {
				exitErr(err)
			}
			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(sum); err != nil {
					exitErr(err)
				}
				return
			}
			fmt.Printf("state file: %s\n", sum.StatePath)
			fmt.Printf("docker:     %s (reachable=%v)\n", sum.DockerHost, sum.DockerReachable)
			if sum.DockerError != "" {
				fmt.Printf("docker err: %s\n", sum.DockerError)
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "\nSSH ENABLED")
			if len(sum.SSHEnabled) == 0 {
				fmt.Fprintln(tw, "(none)")
			} else {
				for alias, on := range sum.SSHEnabled {
					if on {
						fmt.Fprintf(tw, "%s\ttrue\n", alias)
					}
				}
			}
			_ = tw.Flush()
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit JSON")
	return cmd
}

func newConfigEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable <ssh-host-alias>",
		Short: "Enable an SSH Host alias for procpane",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store, err := config.DefaultStore()
			if err != nil {
				exitErr(err)
			}
			if err := store.SetSSHEnabled(args[0], true); err != nil {
				exitErr(err)
			}
			fmt.Printf("enabled SSH host %q\n", args[0])
		},
	}
}

func newConfigDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <ssh-host-alias>",
		Short: "Disable an SSH Host alias",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store, err := config.DefaultStore()
			if err != nil {
				exitErr(err)
			}
			if err := store.SetSSHEnabled(args[0], false); err != nil {
				exitErr(err)
			}
			fmt.Printf("disabled SSH host %q\n", args[0])
		},
	}
}
