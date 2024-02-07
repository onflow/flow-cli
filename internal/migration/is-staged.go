package migration

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"
	"github.com/spf13/cobra"
)

type IsStagedFlags struct {
	Network string `default:"" flag:"network" info:"network to stage the contract on"`
	Signer string `default:"" flag:"signer" info:"signer account to stage the contract with"`
}

var isStagedflags = IsStagedFlags{}

var IsStagedCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flow is-staged <CONTRACT_ADDRESS> <CONTRACT_NAME> --network",
		Short:   "checks to see if the contract is staged for migration",
		Example: `flow is-staged HelloWorld hello_world.cdc --network testnet`,
		Args:    cobra.MinimumNArgs(2),
	},
	Flags: &isStagedflags,
	Run:   IsStaged,
}

func IsStaged(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	readerWriter flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	IsStagedCode, err := readerWriter.ReadFile("./scripts/get_staged_contract_code.cdc")
	if err != nil {
		return nil, fmt.Errorf("error loading stag contract file: %w", err)
	}

	return scripts.SendScript(IsStagedCode, args, "", flow, scripts.Flags{})
}