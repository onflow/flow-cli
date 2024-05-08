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

	"github.com/logrusorgru/aurora/v4"
	"github.com/manifoldco/promptui"
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/contract-updater/lib/go/templates"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/project"
	"github.com/onflow/flowkit/v2/transactions"
)

//go:generate mockery --name stagingService --output ./mocks --case underscore --exported
type stagingService interface {
	// StageContracts stages contracts for the network, based on an optional filter
	StageContracts(ctx context.Context, filter func(*project.Contract) bool) (map[common.AddressLocation]error, error)
}

type stagingServiceImpl struct {
	flow              flowkit.Services
	state             *flowkit.State
	logger            output.Logger
	validationEnabled bool
}

var _ stagingService = &stagingServiceImpl{}

func newStagingService(flow flowkit.Services, state *flowkit.State, logger output.Logger, validationEnabled bool) *stagingServiceImpl {
	return &stagingServiceImpl{
		flow:              flow,
		state:             state,
		logger:            logger,
		validationEnabled: validationEnabled,
	}
}

func (s *stagingServiceImpl) StageContracts(ctx context.Context, filter func(*project.Contract) bool) (map[common.AddressLocation]error, error) {
	contracts, err := s.state.DeploymentContractsByNetwork(s.flow.Network())
	if err != nil {
		return nil, err
	}

	// Filter contracts
	filtered := make([]*project.Contract, 0, len(contracts))
	for _, contract := range contracts {
		if filter(contract) {
			filtered = append(filtered, contract)
		}
	}

	return s.validateAndStageContracts(ctx, filtered)
}

func (s *stagingServiceImpl) validateAndStageContracts(ctx context.Context, contracts []*project.Contract) (map[common.AddressLocation]error, error) {
	// If validation is disabled, just stage the contracts
	if !s.validationEnabled {
		s.logger.Info("Skipping contract code validation, you may monitor the status of your contract using the `flow migrate is-validated` command\n")
		s.logger.StartProgress(fmt.Sprintf("Staging %d contracts for accounts: %s", len(contracts), s.state.AccountsForNetwork(s.flow.Network()).String()))
		defer s.logger.StopProgress()

		return s.stageContracts(ctx, contracts), nil
	}

	s.logger.StartProgress(fmt.Sprintf("Validating and staging %d contracts for accounts: %s", len(contracts), s.state.AccountsForNetwork(s.flow.Network()).String()))
	defer s.logger.StopProgress()

	// Create a new validator
	validator := newStagingValidator(s.flow)

	// Collect all staged contracts
	stagedContracts := make([]StagedContract, len(contracts))
	for _, contract := range contracts {
		deployLocation := common.NewAddressLocation(nil, common.Address(contract.AccountAddress), contract.Name)
		sourceLocation := common.StringLocation(contract.Location())

		stagedContracts = append(stagedContracts, StagedContract{
			DeployLocation: deployLocation,
			SourceLocation: sourceLocation,
			Code:           contract.Code(),
		})
	}

	// Validate all contracts
	var validatorError *stagingValidatorError
	err := validator.Validate(stagedContracts)

	// We will handle validation errors separately per contract to allow for partial staging
	if err != nil && !errors.As(err, &validatorError) {
		return nil, fmt.Errorf("failed to validate contracts: %w", err)
	}

	// Collect all staging errors to report to the user
	results := make(map[common.AddressLocation]error)

	// First, stage contracts that passed validation
	newErrors := s.stageValidContracts(ctx, validatorError, contracts)
	for location, err := range newErrors {
		results[location] = err
	}

	// Now, handle contracts that failed validation
	// This will prompt the user to continue staging contracts that have missing dependencies
	// Other validation errors will be fatal
	newErrors = s.maybeStageInvalidContracts(ctx, validatorError, contracts)
	for location, err := range newErrors {
		results[location] = err
	}

	return results, nil
}

func (s *stagingServiceImpl) stageValidContracts(ctx context.Context, validatorError *stagingValidatorError, contracts []*project.Contract) map[common.AddressLocation]error {
	stagingErrors := make(map[common.AddressLocation]error)
	validContracts := make([]*project.Contract, 0, len(contracts))
	for _, contract := range contracts {
		contractLocation := common.NewAddressLocation(nil, common.Address(contract.AccountAddress), contract.Name)
		if _, hasError := validatorError.errors[contractLocation]; !hasError {
			validContracts = append(validContracts, contract)
		}
	}
	for contractLocation, err := range s.stageContracts(ctx, validContracts) {
		if err != nil {
			stagingErrors[contractLocation] = err
		}
	}

	return stagingErrors
}

func (s *stagingServiceImpl) maybeStageInvalidContracts(ctx context.Context, validatorError *stagingValidatorError, contracts []*project.Contract) map[common.AddressLocation]error {
	// Fill results with all validation errors initially
	// These will be overwritten if contracts are staged
	results := make(map[common.AddressLocation]error)
	for contractLocation, contractErr := range validatorError.errors {
		results[contractLocation] = contractErr
	}

	missingDependencyErrors := validatorError.MissingDependencyErrors()
	hasMissingDependencies := len(missingDependencyErrors) > 0
	if !hasMissingDependencies {
		return results
	}

	// Prompt user to continue staging contracts that have missing dependencies
	willStage := s.promptStagingUnvalidatedContracts(validatorError)

	// If user does not want to stage these contracts, we can just return
	// validation errors as-is
	if !willStage {
		return results
	}

	// Otherwise, we will stage the contracts that have missing dependencies
	unvalidatedContracts := make([]*project.Contract, 0, len(missingDependencyErrors))
	for _, contract := range contracts {
		contractLocation := contractDeploymentLocation(contract)
		if _, hasError := missingDependencyErrors[contractLocation]; hasError {
			unvalidatedContracts = append(unvalidatedContracts, contract)
		}
	}

	// Stage contracts that have missing dependencies & add errors to results
	for location, err := range s.stageContracts(ctx, unvalidatedContracts) {
		if err != nil {
			results[location] = err
		}
	}

	return results
}

func (s *stagingServiceImpl) promptStagingUnvalidatedContracts(validatorError *stagingValidatorError) bool {
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
	s.logger.Error(infoMessage.String())

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

// Stage contracts for network with an optional filter
// Returns a map of staged/attempted contracts and errors occuring if any
func (s *stagingServiceImpl) stageContracts(ctx context.Context, contracts []*project.Contract) map[common.AddressLocation]error {
	stagingErrors := make(map[common.AddressLocation]error)
	for _, contract := range contracts {
		targetAccount, err := s.state.Accounts().ByName(contract.AccountName)
		deployLocation := contractDeploymentLocation(contract)

		if err != nil {
			stagingErrors[deployLocation] = fmt.Errorf("failed to get account by contract name: %w", err)
			continue
		}

		txID, err := s.stageContract(
			ctx,
			targetAccount,
			contract.Name,
			contract.Code(),
		)
		if err != nil {
			stagingErrors[deployLocation] = err
			continue
		}

		s.logger.Info(fmt.Sprintf(
			"%s -> 0x%s (%s)",
			output.Green(contract.Name),
			contract.AccountAddress,
			txID.String(),
		))
	}
	return stagingErrors
}

func (s *stagingServiceImpl) stageContract(ctx context.Context, account *accounts.Account, contractName string, contractCode []byte) (flow.Identifier, error) {
	cName := cadence.String(contractName)
	cCode := cadence.String(contractCode)

	tx, _, err := s.flow.SendTransaction(
		context.Background(),
		transactions.SingleAccountRole(*account),
		flowkit.Script{
			Code: templates.GenerateStageContractScript(MigrationContractStagingAddress(s.flow.Network().Name)),
			Args: []cadence.Value{cName, cCode},
		},
		flow.DefaultTransactionGasLimit,
	)
	if err != nil {
		return flow.Identifier{}, err
	}

	return tx.ID(), nil
}

func (s *stagingServiceImpl) prettyPrintValidationResults(contracts []*project.Contract, validatorError *stagingValidatorError) (string, error) {
	var sb strings.Builder

	sb.WriteString("Validation results:\n")

	for _, contract := range contracts {
		sb.WriteString(fmt.Sprintf("  - %s ", contract.Name))

		contractLocation := common.NewAddressLocation(nil, common.Address(contract.AccountAddress), contract.Name)
		contractErr, hasError := validatorError.errors[contractLocation]

		var missingDependenciesErr *missingDependenciesError

		if !hasError {
			sb.WriteString(aurora.Green(fmt.Sprintf("  - ✔ %s\n", contract.Name)).String())
		} else if errors.As(contractErr, &missingDependenciesErr) {
			var yellow strings.Builder
			yellow.WriteString(fmt.Sprintf("  - ⚠ %s ", contract.Name))
			yellow.WriteString(fmt.Sprintf("(%d missing dependencies)\n", len(missingDependenciesErr.MissingContracts)))
			sb.WriteString(aurora.Yellow(yellow.String()).String())

			for _, missingContract := range missingDependenciesErr.MissingContracts {
				sb.WriteString(fmt.Sprintf("    - %s", missingContract))
			}
		} else {
			sb.WriteString(aurora.Red(fmt.Sprintf("  - ✘ %s (validation failed)\n", contract.Name)).String())

			// TODO: too much?
			sb.WriteString(fmt.Sprintf("    %s\n", contractErr.Error()))
		}

		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString("You may still stage your contract, however it will be unable to be migrated until the missing contracts are staged by their respective owners.  It is important to monitor the status of your contract using the `flow migrate is-validated` command\n")

	return sb.String(), nil
}

// helper function to create a common.AddressLocation for a deployment contract
func contractDeploymentLocation(contract *project.Contract) common.AddressLocation {
	return common.NewAddressLocation(nil, common.Address(contract.AccountAddress), contract.Name)
}
