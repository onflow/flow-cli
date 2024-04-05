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

	"github.com/onflow/flow-cli/internal/util"

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
)

type stagingValidator struct {
	flow  flowkit.Services
	state *flowkit.State

	// Cache for account contract names so we don't have to fetch them multiple times
	accountContractNames map[common.Address][]string
	// All resolved contract code
	contracts map[common.Location][]byte
	// Record errors related to missing staged dependencies, as these are reported separately
	missingDependencies []common.AddressLocation
	// Cache for contract elaborations which are reused during program checking & used for the update checker
	elaborations map[common.Location]*sema.Elaboration
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

var chainIdMap = map[string]flow.ChainID{
	"mainnet": flow.Mainnet,
	"testnet": flow.Testnet,
}

func newStagingValidator(flow flowkit.Services, state *flowkit.State) *stagingValidator {
	return &stagingValidator{
		flow:                 flow,
		state:                state,
		contracts:            make(map[common.Location][]byte),
		elaborations:         make(map[common.Location]*sema.Elaboration),
		accountContractNames: make(map[common.Address][]string),
	}
}

func (v *stagingValidator) ValidateContractUpdate(
	// Network location of the contract to be updated
	location common.AddressLocation,
	// Location of the source code, ensures that the error messages reference this instead of a network location
	sourceCodeLocation common.Location,
	// Code of the updated contract
	updatedCode []byte,
) error {
	// Resolve all system contract code & add to cache
	v.loadSystemContracts()

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

	// Store contract code for error pretty printing
	v.contracts[sourceCodeLocation] = updatedCode

	// Parse and check the contract code
	_, newProgramChecker, err := v.parseAndCheckContract(sourceCodeLocation)

	// Errors related to missing dependencies are separate from other errors
	// They may be handled differently by the caller, and it's parsing/checking
	// errors are not relevant/informative if these are present (they are expected)
	if len(v.missingDependencies) > 0 {
		return &missingDependenciesError{MissingContracts: v.missingDependencies}
	}

	if err != nil {
		return err
	}

	// Convert the new program checker to an interpreter program
	interpreterProgram := interpreter.ProgramFromChecker(newProgramChecker)

	// Check if contract code is valid according to Cadence V1 Update Checker
	validator := stdlib.NewCadenceV042ToV1ContractUpdateValidator(
		sourceCodeLocation,
		contractName,
		&accountContractNamesProviderImpl{
			resolverFunc: v.resolveAddressContractNames,
		},
		oldProgram,
		interpreterProgram,
		v.elaborations,
	)
	chainId, ok := chainIdMap[v.flow.Network().Name]
	if !ok {
		return fmt.Errorf("unsupported network: %s", v.flow.Network().Name)
	}
	validator.WithUserDefinedTypeChangeChecker(newUserDefinedTypeChangeCheckerFunc(chainId))

	err = validator.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (v *stagingValidator) parseAndCheckContract(
	location common.Location,
) (*ast.Program, *sema.Checker, error) {
	code := v.contracts[location]

	// Parse the contract code
	program, err := parser.ParseProgram(nil, code, parser.Config{})
	if err != nil {
		return nil, nil, err
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
			LocationHandler: v.resolveLocation,
			ImportHandler:   v.resolveImport,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	err = checker.Check()
	if err != nil {
		return nil, nil, err
	}

	return program, checker, nil
}

func (v *stagingValidator) getStagedContractCode(
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

	strValue, ok := optValue.Value.(cadence.String)
	if !ok {
		return nil, fmt.Errorf("invalid script return value type: %T", value)
	}

	v.contracts[location] = []byte(strValue)
	return v.contracts[location], nil
}

func (v *stagingValidator) resolveImport(checker *sema.Checker, importedLocation common.Location, _ ast.Range) (sema.Import, error) {
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
	elaboration, ok := v.elaborations[importedLocation]

	// If not resolved, parse and check the contract code
	if !ok {
		importedCode, err := v.getStagedContractCode(addrLocation)
		if err != nil {
			v.missingDependencies = append(v.missingDependencies, addrLocation)
			return nil, fmt.Errorf("failed to get staged contract code: %w", err)
		}
		v.contracts[addrLocation] = importedCode

		_, checker, err = v.parseAndCheckContract(addrLocation)
		if err != nil {
			return nil, fmt.Errorf("failed to parse and check contract code: %w", err)
		}

		v.elaborations[importedLocation] = checker.Elaboration
		elaboration = checker.Elaboration
	}

	return sema.ElaborationImport{
		Elaboration: elaboration,
	}, nil
}

func (v *stagingValidator) loadSystemContracts() {
	chainId, ok := chainIdMap[v.flow.Network().Name]
	if !ok {
		return
	}

	stagedSystemContracts := migrations.SystemContractChanges(chainId, migrations.SystemContractChangesOptions{
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

// This is a copy of the resolveLocation function from the linter/language server
func (v *stagingValidator) resolveLocation(
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

func (v *stagingValidator) resolveAddressContractNames(address common.Address) ([]string, error) {
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

// Helper for pretty printing errors
// While it is done by default in checker/parser errors, this has two purposes:
// 1. Add color to the error message
// 2. Use pretty printing on contract update errors which do not do this by default
func (v *stagingValidator) prettyPrintError(err error, location common.Location) string {
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

// TEMPORARY: this is not exported by flow-go and should be removed once it is
// This is for a quick fix to get the validator working
func newUserDefinedTypeChangeCheckerFunc(
	chainID flow.ChainID,
) func(oldTypeID common.TypeID, newTypeID common.TypeID) (checked, valid bool) {

	typeChangeRules := map[common.TypeID]common.TypeID{}

	compositeTypeRules := migrations.NewCompositeTypeConversionRules(chainID)
	for typeID, newStaticType := range compositeTypeRules {
		typeChangeRules[typeID] = newStaticType.ID()
	}

	interfaceTypeRules := migrations.NewInterfaceTypeConversionRules(chainID)
	for typeID, newStaticType := range interfaceTypeRules {
		typeChangeRules[typeID] = newStaticType.ID()
	}

	return func(oldTypeID common.TypeID, newTypeID common.TypeID) (checked, valid bool) {
		expectedNewTypeID, found := typeChangeRules[oldTypeID]
		if found {
			return true, expectedNewTypeID == newTypeID
		}
		return false, false
	}
}
