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

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	goRuntime "runtime"
	"strings"

	cdcTests "github.com/onflow/cadence-tools/test"
	"github.com/onflow/cadence/common"
	"github.com/onflow/cadence/runtime"
	flowGo "github.com/onflow/flow-go/model/flow"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/common/branding"

	"github.com/onflow/flow-cli/build"
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
	CoverProfile string `default:"lcov.info" flag:"coverprofile" info:"Filename to write the calculated coverage report. Supported extensions are .info, .lcov, and .json"`
	CoverCode    string `default:"all" flag:"covercode" info:"Use the covercode flag to calculate coverage report only for certain types of code. Available values are \"all\" & \"contracts\""`
	Random       bool   `default:"false" flag:"random" info:"Use the random flag to execute test cases randomly"`
	Seed         int64  `default:"0" flag:"seed" info:"Use the seed flag to manipulate random execution of test cases"`
	Name         string `default:"" flag:"name" info:"Use the name flag to run only tests that match the given name"`

	// Fork mode flags
	Fork       string // Use definition in init()
	ForkHost   string `default:"" flag:"fork-host" info:"Run tests against a fork of a remote network. Provide the GRPC Access host (host:port)."`
	ForkHeight uint64 `default:"0" flag:"fork-height" info:"Optional block height to pin the fork (if supported)."`
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

func init() {
	// Add default value to --fork flag
	// workaround because config schema via struct tags doesn't support default values
	TestCommand.Cmd.Flags().StringVar(&testFlags.Fork, "fork", "", "Fork tests from a remote network. If provided without a value, defaults to mainnet")
	if f := TestCommand.Cmd.Flags().Lookup("fork"); f != nil {
		f.NoOptDefVal = "mainnet"
	}
}

func run(
	args []string,
	_ command.GlobalFlags,
	logger output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	if !testFlags.Cover && testFlags.CoverProfile != "lcov.info" {
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
		switch ext {
		case ".json":
			file, err = json.MarshalIndent(result.CoverageReport, "", "  ")
		case ".lcov", ".info":
			file, err = result.CoverageReport.MarshalLCOV()
		default:
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

	// Track network resolutions per file for pragma-based fork detection
	// Map: filename -> resolved network name
	fileNetworkResolutions := make(map[string]string)
	var currentTestFile string

	// Resolve network labels using flow.json state
	resolveNetworkFromState := func(label string) (string, bool) {
		normalizedLabel := strings.ToLower(strings.TrimSpace(label))
		network, err := state.Networks().ByName(normalizedLabel)
		if err != nil || network == nil {
			return "", false
		}

		// If network has a fork, resolve the fork network's host
		host := strings.TrimSpace(network.Host)
		if network.Fork != "" {
			forkName := strings.ToLower(strings.TrimSpace(network.Fork))
			forkNetwork, err := state.Networks().ByName(forkName)
			if err == nil && forkNetwork != nil {
				host = strings.TrimSpace(forkNetwork.Host)
			}
		}

		if host == "" {
			return "", false
		}

		// Track network resolution for current test file (indicates pragma-based fork usage)
		// Only track if it's not the default "testing" network
		if currentTestFile != "" && normalizedLabel != "testing" {
			if _, exists := fileNetworkResolutions[currentTestFile]; !exists {
				fileNetworkResolutions[currentTestFile] = normalizedLabel
			}
		}

		return host, true
	}

	// Configure fork mode if requested
	var effectiveForkHost string

	// Determine the fork host
	if flags.ForkHost != "" {
		effectiveForkHost = strings.TrimSpace(flags.ForkHost)
	} else if flags.Fork != "" {
		// Look up network in flow.json
		forkNetwork := strings.ToLower(flags.Fork)
		network, err := state.Networks().ByName(forkNetwork)
		if err != nil {
			return nil, fmt.Errorf("network %q not found in flow.json", flags.Fork)
		}
		effectiveForkHost = network.Host
		if effectiveForkHost == "" {
			return nil, fmt.Errorf("network %q has no host configured", flags.Fork)
		}
	}

	// Determine network label (used by resolver/addresses); default to testing
	networkLabel := "testing"
	if strings.TrimSpace(flags.Fork) != "" {
		networkLabel = strings.ToLower(flags.Fork)
	}

	// If fork mode is enabled, query the host to get chain ID
	var forkCfg *cdcTests.ForkConfig
	if effectiveForkHost != "" {
		forkChainID, err := util.GetChainIDFromHost(effectiveForkHost)
		if err != nil {
			return nil, fmt.Errorf("failed to get chain ID from fork host %q: %w", effectiveForkHost, err)
		}

		cfg := cdcTests.ForkConfig{
			ForkHost:   effectiveForkHost,
			ChainID:    forkChainID,
			ForkHeight: flags.ForkHeight,
		}
		forkCfg = &cfg

		// Map chain ID to a sensible network label if not provided explicitly
		if strings.TrimSpace(flags.Fork) == "" {
			switch forkChainID {
			case flowGo.Mainnet:
				networkLabel = "mainnet"
			case flowGo.Testnet:
				networkLabel = "testnet"
			}
		}
	}

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
	}

	var seed int64
	if flags.Seed > 0 {
		seed = flags.Seed
	} else if flags.Random {
		seed = int64(rand.Intn(150000))
	}

	testResults := make(map[string]cdcTests.Results, 0)
	exitCode := 0
	for scriptPath, code := range testFiles {
		// Set current test file for network resolution tracking
		currentTestFile = scriptPath

		// Create a new test runner per file to ensure complete isolation.
		// Each file gets its own runner with its own backend state.
		fileRunner := cdcTests.NewTestRunner().
			WithLogger(logger).
			WithNetworkResolver(resolveNetworkFromState).
			WithNetworkLabel(networkLabel).
			WithImportResolver(importResolver(scriptPath, state)).
			WithFileResolver(fileResolver(scriptPath, state)).
			WithContractAddressResolver(func(network string, contractName string) (common.Address, error) {
				contractsByName := make(map[string]config.Contract)
				for _, c := range *state.Contracts() {
					contractsByName[c.Name] = c
				}

				contract, exists := contractsByName[contractName]
				if !exists {
					return common.Address{}, fmt.Errorf("contract not found: %s", contractName)
				}

				alias := contract.Aliases.ByNetwork(network)
				if alias != nil {
					return common.Address(alias.Address), nil
				}

				// Fallback to fork network if configured
				networkConfig, err := state.Networks().ByName(network)
				if err == nil && networkConfig != nil && networkConfig.Fork != "" {
					forkAlias := contract.Aliases.ByNetwork(networkConfig.Fork)
					if forkAlias != nil {
						return common.Address(forkAlias.Address), nil
					}
				}

				return common.Address{}, fmt.Errorf("no address for contract %s on network %s", contractName, network)
			})

		if forkCfg != nil {
			fileRunner = fileRunner.WithFork(*forkCfg)
		}
		if coverageReport != nil {
			fileRunner = fileRunner.WithCoverageReport(coverageReport)
		}
		if seed > 0 {
			fileRunner = fileRunner.WithRandomSeed(seed)
		}

		if flags.Name != "" {
			testFunctions, err := fileRunner.GetTests(string(code))
			if err != nil {
				return nil, err
			}

			for _, testFunction := range testFunctions {
				if testFunction != flags.Name {
					continue
				}

				result, err := fileRunner.RunTest(string(code), flags.Name)
				if err != nil {
					return nil, err
				}
				testResults[scriptPath] = []cdcTests.Result{*result}
			}
		} else {
			results, err := fileRunner.RunTests(string(code))
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

		// Clear current test file after processing
		currentTestFile = ""
	}

	// Track fork test usage metrics - aggregate into single event
	hasPragmaFiles := len(fileNetworkResolutions) > 0
	hasStaticFork := forkCfg != nil

	if hasPragmaFiles || hasStaticFork {
		// Determine primary fork source
		forkSource := "none"
		var primaryNetwork string
		var chainID string
		hasHeight := false

		if hasPragmaFiles {
			// Pragma takes priority - collect unique networks
			forkSource = "pragma"
			networkSet := make(map[string]bool)
			for _, network := range fileNetworkResolutions {
				networkSet[network] = true
			}
			// Use first resolved network as primary (for single-value tracking)
			for _, network := range fileNetworkResolutions {
				primaryNetwork = network
				break
			}
			// If multiple networks, note that in source
			if len(networkSet) > 1 {
				forkSource = "pragma-mixed"
			}
		} else if hasStaticFork {
			// Static flags
			if flags.ForkHost != "" {
				forkSource = "fork-host-flag"
			} else if flags.Fork != "" {
				forkSource = "fork-flag"
			}
			primaryNetwork = networkLabel
			chainID = forkCfg.ChainID.String()
			hasHeight = forkCfg.ForkHeight > 0
		}

		command.TrackEvent("test-fork", map[string]any{
			"fork_source":  forkSource,
			"network":      primaryNetwork,
			"chain_id":     chainID,
			"has_height":   hasHeight,
			"pragma_files": len(fileNetworkResolutions),
			"total_files":  len(testFiles),
			"version":      build.Semver(),
			"os":           goRuntime.GOOS,
			"ci":           os.Getenv("CI") != "",
		})
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

	return func(network string, location common.Location) (string, error) {
		contract := config.Contract{}

		switch location := location.(type) {
		case common.AddressLocation:
			contract = contracts[location.Name]

		case common.StringLocation:
			relativePath := location.String()

			if strings.Contains(relativePath, helperScriptSubstr) {
				importedScriptFilePath := util.AbsolutePath(scriptPath, relativePath)
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
		importFilePath := util.AbsolutePath(scriptPath, path)

		content, err := state.ReadFile(importFilePath)
		if err != nil {
			return "", err
		}

		return string(content), nil
	}
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

// colorizeTestOutput adds colors to PASS/FAIL indicators in test output
func colorizeTestOutput(output string) string {
	// Regex patterns for PASS and FAIL
	passPattern := regexp.MustCompile(`(- PASS:)(.*)`)
	failPattern := regexp.MustCompile(`(- FAIL:)(.*)`)

	// Colorize PASS lines
	output = passPattern.ReplaceAllStringFunc(output, func(match string) string {
		parts := passPattern.FindStringSubmatch(match)
		if len(parts) >= 3 {
			passIndicator := branding.GreenStyle.Render(parts[1])
			testName := parts[2]
			return passIndicator + testName
		}
		return match
	})

	// Colorize FAIL lines
	output = failPattern.ReplaceAllStringFunc(output, func(match string) string {
		parts := failPattern.FindStringSubmatch(match)
		if len(parts) >= 3 {
			failIndicator := branding.ErrorStyle.Render(parts[1])
			testName := parts[2]
			return failIndicator + testName
		}
		return match
	})

	return output
}

func (r *result) String() string {
	var b bytes.Buffer
	writer := util.CreateTabWriter(&b)

	if len(r.Results) == 0 {
		_, _ = fmt.Fprint(writer, "No tests found")
	} else {
		for scriptPath, testResult := range r.Results {
			testOutput := cdcTests.PrettyPrintResults(testResult, scriptPath)
			colorizedOutput := colorizeTestOutput(testOutput)
			_, _ = fmt.Fprint(writer, colorizedOutput)
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
