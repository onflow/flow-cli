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

	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"

	"github.com/charmbracelet/lipgloss"

	cdcLint "github.com/onflow/cadence-tools/lint"
	"github.com/onflow/cadence/runtime/ast"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/errors"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/tools/analysis"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
)

type lintFlagsCollection struct {
	Fix bool `default:"false" flag:"fix" info:"Fix linting errors"`
}
type lintResult struct {
	FilePath    string
	Diagnostics []analysis.Diagnostic
}

type lintResults struct {
	Results []lintResult
}

var lintFlags = lintFlagsCollection{}
var status = 0

type convertibleError interface {
	error
	ast.HasPosition
}

var lintCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "lint [files]",
		Short:   "Lint Cadence code",
		Example: "flow cadence lint **/*.cdc",
		Args:    cobra.MinimumNArgs(1),
	},
	Flags:  &lintFlags,
	RunS:   lint,
	Status: &status,
}

func lint(
	args []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	locations := make([]common.Location, 0)

	for _, filePath := range args {
		locations = append(locations, common.StringLocation(filePath))
	}

	results, err := lintFiles(locations, state)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func lintFiles(
	locations []common.Location,
	state *flowkit.State,
) (
	*lintResults,
	error,
) {
	diagnostics := make(map[common.Location][]analysis.Diagnostic)

	processParserCheckerError := func(err error, location common.Location) error {
		unwrapped := err
		errorDiagnostics, err := maybeProcessConvertableError(unwrapped, location)
		if err != nil {
			return err
		}

		diagnostics[location] = append(diagnostics[location], errorDiagnostics...)
		return nil
	}

	config := &analysis.Config{
		Mode: analysis.NeedTypes | analysis.NeedExtendedElaboration | analysis.NeedPositionInfo | analysis.NeedSyntax,
		HandleParserError: func(err analysis.ParsingCheckingError, program *ast.Program) error {
			if program == nil {
				return err
			}
			return processParserCheckerError(err.Unwrap(), err.ImportLocation())
		},
		HandleCheckerError: func(err analysis.ParsingCheckingError, checker *sema.Checker) error {
			if checker == nil {
				return err
			}
			return processParserCheckerError(err.Unwrap(), err.ImportLocation())
		},
		ResolveCode: func(location common.Location, importingLocation common.Location, importRange ast.Range) ([]byte, error) {
			if _, ok := location.(common.StringLocation); !ok {
				return nil, fmt.Errorf("unsupported location type: %T", location)
			}

			originalLocation := location.String()
			resolvedLocation := location

			// If the location is not a .cdc file, it's a contract name
			// Must resolve the contract name to a file location
			if !strings.HasSuffix(originalLocation, ".cdc") {
				contract, err := state.Contracts().ByName(originalLocation)
				if err != nil {
					return nil, fmt.Errorf("unable to resolve contract code: %w", err)
				}

				resolvedLocation = common.StringLocation(contract.Location)
			}

			return state.ReadFile(resolvedLocation.String())
		},
	}

	programs, err := analysis.Load(config, locations...)
	if err != nil {
		return nil, err
	}

	for _, program := range programs {
		analyzers := maps.Values(cdcLint.Analyzers)
		program.Run(analyzers, func(d analysis.Diagnostic) {
			location := program.Location
			diagnostics[location] = append(diagnostics[location], d)
		})
	}

	results := make([]lintResult, 0)
	for location, diagnosticList := range diagnostics {
		results = append(results, lintResult{
			FilePath:    location.String(),
			Diagnostics: diagnosticList,
		})
		if len(diagnosticList) > 0 {
			status = 1
		}
	}

	return &lintResults{
		Results: results,
	}, nil
}

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

	suggestedFixes := make([]analysis.SuggestedFix, 0)

	category := "error"
	if _, ok := err.(sema.SemanticError); ok {
		category = "semantic-error"
	} else if _, ok := err.(*parser.SyntaxError); ok {
		category = "syntax-error"
	}

	diagnostic := analysis.Diagnostic{
		Location:         location,
		Category:         category,
		Message:          message,
		SecondaryMessage: secondaryMessage,
		SuggestedFixes:   suggestedFixes,
		Range: ast.Range{
			StartPos: startPosition,
			EndPos:   endPosition,
		},
	}

	return &diagnostic
}

func (r *lintResults) String() string {
	var sb strings.Builder
	var numProblems int

	for _, result := range r.Results {
		if len(result.Diagnostics) == 0 {
			continue
		}

		numProblems += len(result.Diagnostics)

		relPath, err := filepath.Rel(".", result.FilePath)
		if err != nil {
			relPath = result.FilePath
		}

		for _, diagnostic := range result.Diagnostics {
			startPos := diagnostic.Range.StartPos
			sb.WriteString(fmt.Sprintf("%s:%d:%d: ", relPath, startPos.Line, startPos.Column))
			sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF3E3E")).Render(fmt.Sprintf("%s: %s", diagnostic.Category, diagnostic.Message)))
			sb.WriteString("\n\n")
		}
	}

	if numProblems == 0 {
		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#1FEE8A")).Render("✅ No problems found"))
	} else {
		sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF3E3E")).Render(fmt.Sprintf("❌ Found %d problem(s)", numProblems)))
	}

	return sb.String()
}

func (r *lintResults) JSON() interface{} {
	return r
}

func (r *lintResults) Oneliner() string {
	problems := 0
	for _, result := range r.Results {
		problems += len(result.Diagnostics)
	}

	if problems == 0 {
		return "No problems found"
	}

	return fmt.Sprintf("Found %d problem(s)", problems)
}
