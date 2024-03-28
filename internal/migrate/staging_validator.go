package migrate

import (
	"context"
	"fmt"
	"strings"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/old_parser"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/cadence/runtime/pretty"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/onflow/contract-updater/lib/go/templates"
	internalCadence "github.com/onflow/flow-cli/internal/cadence"
	flowsdk "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go/cmd/util/ledger/migrations"
	"github.com/onflow/flow-go/model/flow"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
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

// Error when not all staged dependencies are found
type missingDependenciesError struct {
	contracts []common.AddressLocation
}

func (e *missingDependenciesError) Error() string {
	contractNames := make([]string, len(e.contracts))
	for i, location := range e.contracts {
		contractNames[i] = location.Name
	}
	return fmt.Sprintf("the following staged contract dependencies could not be found (have they been staged yet?): %v", contractNames)
}

var _ error = &missingDependenciesError{}

var chainIdMap = map[config.Network]flow.ChainID{
	config.MainnetNetwork: flow.Mainnet,
	config.TestnetNetwork: flow.Testnet,
}

func newStagingValidator(flow flowkit.Services, state *flowkit.State) *stagingValidator {
	return &stagingValidator{
		flow:         flow,
		state:        state,
		contracts:    make(map[common.Location][]byte),
		elaborations: make(map[common.Location]*sema.Elaboration),
	}
}

func (v *stagingValidator) ValidateContractUpdate(
	ctx context.Context,
	// Network location of the contract to be updated
	location common.AddressLocation,
	// Location of the source code, ensures that the error messages reference this instead of a network location
	sourceCodeLocation common.Location,
	// Code of the updated contract
	updatedCode []byte,
) error {
	// All dependent code will be resolved first
	// This helps isolate errors related to these missing dependencies
	// Instead of performing this resolution as a part of the checker
	// which would obfuscate the actual error

	// Resolve all system contract code
	v.resolveSystemContracts()

	// Get the account for the contract
	address := flowsdk.Address(location.Address)
	account, err := v.flow.GetAccount(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Get the target contract old code
	contractName := location.Name
	contractCode, ok := account.Contracts[contractName]
	if !ok {
		return fmt.Errorf("old contract code not found")
	}

	// Parse the old contract code
	oldProgram, err := old_parser.ParseProgram(nil, contractCode, old_parser.Config{})
	if err != nil {
		return fmt.Errorf("failed to parse old contract code: %w", err)
	}

	// Parse the new contract code
	newProgram, err := parser.ParseProgram(nil, updatedCode, parser.Config{})
	if err != nil {
		return fmt.Errorf("failed to parse new contract code: %w", err)
	}

	// Reset the missing contract dependencies
	v.missingDependencies = nil

	// Store contract code for error pretty printing
	v.contracts[sourceCodeLocation] = updatedCode

	// Parse and check the contract code
	_, _, err = v.parseAndCheckContract(ctx, sourceCodeLocation)

	// Errors related to missing dependencies are separate from other errors
	// These errors are non-fatal and are only used to inform the user
	// We do not care about checking/parsing errors here, since these are ultimately
	// expected/don't mean anything since the import resolutions are not complete
	if len(v.missingDependencies) > 0 {
		return &missingDependenciesError{contracts: v.missingDependencies}
	}

	if err != nil {
		return err
	}

	// Check if contract code is valid according to Cadence V1 Update Checker
	validator := stdlib.NewCadenceV042ToV1ContractUpdateValidator(
		sourceCodeLocation,
		contractName,
		&accountContractNamesProviderImpl{
			resolverFunc: v.resolveAddressContractNames,
		},
		oldProgram,
		newProgram,
		v.elaborations,
	)

	err = validator.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (v *stagingValidator) parseAndCheckContract(
	ctx context.Context,
	location common.Location,
) (*ast.Program, *sema.Checker, error) {
	code := v.contracts[location]

	// Parse the contract code
	program, err := parser.ParseProgram(nil, code, parser.Config{})
	if err != nil {
		return nil, nil, err
	}

	// Check the contract code
	// TODO: Do we need location handler? Error Message?
	checker, err := sema.NewChecker(
		program,
		location,
		nil,
		&sema.Config{
			AccessCheckMode:    sema.AccessCheckModeStrict,
			AttachmentsEnabled: true,
			BaseValueActivationHandler: func(_ common.Location) *sema.VariableActivation {
				return internalCadence.NewStandardLibrary().BaseValueActivation
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
	ctx context.Context,
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
	addrLocation, ok := importedLocation.(common.AddressLocation)
	if !ok {
		return nil, fmt.Errorf("expected address location")
	}

	// Check if this contract has already been resolved
	elaboration, ok := v.elaborations[importedLocation]

	// If not resolved, parse and check the contract code
	if !ok {
		importedCode, err := v.getStagedContractCode(context.Background(), addrLocation)
		if err != nil {
			v.missingDependencies = append(v.missingDependencies, addrLocation)
			return nil, fmt.Errorf("failed to get staged contract code: %w", err)
		}
		v.contracts[addrLocation] = importedCode

		_, checker, err = v.parseAndCheckContract(context.Background(), addrLocation)
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

func (v *stagingValidator) resolveSystemContracts() {
	chainId, ok := chainIdMap[v.flow.Network()]
	if !ok {
		return
	}

	// TODO: do we need burner/EVM config?
	stagedSystemContracts := migrations.SystemContractChanges(chainId, migrations.SystemContractChangesOptions{})
	for _, stagedSystemContract := range stagedSystemContracts {
		location := common.AddressLocation{
			Address: stagedSystemContract.Address,
			Name:    stagedSystemContract.Name,
		}

		v.contracts[location] = stagedSystemContract.Code
	}
}

// This is a copy of the resolveLocation function from the linter
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
	if v.accountContractNames == nil {
		v.accountContractNames = make(map[common.Address][]string)
	}

	// Check if the contract names are already cached
	if names, ok := v.accountContractNames[address]; ok {
		return names, nil
	}

	// Get the account for the contract
	account, err := v.flow.GetAccount(context.Background(), flowsdk.Address(address))
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Get the contract names
	contractNames := make([]string, 0, len(account.Contracts))
	for name := range account.Contracts {
		contractNames = append(contractNames, name)
	}

	// Cache the contract names
	v.accountContractNames[address] = contractNames

	return contractNames, nil
}

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

func (a *accountContractNamesProviderImpl) GetAccountContractNames(
	address common.Address,
) ([]string, error) {
	return a.resolverFunc(address)
}
