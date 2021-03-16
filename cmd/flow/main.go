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

// Package main implements the entry point for the Flow CLI.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/onflow/flow-cli/cmd/blocks"
	"github.com/onflow/flow-cli/cmd/emulator"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/cmd"
	"github.com/onflow/flow-cli/cmd/accounts"
	"github.com/onflow/flow-cli/cmd/cadence"
	"github.com/onflow/flow-cli/cmd/collections"
	"github.com/onflow/flow-cli/cmd/events"
	"github.com/onflow/flow-cli/cmd/keys"
	"github.com/onflow/flow-cli/cmd/project"
	"github.com/onflow/flow-cli/cmd/scripts"
	"github.com/onflow/flow-cli/cmd/transactions"
	"github.com/onflow/flow-cli/cmd/version"
	"github.com/onflow/flow-cli/flow/gateway"
	"github.com/onflow/flow-cli/flow/lib"
	"github.com/onflow/flow-cli/flow/services"
	"github.com/onflow/flow-cli/flow/util"
	"github.com/onflow/flow-go-sdk/client"
)

var c = &cobra.Command{
	Use:              "flow",
	TraverseChildren: true,
}

var (
	filterFlag      = ""
	formatFlag      = ""
	saveFlag        = ""
	runEmulatorFlag = false
	hostFlag        = "127.0.0.1:3569"
	logFlag         = util.InfoLog
)

func init() {

	c.AddCommand(cadence.Cmd)
	c.AddCommand(version.Cmd)
	c.AddCommand(emulator.Cmd)

	c.AddCommand(accounts.Cmd)
	addCommand(accounts.Cmd, accounts.NewGetCmd())
	addCommand(accounts.Cmd, accounts.NewCreateCmd())
	addCommand(accounts.Cmd, accounts.NewAddContractCmd())
	addCommand(accounts.Cmd, accounts.NewRemoveContractCmd())
	addCommand(accounts.Cmd, accounts.NewUpdateContractCmd())

	c.AddCommand(scripts.Cmd)
	addCommand(scripts.Cmd, scripts.NewExecuteScriptCmd())

	c.AddCommand(transactions.Cmd)
	addCommand(transactions.Cmd, transactions.NewSendCmd())
	addCommand(transactions.Cmd, transactions.NewStatusCmd())

	c.AddCommand(keys.Cmd)
	addCommand(keys.Cmd, keys.NewGenerateCmd())

	c.AddCommand(events.Cmd)
	addCommand(events.Cmd, events.NewGetCmd())

	c.AddCommand(blocks.Cmd)
	addCommand(blocks.Cmd, blocks.NewGetCmd())

	c.AddCommand(collections.Cmd)
	addCommand(collections.Cmd, collections.NewGetCmd())

	c.AddCommand(project.Cmd)
	addCommand(project.Cmd, project.NewInitCmd())
	addCommand(project.Cmd, project.NewDeployCmd())

	c.PersistentFlags().StringVarP(&hostFlag, "host", "", hostFlag, "Flow Access API host address")
	c.PersistentFlags().StringVarP(&filterFlag, "filter", "", filterFlag, "Filter result values by property name")
	c.PersistentFlags().StringVarP(&formatFlag, "format", "", formatFlag, "Format to show result in")
	c.PersistentFlags().StringVarP(&saveFlag, "save", "", saveFlag, "Save result to a filename")
	c.PersistentFlags().StringVarP(&logFlag, "log", "", logFlag, "Logging level")
	c.PersistentFlags().BoolVarP(&runEmulatorFlag, "emulator", "", runEmulatorFlag, "Run in-memory emulator")
	c.PersistentFlags().StringSliceVarP(&lib.ConfigPath, "config-path", "f", lib.ConfigPath, "Path to flow configuration file")
}

// addCommand add new command to main cmd
// and initializes all necessary things as well as take care of errors and output
// here we can do all boilerplate code that is else copied in each command and make sure
// we have one place to handle all errors and ensure commands have consistent results
func addCommand(c *cobra.Command, command cmd.Command) {
	command.GetCmd().RunE = func(cmd *cobra.Command, args []string) error {
		// initialize project but ignore error since config can be missing
		project, _ := lib.LoadProject(lib.ConfigPath)

		gateway, err := createGateway(cmd, project)
		handleError("Gateway Error", err)

		logger := createLogger()

		service := services.NewServices(gateway, project, logger)

		// run command
		result, err := command.Run(cmd, args, project, service)
		handleError("Command Error", err)

		// format output result
		formattedResult, err := formatResult(result)
		handleError("Result", err)

		// output result
		err = outputResult(formattedResult)
		handleError("Output Error", err)

		return nil
	}

	bindFlags(command)
	c.AddCommand(command.GetCmd())
}

// createGateway creates a gateway to be used, defaults to grpc but can support others
func createGateway(cmd *cobra.Command, project *lib.Project) (gateway.Gateway, error) {
	// create in memory emulator client
	if runEmulatorFlag {
		return gateway.NewEmulatorGateway(), nil
	}

	// resolve host
	host := hostFlag
	if host == "" && project != nil {
		host = project.Host("emulator")
	} else if host == "" {
		return nil, fmt.Errorf("Host must be provided using --host flag or in config by initializing project: flow project init")
	}

	// create default grpc client
	return gateway.NewGrpcGateway(host)
}

// create logger utility
func createLogger() util.Logger {
	// disable logging if we user want a specific format like JSON
	//(more common they will not want also to have logs)
	if formatFlag != "" {
		logFlag = util.NoneLog
	}

	return util.NewStdoutLogger(logFlag)
}

// outputResult takes care of showing the result
func formatResult(result cmd.Result) (string, error) {
	if result == nil {
		return "", fmt.Errorf("Missing")
	}

	if filterFlag != "" {
		var jsonResult map[string]interface{}
		val, _ := json.Marshal(result.JSON())
		err := json.Unmarshal(val, &jsonResult)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%v", jsonResult[filterFlag]), nil
	}

	switch formatFlag {
	case "json":
		jsonRes, _ := json.Marshal(result.JSON())
		return string(jsonRes), nil
	case "inline":
		return result.Oneliner(), nil
	default:
		return result.String(), nil
	}
}

// outputResult to selected media
func outputResult(result string) error {
	if saveFlag != "" {
		af := afero.Afero{
			Fs: afero.NewOsFs(),
		}

		fmt.Printf("💾 result saved to: %s \n", saveFlag)
		return af.WriteFile(saveFlag, []byte(result), 0644)
	}

	// default normal output
	fmt.Fprintf(os.Stdout, "%s\n", result)
	return nil
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
		fmt.Fprintf(os.Stderr, "❌  Grpc Error: %s \n", t.GRPCStatus().Err().Error())
	default:
		if strings.Contains(err.Error(), "transport:") {
			fmt.Fprintf(os.Stderr, "❌  %s \n", strings.Split(err.Error(), "transport:")[1])
		} else if strings.Contains(err.Error(), "NotFound desc =") {
			fmt.Fprintf(os.Stderr, "❌  Not Found:%s \n", strings.Split(err.Error(), "NotFound desc =")[1])
		} else if strings.Contains(err.Error(), "code = InvalidArgument desc = ") {
			fmt.Fprintf(os.Stderr, "❌  Invalid argument: %s \n", strings.Split(err.Error(), "code = InvalidArgument desc = ")[1])
		} else {
			fmt.Fprintf(os.Stderr, "❌  %s: %s", description, err)
		}
	}

	fmt.Println()
	os.Exit(1)
}

// bindFlags bind all the flags needed
func bindFlags(command cmd.Command) {
	err := command.GetFlags().
		FromEnvironment(lib.EnvPrefix).
		BindFlags(command.GetCmd().PersistentFlags()).
		Parse()
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	if err := c.Execute(); err != nil {
		lib.Exit(1, err.Error())
	}
}
