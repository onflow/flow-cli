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

	"github.com/onflow/flow-cli/internal/prompt"

	"github.com/onflow/flow-go-sdk"
	"github.com/spf13/cobra"

	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"

	"github.com/onflow/flow-cli/internal/command"
)

type flagsAddContract struct {
	Name          string `flag:"name" info:"Name of the contract"`
	Filename      string `flag:"filename" info:"Filename of the contract source"`
	EmulatorAlias string `flag:"emulator-alias" info:"Address for the emulator alias"`
	TestnetAlias  string `flag:"testnet-alias" info:"Address for the testnet alias"`
	MainnetAlias  string `flag:"mainnet-alias" info:"Address for the mainnet alias"`
}

var addContractFlags = flagsAddContract{}

var addContractCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "contract",
		Short:   "Add contract to configuration",
		Example: "flow config add contract",
		Args:    cobra.NoArgs,
	},
	Flags: &addContractFlags,
	RunS:  addContract,
}

func addContract(
	_ []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	_ flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	raw, flagsProvided, err := flagsToContractData(addContractFlags)
	if err != nil {
		return nil, err
	}

	if !flagsProvided {
		raw = prompt.NewContractPrompt()
	}

	contract := config.Contract{
		Name:     raw.Name,
		Location: raw.Source,
	}

	if raw.Emulator != "" {
		contract.Aliases.Add(
			config.EmulatorNetwork.Name,
			flow.HexToAddress(raw.Emulator),
		)
	}

	if raw.Mainnet != "" {
		contract.Aliases.Add(
			config.MainnetNetwork.Name,
			flow.HexToAddress(raw.Mainnet),
		)
	}

	if raw.Testnet != "" {
		contract.Aliases.Add(
			config.TestnetNetwork.Name,
			flow.HexToAddress(raw.Testnet),
		)
	}

	if raw.Previewnet != "" {
		contract.Aliases.Add(
			config.PreviewnetNetwork.Name,
			flow.HexToAddress(raw.Mainnet),
		)
	}

	state.Contracts().AddOrUpdate(contract)

	err = state.SaveEdited(globalFlags.ConfigPaths)
	if err != nil {
		return nil, err
	}

	return &result{
		result: fmt.Sprintf("Contract %s added to the configuration", raw.Name),
	}, nil
}

func flagsToContractData(flags flagsAddContract) (*prompt.ContractData, bool, error) {
	if flags.Name == "" && flags.Filename == "" {
		return nil, false, nil
	}

	if flags.Name == "" {
		return nil, true, fmt.Errorf("name must be provided")
	}

	if flags.Filename == "" {
		return nil, true, fmt.Errorf("contract file name must be provided")
	}

	if !config.Exists(flags.Filename) {
		return nil, true, fmt.Errorf("contract file doesn't exist: %s", flags.Filename)
	}

	if flags.EmulatorAlias != "" && flow.HexToAddress(flags.EmulatorAlias) == flow.EmptyAddress {
		return nil, true, fmt.Errorf("invalid emulator alias address")
	}

	if flags.TestnetAlias != "" && flow.HexToAddress(flags.TestnetAlias) == flow.EmptyAddress {
		return nil, true, fmt.Errorf("invalid testnet alias address")
	}

	if flags.MainnetAlias != "" && flow.HexToAddress(flags.MainnetAlias) == flow.EmptyAddress {
		return nil, true, fmt.Errorf("invalid mainnnet alias address")
	}

	return &prompt.ContractData{
		Name:     flags.Name,
		Source:   flags.Filename,
		Emulator: flags.EmulatorAlias,
		Testnet:  flags.TestnetAlias,
		Mainnet:  flags.MainnetAlias,
	}, true, nil
}
