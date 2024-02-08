package migration

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/scripts"
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
	scTempl, err := template.ParseFiles("./transactions/get_staged_contract_code.cdc")
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

	res, err := scripts.SendScript(
		txScriptBuf.Bytes(),
		[]string{
			contractName,
			contractAddress,
		},
		"",
		flow,
		isStagedflags,
	)
	if err != nil {
		return nil, fmt.Errorf("error sending script: %w", err)
	}

	return res, nil
}