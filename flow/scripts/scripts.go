package scripts

import (
	"github.com/spf13/cobra"

	"github.com/dapperlabs/flow-cli/flow/scripts/execute"
)

var Cmd = &cobra.Command{
	Use:              "scripts",
	Short:            "Utilities to execute scripts",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(execute.Cmd)
}
