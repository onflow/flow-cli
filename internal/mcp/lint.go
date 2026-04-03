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

package mcp

import (
	"errors"
	"fmt"
	"strings"

	cdclint "github.com/onflow/cadence-tools/lint"
	cdctests "github.com/onflow/cadence-tools/test/helpers"
	"github.com/onflow/cadence/ast"
	"github.com/onflow/cadence/common"
	cdcerrors "github.com/onflow/cadence/errors"
	"github.com/onflow/cadence/parser"
	"github.com/onflow/cadence/sema"
	"github.com/onflow/cadence/stdlib"
	"github.com/onflow/cadence/tools/analysis"
	"golang.org/x/exp/maps"

	"github.com/onflow/flow-cli/internal/util"
)

// lintCode runs all registered cadence-tools lint analyzers on the given code.
// Returns analysis diagnostics (not LSP protocol diagnostics).
func lintCode(code string) ([]analysis.Diagnostic, error) {
	location := common.StringLocation("code.cdc")
	codeBytes := []byte(code)

	// Parse
	program, parseErr := parser.ParseProgram(nil, codeBytes, parser.Config{})
	if parseErr != nil {
		var parentErr cdcerrors.ParentError
		if errors.As(parseErr, &parentErr) {
			return collectPositionedErrors(parentErr, location), nil
		}
		return nil, fmt.Errorf("parse error: %w", parseErr)
	}
	if program == nil {
		return nil, nil
	}

	// Choose standard library based on program type
	var stdLib *util.StandardLibrary
	if program.SoleTransactionDeclaration() != nil || program.SoleContractDeclaration() != nil {
		stdLib = util.NewStandardLibrary()
	} else {
		stdLib = util.NewScriptStandardLibrary()
	}

	checkerConfig := &sema.Config{
		BaseValueActivationHandler: func(_ common.Location) *sema.VariableActivation {
			return stdLib.BaseValueActivation
		},
		AccessCheckMode:            sema.AccessCheckModeNotSpecifiedUnrestricted,
		PositionInfoEnabled:        true,
		ExtendedElaborationEnabled: true,
		ImportHandler:              lintImportHandler,
	}

	// Type check
	checker, err := sema.NewChecker(program, location, nil, checkerConfig)
	if err != nil {
		return nil, fmt.Errorf("checker creation error: %w", err)
	}

	var diagnostics []analysis.Diagnostic

	checkErr := checker.Check()
	if checkErr != nil {
		var checkerErr *sema.CheckerError
		if errors.As(checkErr, &checkerErr) {
			diagnostics = append(diagnostics, collectPositionedErrors(checkerErr, location)...)
		}
	}

	// Run lint analyzers
	analyzers := maps.Values(cdclint.Analyzers)
	analysisProgram := analysis.Program{
		Program:  program,
		Checker:  checker,
		Location: location,
		Code:     codeBytes,
	}
	analysisProgram.Run(analyzers, func(d analysis.Diagnostic) {
		diagnostics = append(diagnostics, d)
	})

	return diagnostics, nil
}

// lintImportHandler resolves standard library imports for lint analysis.
func lintImportHandler(
	checker *sema.Checker,
	importedLocation common.Location,
	_ ast.Range,
) (sema.Import, error) {
	switch importedLocation {
	case stdlib.TestContractLocation:
		return sema.ElaborationImport{
			Elaboration: stdlib.GetTestContractType().Checker.Elaboration,
		}, nil
	case cdctests.BlockchainHelpersLocation:
		return sema.ElaborationImport{
			Elaboration: cdctests.BlockchainHelpersChecker().Elaboration,
		}, nil
	default:
		return nil, fmt.Errorf("cannot resolve import: %s", importedLocation)
	}
}

type positionedError interface {
	error
	ast.HasPosition
}

// collectPositionedErrors extracts positioned errors from a parent error.
func collectPositionedErrors(err cdcerrors.ParentError, location common.Location) []analysis.Diagnostic {
	var diagnostics []analysis.Diagnostic
	for _, childErr := range err.ChildErrors() {
		var posErr positionedError
		if !errors.As(childErr, &posErr) {
			continue
		}
		diagnostics = append(diagnostics, analysis.Diagnostic{
			Location: location,
			Category: "error",
			Message:  posErr.Error(),
			Range: ast.Range{
				StartPos: posErr.StartPosition(),
				EndPos:   posErr.EndPosition(nil),
			},
		})
	}
	return diagnostics
}

// formatLintDiagnostics formats analysis diagnostics as human-readable text.
func formatLintDiagnostics(diagnostics []analysis.Diagnostic) string {
	if len(diagnostics) == 0 {
		return "Lint passed — no issues found."
	}

	var b strings.Builder
	errors := 0
	warnings := 0
	for _, d := range diagnostics {
		severity := "warning"
		if d.Category == "error" || d.Category == "semantic-error" || d.Category == "syntax-error" {
			severity = "error"
			errors++
		} else {
			warnings++
		}
		fmt.Fprintf(&b, "[%s] line %d:%d (%s): %s\n",
			severity,
			d.Range.StartPos.Line,
			d.Range.StartPos.Column,
			d.Category,
			d.Message,
		)
	}
	fmt.Fprintf(&b, "\n%d error(s), %d warning(s)\n", errors, warnings)
	return b.String()
}
