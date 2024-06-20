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
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

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
	Validate(stagedContracts []stagedContractUpdate) error
	PrettyPrintError(err error, location common.Location) string
}

type stagingValidatorImpl struct {
	flow flowkit.Services

	stagedContracts map[common.AddressLocation]stagedContractUpdate

	// Cache for account contract names so we don't have to fetch them multiple times
	accountContractNames map[common.Address][]string
	// All resolved contract code
	contracts map[common.Location][]byte

	// Dependency graph for staged contracts
	// This root level map holds all nodes
	graph map[common.Location]node

	// Cache for contract checkers which are reused during program checking & used for the update checker
	checkingCache map[common.Location]*cachedCheckingResult
}

type node map[common.Location]node

type cachedCheckingResult struct {
	checker *sema.Checker
	err     error
}

type stagedContractUpdate struct {
	DeployLocation common.AddressLocation
	SourceLocation common.StringLocation
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

type upstreamValidationError struct {
	Location        common.Location
	BadDependencies []common.Location
}

func (e *upstreamValidationError) Error() string {
	return fmt.Sprintf("contract %s has upstream validation errors, related to the following dependencies: %v", e.Location, e.BadDependencies)
}

var _ error = &upstreamValidationError{}

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

// MissingDependencies returns the contracts dependended on by the staged contracts that are missing
func (e *stagingValidatorError) MissingDependencies() []common.AddressLocation {
	missingDepsMap := make(map[common.AddressLocation]struct{})
	for _, err := range e.MissingDependencyErrors() {
		for _, missingDep := range err.MissingContracts {
			missingDepsMap[missingDep] = struct{}{}
		}
	}

	missingDependencies := make([]common.AddressLocation, 0)
	for missingDep := range missingDepsMap {
		missingDependencies = append(missingDependencies, missingDep)
	}

	sortAddressLocations(missingDependencies)
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
		checkingCache:        make(map[common.Location]*cachedCheckingResult),
		accountContractNames: make(map[common.Address][]string),
		graph:                make(map[common.Location]node),
	}
}

func (v *stagingValidatorImpl) Validate(stagedContracts []stagedContractUpdate) error {
	v.stagedContracts = make(map[common.AddressLocation]stagedContractUpdate)
	for _, stagedContract := range stagedContracts {
		v.stagedContracts[stagedContract.DeployLocation] = stagedContract

		// Add the contract code to the contracts map for pretty printing
		v.contracts[stagedContract.SourceLocation] = stagedContract.Code
	}

	// Load system contracts
	v.loadSystemContracts()

	// Parse and check all staged contracts
	errs := v.checkAllStaged()

	// Validate all contract updates
	for _, contract := range v.stagedContracts {
		// Don't validate contracts with existing errors
		if errs[contract.SourceLocation] != nil {
			continue
		}

		// Validate the contract update
		checker := v.checkingCache[contract.SourceLocation].checker
		err := v.validateContractUpdate(contract, checker)
		if err != nil {
			errs[contract.SourceLocation] = err
		}
	}

	// Check for any upstream contract update failures
	for _, contract := range v.stagedContracts {
		err := errs[contract.SourceLocation]

		// We will override any errors other than those related
		// to missing dependencies, since they are more specific
		// forms of upstream validation errors
		var missingDependenciesErr *missingDependenciesError
		if errors.As(err, &missingDependenciesErr) {
			continue
		}

		// Leave cyclic import errors to the checker
		var cyclicImportErr *sema.CyclicImportsError
		if errors.As(err, &cyclicImportErr) {
			continue
		}

		badDeps := make([]common.Location, 0)
		v.forEachDependency(contract, func(dependency common.Location) {
			strLocation, ok := dependency.(common.StringLocation)
			if !ok {
				return
			}

			if errs[strLocation] != nil {
				badDeps = append(badDeps, dependency)
			}
		})

		if len(badDeps) > 0 {
			errs[contract.SourceLocation] = &upstreamValidationError{
				Location:        contract.SourceLocation,
				BadDependencies: badDeps,
			}
		}
	}

	// Return a validator error if there are any errors
	if len(errs) > 0 {
		// Map errors to address locations
		errsByAddress := make(map[common.AddressLocation]error)
		for _, contract := range v.stagedContracts {
			err := errs[contract.SourceLocation]
			if err != nil {
				errsByAddress[contract.DeployLocation] = err
			}
		}
		return &stagingValidatorError{errors: errsByAddress}
	}
	return nil
}

func (v *stagingValidatorImpl) checkAllStaged() map[common.StringLocation]error {
	errors := make(map[common.StringLocation]error)
	for _, contract := range v.stagedContracts {
		_, err := v.checkContract(contract.SourceLocation)
		if err != nil {
			errors[contract.SourceLocation] = err
		}
	}

	// Report any missing dependency errors separately
	// These will override any other errors parsing/checking errors
	// Note: nodes are not visited more than once so cyclic imports are not an issue
	// They will be reported, however, by the checker, if they do exist
	for _, contract := range v.stagedContracts {
		// Create a set of all dependencies
		missingDependencies := make([]common.AddressLocation, 0)
		v.forEachDependency(contract, func(dependency common.Location) {
			if code := v.contracts[dependency]; code == nil {
				if dependency, ok := dependency.(common.AddressLocation); ok {
					missingDependencies = append(missingDependencies, dependency)
				}
			}
		})

		if len(missingDependencies) > 0 {
			errors[contract.SourceLocation] = &missingDependenciesError{
				MissingContracts: missingDependencies,
			}
		}
	}
	return errors
}

func (v *stagingValidatorImpl) validateContractUpdate(contract stagedContractUpdate, checker *sema.Checker) (err error) {
	// Gracefully recover from panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during contract update validation: %v", r)
		}
	}()

	// Get the account for the contract
	address := flowsdk.Address(contract.DeployLocation.Address)
	account, err := v.flow.GetAccount(context.Background(), address)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Get the target contract old code
	contractName := contract.DeployLocation.Name
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
	interpreterProgram := interpreter.ProgramFromChecker(checker)

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
	importedLocation common.Location,
) (checker *sema.Checker, err error) {
	// Try to load cached checker
	if cacheItem, ok := v.checkingCache[importedLocation]; ok {
		return cacheItem.checker, cacheItem.err
	}

	// Gracefully recover from panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during contract checking: %v", r)
		}
	}()

	// Cache the checking result
	defer func() {
		var cacheItem *cachedCheckingResult
		if existingCacheItem, ok := v.checkingCache[importedLocation]; ok {
			cacheItem = existingCacheItem
		} else {
			cacheItem = &cachedCheckingResult{}
		}

		cacheItem.checker = checker
		cacheItem.err = err
	}()

	// Resolve the contract code and real location based on whether this is a staged update
	var code []byte

	// If it's an address location, get the staged contract code from the network
	if addressLocation, ok := importedLocation.(common.AddressLocation); ok {
		code, err = v.getStagedContractCode(addressLocation)
		if err != nil {
			return nil, err
		}
	} else {
		// Otherwise, the code is already known
		code = v.contracts[importedLocation]
		if code == nil {
			return nil, fmt.Errorf("contract code not found for location: %s", importedLocation)
		}
	}

	// Parse the contract code
	var program *ast.Program
	program, err = parser.ParseProgram(nil, code, parser.Config{})
	if err != nil {
		return nil, err
	}

	// Check the contract code
	checker, err = sema.NewChecker(
		program,
		importedLocation,
		nil,
		&sema.Config{
			AccessCheckMode:    sema.AccessCheckModeStrict,
			AttachmentsEnabled: true,
			BaseValueActivationHandler: func(_ common.Location) *sema.VariableActivation {
				// Only checking contracts, so no need to consider script standard library
				return util.NewStandardLibrary().BaseValueActivation
			},
			LocationHandler:            v.resolveLocation,
			ImportHandler:              v.resolveImport,
			MemberAccountAccessHandler: v.resolveAccountAccess,
		},
	)
	if err != nil {
		return nil, err
	}

	// We must add this checker to the cache before checking to prevent cyclic imports
	v.checkingCache[importedLocation] = &cachedCheckingResult{
		checker: checker,
	}

	err = checker.Check()
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

	code, err := getStagedContractCode(context.Background(), v.flow, location)
	if err != nil {
		return nil, err
	}

	v.contracts[location] = code
	return v.contracts[location], nil
}

func (v *stagingValidatorImpl) resolveImport(parentChecker *sema.Checker, importedLocation common.Location, _ ast.Range) (sema.Import, error) {
	// Add this to the dependency graph
	if parentChecker != nil {
		v.addDependency(parentChecker.Location, importedLocation)
	}

	// Check if the imported location is the crypto checker
	if importedLocation == stdlib.CryptoCheckerLocation {
		cryptoChecker := stdlib.CryptoChecker()
		return sema.ElaborationImport{
			Elaboration: cryptoChecker.Elaboration,
		}, nil
	}

	// Check if this contract has already been resolved
	subChecker, err := v.checkContract(importedLocation)
	if err != nil {
		return nil, err
	}

	return sema.ElaborationImport{
		Elaboration: subChecker.Elaboration,
	}, nil
}

// This mostly is a copy of the resolveLocation function from the linter/language server
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

		var resolvedLocation common.Location
		resovledAddrLocation := common.AddressLocation{
			Address: addressLocation.Address,
			Name:    identifier.Identifier,
		}

		// If the contract one of our staged contract updates, use the source location
		if stagedUpdate, ok := v.stagedContracts[resovledAddrLocation]; ok {
			resolvedLocation = stagedUpdate.SourceLocation
		} else {
			resolvedLocation = resovledAddrLocation
		}

		resolvedLocations[i] = runtime.ResolvedLocation{
			Location:    resolvedLocation,
			Identifiers: []runtime.Identifier{identifier},
		}
	}

	return resolvedLocations, nil
}

func (v *stagingValidatorImpl) resolveAccountAccess(checker *sema.Checker, memberLocation common.Location) bool {
	if checker == nil {
		return false
	}

	var memberAddress common.Address
	if memberAddressLocation, ok := memberLocation.(common.AddressLocation); ok {
		memberAddress = memberAddressLocation.Address
	} else if memberStringLocation, ok := memberLocation.(common.StringLocation); ok {
		found := false
		for _, stagedContract := range v.stagedContracts {
			if stagedContract.SourceLocation == memberStringLocation {
				memberAddress = stagedContract.DeployLocation.Address
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	var checkerAddress common.Address
	if checkerAddressLocation, ok := checker.Location.(common.AddressLocation); ok {
		checkerAddress = checkerAddressLocation.Address
	} else if checkerStringLocation, ok := checker.Location.(common.StringLocation); ok {
		found := false
		for _, stagedContract := range v.stagedContracts {
			if stagedContract.SourceLocation == checkerStringLocation {
				checkerAddress = stagedContract.DeployLocation.Address
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return memberAddress == checkerAddress
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
		StagedContractsMigrationOptions: migrations.StagedContractsMigrationOptions{
			ChainID: chainId,
		},
		Burner: migrations.BurnerContractChangeUpdate,
		EVM:    migrations.EVMContractChangeUpdate,
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

func (v *stagingValidatorImpl) addDependency(dependent common.Location, dependency common.Location) {
	// Create the dependent node if it does not exist
	if _, ok := v.graph[dependent]; !ok {
		v.graph[dependent] = make(node)
	}

	// Create the dependency node if it does not exist
	if _, ok := v.graph[dependency]; !ok {
		v.graph[dependency] = make(node)
	}

	// Add the dependency
	v.graph[dependent][dependency] = v.graph[dependency]
}

func (v *stagingValidatorImpl) forEachDependency(
	contract stagedContractUpdate,
	visitor func(dependency common.Location),
) {
	seen := make(map[common.Location]bool)
	var traverse func(location common.Location)
	traverse = func(location common.Location) {
		seen[location] = true

		for dep := range v.graph[location] {
			if !seen[dep] {
				visitor(dep)
				traverse(dep)
			}
		}
	}
	traverse(contract.SourceLocation)
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

// util to sort address locations
func sortAddressLocations(locations []common.AddressLocation) {
	slices.SortFunc(locations, func(a common.AddressLocation, b common.AddressLocation) int {
		addrCmp := bytes.Compare(a.Address.Bytes(), b.Address.Bytes())
		if addrCmp != 0 {
			return addrCmp
		}
		return strings.Compare(a.Name, b.Name)
	})
}
