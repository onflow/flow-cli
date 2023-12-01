package evm

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/arguments"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
)

//go:embed get.cdc
var getCode []byte

type flagsGet struct{}

var getFlags = flagsGet{}

var getCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "get-account <evm address>",
		Short:   "Get account by the EVM address",
		Args:    cobra.ExactArgs(1),
		Example: "flow evm get-account 522b3294e6d06aa25ad0f1b8891242e335d3b459",
	},
	Flags: &getFlags,
	RunS:  get,
}

func get(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {

	scriptArgs, err := arguments.ParseWithoutType([]string{args[0]}, getCode, "")
	if err != nil {
		return nil, err
	}

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: getCode,
			Args: scriptArgs,
		},
		flowkit.ScriptQuery{Latest: true},
	)
	if err != nil {
		return nil, err
	}

	fmt.Println(value)

	return nil, nil
}
