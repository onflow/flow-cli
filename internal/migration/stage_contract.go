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

package migration

import (
	"fmt"

	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/output"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/transactions"
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
	RunS:  stageContract,
}

func stageContract(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	code, err := RenderContractTemplate(StageContractTransactionFilepath, globalFlags.Network)
	if err != nil {
		return nil, fmt.Errorf("error loading staging contract file: %w", err)
	}

	contractName, contractPath := args[0], args[1]

	// get the contract code from argument
	contractCode, err := state.ReadFile(contractPath)
	if err != nil {
		return nil, fmt.Errorf("error loading contract file: %w", err)
	}

	res, err := transactions.SendTransaction(
		code,
		[]string{
			contractName,
			string(contractCode),
		},
		"",
		flow,
		state,
		stageContractflags,
	)
	if err != nil {
		return nil, fmt.Errorf("error sending transaction: %w", err)
	}

	return res, nil
}
