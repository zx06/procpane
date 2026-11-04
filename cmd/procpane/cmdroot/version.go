package cmdroot

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

const version = "0.1.0-v1"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("procpane %s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		},
	}
}
