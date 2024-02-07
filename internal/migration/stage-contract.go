package migration

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/transactions"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"
	"github.com/spf13/cobra"
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
	// Run:   stageContract,
}

var MigrationContractStagingAddress = map[string]string{
	"testnet":  "0xSomeAddress",
	"mainnet":  "0xSomeOtherAddress",
}


func stageContract(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (interface{}, error) {
	scTempl, err := template.ParseFiles("./transactions/stage-contract.cdc")
	if err != nil {
		return nil, fmt.Errorf("error loading stag contract file: %w", err)
	}

	// render transaction template with network
	var txScriptBuf bytes.Buffer
	if err := scTempl.Execute(
		&txScriptBuf, 
		map[string]string{
		"MigrationContractStaging": MigrationContractStagingAddress["testnet"],
		}); err != nil {
		return nil, err
	}

	// get the contract code from argument
	contractCode, err := state.ReadFile(args[1])
	if err != nil {
		return nil, fmt.Errorf("error loading contract file: %w", err)
	}

	res, err := transactions.SendTransaction(
		txScriptBuf.Bytes(), 
		[]string{
			args[0],
			string(contractCode),
		}, 
		"", 
		flow, 
		state, 
		stageContractflags,
	)
	if err != nil {
		return nil, err
	}

	return res, nil
}
