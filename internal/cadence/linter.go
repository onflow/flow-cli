/*
 * Flow CLI
 *
 * Copyright Flow Foundation
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

package cadence

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/onflow/flow-cli/internal/util"

	cdclint "github.com/onflow/cadence-tools/lint"
	cdctests "github.com/onflow/cadence-tools/test/helpers"
	"github.com/onflow/cadence/ast"
	"github.com/onflow/cadence/common"
	cdcerrors "github.com/onflow/cadence/errors"
	"github.com/onflow/cadence/parser"
	"github.com/onflow/cadence/sema"
	"github.com/onflow/cadence/stdlib"
	"github.com/onflow/cadence/tools/analysis"
	"github.com/onflow/flow-core-contracts/lib/go/contracts"
	flowGo "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"golang.org/x/exp/maps"
)

type linter struct {
	checkers              map[string]*sema.Checker
	state                 *flowkit.State
	checkerStandardConfig *sema.Config
	checkerScriptConfig   *sema.Config
}

type positionedError interface {
	error
	ast.HasPosition
}

// Error diagnostic categories
const (
	SemanticErrorCategory = "semantic-error"
	SyntaxErrorCategory   = "syntax-error"
	ErrorCategory         = "error"
)

var analyzers = maps.Values(cdclint.Analyzers)

func newLinter(state *flowkit.State) *linter {
	l := &linter{
		checkers: make(map[string]*sema.Checker),
		state:    state,
	}

	// Create checker configs for both standard and script
	// Scripts have a different stdlib than contracts and transactions
	l.checkerStandardConfig = l.newCheckerConfig(util.NewStandardLibrary())
	l.checkerScriptConfig = l.newCheckerConfig(util.NewScriptStandardLibrary())

	return l
}

func (l *linter) lintFile(
	filePath string,
) (diagnostics []analysis.Diagnostic, err error) {
	// Recover from panics in the Cadence checker
	defer func() {
		if r := recover(); r != nil {
			// Convert panic to error instead of crashing
			err = fmt.Errorf("internal error: %v", r)
		}
	}()

	diagnostics = make([]analysis.Diagnostic, 0)
	location := common.StringLocation(filePath)

	code, readErr := l.state.ReadFile(filePath)
	if readErr != nil {
		return nil, readErr
	}
	codeStr := string(code)

	// Parse program & convert any parsing errors to diagnostics
	program, parseProgramErr := parser.ParseProgram(nil, code, parser.Config{})
	if parseProgramErr != nil {
		var parserErr *parser.Error
		if !errors.As(parseProgramErr, &parserErr) {
			return nil, fmt.Errorf("could not process parsing error: %s", parseProgramErr)
		}

		checkerDiagnostics, err := getDiagnosticsFromParentError(parserErr, location, codeStr)
		if err != nil {
			return nil, err
		}

		diagnostics = append(diagnostics, checkerDiagnostics...)
	}

	// If the program is nil, nothing can be checked & analyzed so return early
	if program == nil {
		return diagnostics, nil
	}

	// Create checker based on program type
	checker, err := sema.NewChecker(
		program,
		location,
		nil,
		l.decideCheckerConfig(program),
	)
	if err != nil {
		return nil, err
	}

	// Check the program & convert any checking errors to diagnostics
	checkProgramErr := checker.Check()
	if checkProgramErr != nil {
		var checkerErr *sema.CheckerError
		if !errors.As(checkProgramErr, &checkerErr) {
			return nil, fmt.Errorf("could not process checking error: %s", checkProgramErr)
		}

		checkerDiagnostics, err := getDiagnosticsFromParentError(checkerErr, location, codeStr)
		if err != nil {
			return nil, err
		}

		diagnostics = append(diagnostics, checkerDiagnostics...)
	}

	// Run analysis on the program
	analysisProgram := analysis.Program{
		Program:  program,
		Checker:  checker,
		Location: checker.Location,
		Code:     []byte(code),
	}
	report := func(diagnostic analysis.Diagnostic) {
		diagnostics = append(diagnostics, diagnostic)
	}
	analysisProgram.Run(analyzers, report)

	return diagnostics, nil
}

// isContractName returns true if the location string is a contract name (not a file path)
func isContractName(locationString string) bool {
	return !strings.HasSuffix(locationString, ".cdc")
}

// resolveContractName attempts to resolve a location to a contract name
func (l *linter) resolveContractName(location common.StringLocation) string {
	locationString := location.String()

	// If it's already a contract name, return it
	if isContractName(locationString) {
		return locationString
	}

	// Otherwise, try to find the contract by file path
	if l.state == nil {
		return ""
	}

	contracts := l.state.Contracts()
	if contracts == nil {
		return ""
	}

	// Normalize the location path
	absLocation, err := filepath.Abs(locationString)
	if err != nil {
		absLocation = locationString
	}

	// Search for matching contract
	for _, contract := range *contracts {
		contractPath := contract.Location
		absContractPath, err := filepath.Abs(contractPath)
		if err != nil {
			absContractPath = contractPath
		}

		if absLocation == absContractPath {
			return contract.Name
		}
	}

	return ""
}

// checkAccountAccess determines if checker and member locations are on the same account
func (l *linter) checkAccountAccess(checker *sema.Checker, memberLocation common.Location) bool {
	// If both are AddressLocation, directly compare addresses
	if checkerAddr, ok := checker.Location.(common.AddressLocation); ok {
		if memberAddr, ok := memberLocation.(common.AddressLocation); ok {
			return checkerAddr.Address == memberAddr.Address
		}
	}

	// For StringLocations, resolve to contract names and check deployments
	checkerLocation, ok := checker.Location.(common.StringLocation)
	if !ok {
		return false
	}

	memberStringLocation, ok := memberLocation.(common.StringLocation)
	if !ok {
		return false
	}

	checkerContractName := l.resolveContractName(checkerLocation)
	if checkerContractName == "" {
		return false
	}

	memberContractName := l.resolveContractName(memberStringLocation)
	if memberContractName == "" {
		return false
	}

	if l.state == nil {
		return false
	}

	// Check across all networks if they're deployed to the same account
	networks := l.state.Networks()
	if networks == nil {
		return false
	}

	// Build contract name -> address mapping per network
	for _, network := range *networks {
		contractNameToAddress := make(map[string]flowGo.Address)

		// Add aliases first
		contracts := l.state.Contracts()
		if contracts != nil {
			for _, contract := range *contracts {
				if alias := contract.Aliases.ByNetwork(network.Name); alias != nil {
					contractNameToAddress[contract.Name] = alias.Address
				}
			}
		}

		// Add deployments (overwrites aliases, giving deployments priority)
		deployedContracts, err := l.state.DeploymentContractsByNetwork(network)
		if err == nil {
			for _, deployedContract := range deployedContracts {
				contract, err := l.state.Contracts().ByName(deployedContract.Name)
				if err == nil {
					address, err := l.state.ContractAddress(contract, network)
					if err == nil && address != nil {
						contractNameToAddress[deployedContract.Name] = *address
					}
				}
			}
		}

		// Check if both contracts exist at the same address on this network
		checkerAddress, checkerExists := contractNameToAddress[checkerContractName]
		memberAddress, memberExists := contractNameToAddress[memberContractName]
		if checkerExists && memberExists && checkerAddress == memberAddress {
			return true
		}
	}

	return false
}

// Create a new checker config with the given standard library
func (l *linter) newCheckerConfig(standardLibrary *util.StandardLibrary) *sema.Config {
	return &sema.Config{
		BaseValueActivationHandler: func(location common.Location) *sema.VariableActivation {
			return standardLibrary.BaseValueActivation
		},
		MemberAccountAccessHandler: func(checker *sema.Checker, memberLocation common.Location) bool {
			return l.checkAccountAccess(checker, memberLocation)
		},
		AccessCheckMode:            sema.AccessCheckModeNotSpecifiedUnrestricted,
		PositionInfoEnabled:        true, // Must be enabled for linters
		ExtendedElaborationEnabled: true, // Must be enabled for linters
		ImportHandler:              l.handleImport,
		SuggestionsEnabled:         true, // Must be enabled to offer semantic suggestions
	}
}

// Choose the checker config based on the assumed type of the program
func (l *linter) decideCheckerConfig(program *ast.Program) *sema.Config {
	if program.SoleTransactionDeclaration() != nil || program.SoleContractDeclaration() != nil {
		return l.checkerStandardConfig
	}

	return l.checkerScriptConfig
}

// Resolve any imports found in the program while checking
func (l *linter) handleImport(
	checker *sema.Checker,
	importedLocation common.Location,
	_ ast.Range,
) (
	sema.Import,
	error,
) {
	switch importedLocation {
	case stdlib.TestContractLocation:
		testChecker := stdlib.GetTestContractType().Checker
		return sema.ElaborationImport{
			Elaboration: testChecker.Elaboration,
		}, nil
	case cdctests.BlockchainHelpersLocation:
		helpersChecker := cdctests.BlockchainHelpersChecker()
		return sema.ElaborationImport{
			Elaboration: helpersChecker.Elaboration,
		}, nil
	case stdlib.CryptoContractLocation:
		cryptoChecker, ok := l.checkers[stdlib.CryptoContractLocation.String()]
		if !ok {
			cryptoCode := contracts.Crypto()
			cryptoProgram, err := parser.ParseProgram(nil, cryptoCode, parser.Config{})
			if err != nil {
				return nil, err
			}
			if cryptoProgram == nil {
				return nil, &sema.CheckerError{
					Errors: []error{fmt.Errorf("cannot parse Crypto contract")},
				}
			}

			cryptoChecker, err = sema.NewChecker(
				cryptoProgram,
				stdlib.CryptoContractLocation,
				nil,
				l.checkerStandardConfig,
			)
			if err != nil {
				return nil, err
			}

			err = cryptoChecker.Check()
			if err != nil {
				return nil, err
			}

			l.checkers[stdlib.CryptoContractLocation.String()] = cryptoChecker
		}

		return sema.ElaborationImport{
			Elaboration: cryptoChecker.Elaboration,
		}, nil
	default:
		// Normalize relative path imports to absolute paths
		if util.IsPathLocation(importedLocation) {
			importedLocation = util.NormalizePathLocation(checker.Location, importedLocation)
		}

		filepath, err := l.resolveImportFilepath(importedLocation, checker.Location)
		if err != nil {
			return nil, err
		}

		fileLocation := common.StringLocation(filepath)

		importedChecker, ok := l.checkers[filepath]
		if !ok {
			code, err := l.state.ReadFile(filepath)
			if err != nil {
				return nil, err
			}

			importedProgram, err := parser.ParseProgram(nil, code, parser.Config{})

			if err != nil {
				return nil, err
			}
			if importedProgram == nil {
				return nil, &sema.CheckerError{
					Errors: []error{fmt.Errorf("cannot import %s", importedLocation)},
				}
			}

			importedChecker, err = checker.SubChecker(importedProgram, fileLocation)
			if err != nil {
				return nil, err
			}

			l.checkers[filepath] = importedChecker
			err = importedChecker.Check()
			if err != nil {
				return nil, err
			}
		}

		return sema.ElaborationImport{
			Elaboration: importedChecker.Elaboration,
		}, nil
	}
}

func (l *linter) resolveImportFilepath(
	location common.Location,
	parentLocation common.Location,
) (
	string,
	error,
) {
	switch location := location.(type) {
	case common.StringLocation:
		// Resolve by contract name from flowkit config
		if !strings.Contains(location.String(), ".cdc") {
			contract, err := l.state.Contracts().ByName(location.String())
			if err != nil {
				return "", err
			}

			return contract.Location, nil
		}

		return location.String(), nil
	default:
		return "", fmt.Errorf("unsupported location: %T", location)
	}
}

// helpers

func getDiagnosticsFromParentError(err cdcerrors.ParentError, location common.Location, code string) ([]analysis.Diagnostic, error) {
	diagnostics := make([]analysis.Diagnostic, 0)

	for _, childErr := range err.ChildErrors() {
		var positionedErr positionedError
		if !errors.As(childErr, &positionedErr) {
			return nil, fmt.Errorf("could not process error: %s", childErr)
		}

		diagnostic := convertPositionedErrorToDiagnostic(positionedErr, location, code)
		if diagnostic == nil {
			continue
		}

		diagnostics = append(diagnostics, *diagnostic)
	}

	return diagnostics, nil
}

func convertPositionedErrorToDiagnostic(
	err positionedError,
	location common.Location,
	code string,
) *analysis.Diagnostic {
	message := err.Error()
	startPosition := err.StartPosition()
	endPosition := err.EndPosition(nil)

	var secondaryMessage string
	var secondaryErr cdcerrors.SecondaryError
	if errors.As(err, &secondaryErr) {
		secondaryMessage = secondaryErr.SecondaryError()
	}

	var category string
	var semanticErr sema.SemanticError
	var parseError parser.ParseError
	switch {
	case errors.As(err, &semanticErr):
		category = SemanticErrorCategory
	case errors.As(err, &parseError):
		category = SyntaxErrorCategory
	default:
		category = ErrorCategory
	}

	var suggestedFixes []cdcerrors.SuggestedFix[ast.TextEdit]
	var errWithFixes cdcerrors.HasSuggestedFixes[ast.TextEdit]
	if errors.As(err, &errWithFixes) {
		suggestedFixes = errWithFixes.SuggestFixes(code)
	}

	diagnostic := analysis.Diagnostic{
		Location:         location,
		Category:         category,
		Message:          message,
		SecondaryMessage: secondaryMessage,
		Range: ast.Range{
			StartPos: startPosition,
			EndPos:   endPosition,
		},
		SuggestedFixes: suggestedFixes,
	}

	return &diagnostic
}

func isErrorDiagnostic(diagnostic analysis.Diagnostic) bool {
	return diagnostic.Category == ErrorCategory || diagnostic.Category == SemanticErrorCategory || diagnostic.Category == SyntaxErrorCategory
}
