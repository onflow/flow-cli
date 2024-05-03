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

package command

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/dukex/mixpanel"
	"github.com/getsentry/sentry-go"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/gateway"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/build"
	"github.com/onflow/flow-cli/internal/settings"
	"github.com/onflow/flow-cli/internal/util"
)

// run the command with arguments.
type run func(
	args []string,
	globalFlags GlobalFlags,
	logger output.Logger,
	readerWriter flowkit.ReaderWriter,
	flow flowkit.Services,
) (Result, error)

// RunWithState runs the command with arguments and state.
type RunWithState func(
	args []string,
	globalFlags GlobalFlags,
	logger output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (Result, error)

type Command struct {
	Cmd   *cobra.Command
	Flags any
	Run   run
	RunS  RunWithState
}

const (
	FormatText   = "text"
	FormatInline = "inline"
	FormatJSON   = "json"
)

const (
	logLevelDebug = "debug"
	logLevelInfo  = "info"
	logLevelError = "error"
	logLevelNone  = "none"
)

// AddToParent add new command to main parent cmd
// and initializes all necessary things as well as take care of errors and output
// here we can do all boilerplate code that is else copied in each command and make sure
// we have one place to handle all errors and ensure commands have consistent results.
func (c Command) AddToParent(parent *cobra.Command) {
	// initialize crash reporting for the CLI
	initCrashReporting()

	c.Cmd.Run = func(cmd *cobra.Command, args []string) {
		if !isDevelopment() { // only report crashes in production
			defer sentry.Flush(2 * time.Second)
			defer sentry.Recover()
		}

		// initialize file loader used in commands
		loader := &afero.Afero{Fs: afero.NewOsFs()}

		// if we receive a config error that isn't missing config we should handle it
		state, confErr := flowkit.Load(Flags.ConfigPaths, loader)
		if !errors.Is(confErr, config.ErrDoesNotExist) {
			handleError("Config Error", confErr)
		}

		network, err := resolveHost(state, Flags.Host, Flags.HostNetworkKey, Flags.Network)
		handleError("Host Error", err)

		clientGateway, err := createGateway(*network)
		handleError("Gateway Error", err)

		logger := createLogger(Flags.Log, Flags.Format)

		// initialize services
		flow := flowkit.NewFlowkit(state, *network, clientGateway, logger)

		// skip version check if flag is set
		if !Flags.SkipVersionCheck {
			checkVersion(logger)
		}

		// record command usage
		wg := sync.WaitGroup{}
		go UsageMetrics(c.Cmd, &wg)

		// run command based on requirements for state
		var result Result
		if c.Run != nil {
			result, err = c.Run(args, Flags, logger, loader, flow)
		} else if c.RunS != nil {
			if confErr != nil {
				handleError("Config Error", confErr)
			}

			result, err = c.RunS(args, Flags, logger, flow, state)
		} else {
			panic("command implementation needs to provide run functionality")
		}

		handleError("Command Error", err)

		// Do not print a result if none is provided.
		//
		// This is useful for interactive commands that do not
		// require a printed summary (e.g. flow accounts create).
		if result == nil {
			return
		}

		// format output result
		formattedResult, err := formatResult(result, Flags.Filter, Flags.Format)
		handleError("Result", err)

		// output result
		err = outputResult(formattedResult, Flags.Save, Flags.Format, Flags.Filter)
		handleError("Output Error", err)

		wg.Wait()

		// exit with code if result has it
		exitCode := 0
		if res, ok := result.(ResultWithExitCode); ok {
			exitCode = res.ExitCode()
		}
		os.Exit(exitCode)
	}

	bindFlags(c)
	parent.AddCommand(c.Cmd)
}

// createGateway creates a gateway to be used, defaults to grpc but can support others.
func createGateway(network config.Network) (gateway.Gateway, error) {
	// create secure grpc client if hostNetworkKey provided
	if network.Key != "" {
		return gateway.NewSecureGrpcGateway(network)
	}

	return gateway.NewGrpcGateway(network)
}

// resolveHost from the flags provided.
//
// Resolve the network host in the following order:
// 1. if host flag is provided resolve to that host
// 2. if conf is initialized return host by network flag
// 3. if conf is not initialized and network flag is provided resolve to coded value for that network
// 4. default to emulator network
func resolveHost(state *flowkit.State, hostFlag, networkKeyFlag, networkFlag string) (*config.Network, error) {
	// host flag has the highest priority
	if hostFlag != "" {
		// if network-key was provided validate it
		if networkKeyFlag != "" {
			err := util.ValidateECDSAP256Pub(networkKeyFlag)
			if err != nil {
				return nil, fmt.Errorf("invalid network key %s: %w", networkKeyFlag, err)
			}
		}

		if state != nil {
			_, err := state.Networks().ByName(networkFlag)
			if err != nil {
				return nil, fmt.Errorf("network with name %s does not exist in configuration", networkFlag)
			}
		} else {
			networkFlag = "custom"
		}

		return &config.Network{Name: networkFlag, Host: hostFlag, Key: networkKeyFlag}, nil
	}

	// network flag with project initialized is next
	if state != nil {
		stateNetwork, err := state.Networks().ByName(networkFlag)
		if err != nil {
			return nil, fmt.Errorf("network with name %s does not exist in configuration", networkFlag)
		}

		return stateNetwork, nil
	}

	networks := config.DefaultNetworks
	network, err := networks.ByName(networkFlag)
	if err != nil {
		return nil, fmt.Errorf("invalid network with name %s", networkFlag)
	}

	return network, nil
}

// create logger utility.
func createLogger(logFlag string, formatFlag string) output.Logger {
	// disable logging if we user want a specific format like JSON
	// (more common they will not want also to have logs)
	if formatFlag != FormatText {
		logFlag = logLevelNone
	}

	var logLevel int

	switch logFlag {
	case logLevelDebug:
		logLevel = output.DebugLog
	case logLevelError:
		logLevel = output.ErrorLog
	case logLevelNone:
		logLevel = output.NoneLog
	default:
		logLevel = output.InfoLog
	}

	return output.NewStdoutLogger(logLevel)
}

// checkVersion fetches latest version and compares it to local.
func checkVersion(logger output.Logger) {
	currentVersion := build.Semver()
	if isDevelopment() {
		return // avoid warning in local development
	}

	// If using cadence-v1.0.0 pre-release, check for cadence-v1.0.0 releases instead
	if strings.Contains(currentVersion, "cadence-v1.0.0") {
		checkVersionCadence1(logger)
		return
	}

	resp, err := http.Get("https://formulae.brew.sh/api/formula/flow-cli.json")
	if err != nil || resp.StatusCode >= 400 {
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("error closing request")
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return
	}

	versions, ok := data["versions"].(map[string]interface{})
	if !ok {
		return
	}

	latestVersion, ok := versions["stable"].(string)
	if !ok {
		return
	}

	if currentVersion != latestVersion {
		logger.Info(fmt.Sprintf(
			"\n%s  Version warning: a new version of Flow CLI is available (v%s).\n"+
				"   Read the installation guide for upgrade instructions: https://docs.onflow.org/flow-cli/install\n",
			output.WarningEmoji(),
			strings.ReplaceAll(latestVersion, "\n", ""),
		))
	}
}

// checkVersionCadence1 fetches latest version of cadence-v1.0.0 and compares it to local.
// This is a special case for cadence-v1.0.0 pre-release & should be removed when cadence-v1.0.0 branch is merged.
func checkVersionCadence1(logger output.Logger) {
	resp, err := http.Get("https://api.github.com/repos/onflow/flow-cli/tags?per_page=100")
	if err != nil || resp.StatusCode >= 400 {
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("error closing request")
		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)
	var tags []map[string]interface{}
	err = json.Unmarshal(body, &tags)
	if err != nil {
		return
	}

	var latestVersion string
	for _, tag := range tags {
		tagName, ok := tag["name"].(string)
		if !ok {
			continue
		}

		if strings.Contains(tagName, "cadence-v1.0.0") {
			latestVersion = tagName
			break
		}
	}

	currentVersion := build.Semver()
	if currentVersion != latestVersion && latestVersion != "" {
		logger.Info(fmt.Sprintf(
			"\n%s  Version warning: a new version of Flow CLI is available (%s).\n"+
				"   Read the installation guide for upgrade instructions: https://cadence-lang.org/docs/cadence-migration-guide#install-cadence-10-cli\n",
			output.WarningEmoji(),
			strings.ReplaceAll(latestVersion, "\n", ""),
		))
	}
}

func isDevelopment() bool {
	return build.Semver() == "undefined"
}

// initCrashReporting set-ups sentry as crash reporting tool, it also sets listener for panics
// and asks before sending the error for a permission to do so from the user.
func initCrashReporting() {
	currentVersion := build.Semver()
	sentrySyncTransport := sentry.NewHTTPSyncTransport()
	sentrySyncTransport.Timeout = time.Second * 3

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              "https://f4e84ec91b1645779765bbe249b42311@o114654.ingest.sentry.io/6178538",
		Environment:      "Prod",
		Release:          currentVersion,
		AttachStacktrace: true,
		Transport:        sentrySyncTransport,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// ask for crash report permission
			fmt.Printf("\n%s Crash detected! %s\n\n", output.ErrorEmoji(), event.Message)

			if util.ReportCrash() {
				return event
			} else {
				fmt.Printf("\nPlease help us improve the Flow CLI by opening an issue on https://github.com/onflow/flow-cli/issues, \nand pasting the output as well as a description of the actions you took that resulted in this crash.\n\n")
				fmt.Println(hint.RecoveredException)
				fmt.Println(event.Threads, event.Fingerprint)
				fmt.Println(event.Contexts)
				fmt.Println(string(debug.Stack()))
			}

			return nil
		},
	})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err) // safest output method at this point
	}
}

// The token is injected at build-time using ldflags
var mixpanelToken = ""

func UsageMetrics(command *cobra.Command, wg *sync.WaitGroup) {
	if !settings.MetricsEnabled() || mixpanelToken == "" {
		return
	}
	wg.Add(1)
	client := mixpanel.New(mixpanelToken, "")

	// calculates a user ID that doesn't leak any personal information
	usr, _ := user.Current() // ignore err, just use empty string
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s%s", usr.Username, usr.Uid)))
	userID := base64.StdEncoding.EncodeToString(hash[:])

	_ = client.Track(userID, "cli-command", &mixpanel.Event{
		IP: "0", // do not track IPs
		Properties: map[string]any{
			"command": command.CommandPath(),
			"version": build.Semver(),
			"os":      runtime.GOOS,
		},
	})
	wg.Done()
}

// GlobalFlags contains all global flags definitions.
type GlobalFlags struct {
	Filter           string
	Format           string
	Save             string
	Host             string
	HostNetworkKey   string
	Log              string
	Network          string
	Yes              bool
	ConfigPaths      []string
	SkipVersionCheck bool
}
