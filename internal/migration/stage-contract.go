package migration

import (
	"context"
	"fmt"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/arguments"
	"github.com/onflow/flowkit/output"
	"github.com/spf13/cobra"

	flowsdk "github.com/onflow/flow-go-sdk"
)

type StageContractFlags struct {
	Network string `default:"" flag:"network" info:"network to stage the contract on"`
	Signer string `default:"" flag:"signer" info:"signer account to stage the contract with"`
}

var flags = StageContractFlags{}

var stageContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flow stage-contract <NAME> <CONTRACT_PATH> --network <NETWORK> --signer <HOST_ACCOUNT>",
		Short:   "stage a contract for migration",
		Example: `flow stage-contract HelloWorld hello_world.cdc --network testnet --signer emulator-account`,
		Args:    cobra.MinimumNArgs(2),
	},
	Flags: &flags,
	Run:   stageContract,
}

func stageContract(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	readerWriter flowkit.ReaderWriter,
	flow flowkit.Services,
) (command.Result, error) {
	stageContractCode, err := readerWriter.ReadFile("./transactions/stage-contract.cdc")
	if err != nil {
		return nil, fmt.Errorf("error loading stag contract file: %w", err)
	}

	// TODO: do something about sending transactions instead
	return SendScript(stageContractCode, args[1:], "", flow, flags)
}

func SendScript(code []byte, argsArr []string, location string, flow flowkit.Services, scriptFlags Flags) (command.Result, error) {
	var cadenceArgs []cadence.Value
	var err error
	if scriptFlags.ArgsJSON != "" {
		cadenceArgs, err = arguments.ParseJSON(scriptFlags.ArgsJSON)
	} else {
		cadenceArgs, err = arguments.ParseWithoutType(argsArr, code, location)
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing script arguments: %w", err)
	}

	query := flowkit.ScriptQuery{}
	if scriptFlags.BlockHeight != 0 {
		query.Height = scriptFlags.BlockHeight
	} else if scriptFlags.BlockID != "" {
		query.ID = flowsdk.HexToID(scriptFlags.BlockID)
	} else {
		query.Latest = true
	}

	value, err := flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code:     code,
			Args:     cadenceArgs,
			Location: location,
		},
		query,
	)
	if err != nil {
		return nil, err
	}

	return &scriptResult{value}, nil
}
