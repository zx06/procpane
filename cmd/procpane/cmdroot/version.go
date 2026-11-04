package cmdroot

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/zx06/procpane/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("procpane %s (%s/%s)\n", version.Version, runtime.GOOS, runtime.GOARCH)
		},
	}
}
