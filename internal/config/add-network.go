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

package config

import (
	"fmt"
	"net/url"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowcli/config"
	"github.com/onflow/flow-cli/pkg/flowcli/output"
	"github.com/onflow/flow-cli/pkg/flowcli/services"
	"github.com/spf13/cobra"
)

type flagsAddNetwork struct {
	Name string `flag:"name" info:"Network name"`
	Host string `flag:"host" info:"Flow Access API host address"`
}

var addNetworkFlags = flagsAddNetwork{}

var AddNetworkCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "network",
		Short:   "Add network to configuration",
		Example: "flow config add network",
		Args:    cobra.NoArgs,
	},
	Flags: &addNetworkFlags,
	Run: func(
		cmd *cobra.Command,
		args []string,
		globalFlags command.GlobalFlags,
		services *services.Services,
	) (command.Result, error) {
		networkData, flagsProvided, err := flagsToNetworkData(addNetworkFlags)
		if err != nil {
			return nil, err
		}

		if !flagsProvided {
			networkData = output.NewNetworkPrompt()
		}

		networkData = output.NewNetworkPrompt()
		network := config.StringToNetwork(networkData["name"], networkData["host"])

		err = services.Config.AddNetwork(network)
		if err != nil {
			return nil, err
		}

		return &ConfigResult{
			result: "network added",
		}, nil
	},
}

func init() {
	AddNetworkCommand.AddToParent(AddCmd)
}

func flagsToNetworkData(flags flagsAddNetwork) (map[string]string, bool, error) {
	if flags.Name == "" && flags.Host == "" {
		return nil, false, nil
	}

	if flags.Name == "" {
		return nil, true, fmt.Errorf("name must be provided")
	} else if flags.Host == "" {
		return nil, true, fmt.Errorf("contract file name must be provided")
	}

	_, err := url.ParseRequestURI(flags.Host)
	if err != nil {
		return nil, true, err
	}

	return map[string]string{
		"name": flags.Name,
		"host": flags.Host,
	}, true, nil
}
