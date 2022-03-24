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

	"github.com/onflow/flow-cli/pkg/flowkit"

	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/pkg/flowkit/config"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/onflow/flow-cli/pkg/flowkit/services"
)

type flagsAddContract struct {
	Name          string `flag:"name" info:"Name of the contract"`
	Filename      string `flag:"filename" info:"Filename of the contract source"`
	EmulatorAlias string `flag:"emulator-alias" info:"Address for the emulator alias"`
	TestnetAlias  string `flag:"testnet-alias" info:"Address for the testnet alias"`
	MainnetAlias  string `flag:"mainnet-alias" info:"Address for the mainnet alias"`
}

var addContractFlags = flagsAddContract{}

var AddContractCommand = &command.Command{
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
	_ flowkit.ReaderWriter,
	globalFlags command.GlobalFlags,
	_ *services.Services,
	state *flowkit.State,
) (command.Result, error) {
	contractData, flagsProvided, err := flagsToContractData(addContractFlags)
	if err != nil {
		return nil, err
	}

	if !flagsProvided {
		contractData = output.NewContractPrompt()
	}

	contracts := config.StringToContracts(
		contractData["name"],
		contractData["source"],
		contractData["emulator"],
		contractData["testnet"],
		contractData["mainnet"],
	)

	for _, contract := range contracts {
		state.Contracts().AddOrUpdate(contract.Name, contract)
	}

	err = state.SaveEdited(globalFlags.ConfigPaths)
	if err != nil {
		return nil, err
	}

	return &Result{
		result: fmt.Sprintf("Contract %s added to the configuration", contractData["name"]),
	}, nil
}

func flagsToContractData(flags flagsAddContract) (map[string]string, bool, error) {
	if flags.Name == "" && flags.Filename == "" {
		return nil, false, nil
	}

	if flags.Name == "" {
		return nil, true, fmt.Errorf("name must be provided")
	} else if flags.Filename == "" {
		return nil, true, fmt.Errorf("contract file name must be provided")
	} else if !config.Exists(flags.Filename) {
		return nil, true, fmt.Errorf("contract file doesn't exist: %s", flags.Filename)
	}

	if flags.EmulatorAlias != "" {
		_, err := config.StringToAddress(flags.EmulatorAlias)
		if err != nil {
			return nil, true, err
		}
	}

	if flags.TestnetAlias != "" {
		_, err := config.StringToAddress(flags.TestnetAlias)
		if err != nil {
			return nil, true, err
		}
	}

	if flags.MainnetAlias != "" {
		_, err := config.StringToAddress(flags.MainnetAlias)
		if err != nil {
			return nil, true, err
		}
	}

	return map[string]string{
		"name":     flags.Name,
		"source":   flags.Filename,
		"emulator": flags.EmulatorAlias,
		"testnet":  flags.TestnetAlias,
		"mainnet":  flags.MainnetAlias,
	}, true, nil
}
