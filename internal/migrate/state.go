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

	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-emulator/storage/migration"
	emulatorMigrate "github.com/onflow/flow-emulator/storage/migration"
	"github.com/onflow/flow-emulator/storage/sqlite"
	"github.com/onflow/flow-go/cmd/util/ledger/migrations"
	"github.com/onflow/flowkit/v2"
	"github.com/onflow/flowkit/v2/output"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/onflow/flow-cli/internal/command"
)

var stateFlags struct {
	Contracts []string `default:"" flag:"contracts" info:"contract names to migrate"`
	DBPath    string   `default:"./flowdb" flag:"db-path" info:"path to the database file"`
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
	contracts := make([]migrations.StagedContract, len(stateFlags.Contracts))
	contractNames := stateFlags.Contracts
	for i, contractName := range contractNames {
		contract, err := state.Contracts().ByName(contractName)
		if err != nil {
			return nil, fmt.Errorf("failed to get contract by name: %w", err)
		}

		code, err := state.ReadFile(contract.Location)
		if err != nil {
			return nil, fmt.Errorf("failed to read contract file: %w", err)
		}

		address, err := getAddressByContractName(state, contractName, flow.Network())
		if err != nil {
			return nil, fmt.Errorf("failed to get address by contract name: %w", err)
		}

		contracts[i] = migrations.StagedContract{
			Contract: migrations.Contract{
				Name: contractName,
				Code: code,
			},
			Address: common.Address(address),
		}
	}

	store, err := sqlite.New(stateFlags.DBPath)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	rwf := &migration.NOOPReportWriterFactory{}
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	err = emulatorMigrate.MigrateCadence1(store, contracts, rwf, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil, nil
}
