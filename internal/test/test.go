/*
 * Flow CLI
 *
 * Copyright 2022 Dapper Labs, Inc.
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

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	cdcTests "github.com/onflow/cadence-tools/test"
	"github.com/onflow/cadence/runtime"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
	"github.com/onflow/flow-cli/pkg/flowkit/util"
)

type flagsTests struct {
	Cover        bool   `default:"false" flag:"cover" info:"Use the cover flag to calculate coverage report"`
	CoverProfile string `default:"coverage.json" flag:"coverprofile" info:"Filename to write the calculated coverage report"`
}

var testFlags = flagsTests{}

var Cmd = &cobra.Command{
	Use:     "test <filename>",
	Short:   "Run Cadence tests",
	Example: `flow test script.cdc`,
	Args:    cobra.MinimumNArgs(1),
	GroupID: "tools",
}

var TestCommand = &command.Command{
	Cmd:   Cmd,
	Flags: &testFlags,
	Run:   run,
}

func run(
	args []string,
	readerWriter flowkit.ReaderWriter,
	_ command.GlobalFlags,
	services *services.Services,
) (command.Result, error) {
	testFiles := make(map[string][]byte, 0)
	for _, filename := range args {
		code, err := readerWriter.ReadFile(filename)

		if err != nil {
			return nil, fmt.Errorf("error loading script file: %w", err)
		}

		testFiles[filename] = code
	}

	result, coverageReport, err := services.Tests.Execute(
		testFiles,
		readerWriter,
		testFlags.Cover,
	)

	if err != nil {
		return nil, err
	}

	if coverageReport != nil {
		file, err := json.MarshalIndent(coverageReport, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error serializing coverage report: %w", err)
		}

		err = os.WriteFile(testFlags.CoverProfile, file, 0644)
		if err != nil {
			return nil, fmt.Errorf("error writing coverage report file: %w", err)
		}
	} else if Cmd.Flags().Changed("coverprofile") {
		return nil, fmt.Errorf("the '--coverprofile' flag requires the '--cover' flag")
	}

	return &TestResult{
		Results:        result,
		CoverageReport: coverageReport,
	}, nil
}

type TestResult struct {
	Results        map[string]cdcTests.Results
	CoverageReport *runtime.CoverageReport
}

var _ command.Result = &TestResult{}

func (r *TestResult) JSON() any {
	results := make(map[string]map[string]string, len(r.Results))

	for testFile, testResult := range r.Results {
		testFileResults := make(map[string]string, len(testResult))
		for _, result := range testResult {
			var status string
			if result.Error == nil {
				status = "PASS"
			} else {
				status = fmt.Sprintf("FAIL: %s", result.Error.Error())
			}
			testFileResults[result.TestName] = status
		}
		results[testFile] = testFileResults
	}

	if r.CoverageReport != nil {
		results["meta"] = map[string]string{
			"info": r.CoverageReport.Percentage(),
		}
	}

	return results
}

func (r *TestResult) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	for scriptPath, testResult := range r.Results {
		_, _ = fmt.Fprint(writer, cdcTests.PrettyPrintResults(testResult, scriptPath))
	}
	if r.CoverageReport != nil {
		_, _ = fmt.Fprint(writer, r.CoverageReport.String())
	}

	_ = writer.Flush()

	return b.String()
}

func (r *TestResult) Oneliner() string {
	var builder strings.Builder

	for scriptPath, testResult := range r.Results {
		builder.WriteString(cdcTests.PrettyPrintResults(testResult, scriptPath))
	}
	if r.CoverageReport != nil {
		builder.WriteString(r.CoverageReport.String())
		builder.WriteString("\n")
	}

	return builder.String()
}
