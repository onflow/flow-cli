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

package migrate

import (
	"context"
	"fmt"

	"github.com/onflow/cadence"
	"github.com/onflow/contract-updater/lib/go/templates"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/transactions"
	"github.com/spf13/cobra"

	internaltx "github.com/onflow/flow-cli/internal/transactions"

	"github.com/onflow/flow-cli/internal/command"
)

var stageContractflags struct{}

var stageContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "stage-contract <CONTRACT_NAME>",
		Short:   "stage a contract for migration",
		Example: `flow migrate stage-contract HelloWorld`,
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
	contractName := args[0]

	contract, err := state.Contracts().ByName(contractName)
	if err != nil {
		return nil, fmt.Errorf("no contracts found in state")
	}

	replacedCode, err := replaceImportsIfExists(state, flow, contract.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to replace imports: %w", err)
	}

	cName, err := cadence.NewString(contractName)
	if err != nil {
		return nil, fmt.Errorf("failed to get cadence string from contract name: %w", err)
	}

	cCode, err := cadence.NewString(string(replacedCode))
	if err != nil {
		return nil, fmt.Errorf("failed to get cadence string from contract code: %w", err)
	}

	account, err := getAccountByContractName(state, contractName, flow.Network())
	if err != nil {
		return nil, fmt.Errorf("failed to get account by contract name: %w", err)
	}

	tx, res, err := flow.SendTransaction(
		context.Background(),
		transactions.SingleAccountRole(*account),
		flowkit.Script{
			Code: templates.GenerateStageContractScript(MigrationContractStagingAddress(flow.Network().Name)),
			Args: []cadence.Value{cName, cCode},
		},
		flowsdk.DefaultTransactionGasLimit,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	return internaltx.NewTransactionResult(tx, res), nil
}
