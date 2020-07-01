package blocks

import (
	"github.com/spf13/cobra"

	"github.com/dapperlabs/flow-cli/flow/blocks/get"
)

var Cmd = &cobra.Command{
	Use:              "blocks",
	Short:            "Utilities to read blocks",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(get.Cmd)
}
