package transactions

import (
	"github.com/spf13/cobra"

	"github.com/dapperlabs/flow-cli/flow/transactions/send"
	"github.com/dapperlabs/flow-cli/flow/transactions/status"
)

var Cmd = &cobra.Command{
	Use:              "transactions",
	Short:            "Utilities to send transactions",
	TraverseChildren: true,
}

func init() {
	Cmd.AddCommand(send.Cmd)
	Cmd.AddCommand(status.Cmd)
}
