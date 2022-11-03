/*
 * Flow CLI
 *
 * Copyright 2019 Dapper Labs, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package accounts

import (
	"fmt"

	"github.com/onflow/flow-cli/pkg/flowkit/output"

	"github.com/onflow/cadence"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsUpdateContract struct {
	ArgsJSON string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Signer   string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	Include  []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: contracts."`
}

var updateContractFlags = flagsUpdateContract{}

var UpdateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "update-contract <filename>",
		Short:   "Update a contract deployed to an account",
		Example: `flow accounts update-contract ./FungibleToken.cdc`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &updateContractFlags,
	RunS:  updateContract,
}

func updateContract(
	args []string,
	readerWriter flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	srv *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	filename := ""

	if len(args) == 1 {
		filename = args[0]
	} else {
		output.NewStdoutLogger(output.InfoLog).Info("⚠️Deprecation notice: using name argument in update contract " +
			"command will be deprecated soon.")
		filename = args[1]
	}

	code, err := readerWriter.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading contract file: %w", err)
	}

	to, err := state.Accounts().ByName(updateContractFlags.Signer)
	if err != nil {
		return nil, err
	}

	var contractArgs []cadence.Value
	if updateContractFlags.ArgsJSON != "" {
		contractArgs, err = flowkit.ParseArguments(nil, updateContractFlags.ArgsJSON)
	} else if len(args) > 2 {
		contractArgs, err = flowkit.ParseArgumentsWithoutType(filename, code, args[2:])
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing transaction arguments: %w", err)
	}

	account, err := srv.Accounts.AddContract(
		to,
		&services.Contract{
			Source:   code,
			Args:     contractArgs,
			Filename: filename,
			Network:  globalFlags.Network,
		},
		true,
	)
	if err != nil {
		return nil, err
	}

	return &AccountResult{
		Account: account,
		include: updateContractFlags.Include,
	}, nil
}
