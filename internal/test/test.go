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
	"path/filepath"
	"strings"

	cdcTests "github.com/onflow/cadence-tools/test"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/common"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/config"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

// Import statements with a path that contain this substring,
// are considered to be helper/utility scripts for test files.
const helperScriptSubstr = "_helper"

// When the value of flagsTests.CoverCode equals "contracts",
// scripts and transactions are excluded from coverage report.
const contractsCoverCode = "contracts"

type flagsTests struct {
	Cover        bool   `default:"false" flag:"cover" info:"Use the cover flag to calculate coverage report"`
	CoverProfile string `default:"coverage.json" flag:"coverprofile" info:"Filename to write the calculated coverage report. Supported extensions are .json and .lcov"`
	CoverCode    string `default:"all" flag:"covercode" info:"Use the covercode flag to calculate coverage report only for certain types of code. Available values are \"all\" & \"contracts\""`
}

var testFlags = flagsTests{}

var status = 0

var TestCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "test <filename>",
		Short:   "Run Cadence tests",
		Example: `flow test script.cdc`,
		Args:    cobra.MinimumNArgs(1),
		GroupID: "tools",
	},
	Flags:  &testFlags,
	RunS:   run,
	Status: &status,
}

func run(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	if !testFlags.Cover && testFlags.CoverProfile != "coverage.json" {
		return nil, fmt.Errorf("the '--coverprofile' flag requires the '--cover' flag")
	}

	testFiles := make(map[string][]byte, 0)
	for _, filename := range args {
		code, err := state.ReadFile(filename)

		if err != nil {
			return nil, fmt.Errorf("error loading script file: %w", err)
		}

		testFiles[filename] = code
	}

	res, coverageReport, err := testCode(testFiles, state, testFlags)
	if err != nil {
		return nil, err
	}

	if coverageReport != nil {
		var file []byte
		var err error

		ext := filepath.Ext(testFlags.CoverProfile)
		if ext == ".json" {
			file, err = json.MarshalIndent(coverageReport, "", "  ")
		} else if ext == ".lcov" {
			file, err = coverageReport.MarshalLCOV()
		} else {
			return nil, fmt.Errorf("given format: %v, only .json and .lcov are supported", ext)
		}

		if err != nil {
			return nil, fmt.Errorf("error serializing coverage report: %w", err)
		}

		err = os.WriteFile(testFlags.CoverProfile, file, 0644)
		if err != nil {
			return nil, fmt.Errorf("error writing coverage report file: %w", err)
		}
	}

	return &result{
		Results:        res,
		CoverageReport: coverageReport,
	}, nil
}

func testCode(
	testFiles map[string][]byte,
	state *flowkit.State,
	flags flagsTests,
) (map[string]cdcTests.Results, *runtime.CoverageReport, error) {
	var coverageReport *runtime.CoverageReport
	runner := cdcTests.NewTestRunner()
	if flags.Cover {
		coverageReport = runtime.NewCoverageReport()
		if flags.CoverCode == contractsCoverCode {
			coverageReport.WithLocationFilter(
				func(location common.Location) bool {
					_, addressLoc := location.(common.AddressLocation)
					_, stringLoc := location.(common.StringLocation)
					// We only allow inspection of AddressLocation or StringLocation,
					// since scripts and transactions cannot be attributed to their
					// source files anyway.
					return addressLoc || stringLoc
				},
			)
		}
		runner = runner.WithCoverageReport(coverageReport)
	}

	testResults := make(map[string]cdcTests.Results, 0)
	for scriptPath, code := range testFiles {
		runner := runner.
			WithImportResolver(importResolver(scriptPath, state)).
			WithFileResolver(fileResolver(scriptPath, state))
		results, err := runner.RunTests(string(code))
		if err != nil {
			return nil, nil, err
		}
		testResults[scriptPath] = results
		for _, result := range results {
			if result.Error != nil {
				status = 1
				break
			}
		}
	}
	return testResults, coverageReport, nil
}

func importResolver(scriptPath string, state *flowkit.State) cdcTests.ImportResolver {
	return func(location common.Location) (string, error) {
		stringLocation, isFileImport := location.(common.StringLocation)
		if !isFileImport {
			return "", fmt.Errorf("cannot import from %s", location)
		}
		relativePath := stringLocation.String()

		if strings.Contains(relativePath, helperScriptSubstr) {
			importedScriptFilePath := absolutePath(scriptPath, relativePath)
			scriptCode, err := state.ReadFile(importedScriptFilePath)
			if err != nil {
				return "", nil
			}
			return string(scriptCode), nil
		}

		var contract *config.Contract
		for _, c := range *state.Contracts() {
			if c.Name == relativePath {
				contract = &c
				break
			}
		}
		if contract == nil {
			return "", fmt.Errorf(
				"cannot find contract with location '%s' in configuration",
				relativePath,
			)
		}

		contractCode, err := state.ReadFile(contract.Location)
		if err != nil {
			return "", err
		}

		return string(contractCode), nil
	}
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
	if filepath.IsAbs(filePath) {
		return filePath
	}

	return filepath.Join(filepath.Dir(basePath), filePath)
}

type result struct {
	Results        map[string]cdcTests.Results
	CoverageReport *runtime.CoverageReport
}

var _ command.Result = &result{}

func (r *result) JSON() any {
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

func (r *result) String() string {
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

func (r *result) Oneliner() string {
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
