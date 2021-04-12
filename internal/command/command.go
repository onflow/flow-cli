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
	"os"
	"strings"

	"github.com/onflow/flow-cli/pkg/flowcli/util"

	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/project"

	"github.com/onflow/flow-cli/pkg/flowcli/config"

	"github.com/psiemens/sconfig"

	"github.com/onflow/flow-cli/pkg/flowcli/gateway"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/onflow/flow-go-sdk/client"
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
	Filter  string
	Format  string
	Save    string
	Host    string
	Log     string
	Network string
	Yes     bool
	ConfigPath []string
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

var flags = GlobalFlags{
	Filter:  "",
	Format:     formatText,
	Save:    "",
	Host:    "",
	Log:        logLevelInfo,
	Network: "",
	Yes:     false,
	ConfigPath: project.DefaultConfigPaths,
}

// InitFlags init all the global persistent flags
func InitFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(
		&flags.Filter,
		"filter",
		"x",
		flags.Filter,
		"Filter result values by property name",
	)

	cmd.PersistentFlags().StringVarP(
		&flags.Host,
		"host",
		"",
		flags.Host,
		"Flow Access API host address",
	)

	cmd.PersistentFlags().StringVarP(
		&flags.Format,
		"output",
		"o",
		flags.Format,
		"Output format, options: \"text\", \"json\", \"inline\"",
	)

	cmd.PersistentFlags().StringVarP(
		&flags.Save,
		"save",
		"s",
		flags.Save,
		"Save result to a filename",
	)

	cmd.PersistentFlags().StringVarP(
		&flags.Log,
		"log",
		"l",
		flags.Log,
		"Log level, options: \"debug\", \"info\", \"error\", \"none\"",
	)

	cmd.PersistentFlags().StringSliceVarP(
		&flags.ConfigPath,
		"config-path",
		"f",
		flags.ConfigPath,
		"Path to flow configuration file",
	)

	cmd.PersistentFlags().StringVarP(
		&flags.Network,
		"network",
		"n",
		flags.Network,
		"Network from configuration file",
	)

	cmd.PersistentFlags().BoolVarP(
		&flags.Yes,
		"yes",
		"y",
		flags.Yes,
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
		proj, err := project.Load(flags.ConfigPath)
		// here we ignore if config does not exist as some commands don't require it
		if !errors.Is(err, config.ErrDoesNotExist) {
			handleError("Config Error", err)
		}

		host, err := resolveHost(proj, flags.Host, flags.Network)
		handleError("Host Error", err)

		clientGateway, err := createGateway(host)
		handleError("Gateway Error", err)

		logger := createLogger(flags.Log, flags.Format)

		service := services.NewServices(clientGateway, proj, logger)

		// run command
		result, err := c.Run(cmd, args, flags, service)
		handleError("Command Error", err)

		// format output result
		formattedResult, err := formatResult(result, flags.Filter, flags.Format)
		handleError("Result", err)

		// output result
		err = outputResult(formattedResult, flags.Save, flags.Format, flags.Filter)
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
func resolveHost(proj *project.Project, hostFlag string, networkFlag string) (string, error) {
	host := hostFlag
	if networkFlag != "" && proj != nil {
		check := proj.NetworkByName(networkFlag)
		if check == nil {
			return "", fmt.Errorf("provided network with name %s doesn't exists in condiguration", networkFlag)
		}

		host = proj.NetworkByName(networkFlag).Host
	} else if host == "" {
		host = config.DefaultEmulatorNetwork().Host
	}

	return host, nil
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

		fmt.Printf("💾 result saved to: %s \n", saveFlag)
		return af.WriteFile(saveFlag, []byte(result), 0644)
	}

	if formatFlag == "inline" || filterFlag != "" {
		fmt.Fprintf(os.Stdout, "%s", result)
	} else { // default normal output
		fmt.Fprintf(os.Stdout, "\n%s\n\n", result)
	}
	return nil
}

// filterResultValue returns a value by its name filtered from other result values
func filterResultValue(result Result, filter string) (interface{}, error) {
	var jsonResult map[string]interface{}
	val, _ := json.Marshal(result.JSON())
	err := json.Unmarshal(val, &jsonResult)
	if err != nil {
		return "", err
	}

	possibleFilters := make([]string, 0)
	for key := range jsonResult {
		possibleFilters = append(possibleFilters, key)
	}

	value := jsonResult[filter]

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
		fmt.Fprintf(os.Stderr, "❌ Grpc Error: %s \n", t.GRPCStatus().Err().Error())
	default:
		if errors.Is(err, config.ErrOutdatedFormat) {
			fmt.Fprintf(os.Stderr, "❌ Config Error: %s \n", err.Error())
			fmt.Fprintf(os.Stderr, "🙏 Please reset configuration using: 'flow init --reset'. Read more about new configuration here: https://github.com/onflow/flow-cli/releases/tag/v0.17.0")
		} else if strings.Contains(err.Error(), "transport:") {
			fmt.Fprintf(os.Stderr, "❌ %s \n", strings.Split(err.Error(), "transport:")[1])
			fmt.Fprintf(os.Stderr, "🙏 Make sure your emulator is running or connection address is correct.")
		} else if strings.Contains(err.Error(), "NotFound desc =") {
			fmt.Fprintf(os.Stderr, "❌ Not Found:%s \n", strings.Split(err.Error(), "NotFound desc =")[1])
		} else if strings.Contains(err.Error(), "code = InvalidArgument desc = ") {
			desc := strings.Split(err.Error(), "code = InvalidArgument desc = ")
			fmt.Fprintf(os.Stderr, "❌ Invalid argument: %s \n", desc[len(desc)-1])
			if strings.Contains(err.Error(), "is invalid for chain") {
				fmt.Fprintf(os.Stderr, "🙏 Check you are connecting to the correct network or account address you use is correct.")
			} else {
				fmt.Fprintf(os.Stderr, "🙏 Check your argument and flags value, you can use --help.")
			}
		} else if strings.Contains(err.Error(), "invalid signature:") {
			fmt.Fprintf(os.Stderr, "❌ Invalid signature: %s \n", strings.Split(err.Error(), "invalid signature:")[1])
			fmt.Fprintf(os.Stderr, "🙏 Check the signer private key is provided or is in the correct format. If running emulator, make sure it's using the same configuration as this command.")
		} else if strings.Contains(err.Error(), "signature could not be verified using public key with") {
			fmt.Fprintf(os.Stderr, "❌ %s: %s \n", description, err)
			fmt.Fprintf(os.Stderr, "🙏 If you are running emulator locally make sure that the emulator was started with the same config as used in this command. \nTry restarting the emulator.")
		} else {
			fmt.Fprintf(os.Stderr, "❌ %s: %s", description, err)
		}
	}

	fmt.Println()
	os.Exit(1)
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
