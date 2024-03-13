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

	"github.com/logrusorgru/aurora/v4"
	"github.com/spf13/cobra"

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

type severity string

const (
	errorSeverity   severity = "error"
	warningSeverity severity = "warning"
)

func lint(
	args []string,
	globalFlags command.GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	filePaths := args
	results, err := lintFiles(state, filePaths...)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func lintFiles(
	state *flowkit.State,
	filePaths ...string,
) (
	*lintResults,
	error,
) {
	l := newLinter(state)
	results := make([]lintResult, 0)
	exitCode := 0

	// Only run linter on explicitly provided files
	for _, location := range filePaths {
		diagnostics, err := l.lintFile(location)
		if err != nil {
			return nil, err
		}

		sortDiagnostics(diagnostics)
		results = append(results, lintResult{
			FilePath:    location,
			Diagnostics: diagnostics,
		})

		// Set the exitCode to 1 if any of the diagnostics are error-level
		// In the future, this may be configurable
		for _, diagnostic := range diagnostics {
			severity := getDiagnosticSeverity(diagnostic)
			if severity == errorSeverity {
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

func getDiagnosticSeverity(
	diagnostic analysis.Diagnostic,
) severity {
	if isErrorDiagnostic(diagnostic) {
		return errorSeverity
	}
	return warningSeverity
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
	categoryColor := aurora.RedFg
	if getDiagnosticSeverity(diagnostic) == warningSeverity {
		categoryColor = aurora.YellowFg
	}

	startPos := diagnostic.Range.StartPos
	locationText := fmt.Sprintf("%s:%d:%d:", diagnostic.Location.String(), startPos.Line, startPos.Column)
	categoryText := fmt.Sprintf("%s:", diagnostic.Category)

	return fmt.Sprintf("%s %s %s",
		aurora.Gray(12, locationText).String(),
		aurora.Colorize(categoryText, categoryColor).String(),
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
		sb.WriteString("No problems found")
	} else {
		sb.WriteString(aurora.Red(fmt.Sprintf("Found %d problem(s)", numProblems)).String())
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
