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

package cadence

import (
	"fmt"
	"path/filepath"
	"strings"

	"errors"

	"github.com/onflow/flow-cli/internal/util"

	cdclint "github.com/onflow/cadence-tools/lint"
	cdctests "github.com/onflow/cadence-tools/test/helpers"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	cdcerrors "github.com/onflow/cadence/runtime/errors"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/onflow/cadence/tools/analysis"
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
) (
	[]analysis.Diagnostic,
	error,
) {
	diagnostics := make([]analysis.Diagnostic, 0)
	location := common.StringLocation(filePath)

	code, err := l.state.ReadFile(filePath)
	if err != nil {
		return nil, err
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

// Create a new checker config with the given standard library
func (l *linter) newCheckerConfig(lib util.StandardLibrary) *sema.Config {
	return &sema.Config{
		BaseValueActivationHandler: func(_ common.Location) *sema.VariableActivation {
			return lib.BaseValueActivation
		},
		AccessCheckMode:            sema.AccessCheckModeStrict,
		PositionInfoEnabled:        true, // Must be enabled for linters
		ExtendedElaborationEnabled: true, // Must be enabled for linters
		ImportHandler:              l.handleImport,
		SuggestionsEnabled:         true, // Must be enabled to offer semantic suggestions
		AttachmentsEnabled:         true,
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
	case stdlib.CryptoCheckerLocation:
		cryptoChecker := stdlib.CryptoChecker()
		return sema.ElaborationImport{
			Elaboration: cryptoChecker.Elaboration,
		}, nil
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
	default:
		filepath, err := l.resolveImportFilepath(importedLocation, checker.Location)
		if err != nil {
			return nil, err
		}

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

			importedChecker, err = checker.SubChecker(importedProgram, importedLocation)
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
		// If the location is not a cadence file try getting the code by identifier
		if !strings.Contains(location.String(), ".cdc") {
			contract, err := l.state.Contracts().ByName(location.String())
			if err != nil {
				return "", err
			}

			return contract.Location, nil
		}

		// If the location is a cadence file, resolve relative to the parent location
		parentPath := ""
		if parentLocation != nil {
			parentPath = parentLocation.String()
		}

		resolvedPath := filepath.Join(filepath.Dir(parentPath), location.String())
		return resolvedPath, nil
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
	var syntaxErr *parser.SyntaxError
	var syntaxErrWithSuggestedReplacement *parser.SyntaxErrorWithSuggestedReplacement
	switch {
	case errors.As(err, &semanticErr):
		category = SemanticErrorCategory
	case errors.As(err, &syntaxErr):
		category = SyntaxErrorCategory
	case errors.As(err, &syntaxErrWithSuggestedReplacement):
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
