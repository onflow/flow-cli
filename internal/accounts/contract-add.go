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

	"github.com/onflow/cadence"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsAddContract struct {
	ArgsJSON string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Signer   string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	Include  []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: contracts."`
}

var addContractFlags = flagsAddContract{}

var AddContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "add-contract <filename>",
		Short:   "Deploy a new contract to an account",
		Example: `flow accounts add-contract ./FungibleToken.cdc`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &addContractFlags,
	RunS:  addContract,
}

func addContract(
	args []string,
	readerWriter flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	srv *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	filename := args[0]
	if len(args) > 1 {
		fmt.Println("⚠️Deprecation notice: using name argument in add contract command will be deprecated soon.")
		filename = args[1]
	}

	code, err := readerWriter.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading contract file: %w", err)
	}

	to, err := state.Accounts().ByName(addContractFlags.Signer)
	if err != nil {
		return nil, err
	}

	var contractArgs []cadence.Value
	if addContractFlags.ArgsJSON != "" {
		contractArgs, err = flowkit.ParseArgumentsJSON(addContractFlags.ArgsJSON)
	} else if len(args) > 2 {
		contractArgs, err = flowkit.ParseArgumentsWithoutType(filename, code, args[2:])
	}

	if err != nil {
		return nil, fmt.Errorf("error parsing transaction arguments: %w", err)
	}

	_, _, err = srv.Accounts.AddContract(
		to,
		&services.Contract{
			Script: &services.Script{
				Code:     code,
				Args:     contractArgs,
				Filename: filename,
			},
			Network: globalFlags.Network,
		},
		false,
	)
	if err != nil {
		return nil, err
	}

	account, err := srv.Accounts.Get(to.Address())
	if err != nil {
		return nil, err
	}

	return &AccountResult{
		Account: account,
		include: addContractFlags.Include,
	}, nil
}
