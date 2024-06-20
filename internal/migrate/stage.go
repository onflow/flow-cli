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

	"github.com/onflow/flow-cli/internal/prompt"

	"github.com/logrusorgru/aurora/v4"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/project"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

const stagingLimit = 200

type stagingResults struct {
	Results       map[common.AddressLocation]stagingResult
	prettyPrinter func(err error, location common.Location) string
}

var _ command.ResultWithExitCode = &stagingResults{}

var stageProjectFlags struct {
	Accounts       []string `default:"" flag:"account" info:"Accounts to stage the contract under"`
	SkipValidation bool     `default:"false" flag:"skip-validation" info:"Do not validate the contract code against staged dependencies"`
}

var stageCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "stage [contract names...]",
		Short: "Stage a contract, or many contracts, for migration",
		Example: `# Stage all contracts
flow migrate stage --network testnet

# Stage by contract name(s)
flow migrate stage Foo Bar --network testnet

# Stage by account name(s)
flow migrate stage --account my-account --network testnet`,
		Args:    cobra.ArbitraryArgs,
		Aliases: []string{"stage"},
	},
	Flags: &stageProjectFlags,
	RunS:  stageProject,
}

func stageProject(
	args []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	err := util.CheckNetwork(flow.Network())
	if err != nil {
		return nil, err
	}

	// Validate command arguments
	if len(stageProjectFlags.Accounts) > 0 && len(args) > 0 {
		return nil, fmt.Errorf("only one of contract names or --account can be provided")
	}

	// Stage based on flags
	var v stagingValidator
	if !stageProjectFlags.SkipValidation {
		v = newStagingValidator(flow)
	}
	s := newStagingService(flow, state, logger, v, promptStagingUnvalidatedContracts(logger))

	var result command.Result
	if len(args) == 0 && len(stageProjectFlags.Accounts) == 0 {
		result, err = stageAll(s, state, flow)
	} else if len(stageProjectFlags.Accounts) > 0 {
		result, err = stageByAccountNames(s, state, flow, stageProjectFlags.Accounts)
	} else {
		result, err = stageByContractNames(s, state, flow, args)
	}

	if err != nil {
		var fatalValidationErr *fatalValidationError
		if errors.As(err, &fatalValidationErr) {
			return nil, fmt.Errorf("a fatal error occurred during validation, you may use the --skip-validation flag to disable these checks: %w", fatalValidationErr.err)
		}
		return nil, err
	}
	return result, nil
}

func promptStagingUnvalidatedContracts(logger output.Logger) func(validatorError *stagingValidatorError) bool {
	return func(validatorError *stagingValidatorError) bool {
		infoMessage := strings.Builder{}
		infoMessage.WriteString("Preliminary validation could not be performed on the following contracts:\n")

		// Sort the locations for consistent output
		missingDependencyErrors := validatorError.MissingDependencyErrors()
		sortedLocations := make([]common.AddressLocation, 0, len(missingDependencyErrors))
		for deployLocation := range missingDependencyErrors {
			sortedLocations = append(sortedLocations, deployLocation)
		}
		sortAddressLocations(sortedLocations)

		// Print the locations
		for _, deployLocation := range sortedLocations {
			infoMessage.WriteString(fmt.Sprintf("  - %s\n", deployLocation))
		}

		infoMessage.WriteString("\nThese contracts depend on the following contracts which have not been staged yet:\n")

		// Print the missing dependencies
		missingDependencies := validatorError.MissingDependencies()
		for _, depLocation := range missingDependencies {
			infoMessage.WriteString(fmt.Sprintf("  - %s\n", depLocation))
		}

		infoMessage.WriteString("\nYou may still stage your contract, however it will be unable to be migrated until the missing contracts are staged by their respective owners.  It is important to monitor the status of your contract using the `flow migrate is-validated` command\n")
		logger.Error(infoMessage.String())

		return prompt.GenericBoolPrompt("Do you wish to continue staging your contract?")
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

	if len(contracts) > stagingLimit {
		return nil, fmt.Errorf("cannot stage more than %d contracts at once", stagingLimit)
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

	if len(contracts) > stagingLimit {
		return nil, fmt.Errorf("cannot stage more than %d contracts at once", stagingLimit)
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

	if len(contracts) > stagingLimit {
		return nil, fmt.Errorf("cannot stage more than %d contracts at once", stagingLimit)
	}

	results, err := s.StageContracts(context.Background(), filteredContracts)
	if err != nil {
		return nil, err
	}

	return &stagingResults{Results: results, prettyPrinter: s.PrettyPrintValidationError}, nil
}

func (r *stagingResults) ExitCode() int {
	for _, r := range r.Results {
		if r.Err != nil {
			return 1
		}
	}
	return 0
}

func (r *stagingResults) String() string {
	var sb strings.Builder

	// First print out any errors that occurred during staging
	for _, result := range r.Results {
		if result.Err != nil {
			sb.WriteString(r.prettyPrinter(result.Err, nil))
			sb.WriteString("\n")
		}
	}

	numStaged := 0
	numUnvalidated := 0
	numFailed := 0

	for location, result := range r.Results {
		var color aurora.Color
		var prefix string

		if result.Err == nil {
			if result.WasValidated {
				color = aurora.GreenFg
				prefix = "✔"
				numStaged++
			} else {
				color = aurora.YellowFg
				prefix = "⚠"
				numUnvalidated++
			}
		} else {
			color = aurora.RedFg
			prefix = "✘"
			numFailed++
		}

		sb.WriteString(aurora.Colorize(fmt.Sprintf("%s %s ", prefix, location.String()), color).String())
		if result.TxId != flow.EmptyID {
			sb.WriteString(fmt.Sprintf(" (txId: %s)", result.TxId))
		} else if result.Err == nil {
			sb.WriteString(" (no changes)")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	reports := []string{}
	if numStaged > 0 {
		reports = append(reports, aurora.Green(fmt.Sprintf("%d %s staged & validated", numStaged, util.Pluralize("contract", numStaged))).String())
	}
	if numUnvalidated > 0 {
		reports = append(reports, aurora.Yellow(fmt.Sprintf("%d %s staged without validation", numUnvalidated, util.Pluralize("contract", numStaged))).String())
	}
	if numFailed > 0 {
		reports = append(reports, aurora.Red(fmt.Sprintf("%d %s failed to stage", numFailed, util.Pluralize("contract", numFailed))).String())
	}

	sb.WriteString(fmt.Sprintf("Staging results: %s\n\n", strings.Join(reports, ", ")))

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
