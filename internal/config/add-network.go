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

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
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
	RunS:  addNetwork,
}

func addNetwork(
	_ []string,
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {

	if !config.IsGlobalPath(globalFlags.ConfigPaths) && len(globalFlags.ConfigPaths) > 1 {
		return nil, fmt.Errorf("specifying multiple paths to -f is not supported when updating configuration")
	}

	if state == nil {
		return nil, config.ErrDoesNotExist
	}

	networkData, flagsProvided, err := flagsToNetworkData(addNetworkFlags)
	if err != nil {
		return nil, err
	}

	if !flagsProvided {
		networkData = output.NewNetworkPrompt()
	}

	network := config.StringToNetwork(networkData["name"], networkData["host"])
	state.Networks().AddOrUpdate(network.Name, network)

	err = state.SaveDefault()
	if err != nil {
		return nil, err
	}

	return &Result{
		result: fmt.Sprintf("Network %s added to the configuration", networkData["name"]),
	}, nil
}

func flagsToNetworkData(flags flagsAddNetwork) (map[string]string, bool, error) {
	if flags.Name == "" && flags.Host == "" {
		return nil, false, nil
	}

	if flags.Name == "" {
		return nil, true, fmt.Errorf("name must be provided")
	} else if flags.Host == "" {
		return nil, true, fmt.Errorf("host must be provided")
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
