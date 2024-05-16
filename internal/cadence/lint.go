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
	"strings"

	"golang.org/x/exp/slices"

	"github.com/logrusorgru/aurora/v4"
	"github.com/spf13/cobra"

	"github.com/onflow/cadence/tools/analysis"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type lintFlagsCollection struct{}

type fileResult struct {
	FilePath    string
	Diagnostics []analysis.Diagnostic
}

type lintResult struct {
	Results  []fileResult
	exitCode int
}

var _ command.ResultWithExitCode = &lintResult{}

var lintFlags = lintFlagsCollection{}

var lintCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "lint [files]",
		Short:   "Lint Cadence code to identify potential issues or errors",
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
	result, err := lintFiles(state, filePaths...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func lintFiles(
	state *flowkit.State,
	filePaths ...string,
) (
	*lintResult,
	error,
) {
	l := newLinter(state)
	results := make([]fileResult, 0)
	exitCode := 0

	for _, location := range filePaths {
		diagnostics, err := l.lintFile(location)
		if err != nil {
			return nil, err
		}

		// Sort for consistent output
		sortDiagnostics(diagnostics)
		results = append(results, fileResult{
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

	return &lintResult{
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

func (r *lintResult) countProblems() (int, int) {
	numErrors := 0
	numWarnings := 0
	for _, result := range r.Results {
		for _, diagnostic := range result.Diagnostics {
			if isErrorDiagnostic(diagnostic) {
				numErrors++
			} else {
				numWarnings++
			}
		}
	}
	return numErrors, numWarnings
}

func (r *lintResult) String() string {
	var sb strings.Builder

	for _, result := range r.Results {
		for _, diagnostic := range result.Diagnostics {
			sb.WriteString(fmt.Sprintf("%s\n\n", renderDiagnostic(diagnostic)))
		}
	}

	var color aurora.Color
	numErrors, numWarnings := r.countProblems()
	if numErrors > 0 {
		color = aurora.RedFg
	} else if numWarnings > 0 {
		color = aurora.YellowFg
	}

	total := numErrors + numWarnings
	if total > 0 {
		sb.WriteString(aurora.Colorize(fmt.Sprintf(
			"%d %s (%d %s, %d %s)",
			total,
			util.Pluralize("problem", total),
			numErrors,
			util.Pluralize("error", numErrors),
			numWarnings,
			util.Pluralize("warning", numWarnings),
		), color).String())
	} else {
		sb.WriteString(aurora.Green("Lint passed").String())
	}

	return sb.String()
}

func (r *lintResult) JSON() interface{} {
	return r
}

func (r *lintResult) Oneliner() string {
	numErrors, numWarnings := r.countProblems()
	total := numErrors + numWarnings

	if total > 0 {
		return fmt.Sprintf("%d %s (%d %s, %d %s)", total, util.Pluralize("problem", total), numErrors, util.Pluralize("error", numErrors), numWarnings, util.Pluralize("warning", numWarnings))
	}
	return "Lint passed"
}

func (r *lintResult) ExitCode() int {
	return r.exitCode
}
