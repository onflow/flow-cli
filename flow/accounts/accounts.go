package accounts

import (
	"github.com/spf13/cobra"

	"github.com/dapperlabs/flow-cli/flow/accounts/create"
	"github.com/dapperlabs/flow-cli/flow/accounts/get"
)

var Cmd = &cobra.Command{
	Use:              "accounts",
	Short:            "Utilities to manage accounts",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(create.Cmd)
	Cmd.AddCommand(get.Cmd)
}
