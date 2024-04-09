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
	"strings"

	"github.com/manifoldco/promptui"
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
	SkipValidation bool `default:"false" flag:"skip-validation" info:"Do not validate the contract code against staged dependencies"`
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
	err := checkNetwork(flow.Network())
	if err != nil {
		return nil, err
	}

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
	if !stageContractflags.SkipValidation {
		logger.StartProgress("Validating contract code against any staged dependencies")
		validator := newStagingValidator(flow, state)

		var missingDependenciesErr *missingDependenciesError
		contractLocation := common.NewAddressLocation(nil, common.Address(account.Address), contractName)
		err = validator.ValidateContractUpdate(contractLocation, common.StringLocation(contract.Location), replacedCode)

		logger.StopProgress()

		// Errors when the contract's dependencies have not been staged yet are non-fatal
		// This is because the contract may be dependent on contracts that are not yet staged
		// and we do not want to require in-order staging of contracts
		// Instead, we will prompt the user to continue staging the contract.  Other errors
		// will be fatal and require manual intervention using the --skip-validation flag if desired
		if errors.As(err, &missingDependenciesErr) {
			infoMessage := strings.Builder{}
			infoMessage.WriteString("Validation cannot be performed as some of your contract's dependencies could not be found (have they been staged yet?)\n")
			for _, contract := range missingDependenciesErr.MissingContracts {
				infoMessage.WriteString(fmt.Sprintf("  - %s\n", contract))
			}
			infoMessage.WriteString("\nYou may still stage your contract, however it will be unable to be migrated until the missing contracts are staged by their respective owners.  It is important to monitor the status of your contract using the `flow migrate is-validated` command\n")
			logger.Error(infoMessage.String())

			continuePrompt := promptui.Select{
				Label: "Do you wish to continue staging your contract?",
				Items: []string{"Yes", "No"},
			}

			_, result, err := continuePrompt.Run()
			if err != nil {
				return nil, err
			}

			if result == "No" {
				return nil, fmt.Errorf("staging cancelled")
			}
		} else if err != nil {
			logger.Error(validator.prettyPrintError(err, common.StringLocation(contract.Location)))
			return nil, fmt.Errorf("errors were found while attempting to perform preliminary validation of the contract code, and your contract HAS NOT been staged, however you can use the --skip-validation flag to bypass this check & stage the contract anyway")
		} else {
			logger.Info("No issues found while validating contract code\n")
			logger.Info("DISCLAIMER: Pre-staging validation checks are not exhaustive and do not guarantee the contract will work as expected, please monitor the status of your contract using the `flow migrate is-validated` command\n")
		}
	} else {
		logger.Info("Skipping contract code validation, you may monitor the status of your contract using the `flow migrate is-validated` command\n")
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
