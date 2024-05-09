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
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/manifoldco/promptui"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/project"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

type stagingResults struct {
	Results       map[common.AddressLocation]stagingResult
	prettyPrinter func(err error, location common.Location) string
}

var _ command.ResultWithExitCode = &stagingResults{}

var stageContractflags struct {
	All            bool     `default:"false" flag:"all" info:"Stage all contracts"`
	Accounts       []string `default:"" flag:"account" info:"Accounts to stage the contract under"`
	SkipValidation bool     `default:"false" flag:"skip-validation" info:"Do not validate the contract code against staged dependencies"`
}

var stageContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "stage-contract [contract names...]",
		Short: "Stage a contract, or many contracts, for migration",
		Example: `flow migrate stage-contract Foo Bar --network testnet
flow migrate stage-contract --account my-account --network testnet
flow migrate stage-contract --all --network testnet`,
		Args: cobra.ArbitraryArgs,
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

	// Validate command arguments
	optionCount := boolCount(stageContractflags.All, len(stageContractflags.Accounts) > 0, len(args) > 0)
	if optionCount > 1 {
		return nil, fmt.Errorf("only one of --all, --account, or contract names can be provided")
	} else if optionCount == 0 {
		return nil, fmt.Errorf("at least one of --all, --account, or contract names must be provided")
	}

	// Stage based on flags
	var v stagingValidator
	if !stageContractflags.SkipValidation {
		v = newStagingValidator(flow)
	}
	s := newStagingService(flow, state, logger, v, promptStagingUnvalidatedContracts(logger))

	if stageContractflags.All {
		return stageAll(s, state, flow)
	}

	if len(stageContractflags.Accounts) > 0 {
		return stageByAccountNames(s, state, flow, stageContractflags.Accounts)
	}

	return stageByContractNames(s, state, flow, args)
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

func stageAll(
	s stagingService,
	state *flowkit.State,
	flow flowkit.Services,
) (*stagingResults, error) {
	contracts, err := state.DeploymentContractsByNetwork(flow.Network())
	if err != nil {
		return nil, err
	}

	results, err := s.StageContracts(context.Background(), contracts)
	if err != nil {
		return nil, err
	}

	return &stagingResults{Results: results, prettyPrinter: s.PrettyPrintValidationError}, nil
}

func stageByContractNames(
	s stagingService,
	state *flowkit.State,
	flow flowkit.Services,
	contractNames []string,
) (*stagingResults, error) {
	contracts, err := state.DeploymentContractsByNetwork(flow.Network())
	if err != nil {
		return nil, err
	}

	filteredContracts := make([]*project.Contract, 0)
	for _, name := range contractNames {
		found := false
		for _, contract := range contracts {
			if contract.Name == name {
				filteredContracts = append(filteredContracts, contract)
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("deployment not found for contract %s on network %s", name, flow.Network().Name)
		}
	}

	results, err := s.StageContracts(context.Background(), filteredContracts)
	if err != nil {
		return nil, err
	}

	return &stagingResults{Results: results, prettyPrinter: s.PrettyPrintValidationError}, nil
}

func stageByAccountNames(
	s stagingService,
	state *flowkit.State,
	flow flowkit.Services,
	accountNames []string,
) (*stagingResults, error) {
	contracts, err := state.DeploymentContractsByNetwork(flow.Network())
	if err != nil {
		return nil, err
	}

	filteredContracts := make([]*project.Contract, 0)
	for _, accountName := range accountNames {
		account, err := state.Accounts().ByName(accountName)
		if err != nil {
			return nil, err
		}

		found := false
		for _, contract := range contracts {
			if contract.AccountName == account.Name {
				filteredContracts = append(filteredContracts, contract)
				found = true
			}
		}

		if !found {
			return nil, fmt.Errorf("no deployments found for account %s on network %s", account.Name, flow.Network().Name)
		}
	}

	results, err := s.StageContracts(context.Background(), filteredContracts)
	if err != nil {
		return nil, err
	}

	return &stagingResults{Results: results, prettyPrinter: s.PrettyPrintValidationError}, nil
}

func (r *stagingResults) ExitCode() int {
	for _, r := range r.Results {
		if r.err != nil {
			return 1
		}
	}
	return 0
}

func (r *stagingResults) String() string {
	var sb strings.Builder

	// First print out any errors that occurred during staging
	for _, result := range r.Results {
		if result.err != nil {
			sb.WriteString(r.prettyPrinter(result.err, nil))
			sb.WriteString("\n")
		}
	}

	numStaged := 0
	numUnvalidated := 0
	numFailed := 0

	for location, result := range r.Results {
		if result.err == nil {
			if result.wasValidated {
				sb.WriteString(aurora.Green(fmt.Sprintf("✔ %s ", location.String())).String())
				sb.WriteString("(staged & validated)\n")
				numStaged++
			} else {
				sb.WriteString(aurora.Yellow(fmt.Sprintf("⚠ %s ", location)).String())
				sb.WriteString("(staged, but not validated)\n")
				numUnvalidated++
			}
		} else {
			sb.WriteString(aurora.Red(fmt.Sprintf("✘ %s ", location)).String())
			sb.WriteString("(failed to stage)\n")
			numFailed++
		}
	}

	sb.WriteString("\n")

	if numStaged > 0 {
		sb.WriteString(aurora.Green(fmt.Sprintf("%d %s staged & passed local validation\n", numStaged, pluralize("contract", numStaged))).String())
	}
	if numUnvalidated > 0 {
		sb.WriteString(aurora.Yellow(fmt.Sprintf("%d %s staged, but was not validated\n", numUnvalidated, pluralize("contract", numStaged))).String())
	}
	if numFailed > 0 {
		sb.WriteString(aurora.Red(fmt.Sprintf("%d %s failed to stage\n", numFailed, pluralize("contract", numStaged))).String())
	}

	sb.WriteString("\n")
	sb.WriteString("DISCLAIMER: Pre-staging validation checks are not exhaustive and do not guarantee the contract will work as expected, please monitor the status of your contract using the `flow migrate is-validated` command\n\n")
	sb.WriteString("You may use the --skip-validation flag to disable these checks and stage all contracts regardless")

	return sb.String()
}

func (s *stagingResults) JSON() interface{} {
	return s
}

func (r *stagingResults) Oneliner() string {
	if len(r.Results) == 0 {
		return "no contracts staged"
	}
	return fmt.Sprintf("staged %d contracts", len(r.Results))
}

// helpers
func boolCount(flags ...bool) int {
	count := 0
	for _, flag := range flags {
		if flag {
			count++
		}
	}
	return count
}

func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}
