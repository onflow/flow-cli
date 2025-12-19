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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/logrusorgru/aurora/v4"
	"github.com/spf13/cobra"

	"github.com/onflow/cadence/ast"
	"github.com/onflow/cadence/common"
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
		Use:   "lint [files...]",
		Short: "Lint Cadence code to identify potential issues or errors",
		Example: `# Lint all .cdc files in the project
flow cadence lint

# Lint specific files
flow cadence lint file1.cdc file2.cdc`,
		Args: cobra.ArbitraryArgs,
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
	var filePaths []string
	if len(args) == 0 {
		var err error
		filePaths, err = findAllCadenceFiles(".")
		if err != nil {
			return nil, fmt.Errorf("error finding Cadence files: %w", err)
		}
		if len(filePaths) == 0 {
			return nil, fmt.Errorf("no .cdc files found in the project")
		}
	} else {
		filePaths = args
	}

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
			// If there's an internal error (like a panic), convert it to a diagnostic
			// and continue processing other files
			diagnostics = []analysis.Diagnostic{
				{
					Location: common.StringLocation(location),
					Category: ErrorCategory,
					Message:  err.Error(),
					Range:    ast.Range{},
				},
			}
			exitCode = 1
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

func (r *lintResult) JSON() any {
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

func findAllCadenceFiles(baseDir string) ([]string, error) {
	var filenames []string
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".cdc") {
			return nil
		}

		filenames = append(filenames, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return filenames, nil
}
