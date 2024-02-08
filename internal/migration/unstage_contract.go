package migration

import (
	"fmt"

	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/transactions"
)

var unstageContractflags = transactions.Flags{}

var unstageContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flow unstage-contract <NAME> --network <NETWORK> --signer <HOST_ACCOUNT>",
		Short:   "unstage a contract for migration",
		Example: `flow unstage-contract HelloWorld`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &unstageContractflags,
	RunS:  unstageContract,
}

func unstageContract(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	code, err := RenderContractTemplate(UnstageContractTransactionFilepath, globalFlags.Network)
	if err != nil {
		return nil, fmt.Errorf("error loading staging contract file: %w", err)
	}

	contractName := args[0]

	res, err := transactions.SendTransaction(
		code,
		[]string{
			contractName,
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
