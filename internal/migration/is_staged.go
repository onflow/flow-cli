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

var isStagedflags = scripts.Flags{}

var IsStagedCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flow is-staged <CONTRACT_NAME> <CONTRACT_ADDRESS>",
		Short:   "checks to see if the contract is staged for migration",
		Example: `flow is-staged HelloWorld 0xhello`,
		Args:    cobra.MinimumNArgs(2),
	},
	Flags: &isStagedflags,
	RunS:  isStaged,
}

func isStaged(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	code, err := RenderContractTemplate(IsStagedFileScriptpath, globalFlags.Network)
	if err != nil {
		return nil, fmt.Errorf("error loading staging contract file: %w", err)
	}

	contractName, contractAddress := args[0], args[1]

	caddr := cadence.NewAddress(flowsdk.HexToAddress(contractAddress))

	cname, err := cadence.NewString(contractName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cadence string from contract name: %w", err)
	}

	query := flowkit.ScriptQuery{}
	if isStagedflags.BlockHeight != 0 {
		query.Height = isStagedflags.BlockHeight
	} else if isStagedflags.BlockID != "" {
		query.ID = flowsdk.HexToID(isStagedflags.BlockID)
	} else {
		query.Latest = true
	}

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: code,
			Args: []cadence.Value{caddr, cname},
		},
		query,
	)
	if err != nil {
		return nil, err
	}

	return scripts.NewScriptResult(value), nil
}
