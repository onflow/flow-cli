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
	"path"
	"strings"

	cdcTests "github.com/onflow/cadence-tools/test"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/common"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/flowkit"
	"github.com/onflow/flow-cli/flowkit/output"
	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type flagsTests struct {
	Cover        bool   `default:"false" flag:"cover" info:"Use the cover flag to calculate coverage report"`
	CoverProfile string `default:"coverage.json" flag:"coverprofile" info:"Filename to write the calculated coverage report"`
}

var testFlags = flagsTests{}

var TestCommand = &command.Command{
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

	logger.StartProgress("Running tests...")
	defer logger.StopProgress()

	res, coverageReport, err := testCode(testFiles, state, testFlags.Cover)
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
	}

	return &result{
		Results:        res,
		CoverageReport: coverageReport,
	}, nil
}

func testCode(
	testFiles map[string][]byte,
	state *flowkit.State,
	coverageEnabled bool,
) (map[string]cdcTests.Results, *runtime.CoverageReport, error) {
	var coverageReport *runtime.CoverageReport
	runner := cdcTests.NewTestRunner()
	if coverageEnabled {
		coverageReport = runtime.NewCoverageReport()
		contracts := map[string]string{
			"FlowToken":             "0x0ae53cb6e3f42a79",
			"FlowFees":              "0xe5a8b7f23e8b548f",
			"FungibleToken":         "0xee82856bf20e2aa6",
			"FlowClusterQC":         "0xf8d6e0586b0a20c7",
			"FlowDKG":               "0xf8d6e0586b0a20c7",
			"FlowEpoch":             "0xf8d6e0586b0a20c7",
			"FlowIDTableStaking":    "0xf8d6e0586b0a20c7",
			"FlowServiceAccount":    "0xf8d6e0586b0a20c7",
			"FlowStakingCollection": "0xf8d6e0586b0a20c7",
			"FlowStorageFees":       "0xf8d6e0586b0a20c7",
			"LockedTokens":          "0xf8d6e0586b0a20c7",
			"NodeVersionBeacon":     "0xf8d6e0586b0a20c7",
			"StakingProxy":          "0xf8d6e0586b0a20c7",
			"ExampleNFT":            "0xf8d6e0586b0a20c7",
			"FUSD":                  "0xf8d6e0586b0a20c7",
			"NFTStorefront":         "0xf8d6e0586b0a20c7",
			"NFTStorefrontV2":       "0xf8d6e0586b0a20c7",
		}
		for name, address := range contracts {
			addr, _ := common.HexToAddress(address)
			location := common.AddressLocation{
				Address: addr,
				Name:    name,
			}
			coverageReport.ExcludeLocation(location)
		}
		coverageReport.WithLocationInspectionHandler(func(location common.Location) bool {
			_, addressLoc := location.(common.AddressLocation)
			_, stringLoc := location.(common.StringLocation)
			return addressLoc || stringLoc
		})
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
		contractFound := false
		for _, contract := range *state.Contracts() {
			if strings.Contains(relativePath, contract.Location) {
				contractFound = true
				break
			}
		}
		if !contractFound {
			return "", fmt.Errorf(
				"cannot find contract with location '%s' in configuration",
				relativePath,
			)
		}

		importedContractFilePath := absolutePath(scriptPath, relativePath)
		contractCode, err := state.ReadFile(importedContractFilePath)
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
	if path.IsAbs(filePath) {
		return filePath
	}

	return path.Join(path.Dir(basePath), filePath)
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
