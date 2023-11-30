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

//go:embed deploy.cdc
var deployCode string

type flagsDeploy struct {
	Signer string `default:"" flag:"signer" info:"Account name from configuration used to sign the transaction as proposer, payer and suthorizer"`
}

var deployFlags = flagsDeploy{}

var deployCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "deploy <evm bytecode file>",
		Short:   "Deploy compiled bytecode to the Flow EVM",
		Args:    cobra.ExactArgs(1),
		Example: "flow evm deploy ./hello",
	},
	Flags: &deployFlags,
	RunS:  deploy,
}

func deploy(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	filename := args[0]

	// read file containing hex-encoded evm bytecode
	evmCode, err := state.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading transaction file: %w", err)
	}

	result, err := transactions.SendTransaction(
		[]byte(deployCode),
		[]string{string(evmCode)},
		filename,
		flow,
		state,
		transactions.Flags{
			Signer: deployFlags.Signer,
		},
	)

	txResult := result.(*transactions.TransactionResult)

	fmt.Println("========== EVM ===========")

	fmt.Println("==========================")
	fmt.Println(result)

	return nil, nil
}
