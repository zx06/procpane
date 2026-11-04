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
	cmd.AddCommand(
		newConfigShowCmd(),
		newConfigEnableCmd(),
		newConfigDisableCmd(),
		newConfigDockerContextCmd(),
		newConfigDockerContextsCmd(),
	)
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
			selected, err := store.DockerContext()
			if err != nil {
				exitErr(err)
			}
			host, dockerErr := dockersrc.PingContext(cmd.Context(), selected)
			sum, err := config.ShowSummary(store, selected, host, dockerErr == nil, dockerErr)
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
			if sum.DockerContext != "" {
				fmt.Printf("docker ctx: %s\n", sum.DockerContext)
			} else {
				fmt.Printf("docker ctx: (auto — DOCKER_HOST / local socket)\n")
			}
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

func newConfigDockerContextsCmd() *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "docker-contexts",
		Aliases: []string{"docker-contexts-list"},
		Short:   "List local Docker contexts (same store as docker CLI)",
		Run: func(cmd *cobra.Command, args []string) {
			store, err := config.DefaultStore()
			if err != nil {
				exitErr(err)
			}
			selected, err := store.DockerContext()
			if err != nil {
				exitErr(err)
			}
			ctxs, err := dockersrc.ListContexts(selected)
			if err != nil {
				exitErr(err)
			}
			if asJSON {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				if err := enc.Encode(ctxs); err != nil {
					exitErr(err)
				}
				return
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(tw, "CURRENT\tSELECTED\tNAME\tENDPOINT\tDESCRIPTION")
			for _, c := range ctxs {
				cur, sel := "", ""
				if c.Current {
					cur = "*"
				}
				if c.Selected {
					sel = "*"
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", cur, sel, c.Name, c.Host, c.Description)
			}
			_ = tw.Flush()
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit JSON")
	return cmd
}

func newConfigDockerContextCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docker-context [name]",
		Short: "Set or clear the Docker context procpane uses",
		Long:  "With no name, prints the current selection. Use --clear to follow DOCKER_HOST / local socket.",
		Run: func(cmd *cobra.Command, args []string) {
			store, err := config.DefaultStore()
			if err != nil {
				exitErr(err)
			}
			clear, _ := cmd.Flags().GetBool("clear")
			if clear {
				if err := store.SetDockerContext(""); err != nil {
					exitErr(err)
				}
				fmt.Println("cleared procpane Docker context (auto endpoint)")
				return
			}
			if len(args) == 0 {
				selected, err := store.DockerContext()
				if err != nil {
					exitErr(err)
				}
				if selected == "" {
					fmt.Println("(auto — DOCKER_HOST / local socket)")
				} else {
					fmt.Println(selected)
				}
				return
			}
			if err := store.SetDockerContext(args[0]); err != nil {
				exitErr(err)
			}
			fmt.Printf("set procpane Docker context %q\n", args[0])
		},
	}
	cmd.Flags().Bool("clear", false, "Clear selection (use DOCKER_HOST / local socket)")
	return cmd
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
