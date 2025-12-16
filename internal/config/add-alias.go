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

package config

import (
	"fmt"

	"github.com/onflow/flow-cli/internal/prompt"
	"github.com/onflow/flow-cli/internal/util"

	flow "github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsAddAlias struct {
	Contract string `flag:"contract" info:"Name of the contract to add alias for"`
	Network  string `flag:"network" info:"Network name for the alias"`
	Address  string `flag:"address" info:"Address for the alias"`
}

var addAliasFlags = flagsAddAlias{}

var addAliasCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "alias",
		Short:   "Add alias to contract configuration",
		Example: "flow config add alias --contract MyContract --network testnet --address 0x1234567890abcdef",
		Args:    cobra.NoArgs,
	},
	Flags: &addAliasFlags,
	RunS:  addAlias,
}

func addAlias(
	_ []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	raw, flagsProvided, err := flagsToAliasData(addAliasFlags)
	if err != nil {
		return nil, err
	}

	if !flagsProvided {
		raw = prompt.NewAliasPrompt()
	}

	contract, err := state.Contracts().ByName(raw.Contract)
	if err != nil {
		return nil, fmt.Errorf("contract %s not found in configuration: %w", raw.Contract, err)
	}

	contract.Aliases.Add(
		raw.Network,
		flow.HexToAddress(raw.Address),
	)

	state.Contracts().AddOrUpdate(*contract)

	err = state.SaveEdited(globalFlags.ConfigPaths)
	if err != nil {
		return nil, err
	}

	return &result{
		result: fmt.Sprintf("Alias for contract %s on network %s added to the configuration", raw.Contract, raw.Network),
	}, nil
}

func flagsToAliasData(flags flagsAddAlias) (*prompt.AliasData, bool, error) {
	if flags.Contract == "" && flags.Network == "" && flags.Address == "" {
		return nil, false, nil
	}

	if flags.Contract == "" {
		return nil, true, fmt.Errorf("contract name must be provided")
	}

	if flags.Network == "" {
		return nil, true, fmt.Errorf("network name must be provided")
	}

	if flags.Address == "" {
		return nil, true, fmt.Errorf("address must be provided")
	}

	// Validate address is valid for the specified network
	if !util.IsAddressValidForNetwork(address, flags.Network) {
		return nil, true, fmt.Errorf("address %s is not valid for network %s", flags.Address, flags.Network)
	}

	return &prompt.AliasData{
		Contract: flags.Contract,
		Network:  flags.Network,
		Address:  flags.Address,
	}, true, nil
}
