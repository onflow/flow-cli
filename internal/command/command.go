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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/onflow/flow-cli/build"
	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/onflow/flow-cli/pkg/flowcli/util"

	"github.com/onflow/flow-go-sdk/client"
	"github.com/psiemens/sconfig"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type RunCommand func(
	*cobra.Command,
	[]string,
	GlobalFlags,
	*services.Services,
) (Result, error)

type Command struct {
	Cmd   *cobra.Command
	Flags interface{}
	Run   RunCommand
}

type GlobalFlags struct {
	Filter      string
	Format      string
	Save        string
	Host        string
	Log         string
	Network     string
	Yes         bool
	ConfigPaths []string
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

var Flags = GlobalFlags{
	Filter:      "",
	Format:      formatText,
	Save:        "",
	Host:        "",
	Network:     config.DefaultEmulatorNetwork().Name,
	Log:         logLevelInfo,
	Yes:         false,
	ConfigPaths: config.DefaultPaths(),
}

// InitFlags init all the global persistent flags
func InitFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(
		&Flags.Filter,
		"filter",
		"x",
		Flags.Filter,
		"Filter result values by property name",
	)

	cmd.PersistentFlags().StringVarP(
		&Flags.Host,
		"host",
		"",
		Flags.Host,
		"Flow Access API host address",
	)

	cmd.PersistentFlags().StringVarP(
		&Flags.Format,
		"output",
		"o",
		Flags.Format,
		"Output format, options: \"text\", \"json\", \"inline\"",
	)

	cmd.PersistentFlags().StringVarP(
		&Flags.Save,
		"save",
		"s",
		Flags.Save,
		"Save result to a filename",
	)

	cmd.PersistentFlags().StringVarP(
		&Flags.Log,
		"log",
		"l",
		Flags.Log,
		"Log level, options: \"debug\", \"info\", \"error\", \"none\"",
	)

	cmd.PersistentFlags().StringSliceVarP(
		&Flags.ConfigPaths,
		"config-path",
		"f",
		Flags.ConfigPaths,
		"Path to flow configuration file",
	)

	cmd.PersistentFlags().StringVarP(
		&Flags.Network,
		"network",
		"n",
		Flags.Network,
		"Network from configuration file",
	)

	cmd.PersistentFlags().BoolVarP(
		&Flags.Yes,
		"yes",
		"y",
		Flags.Yes,
		"Approve any prompts",
	)
}

// AddToParent add new command to main parent cmd
// and initializes all necessary things as well as take care of errors and output
// here we can do all boilerplate code that is else copied in each command and make sure
// we have one place to handle all errors and ensure commands have consistent results
func (c Command) AddToParent(parent *cobra.Command) {
	c.Cmd.Run = func(cmd *cobra.Command, args []string) {
		// initialize project but ignore error since config can be missing
		proj, err := project.Load(Flags.ConfigPaths)

		// here we ignore if config does not exist as some commands don't require it
		if !errors.Is(err, config.ErrDoesNotExist) && cmd.CommandPath() != "flow init" { // ignore configs errors if we are doing init config
			handleError("Config Error", err)
		}

		host, err := resolveHost(proj, Flags.Host, Flags.Network)
		handleError("Host Error", err)

		clientGateway, err := createGateway(host)
		handleError("Gateway Error", err)

		logger := createLogger(Flags.Log, Flags.Format)

		service := services.NewServices(clientGateway, proj, logger)

		checkVersion(logger)

		// run command
		result, err := c.Run(cmd, args, Flags, service)
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

// createGateway creates a gateway to be used, defaults to grpc but can support others
func createGateway(host string) (gateway.Gateway, error) {
	// TODO implement emulator gateway and check emulator flag here

	// create default grpc client
	return gateway.NewGrpcGateway(host)
}

// resolveHost from the flags provided
//
// Resolve the network host in the following order:
// 1. if host flag is provided resolve to that host
// 2. if conf is initialized return host by network flag
// 3. if conf is not initialized and network flag is provided resolve to coded value for that network
// 4. default to emulator network
func resolveHost(proj *project.Project, hostFlag string, networkFlag string) (string, error) {
	// don't allow both network and host flag as the host might be different
	if networkFlag != config.DefaultEmulatorNetwork().Name && hostFlag != "" {
		return "", fmt.Errorf("shouldn't use both host and network flags, better to use network flag")
	}

	// host flag has highest priority
	if hostFlag != "" {
		return hostFlag, nil
	}
	// network flag with project initialized is next
	if proj != nil {
		check := proj.NetworkByName(networkFlag)
		if check == nil {
			return "", fmt.Errorf("network with name %s does not exist in configuration", networkFlag)
		}

		return proj.NetworkByName(networkFlag).Host, nil
	}

	networks := config.DefaultNetworks()
	network := networks.GetByName(networkFlag)

	if network != nil {
		return network.Host, nil
	}

	return "", fmt.Errorf("invalid network with name %s", networkFlag)
}

// create logger utility
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

// formatResult formats a result for printing.
func formatResult(result Result, filterFlag string, formatFlag string) (string, error) {
	if result == nil {
		return "", fmt.Errorf("missing result")
	}

	if filterFlag != "" {
		value, err := filterResultValue(result, filterFlag)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%v", value), nil
	}

	switch strings.ToLower(formatFlag) {
	case formatJSON:
		jsonRes, _ := json.Marshal(result.JSON())
		return string(jsonRes), nil
	case formatInline:
		return result.Oneliner(), nil
	default:
		return result.String(), nil
	}
}

// outputResult to selected media
func outputResult(result string, saveFlag string, formatFlag string, filterFlag string) error {
	if saveFlag != "" {
		af := afero.Afero{
			Fs: afero.NewOsFs(),
		}

		fmt.Printf("%s result saved to: %s \n", output.SaveEmoji(), saveFlag)
		return af.WriteFile(saveFlag, []byte(result), 0644)
	}

	if formatFlag == formatInline || filterFlag != "" {
		_, _ = fmt.Fprintf(os.Stdout, "%s", result)
	} else { // default normal output
		_, _ = fmt.Fprintf(os.Stdout, "\n%s\n\n", result)
	}
	return nil
}

// filterResultValue returns a value by its name filtered from other result values
func filterResultValue(result Result, filter string) (interface{}, error) {
	var jsonResult map[string]interface{}
	val, err := json.Marshal(result.JSON())
	if err != nil {
		return "", fmt.Errorf("not possible to filter by the value")
	}

	err = json.Unmarshal(val, &jsonResult)
	if err != nil {
		return "", fmt.Errorf("not possible to filter by the value")
	}

	possibleFilters := make([]string, 0)
	for key := range jsonResult {
		possibleFilters = append(possibleFilters, key)
	}

	value := jsonResult[filter]
	if value == nil {
		value = jsonResult[strings.ToLower(filter)]
	}

	if value == nil {
		return nil, fmt.Errorf("value for filter: '%s' doesn't exists, possible values to filter by: %s", filter, possibleFilters)
	}

	return value, nil
}

// handleError handle errors
func handleError(description string, err error) {
	if err == nil {
		return
	}

	// TODO: refactor this to better handle errors not by string matching
	// handle rpc error
	switch t := err.(type) {
	case *client.RPCError:
		_, _ = fmt.Fprintf(os.Stderr, "%s Grpc Error: %s \n", output.ErrorEmoji(), t.GRPCStatus().Err().Error())
	default:
		if errors.Is(err, config.ErrOutdatedFormat) {
			_, _ = fmt.Fprintf(os.Stderr, "%s Config Error: %s \n", output.ErrorEmoji(), err.Error())
			_, _ = fmt.Fprintf(os.Stderr, "%s Please reset configuration using: 'flow init --reset'. Read more about new configuration here: https://github.com/onflow/flow-cli/releases/tag/v0.17.0", output.TryEmoji())
		} else if strings.Contains(err.Error(), "transport:") {
			_, _ = fmt.Fprintf(os.Stderr, "%s %s \n", output.ErrorEmoji(), strings.Split(err.Error(), "transport:")[1])
			_, _ = fmt.Fprintf(os.Stderr, "%s Make sure your emulator is running or connection address is correct.", output.TryEmoji())
		} else if strings.Contains(err.Error(), "NotFound desc =") {
			_, _ = fmt.Fprintf(os.Stderr, "%s Not Found:%s \n", output.ErrorEmoji(), strings.Split(err.Error(), "NotFound desc =")[1])
		} else if strings.Contains(err.Error(), "code = InvalidArgument desc = ") {
			desc := strings.Split(err.Error(), "code = InvalidArgument desc = ")
			_, _ = fmt.Fprintf(os.Stderr, "%s Invalid argument: %s \n", output.ErrorEmoji(), desc[len(desc)-1])
			if strings.Contains(err.Error(), "is invalid for chain") {
				_, _ = fmt.Fprintf(os.Stderr, "%s Check you are connecting to the correct network or account address you use is correct.", output.TryEmoji())
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "%s Check your argument and flags value, you can use --help.", output.TryEmoji())
			}
		} else if strings.Contains(err.Error(), "invalid signature:") {
			_, _ = fmt.Fprintf(os.Stderr, "%s Invalid signature: %s \n", output.ErrorEmoji(), strings.Split(err.Error(), "invalid signature:")[1])
			_, _ = fmt.Fprintf(os.Stderr, "%s Check the signer private key is provided or is in the correct format. If running emulator, make sure it's using the same configuration as this command.", output.TryEmoji())
		} else if strings.Contains(err.Error(), "signature could not be verified using public key with") {
			_, _ = fmt.Fprintf(os.Stderr, "%s %s: %s \n", output.ErrorEmoji(), description, err)
			_, _ = fmt.Fprintf(os.Stderr, "%s If you are running emulator locally make sure that the emulator was started with the same config as used in this command. \nTry restarting the emulator.", output.TryEmoji())
		} else {
			_, _ = fmt.Fprintf(os.Stderr, "%s %s: %s", output.ErrorEmoji(), description, err)
		}
	}

	fmt.Println()
	os.Exit(1)
}

// checkVersion fetches latest version and compares it to local
func checkVersion(logger output.Logger) {
	resp, err := http.Get("https://raw.githubusercontent.com/onflow/flow-cli/master/version.txt")
	if err != nil || resp.StatusCode >= 400 {
		return
	}

	defer resp.Body.Close()
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
			strings.ReplaceAll(string(latestVersion), "\n", ""),
		))
	}
}

// bindFlags bind all the flags needed
func bindFlags(command Command) {
	err := sconfig.New(command.Flags).
		FromEnvironment(util.EnvPrefix).
		BindFlags(command.Cmd.PersistentFlags()).
		Parse()
	if err != nil {
		fmt.Println(err)
	}
}
