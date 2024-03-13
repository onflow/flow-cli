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

	cadenceLint "github.com/onflow/cadence-tools/lint"
	cdcTests "github.com/onflow/cadence-tools/test/helpers"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/errors"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/runtime/stdlib"
	"github.com/onflow/cadence/tools/analysis"
	"github.com/onflow/flowkit/v2"
	"golang.org/x/exp/maps"
)

type linter struct {
	checkers              map[common.Location]*sema.Checker
	state                 *flowkit.State
	checkerStandardConfig *sema.Config
	checkerScriptConfig   *sema.Config
}

type convertibleError interface {
	error
	ast.HasPosition
}

// Error diagnostic categories
const (
	SemanticErrorCategory = "semantic-error"
	SyntaxErrorCategory   = "syntax-error"
	ErrorCategory         = "error"
)

var analyzers = maps.Values(cadenceLint.Analyzers)

func newLinter(state *flowkit.State) *linter {
	l := &linter{
		checkers: make(map[common.Location]*sema.Checker),
		state:    state,
	}

	l.checkerStandardConfig = l.newCheckerConfig(newStandardLibrary())
	l.checkerScriptConfig = l.newCheckerConfig(newScriptStandardLibrary())

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

	program, parseErr := parser.ParseProgram(nil, code, parser.Config{})
	if parseErr != nil {
		errorDiagnostics, err := maybeProcessConvertableError(parseErr, location)
		if err != nil {
			return nil, err
		}

		diagnostics = append(diagnostics, errorDiagnostics...)
	}

	// If the program is nil, nothing can be checked & analyzed so return early
	if program == nil {
		return diagnostics, nil
	}

	checker, err := sema.NewChecker(
		program,
		location,
		nil,
		l.decideCheckerConfig(program),
	)
	if err != nil {
		return nil, err
	}

	checkError := checker.Check()
	if checkError != nil {
		errorDiagnostics, err := maybeProcessConvertableError(checkError, location)
		if err != nil {
			return nil, err
		}

		diagnostics = append(diagnostics, errorDiagnostics...)
	}

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

func (l *linter) newCheckerConfig(lib standardLibrary) *sema.Config {
	return &sema.Config{
		BaseValueActivationHandler: func(_ common.Location) *sema.VariableActivation {
			return lib.baseValueActivation
		},
		AccessCheckMode:            sema.AccessCheckModeStrict,
		PositionInfoEnabled:        true,
		ExtendedElaborationEnabled: true,
		ImportHandler:              l.handleImport,
		AttachmentsEnabled:         true,
	}
}

func (l *linter) decideCheckerConfig(program *ast.Program) *sema.Config {
	if program.SoleTransactionDeclaration() != nil || program.SoleContractDeclaration() != nil {
		return l.checkerStandardConfig
	}

	return l.checkerScriptConfig
}

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
	case cdcTests.BlockchainHelpersLocation:
		helpersChecker := cdcTests.BlockchainHelpersChecker()
		return sema.ElaborationImport{
			Elaboration: helpersChecker.Elaboration,
		}, nil
	default:
		importedChecker, ok := l.checkers[importedLocation]
		if !ok {
			filepath, err := l.resolveImportFilepath(importedLocation, checker.Location)
			if err != nil {
				return nil, err
			}

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
			l.checkers[importedLocation] = importedChecker
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

func maybeProcessConvertableError(
	err error,
	location common.Location,
) (
	[]analysis.Diagnostic,
	error,
) {
	diagnostics := make([]analysis.Diagnostic, 0)
	if parentErr, ok := err.(errors.ParentError); ok {
		checkerDiagnostics, err := getDiagnosticsForParentError(parentErr, location)
		if err != nil {
			return nil, err
		}

		diagnostics = append(diagnostics, checkerDiagnostics...)
	}
	return diagnostics, nil
}

func getDiagnosticsForParentError(err errors.ParentError, location common.Location) ([]analysis.Diagnostic, error) {
	diagnostics := make([]analysis.Diagnostic, 0)

	for _, childErr := range err.ChildErrors() {
		convertibleErr, ok := childErr.(convertibleError)
		if !ok {
			return nil, fmt.Errorf("unable to convert non-convertable error to diagnostic: %T", childErr)
		}
		diagnostic := convertError(convertibleErr, location)
		if diagnostic == nil {
			continue
		}

		diagnostics = append(diagnostics, *diagnostic)
	}

	return diagnostics, nil
}

func convertError(
	err convertibleError,
	location common.Location,
) *analysis.Diagnostic {
	startPosition := err.StartPosition()
	endPosition := err.EndPosition(nil)

	var message string
	var secondaryMessage string

	message = err.Error()
	if secondaryError, ok := err.(errors.SecondaryError); ok {
		secondaryMessage = secondaryError.SecondaryError()
	}

	category := ErrorCategory
	if _, ok := err.(sema.SemanticError); ok {
		category = SemanticErrorCategory
	} else if _, ok := err.(*parser.SyntaxError); ok {
		category = SyntaxErrorCategory
	} else if _, ok := err.(*parser.SyntaxErrorWithSuggestedReplacement); ok {
		category = SyntaxErrorCategory
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
	}

	return &diagnostic
}

func isErrorDiagnostic(diagnostic analysis.Diagnostic) bool {
	return diagnostic.Category == ErrorCategory || diagnostic.Category == SemanticErrorCategory || diagnostic.Category == SyntaxErrorCategory
}
