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

/*
TODO: handle the broken dependency graph case.

e.g. Foo -> Bar -> Baz
but only Foo & Baz are staged, so how to build the contract graph?
*/

package migrate

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/contract-updater/lib/go/templates"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/project"
	"github.com/onflow/flowkit/v2/transactions"
	"github.com/spf13/cobra"

	internaltx "github.com/onflow/flow-cli/internal/transactions"

	"github.com/onflow/flow-cli/internal/command"
)

type stagingResult struct {
	// Error will be nil if the contract was successfully staged
	Contracts map[common.AddressLocation]error
}

var _ command.ResultWithExitCode = &stagingResult{}

var stageContractflags struct {
	All            bool     `default:"false" flag:"all" info:"Stage all contracts"`
	Accounts       []string `default:"" flag:"accounts" info:"Accounts to stage the contract under"`
	SkipValidation bool     `default:"false" flag:"skip-validation" info:"Do not validate the contract code against staged dependencies"`
}

var stageContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "stage-contract <CONTRACT_NAME>",
		Short:   "stage a contract for migration",
		Example: `flow migrate stage-contract HelloWorld`,
		Args:    cobra.MinimumNArgs(1),
	},
	Flags: &stageContractflags,
	RunS:  stageContracts,
}

func buildContract(state *flowkit.State, flow flowkit.Services, contract *config.Contract) (*project.Contract, error) {
	contractName := contract.Name

	replacedCode, err := replaceImportsIfExists(state, flow, contract.Location)
	if err != nil {
		return nil, fmt.Errorf("failed to replace imports: %w", err)
	}

	account, err := getAccountByContractName(state, contractName, flow.Network())
	if err != nil {
		return nil, fmt.Errorf("failed to get account by contract name: %w", err)
	}

	return project.NewContract(contractName, filepath.Clean(contract.Location), replacedCode, account.Address, account.Name, nil), nil
}

func stageContracts(
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

	s := newStagingService(flow, state, logger, !stageContractflags.SkipValidation)

	// Stage contracts based on the provided flags
	var results map[common.AddressLocation]error
	if stageContractflags.All {
		results, err = s.StageAllContracts(context.Background())
		if err != nil {
			return nil, err
		}
	} else if len(args) > 0 {
		results, err = s.StageContractsByName(context.Background(), args)
		if err != nil {
			return nil, err
		}
	} else if len(stageContractflags.Accounts) > 0 {
		results, err = s.StageContractsByAccounts(context.Background(), stageContractflags.Accounts)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("no contracts specified, please provide contract names or use the --all or --accounts flags")
	}

	return &stagingResult{
		Contracts: results,
	}, nil
}

func stageContract() {
	// Validate the contract code by default
	if !stageContractflags.SkipValidation {

		// Errors when the contract's dependencies have not been staged yet are non-fatal
		// This is because the contract may be dependent on contracts that are not yet staged
		// and we do not want to require in-order staging of contracts
		// Instead, we will prompt the user to continue staging the contract.  Other errors
		// will be fatal and require manual intervention using the --skip-validation flag if desired
		if errors.As(err, &missingDependenciesErr) {

		} else if err != nil {
			logger.Error(validator.prettyPrintError(err, common.StringLocation(contract.Location)))
			return nil, fmt.Errorf("errors were found while attempting to perform preliminary validation of the contract code, and your contract HAS NOT been staged, however you can use the --skip-validation flag to bypass this check & stage the contract anyway")
		} else {
			logger.Info("No issues found while validating contract code\n")
		}
	} else {
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

func (r *stagingResult) ExitCode() int {
	for _, err := range r.Contracts {
		if err != nil {
			return 1
		}
	}
	return 0
}

func (s *stagingResult) String() string {
	if len(s.Contracts) == 0 {
		return "no contracts staged"
	}
}

func (s *stagingResult) JSON() interface{} {
	return s
}

func (r *stagingResult) Oneliner() string {
	if len(r.Contracts) == 0 {
		return "no contracts staged"
	}
	return fmt.Sprintf("staged %d contracts", len(r.Contracts))
}
