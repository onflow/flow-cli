package evm

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/transactions"
)

//go:embed create.cdc
var createCode []byte

type flagsCreate struct {
	Signer string `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
}

var createFlags = flagsCreate{}

var createCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "create-account <amount>",
		Short:   "Create a new EVM account and fund it with the amount as well as store the bridged account resource",
		Args:    cobra.ExactArgs(1),
		Example: "flow evm create-account 1.0",
	},
	Flags: &createFlags,
	RunS:  create,
}

func create(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	amount := args[0]
	result, err := transactions.SendTransaction(
		createCode,
		[]string{amount},
		"",
		flow,
		state,
		transactions.Flags{
			Signer: deployFlags.Signer,
		},
	)
	if err != nil {
		return nil, err
	}

	printCreateResult(result)

	return nil, nil
}

func printCreateResult(result command.Result) {
	fmt.Printf("\n🔥🔥🔥🔥🔥🔥🔥 EVM Account Creation Summary 🔥🔥🔥🔥🔥🔥🔥")
	fmt.Printf("\n-------------------------------------------------------------\n\n")

	fmt.Println(result)
	//txResult := result.(*transactions.TransactionResult)
	//events := flowkit.EventsFromTransaction(txResult.Result)

}
