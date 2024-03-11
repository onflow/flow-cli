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

package config

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

type flagsAddNetwork struct {
	Name string `flag:"name" info:"Network name"`
	Host string `flag:"host" info:"Flow Access API host address"`
	Key  string `flag:"network-key" info:"Flow Access API host network key for secure client connections"`
}

var addNetworkFlags = flagsAddNetwork{}

var addNetworkCommand = &command.Command{
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
	globalFlags command.GlobalFlags,
	_ output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	raw, flagsProvided, err := flagsToNetworkData(addNetworkFlags)
	if err != nil {
		return nil, err
	}

	if !flagsProvided {
		raw = util.NewNetworkPrompt()
	}

	state.Networks().AddOrUpdate(config.Network{
		Name: raw["name"],
		Host: raw["host"],
		Key:  raw["key"],
	})

	err = state.SaveEdited(globalFlags.ConfigPaths)
	if err != nil {
		return nil, err
	}

	return &result{
		result: fmt.Sprintf("Network %s added to the configuration", raw["name"]),
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

	err = util.ValidateECDSAP256Pub(flags.Key)
	if err != nil {
		return nil, true, fmt.Errorf("invalid network-key provided")
	}

	return map[string]string{
		"name": flags.Name,
		"host": flags.Host,
		"key":  flags.Key,
	}, true, nil
}
