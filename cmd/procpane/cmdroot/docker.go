package cmdroot

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/zx06/procpane/internal/format"
	dockersrc "github.com/zx06/procpane/internal/source/docker"
)

func newDockerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docker",
		Short: "Docker published ports (Engine API)",
	}
	cmd.AddCommand(newDockerLsCmd())
	return cmd
}

func newDockerLsCmd() *cobra.Command {
	var asJSON bool
	var watch time.Duration
	src := dockersrc.New()

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

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"ps"},
		Short:   "List published container ports",
		Run: func(cmd *cobra.Command, args []string) {
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
