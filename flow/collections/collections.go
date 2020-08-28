package collections

import (
	"github.com/spf13/cobra"

	"github.com/dapperlabs/flow-cli/flow/collections/get"
)

var Cmd = &cobra.Command{
	Use:              "collections",
	Short:            "Utilities to read collections",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(get.Cmd)
}
