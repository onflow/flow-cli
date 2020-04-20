package emulator

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "emulator",
	Short:            "Flow emulator server",
	TraverseChildren: true,
}

func init() {
	// Cmd.AddCommand(start.Cmd)
}
