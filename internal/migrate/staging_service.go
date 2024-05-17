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
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/project"
	"github.com/onflow/flowkit/v2/transactions"
)

//go:generate mockery --name stagingService --inpackage --testonly --case underscore
type stagingService interface {
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
	Err          error
	WasValidated bool
	TxId         flow.Identifier
}

var _ stagingService = &stagingServiceImpl{}

type fatalValidationError struct {
	err error
}

func (e *fatalValidationError) Error() string {
	return e.err.Error()
}

var _ error = &fatalValidationError{}

func newStagingService(
	flow flowkit.Services,
	state *flowkit.State,
	logger output.Logger,
	validator stagingValidator,
	unvalidatedContractsHandler func(*stagingValidatorError) bool,
) *stagingServiceImpl {
	handler := func(err *stagingValidatorError) bool {
		return false
	}
	if unvalidatedContractsHandler != nil {
		handler = unvalidatedContractsHandler
	}

	return &stagingServiceImpl{
		flow:                        flow,
		state:                       state,
		logger:                      logger,
		validator:                   validator,
		unvalidatedContractsHandler: handler,
	}
}

func (s *stagingServiceImpl) StageContracts(ctx context.Context, contracts []*project.Contract) (map[common.AddressLocation]stagingResult, error) {
	// Convert contracts to staged contracts
	stagedContracts, err := s.convertToStagedContracts(contracts)
	if err != nil {
		return nil, err
	}

	// If validation is disabled, just stage the contracts
	if s.validator == nil {
		s.logger.Info("Skipping contract code validation, you may monitor the status of your contract using the `flow migrate is-validated` command\n")
		s.logger.StartProgress(fmt.Sprintf("Staging %d contracts for accounts: %s", len(contracts), s.state.AccountsForNetwork(s.flow.Network()).String()))
		defer s.logger.StopProgress()

		results := s.stageContracts(ctx, stagedContracts)
		return results, nil
	}

	// Otherwise, validate and stage the contracts
	return s.validateAndStageContracts(ctx, stagedContracts)
}

func (s *stagingServiceImpl) validateAndStageContracts(ctx context.Context, contracts []stagedContractUpdate) (map[common.AddressLocation]stagingResult, error) {
	s.logger.StartProgress(fmt.Sprintf("Validating and staging %d contracts", len(contracts)))
	defer s.logger.StopProgress()

	// Validate all contracts
	var validatorError *stagingValidatorError
	err := s.validator.Validate(contracts)

	// We will handle validation errors separately per contract to allow for partial staging
	if err != nil && !errors.As(err, &validatorError) {
		return nil, &fatalValidationError{err: err}
	}

	// Collect all staging errors to report to the user
	results := make(map[common.AddressLocation]stagingResult)
	if validatorError != nil {
		for location, err := range validatorError.errors {
			results[location] = stagingResult{
				Err:          err,
				WasValidated: true,
			}
		}
	}

	// Split contracts into valid, and contracts with missing dependencies
	missingDepsContracts := make([]stagedContractUpdate, 0)
	validContracts := make([]stagedContractUpdate, 0)
	if validatorError == nil {
		validContracts = contracts
	} else {
		for _, contract := range contracts {
			contractErr := validatorError.errors[contract.DeployLocation]

			var missingDepsError *missingDependenciesError
			if errors.As(contractErr, &missingDepsError) {
				missingDepsContracts = append(missingDepsContracts, contract)
			} else if contractErr == nil {
				validContracts = append(validContracts, contract)
			}
		}
	}

	s.logger.StopProgress()

	// Now, handle contracts that were not validated due to missing dependencies
	if len(missingDepsContracts) > 0 && s.unvalidatedContractsHandler(validatorError) {
		for location, res := range s.stageContracts(ctx, missingDepsContracts) {
			res.WasValidated = false
			results[location] = res
		}
	}

	// Stage contracts that passed validation
	for contractLocation, res := range s.stageContracts(ctx, validContracts) {
		res.WasValidated = true
		results[contractLocation] = res
	}

	return results, nil
}

func (s *stagingServiceImpl) stageContracts(ctx context.Context, contracts []stagedContractUpdate) map[common.AddressLocation]stagingResult {
	results := make(map[common.AddressLocation]stagingResult)
	for _, contract := range contracts {
		txId, err := s.stageContract(
			ctx,
			contract,
		)
		if err != nil {
			results[contract.DeployLocation] = stagingResult{
				Err:  fmt.Errorf("failed to stage contract: %w", err),
				TxId: txId,
			}
		} else {
			results[contract.DeployLocation] = stagingResult{
				Err:  nil,
				TxId: txId,
			}
		}
	}

	return results
}

func (s *stagingServiceImpl) stageContract(ctx context.Context, contract stagedContractUpdate) (flow.Identifier, error) {
	s.logger.StartProgress(fmt.Sprintf("Staging contract %s", contract.DeployLocation))
	defer s.logger.StopProgress()

	// Check if the staged contract has changed
	if !s.hasStagedContractChanged(contract) {
		return flow.EmptyID, nil
	}

	cName := cadence.String(contract.DeployLocation.Name)
	cCode := cadence.String(contract.Code)

	// Get the account for the contract
	account, err := s.state.Accounts().ByAddress(flow.Address(contract.DeployLocation.Address))
	if err != nil {
		return flow.Identifier{}, fmt.Errorf("failed to get account for contract %s: %w", contract.DeployLocation.Name, err)
	}

	_, res, err := s.flow.SendTransaction(
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

	if res.Error != nil {
		return res.TransactionID, fmt.Errorf("failed to stage contract: %w", res.Error)
	}

	return res.TransactionID, nil
}

func (s *stagingServiceImpl) hasStagedContractChanged(contract stagedContractUpdate) bool {
	// Get the staged contract code
	stagedCode, err := getStagedContractCode(context.Background(), s.flow, contract.DeployLocation)
	if err != nil {
		// swallow error, if we can't get the staged contract code, we should stage
		return true
	}

	if stagedCode == nil {
		return true
	}

	// If the staged contract code is different from the contract code, we need to stage it
	if string(stagedCode) != string(contract.Code) {
		return true
	}

	return false
}

func (s *stagingServiceImpl) convertToStagedContracts(contracts []*project.Contract) ([]stagedContractUpdate, error) {
	// Collect all staged contracts
	stagedContracts := make([]stagedContractUpdate, 0)
	for _, contract := range contracts {
		rawScript := flowkit.Script{
			Code:     contract.Code(),
			Location: contract.Location(),
			Args:     contract.Args,
		}

		// Replace imports in the contract
		script, err := s.flow.ReplaceImportsInScript(context.Background(), rawScript)
		if err != nil {
			return nil, fmt.Errorf("failed to replace imports in contract %s: %w", contract.Name, err)
		}

		// We need the real name of the contract, not the name in flow.json
		program, err := project.NewProgram(script.Code, script.Args, script.Location)
		if err != nil {
			return nil, fmt.Errorf("failed to parse contract %s: %w", contract.Name, err)
		}

		name, err := program.Name()
		if err != nil {
			return nil, fmt.Errorf("failed to parse contract name: %w", err)
		}

		// Convert relevant information to Cadence types
		deployLocation := common.NewAddressLocation(nil, common.Address(contract.AccountAddress), name)
		sourceLocation := common.StringLocation(contract.Location())

		stagedContracts = append(stagedContracts, stagedContractUpdate{
			DeployLocation: deployLocation,
			SourceLocation: sourceLocation,
			Code:           script.Code,
		})
	}

	return stagedContracts, nil
}

func (v *stagingServiceImpl) PrettyPrintValidationError(err error, location common.Location) string {
	return v.validator.PrettyPrintError(err, location)
}
