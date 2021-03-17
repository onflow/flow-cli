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
	"fmt"
	"os"
	"strings"

	"github.com/onflow/flow-cli/pkg/flow"
	"github.com/onflow/flow-cli/pkg/flow/gateway"
	"github.com/onflow/flow-cli/pkg/flow/services"
	"github.com/onflow/flow-cli/pkg/flow/util"
	"github.com/onflow/flow-go-sdk/client"
	"github.com/psiemens/sconfig"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type Command interface {
	GetCmd() *cobra.Command
	GetFlags() *sconfig.Config
	Run(*cobra.Command, []string, *flow.Project, *services.Services) (Result, error)
}

var (
	FilterFlag      = ""
	FormatFlag      = ""
	SaveFlag        = ""
	RunEmulatorFlag = false
	HostFlag        = "127.0.0.1:3569"
	LogFlag         = util.InfoLog
)

// addCommand add new command to main cmd
// and initializes all necessary things as well as take care of errors and output
// here we can do all boilerplate code that is else copied in each command and make sure
// we have one place to handle all errors and ensure commands have consistent results
func Add(c *cobra.Command, command Command) {
	command.GetCmd().RunE = func(cmd *cobra.Command, args []string) error {
		// initialize project but ignore error since config can be missing
		project, _ := flow.LoadProject(flow.ConfigPath)

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
func createGateway(cmd *cobra.Command, project *flow.Project) (gateway.Gateway, error) {
	// create in memory emulator client
	if RunEmulatorFlag {
		return gateway.NewEmulatorGateway(), nil
	}

	// resolve host
	host := HostFlag
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
	if FormatFlag != "" {
		LogFlag = util.NoneLog
	}

	return util.NewStdoutLogger(LogFlag)
}

// outputResult takes care of showing the result
func formatResult(result Result) (string, error) {
	if result == nil {
		return "", fmt.Errorf("Missing")
	}

	if FilterFlag != "" {
		var jsonResult map[string]interface{}
		val, _ := json.Marshal(result.JSON())
		err := json.Unmarshal(val, &jsonResult)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%v", jsonResult[FilterFlag]), nil
	}

	switch FormatFlag {
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
	if SaveFlag != "" {
		af := afero.Afero{
			Fs: afero.NewOsFs(),
		}

		fmt.Printf("💾 result saved to: %s \n", SaveFlag)
		return af.WriteFile(SaveFlag, []byte(result), 0644)
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
func bindFlags(command Command) {
	err := command.GetFlags().
		FromEnvironment(flow.EnvPrefix).
		BindFlags(command.GetCmd().PersistentFlags()).
		Parse()
	if err != nil {
		fmt.Println(err)
	}
}
