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
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	cdcTests "github.com/onflow/cadence-tools/test"
	"github.com/onflow/cadence/runtime"
	"github.com/onflow/cadence/runtime/common"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

// The key where meta information for a test report in JSON
// format can be found.
const TestReportMetaKey = "meta"

// The key where coverage meta information for a test report
// in JSON format can be found.
const TestReportCoverageKey = "coverage"

// The key where seed meta information for a test report
// in JSON format can be found.
const TestReportSeedKey = "seed"

// Import statements with a path that contain this substring,
// are considered to be helper/utility scripts for test files.
const helperScriptSubstr = "_helper"

// When the value of flagsTests.CoverCode equals "contracts",
// scripts and transactions are excluded from coverage report.
const contractsCoverCode = "contracts"

// The default glob pattern to find test files.
const defaultTestSuffix = "_test.cdc"

type flagsTests struct {
	Cover        bool   `default:"false" flag:"cover" info:"Use the cover flag to calculate coverage report"`
	CoverProfile string `default:"coverage.json" flag:"coverprofile" info:"Filename to write the calculated coverage report. Supported extensions are .json and .lcov"`
	CoverCode    string `default:"all" flag:"covercode" info:"Use the covercode flag to calculate coverage report only for certain types of code. Available values are \"all\" & \"contracts\""`
	Random       bool   `default:"false" flag:"random" info:"Use the random flag to execute test cases randomly"`
	Seed         int64  `default:"0" flag:"seed" info:"Use the seed flag to manipulate random execution of test cases"`
	Name         string `default:"" flag:"name" info:"Use the name flag to run only tests that match the given name"`
}

var testFlags = flagsTests{}

var TestCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:   "test [files...]",
		Short: "Run Cadence tests",
		Example: `# Run tests in files matching default pattern **/*_test.cdc
flow test

# Run tests in the specified files
flow test test1.cdc test2.cdc`,
		Args:    cobra.ArbitraryArgs,
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
	if testFlags.Random && testFlags.Seed > 0 {
		fmt.Printf(
			"%s Both '--seed' and '--random' flags are used. Hence, the '--random' flag will be ignored.\n",
			output.WarningEmoji(),
		)
	}

	var filenames []string
	if len(args) == 0 {
		var err error
		filenames, err = findAllTestFiles(".")
		if err != nil {
			return nil, fmt.Errorf("error loading script files: %w", err)
		}
	} else {
		filenames = args
	}

	testFiles := make(map[string][]byte, 0)
	for _, filename := range filenames {
		code, err := state.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("error loading script file: %w", err)
		}

		testFiles[filename] = code
	}

	result, err := testCode(testFiles, state, testFlags)
	if err != nil {
		return nil, err
	}

	if result.CoverageReport != nil {
		var file []byte
		var err error

		ext := filepath.Ext(testFlags.CoverProfile)
		if ext == ".json" {
			file, err = json.MarshalIndent(result.CoverageReport, "", "  ")
		} else if ext == ".lcov" {
			file, err = result.CoverageReport.MarshalLCOV()
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

	return result, nil
}

func testCode(
	testFiles map[string][]byte,
	state *flowkit.State,
	flags flagsTests,
) (*result, error) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	runner := cdcTests.NewTestRunner().WithLogger(logger)

	var coverageReport *runtime.CoverageReport
	if flags.Cover {
		coverageReport = state.CreateCoverageReport("testing")
		if flags.CoverCode == contractsCoverCode {
			coverageReport.WithLocationFilter(
				func(location common.Location) bool {
					_, addressLoc := location.(common.AddressLocation)
					// We only allow inspection of AddressLocation,
					// since scripts and transactions cannot be
					// attributed to their source files anyway.
					return addressLoc
				},
			)
		}
		runner = runner.WithCoverageReport(coverageReport)
	}

	var seed int64
	if flags.Seed > 0 {
		seed = flags.Seed
		runner = runner.WithRandomSeed(seed)
	} else if flags.Random {
		seed = int64(rand.Intn(150000))
		runner = runner.WithRandomSeed(seed)
	}

	contractsConfig := *state.Contracts()
	contracts := make(map[string]common.Address, len(contractsConfig))
	for _, contract := range contractsConfig {
		alias := contract.Aliases.ByNetwork("testing")
		if alias != nil {
			contracts[contract.Name] = common.Address(alias.Address)
		}
	}

	testResults := make(map[string]cdcTests.Results, 0)
	exitCode := 0
	for scriptPath, code := range testFiles {
		runner := runner.
			WithImportResolver(importResolver(scriptPath, state)).
			WithFileResolver(fileResolver(scriptPath, state)).
			WithContracts(contracts)

		if flags.Name != "" {
			testFunctions, err := runner.GetTests(string(code))
			if err != nil {
				return nil, err
			}

			for _, testFunction := range testFunctions {
				if testFunction != flags.Name {
					continue
				}

				result, err := runner.RunTest(string(code), flags.Name)
				if err != nil {
					return nil, err
				}
				testResults[scriptPath] = []cdcTests.Result{*result}
			}
		} else {
			results, err := runner.RunTests(string(code))
			if err != nil {
				return nil, err
			}
			testResults[scriptPath] = results
		}

		for _, result := range testResults[scriptPath] {
			if result.Error != nil {
				exitCode = 1
				break
			}
		}
	}

	return &result{
		Results:        testResults,
		CoverageReport: coverageReport,
		RandomSeed:     seed,
		exitCode:       exitCode,
	}, nil
}

func importResolver(scriptPath string, state *flowkit.State) cdcTests.ImportResolver {
	contracts := make(map[string]config.Contract, 0)
	for _, contract := range *state.Contracts() {
		contracts[contract.Name] = contract
	}

	return func(location common.Location) (string, error) {
		contract := config.Contract{}

		switch location := location.(type) {
		case common.AddressLocation:
			contract = contracts[location.Name]

		case common.StringLocation:
			relativePath := location.String()

			if strings.Contains(relativePath, helperScriptSubstr) {
				importedScriptFilePath := absolutePath(scriptPath, relativePath)
				scriptCode, err := state.ReadFile(importedScriptFilePath)
				if err != nil {
					return "", nil
				}
				return string(scriptCode), nil
			}

			contract = contracts[relativePath]
		}

		if contract.Location == "" {
			return "", fmt.Errorf(
				"cannot find contract with location '%s' in configuration",
				location,
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
	RandomSeed     int64
	exitCode       int
}

var _ command.ResultWithExitCode = &result{}

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

	meta := map[string]string{}
	if r.CoverageReport != nil {
		meta[TestReportCoverageKey] = r.CoverageReport.Percentage()
	}
	if r.RandomSeed > 0 {
		meta[TestReportSeedKey] = fmt.Sprint(r.RandomSeed)
	}
	results[TestReportMetaKey] = meta

	return results
}

func (r *result) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	if len(r.Results) == 0 {
		_, _ = fmt.Fprint(writer, "No tests found")
	} else {
		for scriptPath, testResult := range r.Results {
			_, _ = fmt.Fprint(writer, cdcTests.PrettyPrintResults(testResult, scriptPath))
		}
		if r.CoverageReport != nil {
			_, _ = fmt.Fprint(writer, r.CoverageReport.String())
		}
		if r.RandomSeed > 0 {
			_, _ = fmt.Fprintf(writer, "\nSeed: %d", r.RandomSeed)
		}
	}

	_ = writer.Flush()

	return b.String()
}

func (r *result) Oneliner() string {
	var builder strings.Builder

	if len(r.Results) == 0 {
		builder.WriteString("No tests found")
		return builder.String()
	}

	for scriptPath, testResult := range r.Results {
		builder.WriteString(cdcTests.PrettyPrintResults(testResult, scriptPath))
	}
	if r.CoverageReport != nil {
		builder.WriteString(r.CoverageReport.String())
		builder.WriteString("\n")
	}
	if r.RandomSeed > 0 {
		builder.WriteString(fmt.Sprintf("Seed: %d", r.RandomSeed))
		builder.WriteString("\n")
	}

	return builder.String()
}

func (r *result) ExitCode() int {
	return r.exitCode
}

func findAllTestFiles(baseDir string) ([]string, error) {
	var filenames []string
	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, defaultTestSuffix) {
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
