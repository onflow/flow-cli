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
	"github.com/onflow/flowkit/v2/project"
	"github.com/spf13/cobra"
)

func init() {
	getStagedCodeCommand.AddToParent(Cmd)
	IsStagedCommand.AddToParent(Cmd)
	listStagedContractsCommand.AddToParent(Cmd)
	stageCommand.AddToParent(Cmd)
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
