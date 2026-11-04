package cmdroot

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "procpane",
	Short: "Local LISTEN port and process panel",
	Long:  "Cross-platform CLI for listing listening ports, inspecting processes, and killing with safety checks.",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newFindCmd())
	rootCmd.AddCommand(newShowCmd())
	rootCmd.AddCommand(newKillCmd())
	rootCmd.AddCommand(newVersionCmd())
}

func exitErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
