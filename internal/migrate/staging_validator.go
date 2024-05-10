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

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/interpreter"
	"github.com/onflow/cadence/runtime/old_parser"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/cadence/runtime/pretty"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/onflow/contract-updater/lib/go/templates"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go/cmd/util/ledger/migrations"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flowkit/v2"

	"github.com/onflow/flow-cli/internal/util"
)

//go:generate mockery --name stagingValidator --inpackage --testonly --case underscore
type stagingValidator interface {
	Validate(stagedContracts []StagedContract) error
	PrettyPrintError(err error, location common.Location) string
}

type stagingValidatorImpl struct {
	flow flowkit.Services

	stagedContracts map[common.AddressLocation]StagedContract

	// Cache for account contract names so we don't have to fetch them multiple times
	accountContractNames map[common.Address][]string
	// All resolved contract code
	contracts map[common.Location][]byte
	// Record errors related to missing staged dependencies, as these are reported separately
	// dependent -> missing dependencies
	missingDependencies map[common.Location][]common.AddressLocation
	// Cache for contract checkers which are reused during program checking & used for the update checker
	checkingCache map[common.Location]*cachedCheckingResult
}

type cachedCheckingResult struct {
	checker *sema.Checker
	err     error
}

type StagedContract struct {
	DeployLocation common.AddressLocation
	SourceLocation common.Location
	Code           []byte
}

type accountContractNamesProviderImpl struct {
	resolverFunc func(address common.Address) ([]string, error)
}

var _ stdlib.AccountContractNamesProvider = &accountContractNamesProviderImpl{}

type missingDependenciesError struct {
	MissingContracts []common.AddressLocation
}

func (e *missingDependenciesError) Error() string {
	contractNames := make([]string, len(e.MissingContracts))
	for i, location := range e.MissingContracts {
		contractNames[i] = location.Name
	}
	return fmt.Sprintf("the following staged contract dependencies could not be found (have they been staged yet?): %v", contractNames)
}

var _ error = &missingDependenciesError{}

type stagingValidatorError struct {
	errors map[common.AddressLocation]error
}

func (e *stagingValidatorError) Error() string {
	var sb strings.Builder
	for location, err := range e.errors {
		sb.WriteString(fmt.Sprintf("error for contract %s: %s\n", location, err))
	}
	return sb.String()
}

func (e *stagingValidatorError) Unwrap() []error {
	var errs []error
	for _, err := range e.errors {
		errs = append(errs, err)
	}
	return errs
}

// MissingDependencies returns the contracts dependended on by the staged contracts that are missing
func (e *stagingValidatorError) MissingDependencies() []common.AddressLocation {
	missingDependencies := make([]common.AddressLocation, 0)
	for _, err := range e.MissingDependencyErrors() {
		missingDependencies = append(missingDependencies, err.MissingContracts...)
	}

	return missingDependencies
}

// ContractsMissingDependencies returns the contracts attempted to be validated that are missing dependencies
func (e *stagingValidatorError) MissingDependencyErrors() map[common.AddressLocation]*missingDependenciesError {
	missingDependencyErrors := make(map[common.AddressLocation]*missingDependenciesError)
	for location := range e.errors {
		var missingDependenciesErr *missingDependenciesError
		if errors.As(e.errors[location], &missingDependenciesErr) {
			missingDependencyErrors[location] = missingDependenciesErr
		}
	}
	return missingDependencyErrors
}

var _ error = &stagingValidatorError{}

var chainIdMap = map[string]flow.ChainID{
	"mainnet": flow.Mainnet,
	"testnet": flow.Testnet,
}

func newStagingValidator(flow flowkit.Services) *stagingValidatorImpl {
	return &stagingValidatorImpl{
		flow:                 flow,
		contracts:            make(map[common.Location][]byte),
		missingDependencies:  make(map[common.Location][]common.AddressLocation),
		checkingCache:        make(map[common.Location]*cachedCheckingResult),
		accountContractNames: make(map[common.Address][]string),
	}
}

func (v *stagingValidatorImpl) Validate(stagedContracts []StagedContract) error {
	v.stagedContracts = make(map[common.AddressLocation]StagedContract)
	for _, stagedContract := range stagedContracts {
		v.stagedContracts[stagedContract.DeployLocation] = stagedContract

		// Add the contract code to the contracts map for pretty printing
		v.contracts[stagedContract.SourceLocation] = stagedContract.Code
	}

	// Load system contracts
	v.loadSystemContracts()

	// Parse and check all staged contracts
	errors := v.parseAndCheckAllStaged()

	for location := range v.stagedContracts {
		// Don't validate contracts with existing errors
		if errors[location] != nil {
			continue
		}

		// Validate the contract update
		err := v.validateContractUpdate(location)
		if err != nil {
			errors[location] = err
		}
	}

	// Return a validator error if there are any errors
	if len(errors) > 0 {
		return &stagingValidatorError{errors: errors}
	}
	return nil
}

func (v *stagingValidatorImpl) parseAndCheckAllStaged() map[common.AddressLocation]error {
	errors := make(map[common.AddressLocation]error)
	for location := range v.stagedContracts {
		_, err := v.checkContract(location)

		// First, check if the contract has missing dependencies
		// This error case takes precedence over any other errors
		// Since it is a prerequisite for the contract to be checked
		// and any other errors are misleading
		if len(v.missingDependencies[location]) > 0 {
			errors[location] = &missingDependenciesError{MissingContracts: v.missingDependencies[location]}
			continue
		}

		// Otherwise, check if there was an error checking the contract
		if err != nil {
			errors[location] = err
			continue
		}
	}
	return errors
}

func (v *stagingValidatorImpl) validateContractUpdate(location common.AddressLocation) error {
	// Get the staged update
	contract, ok := v.stagedContracts[location]
	if !ok {
		return fmt.Errorf("staged contract not found for location: %s", location)
	}

	// Get the account for the contract
	address := flowsdk.Address(location.Address)
	account, err := v.flow.GetAccount(context.Background(), address)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Get the target contract old code
	contractName := location.Name
	contractCode, ok := account.Contracts[contractName]
	if !ok {
		return fmt.Errorf("old contract code not found for contract: %s", contractName)
	}

	// Parse the old contract code
	oldProgram, err := old_parser.ParseProgram(nil, contractCode, old_parser.Config{})
	if err != nil {
		return fmt.Errorf("failed to parse old contract code: %w", err)
	}

	// Convert the new program checker to an interpreter program
	interpreterProgram := interpreter.ProgramFromChecker(v.checkingCache[location].checker)

	// Check if contract code is valid according to Cadence V1 Update Checker
	validator := stdlib.NewCadenceV042ToV1ContractUpdateValidator(
		contract.SourceLocation,
		contractName,
		&accountContractNamesProviderImpl{
			resolverFunc: v.resolveAddressContractNames,
		},
		oldProgram,
		interpreterProgram,
		v.elaborations(),
	)

	// Set the user defined type change checker
	chainId, ok := chainIdMap[v.flow.Network().Name]
	if !ok {
		return fmt.Errorf("unsupported network: %s", v.flow.Network().Name)
	}
	validator.WithUserDefinedTypeChangeChecker(migrations.NewUserDefinedTypeChangeCheckerFunc(chainId))

	err = validator.Validate()
	if err != nil {
		return err
	}

	return nil
}

// Check a contract by location
func (v *stagingValidatorImpl) checkContract(
	importedLocation common.AddressLocation,
	stack ...common.Location,
) (*sema.Checker, error) {
	// Try to load cached checker
	if cacheItem, ok := v.checkingCache[importedLocation]; ok {
		// Inherit missing dependencies from the imported contract
		// This is important because these missing upstream dependencies will not be encountered
		// when checking this dependency tree, as this intermediate checker is cached
		for _, dependent := range stack {
			v.missingDependencies[dependent] = append(v.missingDependencies[dependent], v.missingDependencies[importedLocation]...)
		}
		return cacheItem.checker, cacheItem.err
	}

	// Check the contract code
	checker, err := (func() (*sema.Checker, error) {
		// Resolve the contract code and real location based on whether this is a staged update
		var location common.Location
		var code []byte

		stagedContract, ok := v.stagedContracts[importedLocation]
		if ok {
			location = stagedContract.SourceLocation
			code = stagedContract.Code
		} else {
			var err error

			location = importedLocation
			code, err = v.getStagedContractCode(importedLocation)
			if err != nil {
				return nil, fmt.Errorf("failed to get staged contract code: %w", err)
			}

			// Handle the case where the contract has not been staged yet
			// This missing dependency will be tracked for all dependents
			if code == nil {
				for _, dependent := range stack {
					v.missingDependencies[dependent] = append(v.missingDependencies[dependent], importedLocation)
				}
				err := fmt.Errorf("the following contract has not been staged: %s", importedLocation)
				return nil, err
			}
		}

		// Parse the contract code
		var program *ast.Program
		program, err := parser.ParseProgram(nil, code, parser.Config{})
		if err != nil {
			return nil, err
		}

		// Check the contract code
		checker, err := sema.NewChecker(
			program,
			location,
			nil,
			&sema.Config{
				AccessCheckMode:    sema.AccessCheckModeStrict,
				AttachmentsEnabled: true,
				BaseValueActivationHandler: func(_ common.Location) *sema.VariableActivation {
					// Only checking contracts, so no need to consider script standard library
					return util.NewStandardLibrary().BaseValueActivation
				},
				LocationHandler:            v.resolveLocation,
				ImportHandler:              v.resolveImport(append(stack, importedLocation)),
				MemberAccountAccessHandler: v.resolveAccountAccess,
			},
		)
		if err != nil {
			return nil, err
		}

		err = checker.Check()
		return checker, err
	})()

	// Cache the checking result
	v.checkingCache[importedLocation] = &cachedCheckingResult{
		checker: checker,
		err:     err,
	}
	return checker, err
}

func (v *stagingValidatorImpl) getStagedContractCode(
	location common.AddressLocation,
) ([]byte, error) {
	// First check if the code is already known
	// This may be true for system contracts since they are not staged
	// Or any other staged contracts that have been resolved
	if code, ok := v.contracts[location]; ok {
		return code, nil
	}

	cAddr := cadence.BytesToAddress(location.Address.Bytes())
	cName, err := cadence.NewString(location.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get cadence string from contract name: %w", err)
	}

	value, err := v.flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: templates.GenerateGetStagedContractCodeScript(MigrationContractStagingAddress(v.flow.Network().Name)),
			Args: []cadence.Value{cAddr, cName},
		},
		flowkit.LatestScriptQuery,
	)
	if err != nil {
		return nil, err
	}

	optValue, ok := value.(cadence.Optional)
	if !ok {
		return nil, fmt.Errorf("invalid script return value type: %T", value)
	}

	// If the contract code is nil, the contract has not been staged yet
	// Return nil to indicate this
	if optValue.Value == nil {
		v.contracts[location] = nil
		return nil, nil
	}

	strValue, ok := optValue.Value.(cadence.String)
	if !ok {
		return nil, fmt.Errorf("invalid script return value type: %T", value)
	}

	v.contracts[location] = []byte(strValue)
	return v.contracts[location], nil
}

func (v *stagingValidatorImpl) resolveImport(stack []common.Location) func(*sema.Checker, common.Location, ast.Range) (sema.Import, error) {
	return func(_ *sema.Checker, importedLocation common.Location, _ ast.Range) (sema.Import, error) {
		// Check if the imported location is the crypto checker
		if importedLocation == stdlib.CryptoCheckerLocation {
			cryptoChecker := stdlib.CryptoChecker()
			return sema.ElaborationImport{
				Elaboration: cryptoChecker.Elaboration,
			}, nil
		}

		// Check if the imported location is an address location
		// No other location types are supported (as is the case with code on-chain)
		addrLocation, ok := importedLocation.(common.AddressLocation)
		if !ok {
			return nil, fmt.Errorf("expected address location")
		}

		// Check if this contract has already been resolved
		subChecker, err := v.checkContract(addrLocation, stack...)
		if err != nil {
			return nil, err
		}

		return sema.ElaborationImport{
			Elaboration: subChecker.Elaboration,
		}, nil
	}
}

// This is a copy of the resolveLocation function from the linter/language server
func (v *stagingValidatorImpl) resolveLocation(
	identifiers []ast.Identifier,
	location common.Location,
) (
	[]sema.ResolvedLocation,
	error,
) {
	addressLocation, isAddress := location.(common.AddressLocation)

	// if the location is not an address location, e.g. an identifier location (`import Crypto`),
	// then return a single resolved location which declares all identifiers.

	if !isAddress {
		return []runtime.ResolvedLocation{
			{
				Location:    location,
				Identifiers: identifiers,
			},
		}, nil
	}

	// if the location is an address,
	// and no specific identifiers where requested in the import statement,
	// then fetch all identifiers at this address

	if len(identifiers) == 0 {
		// if there is no contract name resolver,
		// then return no resolved locations
		contractNames, err := v.resolveAddressContractNames(addressLocation.Address)
		if err != nil {
			return nil, err
		}

		// if there are no contracts deployed,
		// then return no resolved locations

		if len(contractNames) == 0 {
			return nil, nil
		}

		identifiers = make([]ast.Identifier, len(contractNames))

		for i := range identifiers {
			identifiers[i] = runtime.Identifier{
				Identifier: contractNames[i],
			}
		}
	}

	// return one resolved location per identifier.
	// each resolved location is an address contract location

	resolvedLocations := make([]runtime.ResolvedLocation, len(identifiers))
	for i := range resolvedLocations {
		identifier := identifiers[i]
		resolvedLocations[i] = runtime.ResolvedLocation{
			Location: common.AddressLocation{
				Address: addressLocation.Address,
				Name:    identifier.Identifier,
			},
			Identifiers: []runtime.Identifier{identifier},
		}
	}

	return resolvedLocations, nil
}

func (v *stagingValidatorImpl) resolveAccountAccess(checker *sema.Checker, memberLocation common.Location) bool {
	if checker == nil {
		return false
	}

	checkerLocation, ok := checker.Location.(common.StringLocation)
	if !ok {
		return false
	}

	memberAddressLocation, ok := memberLocation.(common.AddressLocation)
	if !ok {
		return false
	}

	// If the source code of the update is being checked, we should check account access based on the
	// targeted network location of the contract & not the source code location
	for deployLocation, stagedContract := range v.stagedContracts {
		if stagedContract.SourceLocation == checkerLocation {
			return memberAddressLocation.Address == deployLocation.Address
		}
	}

	return false
}

func (v *stagingValidatorImpl) resolveAddressContractNames(address common.Address) ([]string, error) {
	// Check if the contract names are already cached
	if names, ok := v.accountContractNames[address]; ok {
		return names, nil
	}

	cAddr := cadence.BytesToAddress(address.Bytes())
	value, err := v.flow.ExecuteScript(
		context.Background(),
		flowkit.Script{
			Code: templates.GenerateGetStagedContractNamesForAddressScript(MigrationContractStagingAddress(v.flow.Network().Name)),
			Args: []cadence.Value{cAddr},
		},
		flowkit.LatestScriptQuery,
	)

	if err != nil {
		return nil, err
	}

	optValue, ok := value.(cadence.Optional)
	if !ok {
		return nil, fmt.Errorf("invalid script return value type: %T", value)
	}

	arrValue, ok := optValue.Value.(cadence.Array)
	if !ok {
		return nil, fmt.Errorf("invalid script return value type: %T", value)
	}

	// Cache the contract names
	for _, name := range arrValue.Values {
		strName, ok := name.(cadence.String)
		if !ok {
			return nil, fmt.Errorf("invalid array value type: %T", name)
		}
		v.accountContractNames[address] = append(v.accountContractNames[address], string(strName))
	}

	return v.accountContractNames[address], nil
}

func (v *stagingValidatorImpl) loadSystemContracts() {
	chainId, ok := chainIdMap[v.flow.Network().Name]
	if !ok {
		return
	}

	stagedSystemContracts := migrations.SystemContractChanges(chainId, migrations.SystemContractsMigrationOptions{
		Burner: migrations.BurnerContractChangeUpdate, // needs to be update for now since BurnerChangeDeploy is a no-op in flow-go
		EVM:    migrations.EVMContractChangeFull,
	})
	for _, stagedSystemContract := range stagedSystemContracts {
		location := common.AddressLocation{
			Address: stagedSystemContract.Address,
			Name:    stagedSystemContract.Name,
		}

		v.contracts[location] = stagedSystemContract.Code
		v.accountContractNames[stagedSystemContract.Address] = append(v.accountContractNames[stagedSystemContract.Address], stagedSystemContract.Name)
	}
}

func (v *stagingValidatorImpl) elaborations() map[common.Location]*sema.Elaboration {
	elaborations := make(map[common.Location]*sema.Elaboration)
	for location, cacheItem := range v.checkingCache {
		checker := cacheItem.checker
		if checker == nil {
			continue
		}
		elaborations[location] = checker.Elaboration
	}
	return elaborations
}

// Helper for pretty printing errors
// While it is done by default in checker/parser errors, this has two purposes:
// 1. Add color to the error message
// 2. Use pretty printing on contract update errors which do not do this by default
func (v *stagingValidatorImpl) PrettyPrintError(err error, location common.Location) string {
	var sb strings.Builder
	printErr := pretty.NewErrorPrettyPrinter(&sb, true).
		PrettyPrintError(err, location, v.contracts)
	if printErr != nil {
		return fmt.Sprintf("failed to pretty print error: %v", printErr)
	} else {
		return sb.String()
	}
}

// Stdlib handler used by the Cadence V1 Update Checker to resolve contract names
// When an address import with no identifiers is used, the contract names are resolved
func (a *accountContractNamesProviderImpl) GetAccountContractNames(
	address common.Address,
) ([]string, error) {
	return a.resolverFunc(address)
}
