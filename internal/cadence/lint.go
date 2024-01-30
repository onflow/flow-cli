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

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/sema"
	"github.com/onflow/cadence/tools/analysis"
	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
)

type lintFlagsCollection struct {
	Fix bool `default:"false" flag:"fix" info:"Fix linting errors"`
}
type lintResult struct {
	FilePath string
	Diagnostics []analysis.Diagnostic
}

type lintResults struct {
	Results []lintResult
}

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

func lint(
	args []string,
	_ command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	linter := &linter{
		checkers: make(map[common.Location]*sema.Checker),
		state: state,
		filePaths: make(map[string]bool),
	}
	results := make([]lintResult, 0)

	for _, filePath := range args {
		diagnostics, err := linter.lintFile(filePath)
		if err != nil {
			return nil, err
		}

		results = append(results, lintResult{
			FilePath: filePath,
			Diagnostics: diagnostics,
		})
	}

	return &lintResults{
		Results: results,
	}, nil
}

func (r *lintResults) String() string {
	var sb strings.Builder
	var numProblems int

	for _, result := range r.Results {
		if len(result.Diagnostics) == 0 {
			continue
		}

		numProblems += len(result.Diagnostics)

		filenameStyle := lipgloss.NewStyle().
			Underline(true)

		absPath, err := filepath.Abs(result.FilePath)
		if err != nil {
			absPath = result.FilePath
		}

		sb.WriteString(filenameStyle.Render(absPath))
		sb.WriteString("\n")

		rows := make([][]string, 0)
		for i, diagnostic := range result.Diagnostics {
			columns := make([]string, 0)
			positionString := fmt.Sprintf("%d:%d", diagnostic.Range.StartPos.Line, diagnostic.Range.StartPos.Column)

			paddingTop := 1
			if i == 0 {
				paddingTop = 0
			}

			columns = append(columns, lipgloss.NewStyle().Width(7).Align(lipgloss.Center).PaddingTop(paddingTop).Render(positionString))
			columns = append(columns, lipgloss.NewStyle().Width(56).Align(lipgloss.Left).PaddingTop(paddingTop).PaddingLeft(1).Render(diagnostic.Message))
			columns = append(columns, lipgloss.NewStyle().Width(13).Align(lipgloss.Center).PaddingTop(paddingTop).Render(diagnostic.Category))

			rows = append(rows, columns)
		}

		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#1FEE8A"))).
			Rows(rows...)

		sb.WriteString(t.Render())
		sb.WriteString("\n\n")
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
	return ""
}