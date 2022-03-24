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

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsUpdateContract struct {
	Signer  string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	Include []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: contracts."`
}

var updateFlags = flagsUpdateContract{}

var UpdateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "update-contract <name> <filename>",
		Short:   "Update a contract deployed to an account",
		Example: `flow accounts update-contract FungibleToken ./FungibleToken.cdc`,
		Args:    cobra.ExactArgs(2),
	},
	Flags: &updateFlags,
	RunS:  updateContract,
}

func updateContract(
	args []string,
	readerWriter flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	name := args[0]
	filename := args[1]

	code, err := readerWriter.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error loading contract file: %w", err)
	}

	to, err := state.Accounts().ByName(updateFlags.Signer)
	if err != nil {
		return nil, err
	}

	account, err := services.Accounts.AddContract(to, name, code, true)
	if err != nil {
		return nil, err
	}

	return &AccountResult{
		Account: account,
		include: updateFlags.Include,
	}, nil
}
