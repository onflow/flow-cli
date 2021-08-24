/*
 * Flow CLI
 *
 * Copyright 2019-2021 Dapper Labs, Inc.
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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/spf13/afero"

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/onflow/flow-cli/build"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/gateway"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"

	"github.com/spf13/cobra"
)

// Run the command with arguments.
type Run func(
	args []string,
	readerWriter flowkit.ReaderWriter,
	globalFlags GlobalFlags,
	services *services.Services,
) (Result, error)

// RunWithState runs the command with arguments and state.
type RunWithState func(
	args []string,
	readerWriter flowkit.ReaderWriter,
	globalFlags GlobalFlags,
	services *services.Services,
	state *flowkit.State,
) (Result, error)

type Command struct {
	Cmd   *cobra.Command
	Flags interface{}
	Run   Run
	RunS  RunWithState
}

const (
	formatText   = "text"
	formatInline = "inline"
	formatJSON   = "json"
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
	c.Cmd.Run = func(cmd *cobra.Command, args []string) {
		// initialize file loader used in commands
		loader := &afero.Afero{Fs: afero.NewOsFs()}

		// if we receive a config error that isn't missing config we should handle it
		state, confErr := flowkit.Load(Flags.ConfigPaths, loader)
		if !errors.Is(confErr, config.ErrDoesNotExist) {
			handleError("Config Error", confErr)
		}

		host, err := resolveHost(state, Flags.Host, Flags.Network)
		handleError("Host Error", err)

		clientGateway, err := createGateway(host)
		handleError("Gateway Error", err)

		logger := createLogger(Flags.Log, Flags.Format)

		// initialize services
		service := services.NewServices(clientGateway, state, logger)

		checkVersion(logger)

		// run command based on requirements for state
		var result Result
		if c.Run != nil {
			result, err = c.Run(args, loader, Flags, service)
		} else if c.RunS != nil {
			if confErr != nil {
				handleError("Config Error", confErr)
			}

			result, err = c.RunS(args, loader, Flags, service, state)
		} else {
			panic("command implementation needs to provide run functionality")
		}

		handleError("Command Error", err)

		// format output result
		formattedResult, err := formatResult(result, Flags.Filter, Flags.Format)
		handleError("Result", err)

		// output result
		err = outputResult(formattedResult, Flags.Save, Flags.Format, Flags.Filter)
		handleError("Output Error", err)
	}

	bindFlags(c)
	parent.AddCommand(c.Cmd)
}

// createGateway creates a gateway to be used, defaults to grpc but can support others.
func createGateway(host string) (gateway.Gateway, error) {
	// create default grpc client
	return gateway.NewGrpcGateway(host)
}

// resolveHost from the flags provided.
//
// Resolve the network host in the following order:
// 1. if host flag is provided resolve to that host
// 2. if conf is initialized return host by network flag
// 3. if conf is not initialized and network flag is provided resolve to coded value for that network
// 4. default to emulator network
func resolveHost(state *flowkit.State, hostFlag string, networkFlag string) (string, error) {
	// don't allow both network and host flag as the host might be different
	if networkFlag != config.DefaultEmulatorNetwork().Name && hostFlag != "" {
		return "", fmt.Errorf("shouldn't use both host and network flags, better to use network flag")
	}

	// host flag has highest priority
	if hostFlag != "" {
		return hostFlag, nil
	}
	// network flag with project initialized is next
	if state != nil {
		stateNetwork, err := state.Networks().ByName(networkFlag)
		if err != nil {
			return "", fmt.Errorf("network with name %s does not exist in configuration", networkFlag)
		}

		return stateNetwork.Host, nil
	}

	networks := config.DefaultNetworks()
	network, err := networks.ByName(networkFlag)

	if err != nil {
		return "", fmt.Errorf("invalid network with name %s", networkFlag)
	}

	return network.Host, nil

}

// create logger utility.
func createLogger(logFlag string, formatFlag string) output.Logger {
	// disable logging if we user want a specific format like JSON
	// (more common they will not want also to have logs)
	if formatFlag != formatText {
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
	resp, err := http.Get("https://raw.githubusercontent.com/onflow/flow-cli/master/version.txt")
	if err != nil || resp.StatusCode >= 400 {
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("error closing request")
		}
	}(resp.Body)

	body, _ := ioutil.ReadAll(resp.Body)
	latestVersion := strings.TrimSpace(string(body))

	currentVersion := build.Semver()
	if currentVersion == "undefined" {
		return // avoid warning in local development
	}

	if currentVersion != latestVersion {
		logger.Info(fmt.Sprintf(
			"\n%s  Version warning: a new version of Flow CLI is available (%s).\n"+
				"   Read the installation guide for upgrade instructions: https://docs.onflow.org/flow-cli/install\n",
			output.WarningEmoji(),
			strings.ReplaceAll(latestVersion, "\n", ""),
		))
	}
}
