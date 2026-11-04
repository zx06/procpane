package cmdroot

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/zx06/procpane/internal/app"
	"github.com/zx06/procpane/internal/format"
)

func newListCmd() *cobra.Command {
	var proto string
	var asJSON bool
	var watch time.Duration
	svc := app.NewService()

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List LISTEN ports (tcp/udp)",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			run := func() error {
				listeners, err := svc.List(ctx, proto)
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
				fmt.Fprintln(os.Stderr, "---")
				time.Sleep(watch)
			}
		},
	}
	cmd.Flags().StringVar(&proto, "proto", "all", "Filter protocol: tcp, udp, or all")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit JSON")
	cmd.Flags().DurationVarP(&watch, "watch", "w", 0, "Refresh interval (e.g. 2s)")
	return cmd
}
