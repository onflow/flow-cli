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
	"os"

	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-emulator/storage/migration"
	emulatorMigrate "github.com/onflow/flow-emulator/storage/migration"
	"github.com/onflow/flow-emulator/storage/sqlite"
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go/cmd/util/ledger/migrations"
	"github.com/onflow/flow-go/cmd/util/ledger/reporters"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/config"
	"github.com/onflow/flowkit/v2/output"
	"github.com/onflow/flowkit/v2/project"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
	"github.com/onflow/flow-cli/internal/util"
)

var stateFlags struct {
	Contracts  []string `default:"" flag:"contracts" info:"contract names to migrate"`
	SaveReport string   `default:"" flag:"save-report" info:"save migration report to a given directory if provided"`
	DBPath     string   `default:"./flowdb" flag:"db-path" info:"path to the sqlite database file"`
}

var stateCommand = &command.Command{
	Cmd: &cobra.Command{
		Use:     "state",
		Short:   "Migrate the state of a SQLite Flow Emulator database",
		Example: `flow migrate state`,
		Args:    cobra.MinimumNArgs(0),
	},
	Flags: &stateFlags,
	RunS:  migrateState,
}

func migrateState(
	args []string,
	globalFlags command.GlobalFlags,
	_ output.Logger,
	flow flowkit.Services,
	state *flowkit.State,
) (command.Result, error) {
	if globalFlags.Network != config.EmulatorNetwork.Name {
		return nil, fmt.Errorf("state migration is only supported for the emulator network")
	}

	contracts, err := resolveStagedContracts(state, stateFlags.Contracts)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve staged contracts: %w", err)
	}

	store, err := sqlite.New(stateFlags.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	// Create a report writer factory if a report path is provided
	var rwf reporters.ReportWriterFactory
	if stateFlags.SaveReport != "" {
		err := state.ReaderWriter().MkdirAll(stateFlags.SaveReport, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("failed to create report directory: %w", err)
		}
		rwf = reporters.NewReportFileWriterFactory(stateFlags.SaveReport, logger)
	} else {
		rwf = &migration.NOOPReportWriterFactory{}
	}

	err = emulatorMigrate.MigrateCadence1(
		store,
		stateFlags.SaveReport,
		migrations.EVMContractChangeNone,
		migrations.BurnerContractChangeDeploy,
		contracts,
		rwf,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil, nil
}

func resolveStagedContracts(state *flowkit.State, contractNames []string) ([]migrations.StagedContract, error) {
	stagedContracts := make([]migrations.StagedContract, len(contractNames))

	network := config.EmulatorNetwork

	contracts, err := state.DeploymentContractsByNetwork(network)
	if err != nil {
		return nil, err
	}

	importReplacer := project.NewImportReplacer(
		contracts,
		state.AliasesForNetwork(network),
	)

	for i, contractName := range contractNames {
		// First try to get contract address from aliases
		contract, err := state.Contracts().ByName(contractName)
		if err != nil {
			return nil, fmt.Errorf("failed to get contract by name: %w", err)
		}

		var address flow.Address
		alias := contract.Aliases.ByNetwork(network.Name)
		if alias != nil {
			address = alias.Address
		}

		code, err := state.ReadFile(contract.Location)
		if err != nil {
			return nil, fmt.Errorf("failed to read contract file: %w", err)
		}

		// If contract is not aliased, try to get address by deployment account
		if address == flow.EmptyAddress {
			address, err = util.GetAddressByContractName(state, contractName, network)
			if err != nil {
				return nil, fmt.Errorf("failed to get address by contract name: %w", err)
			}
		}

		program, err := project.NewProgram(code, []cadence.Value{}, contract.Location)
		if err != nil {
			return nil, err
		}

		if program.HasImports() {
			program, err = importReplacer.Replace(program)
			if err != nil {
				return nil, err
			}
		}

		updatedCode := program.Code()

		stagedContracts[i] = migrations.StagedContract{
			Contract: migrations.Contract{
				Name: contractName,
				Code: updatedCode,
			},
			Address: common.Address(address),
		}
	}

	return stagedContracts, nil
}
