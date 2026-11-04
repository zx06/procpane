package cmdroot

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zx06/procpane/internal/app"
	"github.com/zx06/procpane/internal/format"
)

func newFindCmd() *cobra.Command {
	var asJSON bool
	var proto string
	svc := app.NewService()

	cmd := &cobra.Command{
		Use:   "find <port|:8080|host:port>",
		Short: "Find listeners on a port",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			matched, err := svc.Find(ctx, args[0], proto)
			if err != nil {
				exitErr(err)
			}
			if asJSON {
				if err := format.WriteListenersJSON(os.Stdout, matched); err != nil {
					exitErr(err)
				}
				return
			}
			if len(matched) == 0 {
				fmt.Println("no listeners on port")
				return
			}
			if err := format.WriteListenersTable(os.Stdout, matched); err != nil {
				exitErr(err)
			}
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit JSON")
	cmd.Flags().StringVar(&proto, "proto", "all", "Filter protocol: tcp, udp, or all")
	return cmd
}
