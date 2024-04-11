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

package migrate

import (
	"fmt"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/accounts"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/project"
	"github.com/spf13/cobra"
)

func init() {
	getStagedCodeCommand.AddToParent(Cmd)
	IsStagedCommand.AddToParent(Cmd)
	listStagedContractsCommand.AddToParent(Cmd)
	stageContractCommand.AddToParent(Cmd)
	unstageContractCommand.AddToParent(Cmd)
	stateCommand.AddToParent(Cmd)
	IsValidatedCommand.AddToParent(Cmd)
}

var Cmd = &cobra.Command{
	Use:              "migrate",
	Short:            "Migrate your Cadence project to 1.0",
	TraverseChildren: true,
	GroupID:          "migrate",
}

// address of the migration contract on each network
var migrationContractStagingAddress = map[string]string{
	"testnet":   "0x2ceae959ed1a7e7a",
	"crescendo": "0x27b2302520211b67",
	"mainnet":   "0x56100d46aa9b0212",
}

// MigrationContractStagingAddress returns the address of the migration contract on the given network
func MigrationContractStagingAddress(network string) flow.Address {
	return flow.HexToAddress(migrationContractStagingAddress[network])
}

// replaceImportsIfExists replaces imports in the given contract file with the actual contract code
func replaceImportsIfExists(state *flowkit.State, flow flowkit.Services, location string) ([]byte, error) {
	code, err := state.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("error loading contract file: %w", err)
	}

	contracts, err := state.DeploymentContractsByNetwork(flow.Network())
	if err != nil {
		return nil, err
	}

	importReplacer := project.NewImportReplacer(
		contracts,
		state.AliasesForNetwork(flow.Network()),
	)

	program, err := project.NewProgram(code, []cadence.Value{}, location)
	if err != nil {
		return nil, err
	}

	if program.HasImports() {
		program, err = importReplacer.Replace(program)
		if err != nil {
			return nil, err
		}
	}

	return program.Code(), nil
}

func getAccountByContractName(state *flowkit.State, contractName string, network config.Network) (*accounts.Account, error) {
	deployments := state.Deployments().ByNetwork(network.Name)
	var accountName string
	for _, d := range deployments {
		for _, c := range d.Contracts {
			if c.Name == contractName {
				accountName = d.Account
				break
			}
		}
	}
	if accountName == "" {
		return nil, fmt.Errorf("contract not found in state")
	}

	accs := state.Accounts()
	if accs == nil {
		return nil, fmt.Errorf("no accounts found in state")
	}

	var account *accounts.Account
	for _, a := range *accs {
		if accountName == a.Name {
			account = &a
			break
		}
	}
	if account == nil {
		return nil, fmt.Errorf("account %s not found in state", accountName)
	}

	return account, nil
}

func getAddressByContractName(state *flowkit.State, contractName string, network config.Network) (flow.Address, error) {
	account, err := getAccountByContractName(state, contractName, network)
	if err != nil {
		return flow.Address{}, err
	}

	return flow.HexToAddress(account.Address.Hex()), nil
}

func checkNetwork(network config.Network) error {
	if network.Name != config.TestnetNetwork.Name && network.Name != config.MainnetNetwork.Name {
		return fmt.Errorf("staging contracts is only supported on testnet & mainnet networks, see https://cadence-lang.org/docs/cadence-migration-guide for more information")
	}
	return nil
}
