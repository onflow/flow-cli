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
	"context"
	"fmt"

	"github.com/onflow/flow-cli/internal/prompt"

	"github.com/onflow/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/arguments"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type deployContractFlags struct {
	ArgsJSON string   `default:"" flag:"args-json" info:"arguments in JSON-Cadence format"`
	Signer   string   `default:"emulator-account" flag:"signer" info:"Account name from configuration used to sign the transaction"`
	Include  []string `default:"" flag:"include" info:"Fields to include in the output. Valid values: contracts."`
	ShowDiff bool     `default:"false" flag:"show-diff" info:"Shows diff between existing and new contracts on update"`
}

var addContractFlags = deployContractFlags{}

var addContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "add-contract <filename> <args>",
		Short:   "Deploy a new contract to an account",
		Example: `flow accounts add-contract ./FungibleToken.cdc helloArg`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &addContractFlags,
	RunS:  deployContract(false, &addContractFlags),
}

func deployContract(update bool, flags *deployContractFlags) command.RunWithState {
	return func(
		args []string,
		globalFlags command.GlobalFlags,
		logger output.Logger,
		flow flowkit.Services,
		state *flowkit.State,
	) (command.Result, error) {
		filename := args[0]

		code, err := state.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("error loading contract file: %w", err)
		}

		to, err := state.Accounts().ByName(flags.Signer)
		if err != nil {
			return nil, err
		}

		var contractArgs []cadence.Value
		if flags.ArgsJSON != "" {
			contractArgs, err = arguments.ParseJSON(flags.ArgsJSON)
		} else if len(args) > 1 {
			contractArgs, err = arguments.ParseWithoutType(args[1:], code, filename)
		}

		if err != nil {
			return nil, fmt.Errorf("error parsing transaction arguments: %w", err)
		}

		deployFunc := flowkit.UpdateExistingContract(update)
		if updateContractFlags.ShowDiff {
			deployFunc = prompt.ShowContractDiffPrompt(logger)
		}

		txID, _, err := flow.AddContract(
			context.Background(),
			to,
			flowkit.Script{
				Code:     code,
				Args:     contractArgs,
				Location: filename,
			},
			deployFunc,
		)

		if err != nil {
			if txID != flowsdk.EmptyID {
				logger.Info(fmt.Sprintf(
					"Failed to %s contract on the account '%s' with transaction ID: %s",
					map[bool]string{true: "updated", false: "created"}[update],
					to.Address,
					txID.String(),
				))
			}

			return nil, err
		}

		err = state.SaveDefault()

		if err != nil {
			return nil, err
		}

		logger.Info(fmt.Sprintf(
			"Contract %s on the account '%s' with transaction ID %s.",
			map[bool]string{true: "updated", false: "created"}[update],
			to.Address,
			txID.String(),
		))

		account, err := flow.GetAccount(context.Background(), to.Address)
		if err != nil {
			return nil, err
		}

		return &accountResult{
			Account: account,
			include: flags.Include,
		}, nil
	}
}
