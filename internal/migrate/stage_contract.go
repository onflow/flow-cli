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
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/project"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

type stagingResult struct {
	// Error will be nil if the contract was successfully staged
	Contracts map[common.AddressLocation]error
}

var _ command.ResultWithExitCode = &stagingResult{}

var stageContractflags struct {
	All            bool     `default:"false" flag:"all" info:"Stage all contracts"`
	Accounts       []string `default:"" flag:"account" info:"Accounts to stage the contract under"`
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
	RunS:  stageContract,
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

	var v stagingValidator
	if !stageContractflags.SkipValidation {
		v = newStagingValidator(flow)
	}

	s := newStagingService(flow, state, logger, v, promptStagingUnvalidatedContracts(logger))
	return stageWithFilters(s, stageContractflags.All, args, stageContractflags.Accounts)
}

func promptStagingUnvalidatedContracts(logger output.Logger) func(validatorError *stagingValidatorError) bool {
	return func(validatorError *stagingValidatorError) bool {
		infoMessage := strings.Builder{}

		infoMessage.WriteString("Preliminary validation could not be performed on the following contracts:\n")
		missingDependencyErrors := validatorError.MissingDependencyErrors()
		for deployLocation := range missingDependencyErrors {
			infoMessage.WriteString(fmt.Sprintf("  - %s\n", deployLocation))
		}

		infoMessage.WriteString("\nThese contracts depend on the following contracts which have not been staged yet:\n")
		missingDependencies := validatorError.MissingDependencies()
		for _, depLocation := range missingDependencies {
			infoMessage.WriteString(fmt.Sprintf("  - %s\n", depLocation))
		}

		infoMessage.WriteString("\nYou may still stage your contract, however it will be unable to be migrated until the missing contracts are staged by their respective owners.  It is important to monitor the status of your contract using the `flow migrate is-validated` command\n")
		logger.Error(infoMessage.String())

		continuePrompt := promptui.Select{
			Label: "Do you wish to continue staging your contract?",
			Items: []string{"Yes", "No"},
		}

		_, result, err := continuePrompt.Run()
		if err != nil || result != "Yes" {
			return false
		}
		return true
	}
}

func stageWithFilters(
	s stagingService,
	allContracts bool,
	contractNames []string,
	accountNames []string,
) (*stagingResult, error) {
	var results map[common.AddressLocation]error
	var err error

	// Stage all contracts
	if allContracts {
		if len(contractNames) > 0 || len(accountNames) > 0 {
			return nil, fmt.Errorf("cannot use --all flag with contract names or --accounts flag")
		}

		results, err = s.StageAllContracts(context.Background())
	}

	// Filter by contract names
	if len(contractNames) > 0 {
		if len(accountNames) > 0 {
			return nil, fmt.Errorf("cannot use --account flag with contract names")
		}

		results, err = s.StageAllContracts(context.Background())
	}

	// Filter by accounts
	if len(accountNames) > 0 {
		results, err = s.StageAllContracts(context.Background())
	}

	if err != nil {
		return nil, err
	}

	// Print the results
	return &stagingResult{
		Contracts: results,
	}, nil
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

	sb := &strings.Builder{}

	// First, print the failing contracts
	for location, err := range s.Contracts {
		if err != nil {
			sb.WriteString(fmt.Sprintf("failed to stage contract %s: %s\n", location, err))
		}
	}

	// Then, print the successfully staged contracts
	for location, err := range s.Contracts {
		if err == nil {
			sb.WriteString(fmt.Sprintf("staged contract %s\n", location))
		}
	}

	sb.WriteString(fmt.Sprintf("staged %d contracts", len(s.Contracts)))

	return sb.String()
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
