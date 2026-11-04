package cmdroot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/zx06/procpane/internal/app"
)

func newShowCmd() *cobra.Command {
	svc := app.NewService()

	cmd := &cobra.Command{
		Use:   "show <pid>",
		Short: "Show process details",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			pid64, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				exitErr(fmt.Errorf("invalid pid: %w", err))
			}
			info := svc.Show(context.Background(), int32(pid64))
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(info); err != nil {
				exitErr(err)
			}
		},
	}
	return cmd
}
