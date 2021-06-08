/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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

	"github.com/onflow/flow-cli/pkg/flowcli/project"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
)

type flagsRemoveContract struct {
	Signer  string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	Results bool     `default:"false" flag:"results" info:"⚠️  Deprecated: results are provided by default"`
	Include []string `default:"" flag:"include" info:"Fields to include in the output"`
}

var flagsRemove = flagsRemoveContract{}

var RemoveCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "remove-contract <name>",
		Short:   "Remove a contract deployed to an account",
		Example: `flow accounts remove-contract FungibleToken`,
		Args:    cobra.ExactArgs(1),
	},
	Flags: &flagsRemove,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
		project *project.Project,
	) (command.Result, error) {
		if flagsRemove.Results {
			fmt.Println("⚠️ DEPRECATION WARNING: results flag is deprecated, results are by default included in all executions")
		}

		contractName := args[0]
		from := nil // todo refactor project

		account, err := services.Accounts.RemoveContract(contractName, from)
		if err != nil {
			return nil, err
		}

		return &AccountResult{
			Account:  account,
			showCode: false,
			include:  flagsRemove.Include,
		}, nil
	},
}
