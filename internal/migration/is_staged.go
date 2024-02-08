package migration

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"
	"github.com/spf13/cobra"
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
	RunS:   isStaged,
}


func isStaged(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	scTempl, err := template.ParseFiles("./scripts/get_staged_contract_code.cdc")
	if err != nil {
		return nil, fmt.Errorf("error loading staging contract file: %w", err)
	}

	// render transaction template with network
	var txScriptBuf bytes.Buffer
	if err := scTempl.Execute(
		&txScriptBuf, 
		map[string]string{
		"MigrationContractStaging": MigrationContractStagingAddress[globalFlags.Network],
		}); err != nil {
		return nil, fmt.Errorf("error rendering staging contract template: %w", err)
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
			Code:     txScriptBuf.Bytes(),
			Args:     []cadence.Value{caddr, cname},
		},
		query,
	)
	if err != nil {
		return nil, err
	}

	// check if valid is returned. If no value is returned, return false 
	if value.String() == "" {
		return nil, nil
	}

	return nil, nil
}