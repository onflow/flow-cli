package migration

import (
	"fmt"

	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/transactions"
)

var stageContractflags = transactions.Flags{}

var stageContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flow stage-contract <NAME> <CONTRACT_PATH> --network <NETWORK> --signer <HOST_ACCOUNT>",
		Short:   "stage a contract for migration",
		Example: `flow stage-contract HelloWorld hello_world.cdc --network testnet --signer emulator-account`,
		Args:    cobra.MinimumNArgs(2),
	},
	Flags: &stageContractflags,
	RunS:  stageContract,
}

func stageContract(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	code, err := RenderContractTemplate(StageContractTransactionFilepath, globalFlags.Network)
	if err != nil {
		return nil, fmt.Errorf("error loading staging contract file: %w", err)
	}

	contractName, contractPath := args[0], args[1]

	// get the contract code from argument
	contractCode, err := state.ReadFile(contractPath)
	if err != nil {
		return nil, fmt.Errorf("error loading contract file: %w", err)
	}

	res, err := transactions.SendTransaction(
		code,
		[]string{
			contractName,
			string(contractCode),
		},
		"",
		flow,
		state,
		stageContractflags,
	)
	if err != nil {
		return nil, fmt.Errorf("error sending transaction: %w", err)
	}

	return res, nil
}
