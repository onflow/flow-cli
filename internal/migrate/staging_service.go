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
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/project"
	"github.com/onflow/flowkit/v2/transactions"
)

//go:generate mockery --name stagingService --inpackage --testonly --case underscore
type stagingService interface {
	// StageContracts stages contracts for the network
	StageContracts(ctx context.Context, contracts []*project.Contract) (map[common.AddressLocation]stagingResult, error)
	PrettyPrintValidationError(err error, location common.Location) string
}

type stagingServiceImpl struct {
	flow                        flowkit.Services
	state                       *flowkit.State
	logger                      output.Logger
	validator                   stagingValidator
	unvalidatedContractsHandler func(*stagingValidatorError) bool
}

type stagingResult struct {
	err          error
	wasValidated bool
}

var _ stagingService = &stagingServiceImpl{}

func newStagingService(
	flow flowkit.Services,
	state *flowkit.State,
	logger output.Logger,
	validator stagingValidator,
	unvalidatedContractsHandler func(*stagingValidatorError) bool,
) *stagingServiceImpl {
	return &stagingServiceImpl{
		flow:                        flow,
		state:                       state,
		logger:                      logger,
		validator:                   validator,
		unvalidatedContractsHandler: unvalidatedContractsHandler,
	}
}

func (s *stagingServiceImpl) StageContracts(ctx context.Context, contracts []*project.Contract) (map[common.AddressLocation]stagingResult, error) {
	// If validation is disabled, just stage the contracts
	if s.validator == nil {
		s.logger.Info("Skipping contract code validation, you may monitor the status of your contract using the `flow migrate is-validated` command\n")
		s.logger.StartProgress(fmt.Sprintf("Staging %d contracts for accounts: %s", len(contracts), s.state.AccountsForNetwork(s.flow.Network()).String()))
		defer s.logger.StopProgress()

		results := make(map[common.AddressLocation]stagingResult)
		errorMap := s.stageContracts(ctx, contracts)
		for location, err := range errorMap {
			results[location] = stagingResult{
				err:          err,
				wasValidated: false,
			}
		}

		return results, nil
	}

	return s.validateAndStageContracts(ctx, contracts)
}

func (s *stagingServiceImpl) validateAndStageContracts(ctx context.Context, contracts []*project.Contract) (map[common.AddressLocation]stagingResult, error) {
	s.logger.StartProgress(fmt.Sprintf("Validating and staging %d contracts", len(contracts)))
	defer s.logger.StopProgress()

	// Collect all staged contracts
	stagedContracts := make([]StagedContract, 0)
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
	err := s.validator.Validate(stagedContracts)

	// We will handle validation errors separately per contract to allow for partial staging
	if err != nil && !errors.As(err, &validatorError) {
		return nil, fmt.Errorf("failed to validate contracts: %w", err)
	}

	// Collect all staging errors to report to the user
	results := make(map[common.AddressLocation]stagingResult)
	if validatorError != nil {
		for location, err := range validatorError.errors {
			results[location] = stagingResult{
				err:          err,
				wasValidated: true,
			}
		}
	}

	// First, stage contracts that passed validation
	newResults := s.stageValidContracts(ctx, contracts, validatorError)
	for location, err := range newResults {
		results[location] = stagingResult{
			err:          err,
			wasValidated: true,
		}
	}

	// Now, handle contracts that failed validation
	// This will prompt the user to continue staging contracts that have missing dependencies
	// Other validation errors will be fatal
	newResults = s.maybeStageInvalidContracts(ctx, contracts, validatorError)
	for location, err := range newResults {
		results[location] = stagingResult{
			err:          err,
			wasValidated: false,
		}
	}

	return results, nil
}

func (s *stagingServiceImpl) stageValidContracts(ctx context.Context, contracts []*project.Contract, validatorError *stagingValidatorError) map[common.AddressLocation]error {
	validContracts := make([]*project.Contract, 0)

	if validatorError == nil || validatorError.errors == nil {
		validContracts = contracts
	} else {
		for _, contract := range contracts {
			contractLocation := contractDeploymentLocation(contract)
			if _, hasError := validatorError.errors[contractLocation]; !hasError {
				validContracts = append(validContracts, contract)
			}
		}
	}

	// Stage contracts that passed validation
	results := make(map[common.AddressLocation]error)
	for contractLocation, err := range s.stageContracts(ctx, validContracts) {
		results[contractLocation] = err
	}

	return results
}

func (s *stagingServiceImpl) maybeStageInvalidContracts(ctx context.Context, contracts []*project.Contract, validatorErr *stagingValidatorError) map[common.AddressLocation]error {
	if validatorErr == nil || validatorErr.errors == nil {
		return nil
	}

	// Fill results with all validation errors initially
	// These will be overwritten if contracts are staged
	results := make(map[common.AddressLocation]error)

	missingDependencyErrors := validatorErr.MissingDependencyErrors()
	if len(missingDependencyErrors) == 0 {
		return results
	}

	// Prompt user to continue staging contracts that have missing dependencies
	willStage := s.unvalidatedContractsHandler(validatorErr)

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
			results[contractLocation] = missingDependencyErrors[contractLocation]
		}
	}

	// Stage contracts that have missing dependencies & add errors to results
	for location, err := range s.stageContracts(ctx, unvalidatedContracts) {
		results[location] = err
	}

	return results
}

// Stage contracts for network with an optional filter
// Returns a map of staged/attempted contracts and errors occuring if any
func (s *stagingServiceImpl) stageContracts(ctx context.Context, contracts []*project.Contract) map[common.AddressLocation]error {
	results := make(map[common.AddressLocation]error)
	for _, contract := range contracts {
		targetAccount, err := s.state.Accounts().ByName(contract.AccountName)
		deployLocation := contractDeploymentLocation(contract)

		if err != nil {
			results[deployLocation] = fmt.Errorf("failed to get account by contract name: %w", err)
			continue
		}

		s.logger.StartProgress(fmt.Sprintf("Staging contract %s for account %s", contract.Name, targetAccount.Name))

		_, err = s.stageContract(
			ctx,
			targetAccount,
			contract.Name,
			contract.Code(),
		)
		if err != nil {
			results[deployLocation] = fmt.Errorf("failed to stage contract: %w", err)
		} else {
			results[deployLocation] = nil
		}

		s.logger.StopProgress()
	}
	return results
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

func (v *stagingServiceImpl) PrettyPrintValidationError(err error, location common.Location) string {
	return v.validator.PrettyPrintError(err, location)
}

// helper function to create a common.AddressLocation for a deployment contract
func contractDeploymentLocation(contract *project.Contract) common.AddressLocation {
	return common.NewAddressLocation(nil, common.Address(contract.AccountAddress), contract.Name)
}
