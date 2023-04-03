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
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"path"
	"os"
	"strings"

	cdcTests "github.com/onflow/cadence-tools/test"
	"github.com/onflow/cadence/runtime"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
	"github.com/onflow/flow-cli/pkg/flowkit"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
)

type flagsTests struct {
	Cover        bool   `default:"false" flag:"cover" info:"Use the cover flag to calculate coverage report"`
	CoverProfile string `default:"coverage.json" flag:"coverprofile" info:"Filename to write the calculated coverage report"`
}

var testFlags = flagsTests{}

var Command = &command.Command{
	Cmd: &cobra.Command{
		Use:     "test <filename>",
		Short:   "Run Cadence tests",
		Example: `flow test script.cdc`,
		Args:    cobra.MinimumNArgs(1),
		GroupID: "tools",
	},
	Flags: &testFlags,
	RunS:  run,
}

func run(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	testFiles := make(map[string][]byte, 0)
	for _, filename := range args {
		code, err := state.ReadFile(filename)

		if err != nil {
			return nil, fmt.Errorf("error loading script file: %w", err)
		}

		testFiles[filename] = code
	}

	logger.StartProgress("Running tests...")
	defer logger.StopProgress()

	result, err := testCode(code, filename, state)
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

	return &Result{
		Results:        result,
		CoverageReport: coverageReport,
	}, nil
}

func testCode(code []byte, scriptPath string, state *flowkit.State) (cdcTests.Results, error) {
	runner := cdcTests.NewTestRunner().
		WithImportResolver(importResolver(scriptPath, state)).
		WithFileResolver(fileResolver(scriptPath, state))

	return runner.RunTests(string(code))
}

func importResolver(scriptPath string, state *flowkit.State) cdcTests.ImportResolver {
	return func(location common.Location) (string, error) {
		stringLocation, isFileImport := location.(common.StringLocation)
		if !isFileImport {
			return "", fmt.Errorf("cannot import from %s", location)
		}

		importedContract, err := resolveContract(stringLocation, state)
		if err != nil {
			return "", err
		}

		importedContractFilePath := absolutePath(scriptPath, importedContract.Location)

		contractCode, err := state.ReadFile(importedContractFilePath)
		if err != nil {
			return "", err
		}

		return string(contractCode), nil
	}
}

func resolveContract(stringLocation common.StringLocation, state *flowkit.State) (config.Contract, error) {
	relativePath := stringLocation.String()
	for _, contract := range *state.Contracts() {
		if contract.Location == relativePath {
			return contract, nil
		}
	}

	return config.Contract{}, fmt.Errorf("cannot find contract with location '%s' in configuration", relativePath)
}

func fileResolver(scriptPath string, state *flowkit.State) cdcTests.FileResolver {
	return func(path string) (string, error) {
		importFilePath := absolutePath(scriptPath, path)

		content, err := state.ReadFile(importFilePath)
		if err != nil {
			return "", err
		}

		return string(content), nil
	}
}

func absolutePath(basePath, filePath string) string {
	if path.IsAbs(filePath) {
		return filePath
	}

	return path.Join(path.Dir(basePath), filePath)
}

type Result struct {
	Results        map[string]cdcTests.Results
	CoverageReport *runtime.CoverageReport
}

var _ command.Result = &Result{}

func (r *Result) JSON() any {
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
			"info": r.CoverageReport.CoveredStatementsPercentage(),
		}
	}

	return results
}

func (r *Result) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	for scriptPath, testResult := range r.Results {
		_, _ = fmt.Fprint(writer, cdcTests.PrettyPrintResults(testResult, scriptPath))
	}
	if r.CoverageReport != nil {
		_, _ = fmt.Fprint(writer, r.CoverageReport.CoveredStatementsPercentage())
	}

	_ = writer.Flush()

	return b.String()
}

func (r *Result) Oneliner() string {
	var builder strings.Builder

	for scriptPath, testResult := range r.Results {
		builder.WriteString(cdcTests.PrettyPrintResults(testResult, scriptPath))
	}
	if r.CoverageReport != nil {
		builder.WriteString(r.CoverageReport.CoveredStatementsPercentage())
		builder.WriteString("\n")
	}

	return builder.String()
}
