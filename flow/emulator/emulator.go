package emulator

import (
	"github.com/dapperlabs/flow-emulator/cmd/emulator/start"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:              "emulator",
	Short:            "Flow emulator server",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(start.Cmd)
}
