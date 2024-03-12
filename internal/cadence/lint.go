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
	"slices"
	"strings"
	"sync"

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

type lintFlagsCollection struct{}
type lintResult struct {
	FilePath    string
	Diagnostics []analysis.Diagnostic
}

type lintResults struct {
	Results  []lintResult
	exitCode int
}

var _ command.ResultWithExitCode = &lintResults{}

var lintFlags = lintFlagsCollection{}

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
	Flags: &lintFlags,
	RunS:  lint,
}

const (
	SemanticErrorCategory = "semantic-error"
	SyntaxErrorCategory   = "syntax-error"
	ErrorCategory         = "error"
)

const (
	ErrorSeverity   Severity = "error"
	WarningSeverity Severity = "warning"
)

type Severity string

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
	mutex := &sync.Mutex{}
	linterDiagnostics := make(map[common.Location][]analysis.Diagnostic)
	errorDiagnostics := make(map[common.Location][]analysis.Diagnostic)

	processParserCheckerError := func(err error, location common.Location) error {
		unwrapped := err
		diagnostics, err := maybeProcessConvertableError(unwrapped, location)
		if err != nil {
			return err
		}

		mutex.Lock()
		errorDiagnostics[location] = append(errorDiagnostics[location], diagnostics...)
		mutex.Unlock()
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

	// Only run linter on explicitly provided files
	wg := &sync.WaitGroup{}
	analyzers := maps.Values(cdcLint.Analyzers)
	for _, location := range locations {
		program := programs[location]

		wg.Add(1)
		go func() {
			defer wg.Done()
			program.Run(analyzers, func(d analysis.Diagnostic) {
				location := program.Location
				mutex.Lock()
				linterDiagnostics[location] = append(linterDiagnostics[location], d)
				mutex.Unlock()
			})
		}()
	}
	wg.Wait()

	results := make([]lintResult, 0)
	exitCode := 0
	for _, location := range locations {
		// Remove any diagnostics that are expected from the Cadence V1 Analyzer
		// And combine the linter diagnostics with the error diagnostics
		diagnostics := removeExpectedCadenceV1Errors(linterDiagnostics[location], errorDiagnostics[location])
		diagnostics = append(diagnostics, linterDiagnostics[location]...)
		sortDiagnostics(diagnostics)

		results = append(results, lintResult{
			FilePath:    location.String(),
			Diagnostics: diagnostics,
		})

		// Set the exitCode to 1 if any of the diagnostics are error-level
		for _, diagnostic := range diagnostics {
			severity := getDiagnosticSeverity(diagnostic)
			if severity == ErrorSeverity {
				exitCode = 1
				break
			}
		}
	}

	return &lintResults{
		Results:  results,
		exitCode: exitCode,
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

// This function is used to remove expected errors from the diagnostics
// It will remove all diagnostics that overlap wholly with those provided
// by the Cadence V1 Analyzer
//
// It's imperfect, but it reduces redundancy/confusion.  The implementation
// is somewhat inefficient and using an interval tree would be better.
func removeExpectedCadenceV1Errors(
	linterDiagnostics []analysis.Diagnostic,
	errorDiagnostics []analysis.Diagnostic,
) []analysis.Diagnostic {
	cadenceV1Ranges := make([]ast.Range, 0)
	for _, diagnostic := range linterDiagnostics {
		if diagnostic.Category == cdcLint.CadenceV1Category {
			cadenceV1Ranges = append(cadenceV1Ranges, diagnostic.Range)
		}
	}

	// remove any diagnostics that overlap with the Cadence V1 ranges
	filteredDiagnostics := make([]analysis.Diagnostic, 0)
	for _, diagnostic := range errorDiagnostics {
		if diagnostic.Category == cdcLint.CadenceV1Category {
			filteredDiagnostics = append(filteredDiagnostics, diagnostic)
			continue
		}

		overlap := false
		for _, cadenceV1Range := range cadenceV1Ranges {
			if ast.RangeContains(nil, cadenceV1Range, diagnostic.Range) {
				overlap = true
				break
			}
		}

		if !overlap {
			filteredDiagnostics = append(filteredDiagnostics, diagnostic)
		}
	}

	return filteredDiagnostics
}

func getDiagnosticSeverity(
	diagnostic analysis.Diagnostic,
) Severity {
	switch diagnostic.Category {
	case SemanticErrorCategory:
		return ErrorSeverity
	case SyntaxErrorCategory:
		return ErrorSeverity
	case ErrorCategory:
		return ErrorSeverity
	case cdcLint.CadenceV1Category:
		return ErrorSeverity
	default:
		return WarningSeverity
	}
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
		SuggestedFixes:   suggestedFixes,
		Range: ast.Range{
			StartPos: startPosition,
			EndPos:   endPosition,
		},
	}

	return &diagnostic
}

// Sort diagnostics in order of precedence: start pos -> category -> message
// Ensures that diagnostics are always in a consistent order
func sortDiagnostics(
	diagnostics []analysis.Diagnostic,
) {
	slices.SortFunc(diagnostics, func(a analysis.Diagnostic, b analysis.Diagnostic) int {
		if r := a.Range.StartPos.Offset - b.Range.StartPos.Offset; r != 0 {
			return r
		}

		if r := strings.Compare(a.Category, b.Category); r != 0 {
			return r
		}

		return strings.Compare(a.Message, b.Message)
	})
}

func renderDiagnostic(diagnostic analysis.Diagnostic) string {
	locationColor := lipgloss.Color("#A9B7C6")
	categoryColor := lipgloss.Color("#FF3E3E")
	if getDiagnosticSeverity(diagnostic) == WarningSeverity {
		categoryColor = lipgloss.Color("#FFA500")
	}

	startPos := diagnostic.Range.StartPos
	locationText := fmt.Sprintf("%s:%d:%d:", diagnostic.Location.String(), startPos.Line, startPos.Column)
	categoryText := fmt.Sprintf("%s:", diagnostic.Category)

	return fmt.Sprintf("%s %s %s",
		lipgloss.NewStyle().Foreground(locationColor).Render(locationText),
		lipgloss.NewStyle().Foreground(categoryColor).Render(categoryText),
		diagnostic.Message,
	)
}

func (r *lintResults) String() string {
	var sb strings.Builder
	var numProblems int

	for _, result := range r.Results {
		if len(result.Diagnostics) == 0 {
			continue
		}

		numProblems += len(result.Diagnostics)

		for _, diagnostic := range result.Diagnostics {
			sb.WriteString(fmt.Sprintf("%s\n\n", renderDiagnostic(diagnostic)))
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

func (r *lintResults) ExitCode() int {
	return r.exitCode
}
