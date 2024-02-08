package migration

import (
	"context"
	"fmt"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
)

var listStagedContractsflags = scripts.Flags{}

var listStagedContractsCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flow list-staged <CONTRACT_ADDRESS>",
		Short:   "returns back the a list of staged contracts given a contract address",
		Example: `flow list-staged 0xhello`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &listStagedContractsflags,
	RunS:  listStagedContracts,
}

func listStagedContracts(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	code, err := RenderContractTemplate(GetStagedCodeForAddressScriptFilepath, globalFlags.Network)
	if err != nil {
		return nil, fmt.Errorf("error loading staging contract file: %w", err)
	}

	contractAddress := args[0]

	caddr := cadence.NewAddress(flowsdk.HexToAddress(contractAddress))

	query := flowkit.ScriptQuery{}
	if listStagedContractsflags.BlockHeight != 0 {
		query.Height = listStagedContractsflags.BlockHeight
	} else if listStagedContractsflags.BlockID != "" {
		query.ID = flowsdk.HexToID(listStagedContractsflags.BlockID)
	} else {
		query.Latest = true
	}

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: code,
			Args: []cadence.Value{caddr},
		},
		query,
	)
	if err != nil {
		return nil, err
	}

	return scripts.NewScriptResult(value), nil
}