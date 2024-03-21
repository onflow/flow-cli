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
	"errors"
	"fmt"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/contract-updater/lib/go/templates"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/transactions"
	"github.com/spf13/cobra"

	internaltx "github.com/onflow/flow-cli/internal/transactions"

	"github.com/onflow/flow-cli/internal/command"
)

var stageContractflags struct {
	Force bool `default:"false" flag:"force" info:"Force staging the contract without validation"`
}

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
	logger output.Logger,
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

	// Validate the contract code by default
	if !stageContractflags.Force {
		validator := newStagingValidator(flow, state)

		var missingDependenciesErr *contractsNotStagedError
		contractLocation := common.NewAddressLocation(nil, common.Address(account.Address), contractName)
		err = validator.ValidateContractUpdate(context.Background(), contractLocation, replacedCode)

		// Errors when the contract's dependencies have not been staged yet are non-fatal
		// This is because the contract may be dependent on contracts that are not yet staged
		// and we do not want to require in-order staging of contracts
		// Any other errors are fatal, however, can be bypassed with the --force flag
		if errors.As(err, &missingDependenciesErr) {
			logger.Info(fmt.Sprintf("The staged contract cannot be validated: %s", err))
			logger.Info("Staging will continue without validation, please monitor the status of the contract to ensure it is staged correctly")
		} else if err != nil {
			return nil, fmt.Errorf("the contract code does not appear to be valid, you can use the --force flag to bypass this check: %w", err)
		}
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
