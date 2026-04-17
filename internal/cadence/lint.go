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

type lintFlagsCollection struct {
	WarningsAsErrors bool   `default:"false" flag:"warnings-as-errors" info:"Treat warnings as errors"`
	BaseDir          string `default:"" flag:"base-dir" info:"Directory to search for .cdc files (defaults to current directory)"`
	ShowIgnored      bool   `default:"false" flag:"show-ignored" info:"Show diagnostics suppressed by //lint:ignore <category> directives"`
}

type fileResult struct {
	FilePath           string
	Diagnostics        []analysis.Diagnostic
	IgnoredDiagnostics []analysis.Diagnostic
}

type lintResult struct {
	Results     []fileResult
	exitCode    int
	showIgnored bool
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
		baseDir := "."
		if lintFlags.BaseDir != "" {
			baseDir = lintFlags.BaseDir
		}
		var err error
		filePaths, err = findAllCadenceFiles(baseDir)
		if err != nil {
			return nil, fmt.Errorf("error finding Cadence files: %w", err)
		}
		if len(filePaths) == 0 {
			return nil, fmt.Errorf("no .cdc files found in the project")
		}
	} else {
		filePaths = args
	}

	result, err := lintFiles(state, lintFlags.WarningsAsErrors, filePaths...)
	if err != nil {
		return nil, err
	}

	result.showIgnored = lintFlags.ShowIgnored
	return result, nil
}

func lintFiles(
	state *flowkit.State,
	warningsAsErrors bool,
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
			sortDiagnostics(diagnostics)
			results = append(results, fileResult{
				FilePath:    location,
				Diagnostics: diagnostics,
			})
			continue
		}

		active, ignored := filterIgnoredDiagnostics(diagnostics, state, location)
		sortDiagnostics(active)
		sortDiagnostics(ignored)
		results = append(results, fileResult{
			FilePath:           location,
			Diagnostics:        active,
			IgnoredDiagnostics: ignored,
		})

		// Set the exitCode to 1 if any of the active diagnostics are error-level,
		// or warning-level when warningsAsErrors is set.
		for _, diagnostic := range active {
			severity := getDiagnosticSeverity(diagnostic)
			if severity == errorSeverity {
				exitCode = 1
				break
			}
			if severity == warningSeverity && warningsAsErrors {
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

// parseLintIgnoreDirectives returns a map of 1-indexed line number to the set of
// categories ignored on that line via //lint:ignore <category> comments.
func parseLintIgnoreDirectives(code string) map[int]map[string]bool {
	directives := make(map[int]map[string]bool)
	for i, line := range strings.Split(code, "\n") {
		lineNum := i + 1
		_, after, found := strings.Cut(line, "//lint:ignore ")
		if !found {
			continue
		}
		category := strings.TrimSpace(after)
		if sp := strings.IndexByte(category, ' '); sp >= 0 {
			category = category[:sp]
		}
		if category == "" {
			continue
		}
		if directives[lineNum] == nil {
			directives[lineNum] = make(map[string]bool)
		}
		directives[lineNum][category] = true
	}
	return directives
}

func isDiagnosticIgnored(d analysis.Diagnostic, directives map[int]map[string]bool) bool {
	line := d.Range.StartPos.Line
	for _, l := range []int{line, line - 1} {
		if cats, ok := directives[l]; ok && cats[d.Category] {
			return true
		}
	}
	return false
}

// filterIgnoredDiagnostics reads the source for location and splits diagnostics
// into active and ignored based on //lint:ignore directives. If the source cannot
// be read, all diagnostics are returned as active.
func filterIgnoredDiagnostics(
	diagnostics []analysis.Diagnostic,
	state *flowkit.State,
	location string,
) (active, ignored []analysis.Diagnostic) {
	code, err := state.ReadFile(location)
	if err != nil {
		if diagnostics == nil {
			return []analysis.Diagnostic{}, nil
		}
		return diagnostics, nil
	}

	directives := parseLintIgnoreDirectives(string(code))
	for _, d := range diagnostics {
		if isDiagnosticIgnored(d, directives) {
			ignored = append(ignored, d)
		} else {
			active = append(active, d)
		}
	}
	if active == nil {
		active = []analysis.Diagnostic{}
	}
	return active, ignored
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

func renderIgnoredDiagnostic(diagnostic analysis.Diagnostic) string {
	startPos := diagnostic.Range.StartPos
	locationText := fmt.Sprintf("%s:%d:%d:", diagnostic.Location.String(), startPos.Line, startPos.Column)
	categoryText := fmt.Sprintf("%s:", diagnostic.Category)

	return fmt.Sprintf("%s %s %s (ignored)",
		aurora.Gray(12, locationText).String(),
		aurora.Gray(12, categoryText).String(),
		aurora.Gray(12, diagnostic.Message).String(),
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

func (r *lintResult) countIgnored() int {
	n := 0
	for _, result := range r.Results {
		n += len(result.IgnoredDiagnostics)
	}
	return n
}

func (r *lintResult) String() string {
	var sb strings.Builder

	for _, result := range r.Results {
		for _, diagnostic := range result.Diagnostics {
			fmt.Fprintf(&sb, "%s\n\n", renderDiagnostic(diagnostic))
		}
		if r.showIgnored {
			for _, diagnostic := range result.IgnoredDiagnostics {
				fmt.Fprintf(&sb, "%s\n\n", renderIgnoredDiagnostic(diagnostic))
			}
		}
	}

	numErrors, numWarnings := r.countProblems()
	numIgnored := r.countIgnored()
	total := numErrors + numWarnings + numIgnored

	var color aurora.Color
	if numErrors > 0 {
		color = aurora.RedFg
	} else if numWarnings > 0 {
		color = aurora.YellowFg
	}

	if total == 0 {
		sb.WriteString(aurora.Green("Lint passed").String())
		return sb.String()
	}

	if numIgnored > 0 {
		sb.WriteString(aurora.Colorize(fmt.Sprintf(
			"%d %s (%d %s, %d %s, %d ignored)",
			total,
			util.Pluralize("problem", total),
			numErrors,
			util.Pluralize("error", numErrors),
			numWarnings,
			util.Pluralize("warning", numWarnings),
			numIgnored,
		), color).String())
	} else {
		sb.WriteString(aurora.Colorize(fmt.Sprintf(
			"%d %s (%d %s, %d %s)",
			total,
			util.Pluralize("problem", total),
			numErrors,
			util.Pluralize("error", numErrors),
			numWarnings,
			util.Pluralize("warning", numWarnings),
		), color).String())
	}

	return sb.String()
}

func (r *lintResult) JSON() any {
	return r
}

func (r *lintResult) Oneliner() string {
	numErrors, numWarnings := r.countProblems()
	numIgnored := r.countIgnored()
	total := numErrors + numWarnings + numIgnored

	if total == 0 {
		return "Lint passed"
	}

	if numIgnored > 0 {
		return fmt.Sprintf("%d %s (%d %s, %d %s, %d ignored)",
			total, util.Pluralize("problem", total),
			numErrors, util.Pluralize("error", numErrors),
			numWarnings, util.Pluralize("warning", numWarnings),
			numIgnored,
		)
	}

	return fmt.Sprintf("%d %s (%d %s, %d %s)",
		total, util.Pluralize("problem", total),
		numErrors, util.Pluralize("error", numErrors),
		numWarnings, util.Pluralize("warning", numWarnings),
	)
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
