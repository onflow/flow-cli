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
	"context"
	"fmt"

	"github.com/onflow/cadence"
	"github.com/onflow/contract-updater/lib/go/templates"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit"
	"github.com/onflow/flowkit/accounts"
	"github.com/onflow/flowkit/output"
	"github.com/onflow/flowkit/transactions"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

var stageContractflags interface{}

var stageContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "flow stage-contract <NAME> --network <NETWORK> --signer <HOST_ACCOUNT>",
		Short:   "stage a contract for migration",
		Example: `flow stage-contract HelloWorld --network testnet --signer emulator-account`,
		Args:    cobra.MinimumNArgs(1),
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
    code := templates.GenerateStageContractScript(flowsdk.HexToAddress(migrationContractStagingAddress[globalFlags.Network]))
	contractName := args[0]

	contract, err := state.Contracts().ByName(contractName)
	if err != nil {
		return nil, fmt.Errorf("no contracts found in state")
	}

	contractCode, err := state.ReadFile(contract.Location)
	if err != nil {
		return nil, fmt.Errorf("error loading contract file: %w", err)
	}

	account, err := getAccountByContractName(state, globalFlags.Network, contractName)
	if err != nil {
		return nil, fmt.Errorf("error getting account by contract name: %w", err)
	}

	cName, err := cadence.NewString(contractName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cadence string from contract name: %w", err)
	}

	cCode, err := cadence.NewString(string(contractCode))
	if err != nil {
		return nil, fmt.Errorf("failed to get cadence string from contract code: %w", err)
	}

	_, _, err = flow.SendTransaction(
		context.Background(),
		transactions.AccountRoles{
			Proposer:    *account,
			Authorizers: []accounts.Account{*account},
			Payer:       *account,
		},
		flowkit.Script{
			Code: code,
			Args: []cadence.Value{cName, cCode},
		},
		flowsdk.DefaultTransactionGasLimit,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return nil, nil
}
