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
	"context"
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
	"golang.org/x/sync/errgroup"

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
	Jobs         int    `default:"0" flag:"jobs" info:"Maximum number of test files to run concurrently (default: number of CPU cores)"`

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

// testRunConfig holds the resolved runtime configuration for a test run.
type testRunConfig struct {
	forkCfg        *cdcTests.ForkConfig
	coverageReport *runtime.CoverageReport
	networkLabel   string
	seed           int64
	jobs           int
	name           string
	// raw flag values retained for telemetry
	forkFlag     string
	forkHostFlag string
}

// concurrencyResult holds the aggregated output of runTestsConcurrently.
type concurrencyResult struct {
	testResults            map[string]cdcTests.Results
	fileNetworkResolutions map[string]string
	exitCode               int
}

func testCode(
	testFiles map[string][]byte,
	state *flowkit.State,
	flags flagsTests,
) (*result, error) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	forkCfg, networkLabel, err := resolveForkConfig(flags, state)
	if err != nil {
		return nil, err
	}

	cfg := testRunConfig{
		forkCfg:        forkCfg,
		coverageReport: buildCoverageReport(flags, state),
		networkLabel:   networkLabel,
		seed:           resolveSeed(flags),
		jobs:           flags.Jobs,
		name:           flags.Name,
		forkFlag:       flags.Fork,
		forkHostFlag:   flags.ForkHost,
	}

	cr, err := runTestsConcurrently(testFiles, state, cfg, logger)
	if err != nil {
		return nil, err
	}

	trackForkMetrics(cr, cfg, len(testFiles))

	return &result{
		Results:        cr.testResults,
		CoverageReport: cfg.coverageReport,
		RandomSeed:     cfg.seed,
		exitCode:       cr.exitCode,
	}, nil
}

// resolveForkConfig determines the fork configuration and network label from flags.
func resolveForkConfig(flags flagsTests, state *flowkit.State) (*cdcTests.ForkConfig, string, error) {
	networkLabel := "testing"
	var effectiveForkHost string

	if flags.ForkHost != "" {
		effectiveForkHost = strings.TrimSpace(flags.ForkHost)
	} else if flags.Fork != "" {
		network, err := state.Networks().ByName(strings.ToLower(flags.Fork))
		if err != nil {
			return nil, "", fmt.Errorf("network %q not found in flow.json", flags.Fork)
		}
		effectiveForkHost = network.Host
		if effectiveForkHost == "" {
			return nil, "", fmt.Errorf("network %q has no host configured", flags.Fork)
		}
	}

	if strings.TrimSpace(flags.Fork) != "" {
		networkLabel = strings.ToLower(flags.Fork)
	}

	if effectiveForkHost == "" {
		return nil, networkLabel, nil
	}

	forkChainID, err := util.GetChainIDFromHost(effectiveForkHost)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get chain ID from fork host %q: %w", effectiveForkHost, err)
	}

	// Map chain ID to a sensible network label if not provided explicitly
	if strings.TrimSpace(flags.Fork) == "" {
		switch forkChainID {
		case flowGo.Mainnet:
			networkLabel = "mainnet"
		case flowGo.Testnet:
			networkLabel = "testnet"
		}
	}

	cfg := cdcTests.ForkConfig{
		ForkHost:   effectiveForkHost,
		ChainID:    forkChainID,
		ForkHeight: flags.ForkHeight,
	}
	return &cfg, networkLabel, nil
}

// buildCoverageReport creates a coverage report if coverage is enabled.
func buildCoverageReport(flags flagsTests, state *flowkit.State) *runtime.CoverageReport {
	if !flags.Cover {
		return nil
	}
	coverageReport := state.CreateCoverageReport("testing")
	if flags.CoverCode == contractsCoverCode {
		coverageReport.WithLocationFilter(func(location common.Location) bool {
			// We only allow inspection of AddressLocation,
			// since scripts and transactions cannot be
			// attributed to their source files anyway.
			_, addressLoc := location.(common.AddressLocation)
			return addressLoc
		})
	}
	return coverageReport
}

// resolveSeed returns the random seed to use for test execution.
func resolveSeed(flags flagsTests) int64 {
	if flags.Seed > 0 {
		return flags.Seed
	}
	if flags.Random {
		return int64(rand.Intn(150000))
	}
	return 0
}

// networkResolver returns a function that resolves a network label to its host,
// tracking which non-testing network was resolved via resolvedNetwork.
func networkResolver(state *flowkit.State, resolvedNetwork *string) func(string) (string, bool) {
	return func(label string) (string, bool) {
		normalizedLabel := strings.ToLower(strings.TrimSpace(label))
		network, err := state.Networks().ByName(normalizedLabel)
		if err != nil || network == nil {
			return "", false
		}

		host := strings.TrimSpace(network.Host)
		if network.Fork != "" {
			forkName := strings.ToLower(strings.TrimSpace(network.Fork))
			forkNetwork, err := state.Networks().ByName(forkName)
			if err != nil {
				return "", false
			}
			host = strings.TrimSpace(forkNetwork.Host)
		}

		if host == "" {
			return "", false
		}

		// Track network resolution for the current test file (indicates pragma-based fork usage).
		// Only track if it's not the default "testing" network.
		if *resolvedNetwork == "" && normalizedLabel != "testing" {
			*resolvedNetwork = normalizedLabel
		}

		return host, true
	}
}

// contractAddressResolver returns a function that resolves a contract name to its address on a network.
func contractAddressResolver(state *flowkit.State) func(string, string) (common.Address, error) {
	contractsByName := make(map[string]config.Contract)
	for _, c := range *state.Contracts() {
		contractsByName[c.Name] = c
	}

	return func(network string, contractName string) (common.Address, error) {
		contract, exists := contractsByName[contractName]
		if !exists {
			return common.Address{}, fmt.Errorf("contract not found: %s", contractName)
		}

		if alias := contract.Aliases.ByNetwork(network); alias != nil {
			return common.Address(alias.Address), nil
		}

		// Fallback to fork network if configured.
		networkConfig, err := state.Networks().ByName(network)
		if err == nil && networkConfig != nil && networkConfig.Fork != "" {
			if forkAlias := contract.Aliases.ByNetwork(networkConfig.Fork); forkAlias != nil {
				return common.Address(forkAlias.Address), nil
			}
		}

		return common.Address{}, fmt.Errorf("no address for contract %s on network %s", contractName, network)
	}
}

// buildTestRunner creates an isolated test runner for a single file.
// It also returns a pointer to a string that will be populated with the resolved
// non-testing network name (if any) after the runner executes.
func buildTestRunner(scriptPath string, state *flowkit.State, cfg testRunConfig, logger zerolog.Logger) (*cdcTests.TestRunner, *string) {
	var resolvedNetwork string

	runner := cdcTests.NewTestRunner().
		WithLogger(logger).
		WithNetworkResolver(networkResolver(state, &resolvedNetwork)).
		WithNetworkLabel(cfg.networkLabel).
		WithImportResolver(importResolver(scriptPath, state)).
		WithFileResolver(fileResolver(scriptPath, state)).
		WithContractAddressResolver(contractAddressResolver(state))

	if cfg.forkCfg != nil {
		runner = runner.WithFork(*cfg.forkCfg)
	}
	if cfg.coverageReport != nil {
		runner = runner.WithCoverageReport(cfg.coverageReport)
	}
	if cfg.seed > 0 {
		runner = runner.WithRandomSeed(cfg.seed)
	}

	return runner, &resolvedNetwork
}

// runFileTests runs the tests in code using runner, optionally filtering by name.
func runFileTests(runner *cdcTests.TestRunner, code []byte, name string) (cdcTests.Results, error) {
	if name == "" {
		return runner.RunTests(string(code))
	}

	testFunctions, err := runner.GetTests(string(code))
	if err != nil {
		return nil, err
	}

	for _, fn := range testFunctions {
		if fn != name {
			continue
		}
		r, err := runner.RunTest(string(code), name)
		if err != nil {
			return nil, err
		}
		return cdcTests.Results{*r}, nil
	}

	return nil, nil
}

func runTestsConcurrently(
	testFiles map[string][]byte,
	state *flowkit.State,
	cfg testRunConfig,
	logger zerolog.Logger,
) (*concurrencyResult, error) {
	jobs := cfg.jobs
	if jobs <= 0 {
		jobs = goRuntime.NumCPU()
	}

	type fileResult struct {
		scriptPath        string
		results           cdcTests.Results
		networkResolution string
		err               error
	}

	resultCh := make(chan fileResult, len(testFiles))

	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(jobs)

	for scriptPath, code := range testFiles {
		g.Go(func() error {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			runner, resolvedNetwork := buildTestRunner(scriptPath, state, cfg, logger)
			results, err := runFileTests(runner, code, cfg.name)

			resultCh <- fileResult{
				scriptPath:        scriptPath,
				results:           results,
				networkResolution: *resolvedNetwork,
				err:               err,
			}
			return nil
		})
	}

	waitErr := g.Wait()
	close(resultCh)

	cr := &concurrencyResult{
		testResults:            make(map[string]cdcTests.Results),
		fileNetworkResolutions: make(map[string]string),
	}

	for r := range resultCh {
		if r.err != nil && waitErr == nil {
			waitErr = r.err
		}
		if r.results != nil {
			cr.testResults[r.scriptPath] = r.results
			// Check for individual test failures to set exit code
			for _, res := range r.results {
				if res.Error != nil {
					cr.exitCode = 1
				}
			}
		}
		if r.networkResolution != "" {
			cr.fileNetworkResolutions[r.scriptPath] = r.networkResolution
		}
	}

	return cr, waitErr
}

// trackForkMetrics emits a telemetry event when fork mode is used.
func trackForkMetrics(cr *concurrencyResult, cfg testRunConfig, totalFiles int) {
	hasPragmaFiles := len(cr.fileNetworkResolutions) > 0
	hasStaticFork := cfg.forkCfg != nil

	if !hasPragmaFiles && !hasStaticFork {
		return
	}

	forkSource := "none"
	var primaryNetwork, chainID string
	hasHeight := false

	if hasPragmaFiles {
		forkSource = "pragma"
		networkSet := make(map[string]bool)
		for _, network := range cr.fileNetworkResolutions {
			networkSet[network] = true
		}
		for _, network := range cr.fileNetworkResolutions {
			primaryNetwork = network
			break
		}
		if len(networkSet) > 1 {
			forkSource = "pragma-mixed"
		}
	} else {
		if cfg.forkHostFlag != "" {
			forkSource = "fork-host-flag"
		} else if cfg.forkFlag != "" {
			forkSource = "fork-flag"
		}
		primaryNetwork = cfg.networkLabel
		chainID = cfg.forkCfg.ChainID.String()
		hasHeight = cfg.forkCfg.ForkHeight > 0
	}

	command.TrackEvent("test-fork", map[string]any{
		"fork_source":  forkSource,
		"network":      primaryNetwork,
		"chain_id":     chainID,
		"has_height":   hasHeight,
		"pragma_files": len(cr.fileNetworkResolutions),
		"total_files":  totalFiles,
		"version":      build.Semver(),
		"os":           goRuntime.GOOS,
		"ci":           os.Getenv("CI") != "",
	})
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
